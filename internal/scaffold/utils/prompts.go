package utils

import (
	"fmt"
	"strings"
)

// CreateSystemPrompt creates a specialized system prompt based on the agent's purpose and context
func CreateSystemPrompt(agent AgentInfo, agentIndex, totalAgents int, orchestrationMode string) string {
	var prompt strings.Builder

	// Base introduction
	prompt.WriteString(fmt.Sprintf("You are %s, %s.", agent.DisplayName, strings.ToLower(agent.Purpose)))
	prompt.WriteString("\n\n")

	// Role-specific instructions based on agent name
	prompt.WriteString(getRoleSpecificInstructions(agent, agentIndex, totalAgents))

	// Universal tool usage guidelines
	prompt.WriteString("\nTool Usage Strategy:\n")
	prompt.WriteString("- For stock prices/financial data: Use search tools to find current information\n")
	prompt.WriteString("- For current events/news: Use search tools for latest updates\n")
	prompt.WriteString("- For specific web content: Use fetch_content tool with URLs\n")
	prompt.WriteString("- Always prefer real data over general advice\n")
	prompt.WriteString("- Document tool usage and results clearly\n")

	prompt.WriteString("\nResponse Quality:\n")
	prompt.WriteString("- Provide specific, data-driven answers when possible\n")
	prompt.WriteString("- Extract and present key information clearly\n")
	prompt.WriteString("- Be conversational but professional\n")
	prompt.WriteString("- Integrate tool results naturally into responses\n")

	// Orchestration-specific instructions
	prompt.WriteString(getOrchestrationInstructions(orchestrationMode))

	return prompt.String()
}

// getRoleSpecificInstructions returns specialized instructions based on agent name and position
func getRoleSpecificInstructions(agent AgentInfo, agentIndex, totalAgents int) string {
	var instructions strings.Builder

	switch agent.Name {
	case "analyzer":
		instructions.WriteString("Core Responsibilities:\n")
		instructions.WriteString("- Perform deep analysis of input data and extract meaningful insights\n")
		instructions.WriteString("- Identify patterns, trends, and important characteristics\n")
		instructions.WriteString("- Use available MCP tools to gather additional data when needed\n")
		instructions.WriteString("- For financial queries: analyze market conditions, price movements, and trends\n")
		instructions.WriteString("- Provide structured analysis with clear findings and evidence\n")

	case "processor":
		instructions.WriteString("Core Responsibilities:\n")
		instructions.WriteString("- Process and transform data according to user requirements\n")
		instructions.WriteString("- Apply appropriate operations, calculations, or manipulations\n")
		instructions.WriteString("- Use available MCP tools to enhance processing capabilities\n")
		instructions.WriteString("- Ensure data integrity and accuracy throughout processing\n")
		instructions.WriteString("- Provide clear documentation of processing steps taken\n")

	case "validator":
		instructions.WriteString("Core Responsibilities:\n")
		instructions.WriteString("- Validate the accuracy and quality of previous agents' work\n")
		instructions.WriteString("- Cross-check information using available MCP tools\n")
		instructions.WriteString("- Identify any errors, inconsistencies, or missing information\n")
		instructions.WriteString("- For financial data: verify prices, dates, and calculation accuracy\n")
		instructions.WriteString("- Provide confidence scores and validation reports\n")

	case "transformer":
		instructions.WriteString("Core Responsibilities:\n")
		instructions.WriteString("- Transform data and responses into the desired format\n")
		instructions.WriteString("- Adapt content for different audiences or use cases\n")
		instructions.WriteString("- Use MCP tools to gather additional formatting context\n")
		instructions.WriteString("- Ensure transformed output maintains original meaning and accuracy\n")
		instructions.WriteString("- Provide clear and well-structured final outputs\n")

	case "enricher":
		instructions.WriteString("Core Responsibilities:\n")
		instructions.WriteString("- Enrich responses with additional context and information\n")
		instructions.WriteString("- Use MCP tools to gather supplementary data\n")
		instructions.WriteString("- Add relevant background information and explanations\n")
		instructions.WriteString("- Provide comprehensive answers that go beyond basic requirements\n")
		instructions.WriteString("- Ensure enriched content adds genuine value\n")

	case "researcher":
		instructions.WriteString("Core Responsibilities:\n")
		instructions.WriteString("- Research topics thoroughly using available MCP tools\n")
		instructions.WriteString("- Gather information from multiple reliable sources\n")
		instructions.WriteString("- Synthesize research findings into coherent insights\n")
		instructions.WriteString("- Fact-check information and verify source credibility\n")
		instructions.WriteString("- Present research with proper context and citations\n")

	case "data_collector":
		instructions.WriteString("Core Responsibilities:\n")
		instructions.WriteString("- Collect comprehensive data from specified sources\n")
		instructions.WriteString("- Use MCP tools to access real-time and historical data\n")
		instructions.WriteString("- Organize collected data in structured formats\n")
		instructions.WriteString("- Ensure data quality and completeness\n")
		instructions.WriteString("- Document data collection methodology and sources\n")

	case "report_generator":
		instructions.WriteString("Core Responsibilities:\n")
		instructions.WriteString("- Generate comprehensive reports from processed data\n")
		instructions.WriteString("- Structure information in clear, logical formats\n")
		instructions.WriteString("- Include executive summaries and key findings\n")
		instructions.WriteString("- Use visual elements when appropriate (tables, lists)\n")
		instructions.WriteString("- Ensure reports are actionable and well-organized\n")

	default:
		// Generic instructions based on position in workflow
		if agentIndex == 0 {
			instructions.WriteString("Core Responsibilities:\n")
			instructions.WriteString("- Process the original user request and provide initial analysis\n")
			instructions.WriteString("- Use available MCP tools to gather current and accurate information\n")
			instructions.WriteString("- For financial queries: get current prices, market data, and trends\n")
			instructions.WriteString("- Provide concrete answers with actual data rather than generic advice\n")
			instructions.WriteString("- Set a strong foundation for subsequent agents\n")
		} else if agentIndex == totalAgents-1 {
			instructions.WriteString("Core Responsibilities:\n")
			instructions.WriteString("- Provide the final, comprehensive response to the user\n")
			instructions.WriteString("- Synthesize insights from all previous agents\n")
			instructions.WriteString("- Present information in a clear, organized, and authoritative manner\n")
			instructions.WriteString("- Use MCP tools only if critical information is still missing\n")
			instructions.WriteString("- Ensure the response fully addresses the user's original question\n")
		} else {
			instructions.WriteString("Core Responsibilities:\n")
			instructions.WriteString("- Build upon previous agents' work with additional insights and analysis\n")
			instructions.WriteString("- Use available MCP tools to gather more specific information\n")
			instructions.WriteString("- Enhance the response with greater detail and context\n")
			instructions.WriteString("- Identify and fill any information gaps\n")
			instructions.WriteString("- Maintain focus on the user's original question\n")
		}
	}

	return instructions.String()
}

// getOrchestrationInstructions returns orchestration-specific instructions
func getOrchestrationInstructions(orchestrationMode string) string {
	switch orchestrationMode {
	case "collaborative":
		return "\nCollaborative Mode: You work in parallel with other agents. Focus on your specific expertise area."
	case "sequential":
		return "\nSequential Mode: You process tasks in sequence. Build upon previous agents' work effectively."
	case "loop":
		return "\nLoop Mode: You may process the same task multiple times. Improve your response with each iteration."
	case "mixed":
		return "\nMixed Mode: You operate in a hybrid environment combining parallel and sequential processing."
	default:
		return "\nRoute Mode: Process tasks and route to appropriate next steps in the workflow."
	}
}
