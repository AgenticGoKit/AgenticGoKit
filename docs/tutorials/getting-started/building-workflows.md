---
title: "Building Workflows"
description: "Create sophisticated, real-world agent workflows combining all AgenticGoKit capabilities"
prev:
  text: "Tool Integration"
  link: "./tool-integration"
next:
  text: "Troubleshooting"
  link: "./troubleshooting"
---

# Building Workflows

You've mastered individual AgenticGoKit capabilities: agent configuration, multi-agent orchestration, memory systems, and tool integration. Now let's bring everything together to build sophisticated, real-world workflows that solve complex problems.

## Learning Objectives

By the end of this section, you'll be able to:
- Design comprehensive workflows combining multiple AgenticGoKit capabilities
- Build production-ready systems with error handling and resilience
- Create workflows for common business scenarios (research, content creation, data processing)
- Implement monitoring, logging, and performance optimization
- Apply best practices for scalable, maintainable agent systems
- Troubleshoot and debug complex multi-component workflows

## Prerequisites

Before starting, make sure you've completed:
- ✅ [Tool Integration](./tool-integration.md) - Understanding of MCP tools and external integrations

## What Makes a Great Workflow?

A well-designed agent workflow should be:

- **Purpose-Driven**: Solves a specific, well-defined problem
- **Resilient**: Handles errors gracefully and recovers from failures
- **Efficient**: Uses resources optimally and completes tasks in reasonable time
- **Maintainable**: Easy to understand, modify, and extend
- **Observable**: Provides visibility into operations and performance
- **Scalable**: Can handle increased load and complexity

## Workflow Design Patterns

::: info Pattern Selection Guide
Choose the right pattern based on your use case:
- **Research Pipeline**: Information gathering and analysis
- **Content Creation**: Multi-stage content development
- **Data Processing**: Sequential data transformation
:::

### Research and Analysis Pipeline

Perfect for gathering information, analyzing it, and producing insights:

```mermaid
graph TD
    Input[Research Request] --> Planner[Planning Agent]
    Planner --> Researcher[Research Agent]
    Planner --> Validator[Validation Agent]
    Researcher --> |Web Search| WebTools[Web Tools]
    Researcher --> |Save Data| FileTools[File Tools]
    Researcher --> Analyzer[Analysis Agent]
    Validator --> |Check Sources| WebTools
    Analyzer --> |Read Data| FileTools
    Analyzer --> Writer[Report Writer]
    Writer --> |Save Report| FileTools
    Writer --> Output[Final Report]
```

### Content Creation Workflow

Ideal for creating high-quality content with multiple review stages:

```mermaid
graph TD
    Brief[Content Brief] --> Researcher[Research Agent]
    Researcher --> |Gather Info| Tools[Web/File Tools]
    Researcher --> Outliner[Outline Agent]
    Outliner --> Writer[Content Writer]
    Writer --> Editor[Editor Agent]
    Editor --> Reviewer[Quality Reviewer]
    Reviewer --> |Approved| Final[Final Content]
    Reviewer --> |Needs Work| Writer
```

### Data Processing Pipeline

Great for handling data transformation and analysis tasks:

```mermaid
graph LR
    Data[Raw Data] --> Validator[Data Validator]
    Validator --> Cleaner[Data Cleaner]
    Cleaner --> Processor[Data Processor]
    Processor --> Analyzer[Data Analyzer]
    Analyzer --> Visualizer[Report Generator]
    Visualizer --> Output[Final Reports]
```

## Building a Comprehensive Research Workflow

Let's build a sophisticated research system that combines all AgenticGoKit capabilities:

### Step 1: Create the Project

```bash
agentcli create research-workflow --template advanced
cd research-workflow
```

### Step 2: Design the Complete Configuration

```toml
[agent_flow]
name = "comprehensive-research-workflow"
version = "1.0.0"
description = "Advanced research workflow with memory, tools, and multi-agent orchestration"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000

# Memory Configuration
[agent_memory]
provider = "pgvector"
enable_rag = true
enable_knowledge_base = true
chunk_size = 1000
overlap_size = 200
max_results = 5

[agent_memory.pgvector]
connection_string = "postgresql://agent:agentpass@localhost:5432/research_db"

# MCP Tool Configuration
[mcp]
enabled = true
timeout_seconds = 60

[[mcp.servers]]
name = "web-search"
command = "uvx"
args = ["mcp-server-web-search"]
env = { "SEARCH_ENGINE" = "duckduckgo" }

[[mcp.servers]]
name = "filesystem"
command = "uvx"
args = ["mcp-server-filesystem"]
env = { "ALLOWED_DIRECTORIES" = "./workspace,./reports,./data" }

[[mcp.servers]]
name = "http-client"
command = "uvx"
args = ["mcp-server-fetch"]

# Orchestration Configuration
[orchestration]
mode = "sequential"
sequential_agents = ["planner", "researcher", "analyzer", "validator", "synthesizer", "reporter"]
timeout_seconds = 300
retry_attempts = 2

# Agent Definitions
[agents.planner]
role = "research_planner"
description = "Plans and coordinates research activities"
system_prompt = """
You are a research planning specialist who creates comprehensive research strategies.

Your responsibilities:
1. Analyze research requests to understand scope and objectives
2. Break down complex topics into manageable research areas
3. Identify key questions that need to be answered
4. Plan the research methodology and approach
5. Set priorities and timelines for research activities

When planning research:
- Ask clarifying questions if the request is ambiguous
- Identify primary and secondary research objectives
- Consider multiple perspectives and potential biases
- Plan for validation and fact-checking
- Create a structured research outline

Output your plan as a structured document with:
- Research objectives
- Key questions to investigate
- Methodology and approach
- Priority areas
- Success criteria
"""
enabled = true
memory_enabled = true
tools_enabled = true

[agents.researcher]
role = "primary_researcher"
description = "Conducts comprehensive research using multiple sources"
system_prompt = """
You are a primary researcher who gathers comprehensive information from multiple sources.

Your capabilities include:
- Web search for current information
- Reading and analyzing web content
- Accessing knowledge bases and documents
- Organizing research findings systematically
- Identifying reliable and authoritative sources

Research methodology:
1. Follow the research plan provided by the planner
2. Use multiple search strategies and keywords
3. Verify information from multiple sources
4. Save important findings with proper citations
5. Organize information by topic and relevance
6. Note any gaps or limitations in available information

Always prioritize accuracy, currency, and source reliability.
Save all research findings to organized files for further analysis.
"""
enabled = true
memory_enabled = true
tools_enabled = true

[agents.analyzer]
role = "research_analyst"
description = "Analyzes research data to extract insights and patterns"
system_prompt = """
You are a research analyst who processes information to extract meaningful insights.

Your analytical capabilities:
- Pattern recognition and trend identification
- Statistical analysis and interpretation
- Comparative analysis across sources
- Gap analysis and limitation identification
- Insight generation and hypothesis formation

Analysis process:
1. Read and process all research files systematically
2. Identify key themes, patterns, and trends
3. Compare and contrast different sources and perspectives
4. Evaluate the strength and reliability of evidence
5. Generate insights and draw preliminary conclusions
6. Identify areas needing additional research

Create structured analysis reports with:
- Key findings and insights
- Supporting evidence and sources
- Confidence levels and limitations
- Recommendations for further investigation
"""
enabled = true
memory_enabled = true
tools_enabled = true

[agents.validator]
role = "fact_validator"
description = "Validates research findings and checks source reliability"
system_prompt = """
You are a fact validation specialist who ensures research accuracy and reliability.

Validation responsibilities:
- Verify factual claims against multiple sources
- Assess source credibility and authority
- Identify potential biases or conflicts of interest
- Check for currency and relevance of information
- Flag questionable or unverified claims

Validation process:
1. Review all research findings and analysis
2. Cross-reference key facts with authoritative sources
3. Evaluate source credibility using established criteria
4. Identify any inconsistencies or contradictions
5. Research additional sources to resolve discrepancies
6. Create validation reports with confidence ratings

Provide validation reports that include:
- Verified facts with confidence ratings
- Source reliability assessments
- Identified inconsistencies or gaps
- Recommendations for additional verification
"""
enabled = true
memory_enabled = true
tools_enabled = true

[agents.synthesizer]
role = "information_synthesizer"
description = "Synthesizes research, analysis, and validation into coherent insights"
system_prompt = """
You are an information synthesis specialist who combines multiple inputs into coherent insights.

Synthesis capabilities:
- Integrating information from multiple sources and perspectives
- Resolving contradictions and inconsistencies
- Creating unified narratives from diverse inputs
- Identifying overarching themes and conclusions
- Balancing different viewpoints and evidence levels

Synthesis process:
1. Review all research, analysis, and validation reports
2. Identify common themes and consistent findings
3. Resolve contradictions using evidence strength
4. Create unified insights that address research objectives
5. Acknowledge limitations and areas of uncertainty
6. Prepare synthesized findings for final reporting

Create synthesis reports with:
- Unified insights and conclusions
- Supporting evidence from multiple sources
- Confidence levels and limitations
- Balanced perspectives on controversial topics
"""
enabled = true
memory_enabled = true
tools_enabled = true

[agents.reporter]
role = "report_writer"
description = "Creates comprehensive, professional research reports"
system_prompt = """
You are a professional report writer who creates comprehensive research reports.

Report writing expertise:
- Clear, professional writing style
- Logical structure and organization
- Executive summary creation
- Visual description and data presentation
- Citation and reference management

Report structure:
1. Executive Summary
2. Research Objectives and Methodology
3. Key Findings and Insights
4. Detailed Analysis
5. Conclusions and Recommendations
6. Limitations and Future Research
7. References and Sources

Writing guidelines:
- Use clear, accessible language
- Structure information logically
- Include executive summaries for busy readers
- Cite all sources appropriately
- Provide actionable recommendations
- Acknowledge limitations honestly

Save the final report in multiple formats for different audiences.
"""
enabled = true
memory_enabled = true
tools_enabled = true
```

### Step 3: Test the Complete Workflow

```bash
go run . -m "Conduct a comprehensive research study on the impact of artificial intelligence on job markets in 2024, including current trends, future projections, and policy recommendations"
```

This will trigger the complete workflow:
1. **Planner** creates a research strategy
2. **Researcher** gathers information from web and knowledge bases
3. **Analyzer** processes the data to extract insights
4. **Validator** fact-checks and verifies sources
5. **Synthesizer** combines all inputs into coherent findings
6. **Reporter** creates a professional final report

## Content Creation Workflow Example

Let's build a content creation system for marketing materials:

```toml
[agent_flow]
name = "content-creation-workflow"
version = "1.0.0"

[orchestration]
mode = "collaborative"
collaborative_agents = ["researcher", "strategist", "writer", "editor"]

[agents.researcher]
role = "content_researcher"
system_prompt = """
You research topics for content creation, focusing on:
- Target audience interests and pain points
- Competitor content analysis
- Trending topics and keywords
- Supporting data and statistics
- Expert opinions and quotes

Provide comprehensive research briefs for content creators.
"""
enabled = true
memory_enabled = true
tools_enabled = true

[agents.strategist]
role = "content_strategist"
system_prompt = """
You develop content strategies based on research, including:
- Content positioning and messaging
- Target audience definition
- Content format recommendations
- Distribution channel strategy
- Success metrics and KPIs

Create strategic briefs that guide content creation.
"""
enabled = true
memory_enabled = true

[agents.writer]
role = "content_writer"
system_prompt = """
You create engaging content based on research and strategy:
- Write compelling headlines and hooks
- Develop clear, engaging body content
- Include relevant data and examples
- Optimize for target audience and channels
- Ensure brand voice consistency

Create content that achieves strategic objectives.
"""
enabled = true
memory_enabled = true

[agents.editor]
role = "content_editor"
system_prompt = """
You review and refine content for quality and effectiveness:
- Check grammar, style, and clarity
- Ensure strategic alignment
- Verify facts and sources
- Optimize for readability and engagement
- Provide improvement recommendations

Deliver polished, professional content ready for publication.
"""
enabled = true
memory_enabled = true
```

## Data Processing Workflow Example

Create a system for processing and analyzing business data:

```toml
[agent_flow]
name = "data-processing-workflow"
version = "1.0.0"

[orchestration]
mode = "sequential"
sequential_agents = ["collector", "validator", "processor", "analyzer", "visualizer"]

[agents.collector]
role = "data_collector"
system_prompt = """
You collect data from various sources:
- API endpoints and databases
- File uploads and imports
- Web scraping and data extraction
- Real-time data streams
- Manual data entry validation

Ensure data completeness and initial quality checks.
"""
enabled = true
tools_enabled = true

[agents.validator]
role = "data_validator"
system_prompt = """
You validate and clean collected data:
- Check data completeness and consistency
- Identify and handle missing values
- Detect and correct data quality issues
- Standardize formats and structures
- Flag anomalies and outliers

Prepare clean, validated datasets for processing.
"""
enabled = true
tools_enabled = true

[agents.processor]
role = "data_processor"
system_prompt = """
You process and transform validated data:
- Apply business rules and calculations
- Aggregate and summarize data
- Create derived metrics and KPIs
- Join and merge datasets
- Prepare data for analysis

Generate processed datasets ready for analysis.
"""
enabled = true
tools_enabled = true

[agents.analyzer]
role = "data_analyst"
system_prompt = """
You analyze processed data to extract insights:
- Perform statistical analysis
- Identify trends and patterns
- Compare performance metrics
- Generate predictive insights
- Create analytical summaries

Provide actionable insights and recommendations.
"""
enabled = true
memory_enabled = true

[agents.visualizer]
role = "report_generator"
system_prompt = """
You create reports and visualizations:
- Generate executive dashboards
- Create detailed analytical reports
- Design data visualizations
- Prepare presentation materials
- Document methodology and findings

Deliver comprehensive reporting packages.
"""
enabled = true
tools_enabled = true
```

## Advanced Workflow Features

### Error Handling and Resilience

```toml
[orchestration]
mode = "sequential"
sequential_agents = ["agent1", "agent2", "agent3"]
timeout_seconds = 300
retry_attempts = 3
fallback_strategy = "continue"  # or "abort", "retry"

# Agent-specific error handling
[agents.resilient_agent]
system_prompt = """
You handle errors gracefully by:
1. Attempting alternative approaches when primary methods fail
2. Providing partial results when complete results aren't available
3. Clearly communicating limitations and uncertainties
4. Suggesting manual intervention when needed
5. Logging detailed error information for debugging

Always try to provide value even when facing challenges.
"""
error_recovery = true
max_retries = 2
```

### Performance Monitoring

```toml
[monitoring]
enabled = true
log_level = "info"
metrics_collection = true
performance_tracking = true

[monitoring.alerts]
max_execution_time = 600  # seconds
memory_threshold = "1GB"
error_rate_threshold = 0.1
```

### Workflow Optimization

```toml
# Parallel processing where possible
[orchestration]
mode = "mixed"
parallel_groups = [
    ["researcher1", "researcher2", "researcher3"],
    ["analyzer", "validator"],
    ["writer"]
]

# Resource management
[resources]
max_concurrent_agents = 5
memory_limit = "2GB"
cpu_limit = "4"
```

## Hands-On Exercises

### Exercise 1: Build a Customer Service Workflow

Create a comprehensive customer service system:

1. **Intake Agent**: Categorizes and routes customer inquiries
2. **Specialist Agents**: Handle specific types of issues (technical, billing, etc.)
3. **Knowledge Agent**: Searches knowledge bases for solutions
4. **Escalation Agent**: Handles complex issues requiring human intervention
5. **Follow-up Agent**: Ensures customer satisfaction

### Exercise 2: Create a Market Research Pipeline

Build a system for conducting market research:

1. **Planning Agent**: Defines research objectives and methodology
2. **Data Collection Agents**: Gather data from multiple sources
3. **Analysis Agent**: Processes and analyzes collected data
4. **Insight Agent**: Generates strategic insights and recommendations
5. **Presentation Agent**: Creates executive summaries and reports

### Exercise 3: Develop a Content Audit Workflow

Create a system for auditing and improving existing content:

1. **Inventory Agent**: Catalogs existing content
2. **Analysis Agent**: Evaluates content performance and quality
3. **Gap Agent**: Identifies content gaps and opportunities
4. **Recommendation Agent**: Suggests improvements and new content
5. **Planning Agent**: Creates content improvement roadmaps

## Best Practices for Production Workflows

### 1. Design for Observability

```toml
[logging]
level = "info"
structured = true
include_agent_context = true
include_performance_metrics = true

[agents.observable_agent]
system_prompt = """
Always log important decisions and actions:
- What information you're processing
- What decisions you're making and why
- What actions you're taking
- Any issues or limitations encountered
- Performance and timing information

This helps with debugging and optimization.
"""
```

### 2. Implement Graceful Degradation

```toml
[agents.resilient_agent]
system_prompt = """
When facing issues:
1. Try alternative approaches before failing
2. Provide partial results when possible
3. Clearly communicate what you couldn't accomplish
4. Suggest manual steps or workarounds
5. Continue the workflow when possible

The goal is to provide maximum value even under adverse conditions.
"""
```

### 3. Use Configuration Management

```bash
# Environment-specific configurations
cp agentflow.toml agentflow.dev.toml
cp agentflow.toml agentflow.prod.toml

# Use environment variables for sensitive data
export OPENAI_API_KEY="your-key"
export DATABASE_URL="your-db-url"
```

### 4. Implement Testing Strategies

```bash
# Unit test individual agents
agentcli test-agent researcher "Test research query"

# Integration test workflows
agentcli test-workflow research-workflow "Integration test"

# Performance test under load
agentcli load-test --concurrent 10 --duration 60s
```

### 5. Monitor and Optimize Performance

```bash
# Monitor workflow performance
agentcli monitor --workflow research-workflow

# Analyze performance metrics
agentcli metrics --agent researcher --timeframe 24h

# Optimize based on bottlenecks
agentcli optimize --workflow research-workflow
```

## What You've Learned

✅ **Comprehensive workflow design** combining all AgenticGoKit capabilities  
✅ **Production-ready patterns** with error handling and resilience  
✅ **Real-world workflow examples** for common business scenarios  
✅ **Advanced features** including monitoring, optimization, and scaling  
✅ **Best practices** for maintainable, observable systems  
✅ **Testing and debugging** strategies for complex workflows  
✅ **Performance optimization** techniques for production deployment  

## Understanding Check

Before moving on, make sure you can:
- [ ] Design comprehensive workflows that solve real business problems
- [ ] Implement error handling and resilience in multi-agent systems
- [ ] Configure monitoring and observability for production workflows
- [ ] Apply performance optimization techniques
- [ ] Test and debug complex multi-component systems
- [ ] Choose appropriate orchestration patterns for different scenarios
- [ ] Build scalable, maintainable agent workflows

## Next Steps

You've now built sophisticated, production-ready workflows! The final step in your learning journey is mastering troubleshooting and debugging techniques to keep your systems running smoothly.

**[→ Continue to Troubleshooting](./troubleshooting.md)**

---

::: details Quick Navigation

**Previous:** [Tool Integration](./tool-integration.md) - External tools and services  
**Next:** [Troubleshooting](./troubleshooting.md) - Debug and maintain your systems  
**Related:** [Advanced Patterns](../advanced/README.md) - Complex orchestration techniques  

:::

::: details Production Deployment

**Ready for production?**
- [Deployment Guide](../../guides/deployment/README.md) - Production deployment strategies
- [Monitoring Setup](../../guides/monitoring/README.md) - System health and observability
- [Security Best Practices](../../guides/security/README.md) - Secure your agent systems

**Performance Optimization:**
- [Scaling Patterns](../../guides/scaling/README.md) - Handle increased load
- [Performance Tuning](../../guides/performance/README.md) - Optimize for speed and efficiency
- [Resource Management](../../guides/resources/README.md) - Manage compute and memory

:::

::: tip Workflow Mastery
You've just learned to build sophisticated, real-world agent systems that combine all of AgenticGoKit's capabilities. These patterns and techniques will serve as the foundation for any complex agent system you build in the future.
:::