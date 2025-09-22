package utils

import (
	_ "embed"
	"fmt"
	"strings"
)

// Embedded prompt templates
//
//go:embed templates/prompts/rag_document_ingester.txt
var ragDocumentIngesterTemplate string

//go:embed templates/prompts/rag_query_processor.txt
var ragQueryProcessorTemplate string

//go:embed templates/prompts/rag_response_generator.txt
var ragResponseGeneratorTemplate string

//go:embed templates/prompts/rag_retrieval_agent.txt
var ragRetrievalAgentTemplate string

//go:embed templates/prompts/tool_usage_guidelines.txt
var toolUsageGuidelinesTemplate string

//go:embed templates/prompts/response_quality_standards.txt
var responseQualityStandardsTemplate string

//go:embed templates/prompts/sequential_first_agent.txt
var sequentialFirstAgentTemplate string

//go:embed templates/prompts/sequential_final_agent.txt
var sequentialFinalAgentTemplate string

//go:embed templates/prompts/sequential_middle_agent.txt
var sequentialMiddleAgentTemplate string

// Prompt constants for reusable text blocks
const (
	CoreResponsibilitiesCollaborative = `Core Responsibilities:
1. Focus on your area of specialization while considering the broader context
2. Provide thorough analysis within your domain of expertise
3. Use available tools to gather current and accurate information
4. Contribute insights that complement other agents' work
5. Maintain consistency with the overall response strategy
6. Ensure your contribution adds genuine value to the final output`

	CollaborativeGuidelines = `Collaborative Guidelines:
- Work in parallel with other agents to cover different aspects of the query
- Focus on your specialized capabilities and knowledge area
- Provide comprehensive coverage of your assigned domain
- Ensure your output can be effectively integrated with others`

	IterativeProcessingGuidelines = `Iterative Processing Guidelines:
1. Analyze the current state and any previous iterations
2. Identify areas for improvement or additional detail
3. Use tools to gather new information or verify existing data
4. Build upon previous iterations while addressing gaps
5. Refine accuracy, completeness, and clarity with each pass
6. Determine when the response is sufficiently comprehensive`

	LoopModeStrategy = `Loop Mode Strategy:
- Each iteration should add meaningful value
- Focus on continuous improvement rather than repetition
- Use feedback and context from previous iterations
- Aim for convergence toward an optimal response`

	CoreProcessingGuidelines = `Core Processing Guidelines:
1. Analyze user queries thoroughly to understand intent and requirements
2. Use available tools to gather current and accurate information
3. Provide comprehensive responses that fully address user needs
4. Route tasks to appropriate agents when specialized processing is needed
5. Ensure responses are well-structured and professionally presented
6. Maintain high standards for accuracy and relevance`

	RoutingStrategy = `Routing Strategy:
- Assess the complexity and requirements of each query
- Direct tasks to the most appropriate agent or processing path
- Ensure efficient workflow and optimal resource utilization
- Coordinate between different agents as needed`

	SequentialModeGuidelinesFirst = `Sequential Mode Guidelines:
- You are the foundation agent - set a strong foundation for subsequent processing
- Provide comprehensive initial analysis for the next agent to build upon`

	SequentialModeGuidelinesFinal = `Sequential Mode Guidelines:
- You receive processed information from all previous agents
- Create the final user-facing response that addresses their original query`

	SequentialModeGuidelinesMiddle = `Sequential Mode Guidelines:
- You receive input from previous agents and pass enhanced output to the next agent
- Build meaningfully upon previous work while staying focused on the end goal`
)

// CreateSystemPrompt creates a specialized system prompt based on the agent's purpose and context
func CreateSystemPrompt(agent AgentInfo, agentIndex, totalAgents int, orchestrationMode string) string {
	// Create comprehensive, template-quality system prompts
	switch orchestrationMode {
	case "sequential":
		return createSequentialSystemPrompt(agent, agentIndex, totalAgents)
	case "collaborative":
		return createCollaborativeSystemPrompt(agent, agentIndex, totalAgents)
	case "loop":
		return createLoopSystemPrompt(agent, agentIndex, totalAgents)
	default:
		return createGenericSystemPrompt(agent, agentIndex, totalAgents)
	}
}

// createSequentialSystemPrompt creates detailed system prompts for sequential workflows
func createSequentialSystemPrompt(agent AgentInfo, agentIndex, totalAgents int) string {
	var prompt strings.Builder

	// Base introduction with clear role definition
	if agentIndex == 0 {
		prompt.WriteString(fmt.Sprintf("You are %s, the first agent in a sequential multi-agent system. ", agent.DisplayName))
	} else if agentIndex == totalAgents-1 {
		prompt.WriteString(fmt.Sprintf("You are %s, the final agent in a sequential multi-agent system. ", agent.DisplayName))
	} else {
		prompt.WriteString(fmt.Sprintf("You are %s, agent %d of %d in a sequential multi-agent system. ", agent.DisplayName, agentIndex+1, totalAgents))
	}

	// Add purpose if specified
	if agent.Purpose != "" {
		prompt.WriteString(fmt.Sprintf("Your specialized purpose is to %s. ", agent.Purpose))
	}
	prompt.WriteString("\n")

	// Add role-specific detailed instructions using templates
	switch {
	case agentIndex == 0:
		prompt.WriteString(sequentialFirstAgentTemplate)
	case agentIndex == totalAgents-1:
		prompt.WriteString(sequentialFinalAgentTemplate)
	default:
		prompt.WriteString(sequentialMiddleAgentTemplate)
	}

	// Add universal guidelines
	prompt.WriteString("\n\n")
	prompt.WriteString(toolUsageGuidelinesTemplate)
	prompt.WriteString("\n\n")
	prompt.WriteString(responseQualityStandardsTemplate)
	prompt.WriteString("\n\n")

	// Add sequential mode specific guidelines
	switch {
	case agentIndex == 0:
		prompt.WriteString(SequentialModeGuidelinesFirst)
	case agentIndex == totalAgents-1:
		prompt.WriteString(SequentialModeGuidelinesFinal)
	default:
		prompt.WriteString(SequentialModeGuidelinesMiddle)
	}

	return prompt.String()
}

// createCollaborativeSystemPrompt creates detailed system prompts for collaborative workflows
func createCollaborativeSystemPrompt(agent AgentInfo, agentIndex, totalAgents int) string {
	var prompt strings.Builder

	// Check for RAG-specific agents first and provide specialized prompts
	switch {
	case strings.Contains(agent.Name, "document") && strings.Contains(agent.Name, "ingester"):
		return ragDocumentIngesterTemplate
	case strings.Contains(agent.Name, "query") && strings.Contains(agent.Name, "processor"):
		return ragQueryProcessorTemplate
	case strings.Contains(agent.Name, "response") && strings.Contains(agent.Name, "generator"):
		return ragResponseGeneratorTemplate
	case strings.Contains(agent.Name, "retrieval") || strings.Contains(agent.Name, "retriever"):
		return ragRetrievalAgentTemplate
	}

	prompt.WriteString(fmt.Sprintf("You are %s, one of %d agents working collaboratively to provide the best possible response. ", agent.DisplayName, totalAgents))
	prompt.WriteString("Your role is to contribute your specialized expertise while coordinating with other agents.\n")

	// Add purpose if specified
	if agent.Purpose != "" {
		prompt.WriteString(fmt.Sprintf("Your specialized purpose is to %s.\n", agent.Purpose))
	}
	prompt.WriteString("\n")

	prompt.WriteString(CoreResponsibilitiesCollaborative)
	prompt.WriteString("\n\n")
	prompt.WriteString(CollaborativeGuidelines)
	prompt.WriteString("\n\n")

	addUniversalGuidelines(&prompt)
	return prompt.String()
}

// createLoopSystemPrompt creates detailed system prompts for loop workflows
func createLoopSystemPrompt(agent AgentInfo, agentIndex, totalAgents int) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("You are %s, operating in an iterative loop mode where you may process the same query multiple times. ", agent.DisplayName))
	prompt.WriteString("Your goal is to improve and refine your response with each iteration.\n")

	// Add purpose if specified
	if agent.Purpose != "" {
		prompt.WriteString(fmt.Sprintf("Your specialized purpose is to %s.\n", agent.Purpose))
	}
	prompt.WriteString("\n")

	prompt.WriteString(IterativeProcessingGuidelines)
	prompt.WriteString("\n\n")
	prompt.WriteString(LoopModeStrategy)
	prompt.WriteString("\n\n")

	addUniversalGuidelines(&prompt)
	return prompt.String()
}

// createGenericSystemPrompt creates detailed system prompts for route and other modes
func createGenericSystemPrompt(agent AgentInfo, agentIndex, totalAgents int) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("You are %s, an intelligent agent in a multi-agent system designed to provide comprehensive assistance. ", agent.DisplayName))
	prompt.WriteString("Your role is to process user queries effectively and route tasks appropriately.\n")

	// Add purpose if specified
	if agent.Purpose != "" {
		prompt.WriteString(fmt.Sprintf("Your specialized purpose is to %s.\n", agent.Purpose))
	}
	prompt.WriteString("\n")

	prompt.WriteString(CoreProcessingGuidelines)
	prompt.WriteString("\n\n")
	prompt.WriteString(RoutingStrategy)
	prompt.WriteString("\n\n")

	addUniversalGuidelines(&prompt)
	return prompt.String()
}

// addUniversalGuidelines adds common guidelines to all prompt types
func addUniversalGuidelines(prompt *strings.Builder) {
	prompt.WriteString(toolUsageGuidelinesTemplate)
	prompt.WriteString("\n\n")
	prompt.WriteString(responseQualityStandardsTemplate)
}
