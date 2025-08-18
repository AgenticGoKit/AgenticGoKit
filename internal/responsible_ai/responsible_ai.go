// Package responsible_ai provides the RAI (Responsible AI) content pipeline for automatic content filtering
package responsible_ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// RAIAction represents the action to take after content analysis
type RAIAction string

const (
	RAIActionAllow  RAIAction = "ALLOW"
	RAIActionBlock  RAIAction = "BLOCK"
	RAIActionModify RAIAction = "MODIFY"
	RAIActionWarn   RAIAction = "WARN"
)

// RAIPolicy represents a specific content policy to check
type RAIPolicy string

const (
	RAIPolicySafety     RAIPolicy = "SAFETY"
	RAIPolicyBias       RAIPolicy = "BIAS"
	RAIPolicyCompliance RAIPolicy = "COMPLIANCE"
	RAIPolicyToxicity   RAIPolicy = "TOXICITY"
	RAIPolicyPrivacy    RAIPolicy = "PRIVACY"
)

// RAICheckResult contains the result of content analysis
type RAICheckResult struct {
	Safe            bool        `json:"safe"`
	Action          RAIAction   `json:"action"`
	Violations      []string    `json:"violations"`
	Confidence      float64     `json:"confidence"`
	Policies        []RAIPolicy `json:"policies_checked"`
	ModifiedContent string      `json:"modified_content,omitempty"`
	Timestamp       time.Time   `json:"timestamp"`
}

// RAIConfig configures the RAI content pipeline
type RAIConfig struct {
	CheckInput    bool        `json:"check_input"`
	CheckOutput   bool        `json:"check_output"`
	Policies      []RAIPolicy `json:"policies"`
	StrictMode    bool        `json:"strict_mode"`
	AutoModify    bool        `json:"auto_modify"`
	BlockOnUnsafe bool        `json:"block_on_unsafe"`
	MinConfidence float64     `json:"min_confidence"`
	RAIAgentName  string      `json:"rai_agent_name"`
}

// DefaultRAIConfig returns a default RAI configuration
func DefaultRAIConfig() RAIConfig {
	return RAIConfig{
		CheckInput:    true,
		CheckOutput:   true,
		Policies:      []RAIPolicy{RAIPolicySafety, RAIPolicyToxicity, RAIPolicyBias},
		StrictMode:    false,
		AutoModify:    false,
		BlockOnUnsafe: true,
		MinConfidence: 0.7,
		RAIAgentName:  "responsible_ai",
	}
}

// RAIMiddleware provides content filtering capabilities
type RAIMiddleware struct {
	config    RAIConfig
	provider  core.ModelProvider
	callbacks *core.CallbackRegistry
}

// NewRAIMiddleware creates a new RAI middleware instance
func NewRAIMiddleware(config RAIConfig, provider core.ModelProvider) *RAIMiddleware {
	return &RAIMiddleware{
		config:   config,
		provider: provider,
	}
}

// SetCallbackRegistry sets the callback registry for middleware hooks
func (m *RAIMiddleware) SetCallbackRegistry(registry *core.CallbackRegistry) {
	m.callbacks = registry
}

// CheckContent performs content analysis using the configured policies
func (m *RAIMiddleware) CheckContent(ctx context.Context, content string, contentType string) (*RAICheckResult, error) {
	if content == "" {
		return &RAICheckResult{
			Safe:      true,
			Action:    RAIActionAllow,
			Timestamp: time.Now(),
		}, nil
	}

	// Create RAI analysis prompt
	prompt := m.createRAIPrompt(content, contentType)

	// Call the model provider for analysis
	response, err := m.provider.Call(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("RAI content check failed: %w", err)
	}

	// Parse the response
	result := m.parseRAIResponse(response.Content, content)
	result.Policies = m.config.Policies
	result.Timestamp = time.Now()

	// Apply configuration rules
	m.applyConfigRules(result)

	return result, nil
}

// BeforeModelCall is a callback hook that checks input content
func (m *RAIMiddleware) BeforeModelCall(ctx context.Context, args core.CallbackArgs) (core.State, error) {
	if !m.config.CheckInput {
		return args.State, nil
	}

	// Extract content from the prompt or state
	content := m.extractContentForAnalysis(args)
	if content == "" {
		return args.State, nil
	}

	// Check the content
	result, err := m.CheckContent(ctx, content, "input")
	if err != nil {
		core.Logger().Error().Err(err).Msg("RAI input check failed")
		if m.config.StrictMode {
			return nil, fmt.Errorf("RAI input validation failed: %w", err)
		}
		// In non-strict mode, allow content to pass but log the failure
		return args.State, nil
	}

	// Store RAI result in state
	newState := args.State.Clone()
	newState.Set("rai_input_check", result)

	// Handle unsafe content
	if !result.Safe && m.config.BlockOnUnsafe {
		return nil, fmt.Errorf("content blocked by RAI policy: %s", strings.Join(result.Violations, ", "))
	}

	if result.Action == RAIActionWarn {
		core.Logger().Warn().
			Str("content_type", "input").
			Strs("violations", result.Violations).
			Msg("RAI warning: potentially problematic content detected")
	}

	return newState, nil
}

// AfterModelCall is a callback hook that checks output content
func (m *RAIMiddleware) AfterModelCall(ctx context.Context, args core.CallbackArgs) (core.State, error) {
	if !m.config.CheckOutput {
		return args.State, nil
	}

	// Extract model response from state or result
	content := m.extractOutputForAnalysis(args)
	if content == "" {
		return args.State, nil
	}

	// Check the content
	result, err := m.CheckContent(ctx, content, "output")
	if err != nil {
		core.Logger().Error().Err(err).Msg("RAI output check failed")
		if m.config.StrictMode {
			return nil, fmt.Errorf("RAI output validation failed: %w", err)
		}
		return args.State, nil
	}

	// Store RAI result in state
	newState := args.State.Clone()
	newState.Set("rai_output_check", result)

	// Handle unsafe content
	if !result.Safe {
		if m.config.BlockOnUnsafe {
			return nil, fmt.Errorf("model output blocked by RAI policy: %s", strings.Join(result.Violations, ", "))
		}

		if m.config.AutoModify && result.ModifiedContent != "" {
			// Replace the content with modified version
			newState.Set("model_response_modified", true)
			newState.Set("original_model_response", content)
			newState.Set("model_response", result.ModifiedContent)
			core.Logger().Info().
				Str("action", "auto_modify").
				Msg("RAI automatically modified unsafe model output")
		}
	}

	if result.Action == RAIActionWarn {
		core.Logger().Warn().
			Str("content_type", "output").
			Strs("violations", result.Violations).
			Msg("RAI warning: potentially problematic content in model output")
	}

	return newState, nil
}

// createRAIPrompt creates a prompt for content analysis
func (m *RAIMiddleware) createRAIPrompt(content, contentType string) core.Prompt {
	policies := strings.Join(m.policyStrings(), ", ")

	systemPrompt := fmt.Sprintf(`You are a Responsible AI content analyzer. Analyze the given %s content for violations of these policies: %s.

Respond in this exact format:
SAFE: [true/false]
ACTION: [ALLOW/BLOCK/MODIFY/WARN]
VIOLATIONS: [comma-separated list of violations, or "none"]
CONFIDENCE: [0.0-1.0]
MODIFIED_CONTENT: [only if ACTION is MODIFY, otherwise omit this line]

Be objective and accurate. Consider context and intent.`, contentType, policies)

	return core.Prompt{
		System: systemPrompt,
		User:   fmt.Sprintf("Analyze this %s content: %s", contentType, content),
	}
}

// parseRAIResponse parses the model's RAI analysis response
func (m *RAIMiddleware) parseRAIResponse(response, originalContent string) *RAICheckResult {
	result := &RAICheckResult{
		Safe:       true, // Default to safe
		Action:     RAIActionAllow,
		Violations: []string{},
		Confidence: 0.5, // Default confidence
	}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "SAFE:") {
			safeStr := strings.TrimSpace(strings.TrimPrefix(line, "SAFE:"))
			result.Safe = strings.ToLower(safeStr) == "true"
		} else if strings.HasPrefix(line, "ACTION:") {
			actionStr := strings.TrimSpace(strings.TrimPrefix(line, "ACTION:"))
			result.Action = RAIAction(strings.ToUpper(actionStr))
		} else if strings.HasPrefix(line, "VIOLATIONS:") {
			violationsStr := strings.TrimSpace(strings.TrimPrefix(line, "VIOLATIONS:"))
			if violationsStr != "none" && violationsStr != "" {
				violations := strings.Split(violationsStr, ",")
				for _, v := range violations {
					if v = strings.TrimSpace(v); v != "" {
						result.Violations = append(result.Violations, v)
					}
				}
			}
		} else if strings.HasPrefix(line, "CONFIDENCE:") {
			confStr := strings.TrimSpace(strings.TrimPrefix(line, "CONFIDENCE:"))
			if conf := parseFloat(confStr); conf >= 0 && conf <= 1 {
				result.Confidence = conf
			}
		} else if strings.HasPrefix(line, "MODIFIED_CONTENT:") {
			result.ModifiedContent = strings.TrimSpace(strings.TrimPrefix(line, "MODIFIED_CONTENT:"))
		}
	}

	return result
}

// applyConfigRules applies configuration-based rules to the result
func (m *RAIMiddleware) applyConfigRules(result *RAICheckResult) {
	// Apply minimum confidence threshold
	if result.Confidence < m.config.MinConfidence {
		result.Action = RAIActionWarn
		if !m.config.StrictMode {
			result.Safe = true // Allow in non-strict mode with low confidence
		}
	}

	// Override action based on safety result
	if !result.Safe {
		if m.config.BlockOnUnsafe {
			result.Action = RAIActionBlock
		} else if m.config.AutoModify && result.ModifiedContent != "" {
			result.Action = RAIActionModify
		} else {
			result.Action = RAIActionWarn
		}
	}
}

// extractContentForAnalysis extracts content from callback arguments for input analysis
func (m *RAIMiddleware) extractContentForAnalysis(args core.CallbackArgs) string {
	// Try to get content from various sources in the callback args
	if args.State != nil {
		if content, exists := args.State.Get("input_content"); exists {
			if str, ok := content.(string); ok {
				return str
			}
		}
		if content, exists := args.State.Get("prompt"); exists {
			if str, ok := content.(string); ok {
				return str
			}
		}
	}

	// Could also extract from Event data if available
	if args.Event != nil {
		data := args.Event.GetData()
		if content, ok := data["content"]; ok {
			if str, ok := content.(string); ok {
				return str
			}
		}
		if message, ok := data["message"]; ok {
			if str, ok := message.(string); ok {
				return str
			}
		}
	}

	return ""
}

// extractOutputForAnalysis extracts content from callback arguments for output analysis
func (m *RAIMiddleware) extractOutputForAnalysis(args core.CallbackArgs) string {
	if args.State != nil {
		if content, exists := args.State.Get("model_response"); exists {
			if str, ok := content.(string); ok {
				return str
			}
		}
		if content, exists := args.State.Get("output_content"); exists {
			if str, ok := content.(string); ok {
				return str
			}
		}
	}
	return ""
}

// policyStrings converts policies to string descriptions
func (m *RAIMiddleware) policyStrings() []string {
	var policies []string
	for _, policy := range m.config.Policies {
		switch policy {
		case RAIPolicySafety:
			policies = append(policies, "safety (harmful, violent, dangerous content)")
		case RAIPolicyBias:
			policies = append(policies, "bias (discriminatory, prejudiced content)")
		case RAIPolicyCompliance:
			policies = append(policies, "compliance (legal, regulatory violations)")
		case RAIPolicyToxicity:
			policies = append(policies, "toxicity (offensive, abusive language)")
		case RAIPolicyPrivacy:
			policies = append(policies, "privacy (personal information exposure)")
		}
	}
	return policies
}

// parseFloat safely parses a string to float64
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
