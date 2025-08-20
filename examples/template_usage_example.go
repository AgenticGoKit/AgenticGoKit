package main

import (
	"fmt"
	"log"

	"github.com/kunalkushwaha/agenticgokit/internal/scaffold"
)

func main() {
	fmt.Println("=== AgenticGoKit Template System Demo ===\n")

	// Example 1: List available templates
	fmt.Println("1. Available Templates:")
	fmt.Println("======================")
	templates, err := scaffold.ListAvailableTemplates()
	if err != nil {
		log.Fatalf("Failed to list templates: %v", err)
	}

	for i, template := range templates {
		fmt.Printf("%d. %s\n", i+1, template)
		
		// Get template info
		info, err := scaffold.GetTemplateInfo(template)
		if err != nil {
			fmt.Printf("   Error getting info: %v\n", err)
			continue
		}
		
		fmt.Printf("   Description: %s\n", info.Description)
		fmt.Printf("   Features: %v\n", info.Features)
		fmt.Printf("   Agents: %d\n", info.Config.NumAgents)
		fmt.Printf("   Orchestration: %s\n", info.Config.OrchestrationMode)
		fmt.Printf("   Memory: %t\n", info.Config.MemoryEnabled)
		fmt.Printf("   MCP: %t\n", info.Config.MCPEnabled)
		fmt.Println()
	}

	// Example 2: Show detailed template information
	fmt.Println("2. Detailed Template Information:")
	fmt.Println("=================================")
	
	templateName := "research-assistant"
	if len(templates) > 0 {
		templateName = templates[0] // Use first available template
	}
	
	info, err := scaffold.GetTemplateInfo(templateName)
	if err != nil {
		log.Fatalf("Failed to get template info: %v", err)
	}

	fmt.Printf("Template: %s\n", info.Name)
	fmt.Printf("Description: %s\n", info.Description)
	fmt.Printf("Features: %v\n", info.Features)
	fmt.Println()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  - Agents: %d\n", info.Config.NumAgents)
	fmt.Printf("  - Provider: %s\n", info.Config.Provider)
	fmt.Printf("  - Orchestration: %s\n", info.Config.OrchestrationMode)
	fmt.Printf("  - Memory Enabled: %t\n", info.Config.MemoryEnabled)
	if info.Config.MemoryEnabled {
		fmt.Printf("  - Memory Provider: %s\n", info.Config.MemoryProvider)
		fmt.Printf("  - Embedding Model: %s\n", info.Config.EmbeddingModel)
		fmt.Printf("  - RAG Enabled: %t\n", info.Config.RAGEnabled)
	}
	fmt.Printf("  - MCP Enabled: %t\n", info.Config.MCPEnabled)
	fmt.Println()

	if len(info.Agents) > 0 {
		fmt.Printf("Agents:\n")
		for name, agent := range info.Agents {
			fmt.Printf("  %s:\n", name)
			fmt.Printf("    Role: %s\n", agent.Role)
			fmt.Printf("    Description: %s\n", agent.Description)
			fmt.Printf("    Capabilities: %v\n", agent.Capabilities)
			if agent.LLM != nil {
				fmt.Printf("    LLM Temperature: %.1f\n", agent.LLM.Temperature)
				fmt.Printf("    LLM Max Tokens: %d\n", agent.LLM.MaxTokens)
			}
			if agent.RetryPolicy != nil {
				fmt.Printf("    Max Retries: %d\n", agent.RetryPolicy.MaxRetries)
			}
			if agent.RateLimit != nil {
				fmt.Printf("    Rate Limit: %d req/sec\n", agent.RateLimit.RequestsPerSecond)
			}
			fmt.Println()
		}
	}

	if len(info.MCPServers) > 0 {
		fmt.Printf("MCP Servers:\n")
		for _, server := range info.MCPServers {
			fmt.Printf("  %s:\n", server.Name)
			fmt.Printf("    Type: %s\n", server.Type)
			if server.Command != "" {
				fmt.Printf("    Command: %s\n", server.Command)
			}
			if server.Host != "" {
				fmt.Printf("    Host: %s:%d\n", server.Host, server.Port)
			}
			fmt.Printf("    Enabled: %t\n", server.Enabled)
			fmt.Println()
		}
	}

	// Example 3: Create a project from template (commented out to avoid actual creation)
	fmt.Println("3. Creating Project from Template:")
	fmt.Println("==================================")
	fmt.Printf("To create a project from the '%s' template, you would run:\n", templateName)
	fmt.Printf("  err := scaffold.CreateAgentProjectFromTemplate(\"%s\", \"my-project\")\n", templateName)
	fmt.Println()
	fmt.Println("This would generate:")
	fmt.Println("  - Complete Go project structure")
	fmt.Println("  - Agent implementations with configured capabilities")
	fmt.Println("  - Enhanced agentflow.toml with agent-specific settings")
	fmt.Println("  - Memory and MCP configurations if enabled")
	fmt.Println("  - Retry policies and rate limiting")
	fmt.Println("  - Custom system prompts for each agent")
	fmt.Println("  - Metadata and performance settings")
	fmt.Println()

	// Example 4: Show configuration generation capabilities
	fmt.Println("4. Configuration Features:")
	fmt.Println("=========================")
	fmt.Println("The template system generates comprehensive configurations including:")
	fmt.Println()
	fmt.Println("Agent Configuration:")
	fmt.Println("  ✓ Role-based agent definitions")
	fmt.Println("  ✓ Custom system prompts")
	fmt.Println("  ✓ Capability-based agent design")
	fmt.Println("  ✓ Agent-specific LLM settings")
	fmt.Println("  ✓ Timeout and performance tuning")
	fmt.Println()
	fmt.Println("Advanced Features:")
	fmt.Println("  ✓ Retry policies with exponential backoff")
	fmt.Println("  ✓ Rate limiting and burst control")
	fmt.Println("  ✓ Agent metadata and tagging")
	fmt.Println("  ✓ Performance optimization settings")
	fmt.Println("  ✓ Error handling and recovery")
	fmt.Println()
	fmt.Println("Integration Features:")
	fmt.Println("  ✓ Memory system configuration")
	fmt.Println("  ✓ RAG and embedding settings")
	fmt.Println("  ✓ MCP server definitions")
	fmt.Println("  ✓ Orchestration patterns")
	fmt.Println("  ✓ Provider-specific optimizations")
	fmt.Println()

	// Example 5: Template customization guidance
	fmt.Println("5. Template Customization:")
	fmt.Println("=========================")
	fmt.Println("Templates can be customized by:")
	fmt.Println()
	fmt.Println("1. Modifying existing templates in examples/templates/")
	fmt.Println("2. Creating new YAML template files")
	fmt.Println("3. Adjusting agent configurations:")
	fmt.Println("   - System prompts and roles")
	fmt.Println("   - Capabilities and specializations")
	fmt.Println("   - LLM parameters and settings")
	fmt.Println("   - Performance and reliability settings")
	fmt.Println()
	fmt.Println("4. Configuring integrations:")
	fmt.Println("   - Memory providers and embedding models")
	fmt.Println("   - MCP servers and tools")
	fmt.Println("   - Orchestration patterns")
	fmt.Println("   - Error handling strategies")
	fmt.Println()

	fmt.Println("=== Template System Demo Complete ===")
	fmt.Println()
	fmt.Println("💡 Next Steps:")
	fmt.Println("1. Explore the template files in examples/templates/")
	fmt.Println("2. Create a project using: scaffold.CreateAgentProjectFromTemplate()")
	fmt.Println("3. Customize templates for your specific use cases")
	fmt.Println("4. Leverage the comprehensive validation system")
	fmt.Println("5. Use the generated configurations as starting points")
}