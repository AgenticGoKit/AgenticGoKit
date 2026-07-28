package core

import (
	"strings"
	"testing"
)

// Regression test for the blocker flagged in PR #157 review: a SystemPrompt
// containing a literal "{{" that isn't valid Go template syntax (e.g. a
// prompt documenting Handlebars/Jinja/Vue templating, or a doc example with
// "{{ }}" in it) must not fail the run — it must pass through verbatim.
func TestFormatPromptString_GoTemplate_InvalidBraceSyntaxFallsBackVerbatim(t *testing.T) {
	tpl := "You are an assistant. Use the syntax {{foo}} to reference variables in Handlebars."

	got, err := FormatPromptString(tpl, map[string]any{"Input": "hello"}, FormatGoTemplate)
	if err != nil {
		t.Fatalf("FormatPromptString returned an error for a non-template \"{{\" prompt, want verbatim fallback: %v", err)
	}
	if got != tpl {
		t.Fatalf("FormatPromptString mangled a non-template prompt.\n got:  %q\nwant: %q", got, tpl)
	}
}

// Same failure mode, but via an incomplete/invalid directive rather than a
// non-Go-template dialect example (e.g. "{{if}}" with no condition).
func TestFormatPromptString_GoTemplate_MalformedDirectiveFallsBackVerbatim(t *testing.T) {
	tpl := "Some instructions. {{if}} more text."

	got, err := FormatPromptString(tpl, nil, FormatGoTemplate)
	if err != nil {
		t.Fatalf("FormatPromptString returned an error for a malformed template, want verbatim fallback: %v", err)
	}
	if got != tpl {
		t.Fatalf("FormatPromptString mangled a malformed-template prompt.\n got:  %q\nwant: %q", got, tpl)
	}
}

// Valid Go template syntax must still interpolate correctly — the fallback
// must not mask legitimate templating.
func TestFormatPromptString_GoTemplate_ValidSyntaxStillInterpolates(t *testing.T) {
	tpl := "Hello {{.Input}}!"

	got, err := FormatPromptString(tpl, map[string]any{"Input": "world"}, FormatGoTemplate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Hello world!" {
		t.Fatalf("got %q, want %q", got, "Hello world!")
	}
}

// Plain text with no "{{" at all must pass through unchanged (existing
// guarantee, still covered as a baseline).
func TestFormatPromptString_GoTemplate_PlainTextPassesThroughUnchanged(t *testing.T) {
	tpl := "You are a helpful assistant. Respond in JSON: {\"k\":1}."

	got, err := FormatPromptString(tpl, nil, FormatGoTemplate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != tpl {
		t.Fatalf("got %q, want %q", got, tpl)
	}
}

// A template referencing an undefined execution directive (not a parse
// error, but a runtime execute error) must also fall back verbatim rather
// than erroring, per the same rationale.
func TestFormatPromptString_GoTemplate_ExecuteErrorFallsBackVerbatim(t *testing.T) {
	// {{index .Items 5}} parses fine (valid template syntax) but errors at
	// Execute time (index out of range), reliably reproducing the
	// execute-time (not parse-time) failure path.
	tpl := "Value: {{index .Items 5}}"

	got, err := FormatPromptString(tpl, map[string]any{"Items": []int{1, 2}}, FormatGoTemplate)
	if err != nil {
		t.Fatalf("FormatPromptString returned an error on an execute-time failure, want verbatim fallback: %v", err)
	}
	if !strings.Contains(got, "{{index .Items 5}}") {
		t.Fatalf("expected verbatim fallback containing the original directive, got %q", got)
	}
}
