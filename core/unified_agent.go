package core

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// UnifiedAgent is a concrete Agent implementation used by internal builders.
// It provides basic capability plumbing and a simple Run implementation.
type UnifiedAgent struct {
	name         string
	role         string
	description  string
	systemPrompt string
	timeout      time.Duration
	enabled      bool
	autoLLM      bool // Controls whether to automatically call LLM when provider is configured

	// Config/LLM
	llmConfig *ResolvedLLMConfig

	// Capabilities storage
	capabilities map[CapabilityType]AgentCapability
	handler      AgentHandler

	// Capability-configured dependencies
	llmProvider ModelProvider
	cacheMgr    interface{}
	metricsCfg  MetricsConfig

	// MCP-specific wiring to satisfy MCP capability
	mcpManager     MCPManager
	mcpAgentConfig MCPAgentConfig
	mcpCacheMgr    MCPCacheManager
}

// NewUnifiedAgent constructs a new UnifiedAgent with provided capabilities and optional handler.
func NewUnifiedAgent(name string, caps map[CapabilityType]AgentCapability, handler AgentHandler) *UnifiedAgent {
	if caps == nil {
		caps = map[CapabilityType]AgentCapability{}
	}
	return &UnifiedAgent{
		name:         name,
		role:         "unified_agent",
		description:  "Unified composable agent",
		systemPrompt: "You are a helpful AI agent.",
		timeout:      30 * time.Second,
		enabled:      true,
		autoLLM:      false, // Default to false for safety - user must explicitly enable
		capabilities: caps,
		handler:      handler,
	}
}

// Agent interface
func (u *UnifiedAgent) Name() string                     { return u.name }
func (u *UnifiedAgent) GetRole() string                  { return u.role }
func (u *UnifiedAgent) GetDescription() string           { return u.description }
func (u *UnifiedAgent) GetSystemPrompt() string          { return u.systemPrompt }
func (u *UnifiedAgent) GetTimeout() time.Duration        { return u.timeout }
func (u *UnifiedAgent) IsEnabled() bool                  { return u.enabled }
func (u *UnifiedAgent) GetLLMConfig() *ResolvedLLMConfig { return u.llmConfig }

func (u *UnifiedAgent) GetCapabilities() []string {
	out := make([]string, 0, len(u.capabilities))
	for ct := range u.capabilities {
		out = append(out, string(ct))
	}
	return out
}

func (u *UnifiedAgent) Initialize(ctx context.Context) error { return nil }
func (u *UnifiedAgent) Shutdown(ctx context.Context) error   { return nil }

func (u *UnifiedAgent) Run(ctx context.Context, state State) (State, error) {
	// Basic pre/post without complex hooks for now
	if u.handler != nil {
		res, err := u.handler.Run(ctx, NewEvent(u.name, map[string]any{}, map[string]string{}), state)
		if err != nil {
			return res.OutputState, err
		}
		return res.OutputState, nil
	}
	out := state.Clone()

	// If an LLM provider is configured and auto-LLM is enabled, perform a default completion.
	// This gives UnifiedAgent a sensible behavior out-of-the-box for
	// configuration-driven agents created by the factory/builder.
	if u.llmProvider != nil && u.autoLLM {
		DebugLogWithFields(Logger(), "UnifiedAgent: LLM provider detected; preparing default completion", map[string]interface{}{
			"agent": u.name,
		})
		// Prefer a system prompt set in state (e.g., by a config-aware wrapper),
		// otherwise fall back to the agent's configured systemPrompt.
		system := u.systemPrompt
		if v, ok := state.Get("system_prompt"); ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				system = s
			}
		}

		// Extract user input from state. Orchestrators usually merge event data
		// into state before invoking the agent, so "message" is a common key.
		var user string
		if v, ok := state.Get("message"); ok {
			if s, ok2 := v.(string); ok2 {
				user = s
			}
		}

		// Only call the LLM if we have a non-empty user prompt.
		if strings.TrimSpace(user) != "" {
			DebugLogWithFields(Logger(), "UnifiedAgent: calling LLM provider", map[string]interface{}{
				"agent": u.name,
			})

			// Add memory integration for RAG support
			var memoryContext string
			mem := GetMemory(ctx)
			if mem != nil {
				DebugLogWithFields(Logger(), "UnifiedAgent: Building memory context for RAG", map[string]interface{}{
					"agent":       u.name,
					"memory_type": fmt.Sprintf("%T", mem),
				})

				// Build RAG context from knowledge base
				DebugLogWithFields(Logger(), "UnifiedAgent: Calling BuildContext with query", map[string]interface{}{
					"agent": u.name,
					"query": user,
				})
				ragContext, err := mem.BuildContext(ctx, user,
					WithMaxTokens(1000),
					WithIncludeSources(true))
				if err != nil {
					Logger().Warn().Str("agent", u.name).Err(err).Msg("Failed to build RAG context - continuing without knowledge base context")
				} else if ragContext != nil {
					// Only log detailed context in debug mode to reduce verbosity
					if GetLogLevel() == DEBUG {
						DebugLogWithFields(Logger(), "BuildContext result details", map[string]interface{}{
							"agent":           u.name,
							"personal_count":  len(ragContext.PersonalMemory),
							"knowledge_count": len(ragContext.Knowledge),
							"history_count":   len(ragContext.ChatHistory),
							"context_text_snippet": func() string {
								if len(ragContext.ContextText) > 100 {
									return ragContext.ContextText[:100] + "..."
								}
								return ragContext.ContextText
							}(),
						})
					}

					if ragContext.ContextText != "" {
						memoryContext = "\n\nRelevant Context from Knowledge Base:\n" + ragContext.ContextText
						// Only log RAG context success in debug mode to reduce verbosity
						if GetLogLevel() == DEBUG {
							DebugLogWithFields(Logger(), "RAG context built successfully", map[string]interface{}{
								"agent":          u.name,
								"context_tokens": ragContext.TokenCount,
							})
						}
					} else {
						DebugLogWithFields(Logger(), "No relevant knowledge base context found", map[string]interface{}{
							"agent": u.name,
						})
					}
				} else {
					Logger().Debug().Str("agent", u.name).Msg("BuildContext returned nil ragContext")
				}

				// Query relevant memories
				memoryResults, err := mem.Query(ctx, user, 5)
				if err != nil {
					Logger().Warn().Str("agent", u.name).Err(err).Msg("Failed to query memories - continuing without memory context")
				} else if len(memoryResults) > 0 {
					memoryContext += "\n\nRelevant Memories:\n"
					for _, result := range memoryResults {
						if result.Score >= 0 {
							memoryContext += strings.TrimSpace(result.Content) + "\n"
						}
					}
					Logger().Debug().Str("agent", u.name).Int("memory_count", len(memoryResults)).Msg("Memory context retrieved")
				}

				// Get chat history
				chatHistory, err := mem.GetHistory(ctx, 3)
				if err != nil {
					Logger().Warn().Str("agent", u.name).Err(err).Msg("Failed to get chat history - continuing without history context")
				} else if len(chatHistory) > 0 {
					memoryContext += "\n\nRecent Chat History:\n"
					for _, msg := range chatHistory {
						memoryContext += strings.TrimSpace(msg.Content) + "\n"
					}
					Logger().Debug().Str("agent", u.name).Int("history_count", len(chatHistory)).Msg("Chat history retrieved")
				}

				// Log memory context for debugging
				snippet := memoryContext
				if len(snippet) > 1000 {
					snippet = snippet[:1000] + "...(truncated)"
				}
				// Only log memory context details in debug mode to reduce verbosity
				if GetLogLevel() == DEBUG {
					DebugLogWithFields(Logger(), "UnifiedAgent: RAG context appended to prompt", map[string]interface{}{
						"agent":                  u.name,
						"memory_length":          len(memoryContext),
						"memory_context_snippet": snippet,
					})
				}
			} else {
				Logger().Warn().Str("agent", u.name).Msg("Memory system not available - continuing without memory context")
			}

			// Get available MCP tools if manager is configured
			var toolsPrompt string
			if u.mcpManager != nil {
				availableTools := u.mcpManager.GetAvailableTools()
				if len(availableTools) > 0 {
					DebugLogWithFields(Logger(), "UnifiedAgent: MCP tools discovered", map[string]interface{}{
						"agent":      u.name,
						"tool_count": len(availableTools),
					})
					toolsPrompt = FormatToolsPromptForLLM(availableTools)
				} else {
					DebugLogWithFields(Logger(), "UnifiedAgent: No MCP tools available", map[string]interface{}{
						"agent": u.name,
					})
				}
			}

			// Append memory context and tools prompt to user prompt
			finalUserPrompt := user + memoryContext + toolsPrompt

			// Only make LLM call if provider is configured
			if u.llmProvider != nil {
				params := ModelParameters{}
				if u.llmConfig != nil {
					if u.llmConfig.Temperature != 0 {
						// Convert float64 -> float32 pointer
						t := float32(u.llmConfig.Temperature)
						params.Temperature = &t
					}
					if u.llmConfig.MaxTokens > 0 {
						mt := int32(u.llmConfig.MaxTokens)
						params.MaxTokens = &mt
					}
				}

				resp, err := u.llmProvider.Call(ctx, Prompt{System: system, User: finalUserPrompt, Parameters: params})
				if err != nil {
					return out, err
				}

				// Parse LLM response for tool calls and execute them if MCP is available
				if u.mcpManager != nil && len(toolsPrompt) > 0 {
					toolCalls := ParseLLMToolCalls(resp.Content)
					if len(toolCalls) > 0 {
						DebugLogWithFields(Logger(), "UnifiedAgent: Tool calls detected, executing", map[string]interface{}{
							"agent":      u.name,
							"tool_calls": len(toolCalls),
						})

						var toolResults []string
						for _, toolCall := range toolCalls {
							if toolName, ok := toolCall["name"].(string); ok {
								var args map[string]interface{}
								if toolArgs, exists := toolCall["args"]; exists {
									if argsMap, ok := toolArgs.(map[string]interface{}); ok {
										args = argsMap
									} else {
										args = make(map[string]interface{})
									}
								} else {
									args = make(map[string]interface{})
								}

								DebugLogWithFields(Logger(), "UnifiedAgent: Executing tool", map[string]interface{}{
									"agent":     u.name,
									"tool_name": toolName,
									"args":      args,
								})

								result, err := ExecuteMCPTool(ctx, toolName, args)
								if err != nil {
									Logger().Error().Str("agent", u.name).Str("tool_name", toolName).Err(err).Msg("Tool execution failed")
									toolResults = append(toolResults, fmt.Sprintf("Tool '%s' failed: %v", toolName, err))
								} else if result.Success {
									var resultContent string
									if len(result.Content) > 0 {
										resultContent = result.Content[0].Text
									} else {
										resultContent = "Tool executed successfully but returned no content"
									}
									toolResults = append(toolResults, fmt.Sprintf("Tool '%s' result: %s", toolName, resultContent))
								} else {
									toolResults = append(toolResults, fmt.Sprintf("Tool '%s' was not successful", toolName))
								}
							}
						}

						// If we got tool results, send them back to LLM for final response
						if len(toolResults) > 0 {
							toolResultsText := strings.Join(toolResults, "\n")
							DebugLogWithFields(Logger(), "UnifiedAgent: Sending tool results to LLM", map[string]interface{}{
								"agent":        u.name,
								"result_count": len(toolResults),
							})

							followUpPrompt := finalUserPrompt + "\n\nTool Results:\n" + toolResultsText + "\n\nPlease provide a final response based on the tool results above."

							finalResp, err := u.llmProvider.Call(ctx, Prompt{System: system, User: followUpPrompt, Parameters: params})
							if err != nil {
								Logger().Warn().Str("agent", u.name).Err(err).Msg("Failed to get final response after tools, using tool results")
								resp.Content = "Tool execution completed:\n" + toolResultsText
							} else {
								resp = finalResp
							}
						}
					}
				}

				if resp.Content != "" {
					// Store interaction in memory
					if mem != nil {
						if err := mem.Store(ctx, user, "user-query", u.name); err != nil {
							Logger().Warn().Str("agent", u.name).Err(err).Msg("Failed to store user query in memory")
						}
						if err := mem.Store(ctx, resp.Content, "agent-response", u.name); err != nil {
							Logger().Warn().Str("agent", u.name).Err(err).Msg("Failed to store agent response in memory")
						}
						if err := mem.AddMessage(ctx, "user", user); err != nil {
							Logger().Warn().Str("agent", u.name).Err(err).Msg("Failed to add user message to chat history")
						}
						if err := mem.AddMessage(ctx, "assistant", resp.Content); err != nil {
							Logger().Warn().Str("agent", u.name).Err(err).Msg("Failed to add assistant message to chat history")
						}
					}

					// Standardize keys so downstream result collectors can display output.
					out.Set("response", resp.Content)
					out.Set("message", resp.Content)
				}
			} else {
				// Debug mode: output memory context details without LLM call
				Logger().Debug().Str("agent", u.name).Msg("DEBUG MODE: No LLM provider, outputting memory analysis instead")
				debugResponse := fmt.Sprintf("DEBUG: Memory Analysis for query '%s'\n", user)
				if memoryContext != "" {
					debugResponse += fmt.Sprintf("Memory context found (%d chars):\n%s", len(memoryContext), memoryContext)
				} else {
					debugResponse += "No memory context found"
				}

				out.Set("response", debugResponse)
				out.Set("message", debugResponse)
				Logger().Debug().Str("agent", u.name).Str("debug_response", debugResponse).Msg("DEBUG: Memory analysis completed")
			}
		}

	}
	out.Set("processed_by", u.name)
	out.Set("agent_type", "unified")
	out.Set("capabilities", u.GetCapabilities())
	return out, nil
}

func (u *UnifiedAgent) HandleEvent(ctx context.Context, event Event, state State) (AgentResult, error) {
	start := time.Now()
	out, err := u.Run(ctx, state)
	end := time.Now()
	result := AgentResult{OutputState: out, StartTime: start, EndTime: end, Duration: end.Sub(start)}
	if err != nil {
		result.Error = err.Error()
	}
	return result, nil
}

// CapabilityConfigurable bridging (subset used by internal capabilities)
func (u *UnifiedAgent) SetLLMProvider(provider ModelProvider, config LLMConfig) {
	u.llmProvider = provider
	// Map to ResolvedLLMConfig-lite as available
	u.llmConfig = &ResolvedLLMConfig{
		Provider:         config.Provider,
		Model:            config.Model,
		Temperature:      config.Temperature,
		MaxTokens:        config.MaxTokens,
		Timeout:          TimeoutFromSeconds(config.TimeoutSeconds),
		TopP:             config.TopP,
		FrequencyPenalty: config.FrequencyPenalty,
		PresencePenalty:  config.PresencePenalty,
	}
}

func (u *UnifiedAgent) SetCacheManager(manager interface{}, config interface{}) {
	u.cacheMgr = manager
}

func (u *UnifiedAgent) SetMetricsConfig(config MetricsConfig) { u.metricsCfg = config }

// SetAutoLLM configures whether the agent should automatically call the LLM provider
// when one is configured. Set to true to enable automatic LLM calls, false to disable.
func (u *UnifiedAgent) SetAutoLLM(enabled bool) { u.autoLLM = enabled }

// GetAutoLLM returns whether automatic LLM calls are enabled.
func (u *UnifiedAgent) GetAutoLLM() bool { return u.autoLLM }

// Logger accessor to satisfy internal CapabilityConfigurable usage through core.Logger
func (u *UnifiedAgent) GetLogger() CoreLogger { return Logger() }

// MCP wiring to satisfy MCP capability Configure calls
func (u *UnifiedAgent) SetMCPManager(manager MCPManager, config MCPAgentConfig) {
	u.mcpManager = manager
	u.mcpAgentConfig = config
}

func (u *UnifiedAgent) SetMCPCacheManager(manager MCPCacheManager) { u.mcpCacheMgr = manager }

// GetCapability returns a capability by type if present (helper for internal bridges)
func (u *UnifiedAgent) GetCapability(t CapabilityType) (AgentCapability, bool) {
	if u.capabilities == nil {
		return nil, false
	}
	cap, ok := u.capabilities[t]
	return cap, ok
}
