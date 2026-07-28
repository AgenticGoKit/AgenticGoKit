package core

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"
)

// FormatType selects the substitution syntax used by FormatPromptString.
// Modeled after cloudwego/eino's schema.FormatType (FString/GoTemplate/Jinja2),
// scoped down to the two syntaxes implementable without a new dependency.
type FormatType uint8

const (
	// FormatGoTemplate interpolates using Go's text/template ({{.key}}).
	// A string with no {{ ... }} sequence passes through unchanged. A string
	// that DOES contain "{{" but isn't valid Go template syntax (e.g. a
	// prompt that teaches the model Handlebars/Jinja/Vue syntax, or embeds a
	// doc example containing literal "{{") fails template.Parse/Execute —
	// formatGoTemplate falls back to returning tpl verbatim in that case
	// (with a Warn log) rather than erroring the caller's Run, so this is
	// safe to apply unconditionally even though the syntax collision is not
	// hypothetical: it broke a real skill's frontmatter description in
	// practice. See formatGoTemplate.
	FormatGoTemplate FormatType = iota
	// FormatFString interpolates using Python str.format-style {key} placeholders.
	FormatFString
)

var fstringPlaceholder = regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// FormatPromptString substitutes vs into tpl according to formatType.
//
// Variable keys present in tpl but absent from vs produce an error for
// FormatFString, and render as "<no value>" for FormatGoTemplate (the
// text/template default) — there is no compile-time safety, consistent with
// eino's ChatTemplate.Format contract.
func FormatPromptString(tpl string, vs map[string]any, formatType FormatType) (string, error) {
	switch formatType {
	case FormatFString:
		return formatFString(tpl, vs)
	case FormatGoTemplate:
		return formatGoTemplate(tpl, vs)
	default:
		return "", fmt.Errorf("core: unsupported FormatType %d", formatType)
	}
}

// formatGoTemplate parses tpl as a Go template and executes it against vs.
//
// tpl containing no "{{" is unaffected by this parse/execute round-trip in
// practice, but is not special-cased: it still goes through Parse/Execute
// below, which is a correctness statement, not just a performance one —
// there is no separate "fast path" that could drift from the template-path
// behavior.
//
// tpl containing a literal "{{" that ISN'T valid Go template syntax (a
// prompt documenting Handlebars/Jinja/Vue templating to the model, or a doc
// example with "{{ }}" in it) is a real, observed failure mode, not a
// hypothetical one — it broke a real skill's frontmatter description. Rather
// than hard-failing the caller's Run on that input (the previous behavior:
// return "", err), Parse/Execute failures fall back to tpl unchanged, so
// FormatPromptString stays safe to apply unconditionally to every
// SystemPrompt as the type's doc comment promises. The fallback is logged
// at Warn so a real templating typo (an unclosed "{{if}}", a genuinely
// misspelled ".Fieldname") is still visible instead of silently mis-rendering
// forever — callers that need a hard error on bad template syntax (rather
// than degrade-to-verbatim) should validate tpl with template.New(...).Parse
// themselves before handing it to FormatPromptString.
func formatGoTemplate(tpl string, vs map[string]any) (string, error) {
	t, err := template.New("prompt").Parse(tpl)
	if err != nil {
		Logger().Warn().Err(err).Msg("core: system prompt contains \"{{\" that isn't valid Go template syntax; using it verbatim instead of failing the run")
		return tpl, nil
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, vs); err != nil {
		Logger().Warn().Err(err).Msg("core: system prompt template failed to execute; using it verbatim instead of failing the run")
		return tpl, nil
	}
	return buf.String(), nil
}

func formatFString(tpl string, vs map[string]any) (string, error) {
	var missing error
	result := fstringPlaceholder.ReplaceAllStringFunc(tpl, func(match string) string {
		key := match[1 : len(match)-1]
		v, ok := vs[key]
		if !ok {
			if missing == nil {
				missing = fmt.Errorf("core: missing variable %q for fstring template", key)
			}
			return match
		}
		return fmt.Sprintf("%v", v)
	})
	if missing != nil {
		return "", missing
	}
	return result, nil
}

// StateVars snapshots a State's data map into a plain map[string]any suitable
// for FormatPromptString. State does not expose its underlying map directly,
// so this walks Keys()+Get().
func StateVars(state State) map[string]any {
	keys := state.Keys()
	vs := make(map[string]any, len(keys))
	for _, k := range keys {
		if v, ok := state.Get(k); ok {
			vs[k] = v
		}
	}
	return vs
}
