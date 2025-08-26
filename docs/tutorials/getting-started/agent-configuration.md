---
title: "Agent Configuration"
description: "Master AgenticGoKit's configuration system to customize your agents"
prev:
  text: "Your First Agent"
  link: "./first-agent"
next:
  text: "Multi-Agent Basics"
  link: "./multi-agent-basics"
---

# Agent Configuration

Now that you've created your first agent, let's dive deep into AgenticGoKit's powerful configuration system. Understanding configuration is key to building sophisticated agent systems that behave exactly as you need them to.

## Learning Objectives

By the end of this section, you'll be able to:
- Master the `agentflow.toml` configuration structure
- Configure different LLM providers and their settings
- Create agents with distinct personalities and capabilities
- Use environment variables for secure and flexible configuration
- Validate and troubleshoot configuration issues
- Apply configuration best practices for different scenarios

## Prerequisites

Before starting, make sure you've completed:
- ✅ [Your First Agent](./first-agent.md) - Basic agent creation and execution

## The Configuration-First Philosophy

AgenticGoKit's configuration-first approach means you can create sophisticated agent behaviors without writing complex code. Everything from agent personalities to LLM settings is defined in your `agentflow.toml` file.

### Benefits of Configuration-First Design

- **Easy Experimentation**: Try different agent behaviors immediately
- **Non-Developer Friendly**: Team members can customize agents without coding
- **Environment Management**: Different configs for development, testing, production
- **Version Control**: Track agent behavior changes over time
- **Efficient Prototyping**: Test ideas efficiently without code changes

## Understanding agentflow.toml Structure

Let's explore each section of the configuration file:

```toml
# Project metadata
[agent_flow]
name = "my-agent-system"
version = "1.0.0"
description = "A sophisticated multi-agent system"

# LLM provider configuration
[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 1000

# Orchestration settings
[orchestration]
mode = "collaborative"
timeout_seconds = 30

# Individual agent definitions
[agents.agent_name]
role = "specialist_role"
description = "What this agent does"
system_prompt = "Agent's personality and instructions"
enabled = true
```

## LLM Provider Configuration

### OpenAI Configuration

```toml
[llm]
provider = "openai"
model = "gpt-4"                    # or "gpt-3.5-turbo", "gpt-4-turbo"
temperature = 0.7                  # 0.0 = focused, 1.0 = creative
max_tokens = 1000                  # Maximum response length
top_p = 1.0                       # Nucleus sampling parameter
frequency_penalty = 0.0           # Reduce repetition
presence_penalty = 0.0            # Encourage new topics
```

**Environment Variables:**
```bash
export OPENAI_API_KEY="your-api-key-here"
export OPENAI_ORG_ID="your-org-id"  # Optional
```

### Azure OpenAI Configuration

```toml
[llm]
provider = "azure"
model = "gpt-4"
deployment_name = "your-deployment"
temperature = 0.7
max_tokens = 1000
```

**Environment Variables:**
```bash
export AZURE_OPENAI_API_KEY="your-api-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export AZURE_OPENAI_DEPLOYMENT="your-deployment-name"
export AZURE_OPENAI_API_VERSION="2024-02-15-preview"  # Optional
```

### Ollama Configuration

```toml
[llm]
provider = "ollama"
model = "llama3.1:8b"             # or "gemma2:2b", "mistral:7b"
temperature = 0.7
max_tokens = 1000
host = "http://localhost:11434"   # Default Ollama host
```

**Environment Variables:**
```bash
export OLLAMA_HOST="http://localhost:11434"
```

## Agent Configuration Deep Dive

### Basic Agent Structure

```toml
[agents.my_agent]
role = "helpful_assistant"
description = "A general-purpose helpful assistant"
system_prompt = """
You are a helpful assistant that provides clear, accurate information.
Always be polite, professional, and thorough in your responses.
"""
enabled = true
```

### Advanced Agent Configuration

```toml
[agents.research_specialist]
role = "research_specialist"
description = "Expert at gathering and analyzing information"
system_prompt = """
You are a research specialist with expertise in:
- Finding reliable sources and data
- Analyzing information for accuracy and relevance
- Synthesizing complex information into clear summaries
- Identifying gaps in knowledge or research

When conducting research:
1. Always verify information from multiple sources
2. Cite your sources when possible
3. Acknowledge limitations or uncertainties
4. Provide structured, well-organized responses
"""
enabled = true
capabilities = ["research", "analysis", "synthesis"]
max_context_length = 4000
```

## Hands-On Configuration Examples

Let's create different types of agents to see configuration in action:

### Example 1: Technical Documentation Assistant

Create a new project and configure it as a technical writing assistant:

```bash
agentcli create tech-writer --template basic
cd tech-writer
```

Edit `agentflow.toml`:

```toml
[agent_flow]
name = "technical-documentation-assistant"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.3  # Lower temperature for more consistent technical writing
max_tokens = 1500

[agents.tech_writer]
role = "technical_writer"
description = "Specialist in creating clear, accurate technical documentation"
system_prompt = """
You are an expert technical writer who specializes in creating clear, 
comprehensive documentation for software projects. Your expertise includes:

- Writing clear, step-by-step instructions
- Creating well-structured documentation with proper headings
- Explaining complex technical concepts in accessible language
- Following documentation best practices and standards
- Including relevant code examples and snippets

When writing documentation:
1. Start with a clear overview of what you're documenting
2. Use consistent formatting and structure
3. Include practical examples and use cases
4. Anticipate common questions and edge cases
5. Write for your target audience's technical level
"""
enabled = true
```

Test your technical writer:

```bash
go run . -m "Write documentation for a REST API endpoint that creates a new user account"
```

### Example 2: Creative Writing Assistant

```toml
[agent_flow]
name = "creative-writing-assistant"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.8  # Higher temperature for more creativity
max_tokens = 2000

[agents.creative_writer]
role = "creative_writer"
description = "Imaginative assistant for creative writing projects"
system_prompt = """
You are a creative writing assistant with a passion for storytelling. 
You help writers with:

- Character development and backstories
- Plot structure and story arcs
- Dialogue writing and voice development
- Setting and world-building
- Creative problem-solving for story challenges

Your approach is:
- Encouraging and supportive
- Full of creative ideas and alternatives
- Focused on helping writers find their unique voice
- Knowledgeable about different genres and writing techniques

Always ask follow-up questions to better understand the writer's vision
and provide multiple creative options when possible.
"""
enabled = true
```

### Example 3: Data Analysis Assistant

```toml
[agent_flow]
name = "data-analysis-assistant"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.2  # Very low temperature for analytical precision
max_tokens = 1200

[agents.data_analyst]
role = "data_analyst"
description = "Expert in data analysis and statistical interpretation"
system_prompt = """
You are a data analyst with expertise in:

- Statistical analysis and interpretation
- Data visualization recommendations
- Identifying patterns and trends in data
- Explaining statistical concepts clearly
- Recommending appropriate analytical methods

When analyzing data or answering data-related questions:
1. Ask clarifying questions about the data and objectives
2. Suggest appropriate analytical approaches
3. Explain statistical concepts in accessible terms
4. Recommend visualizations that best represent the data
5. Highlight important caveats and limitations
6. Provide actionable insights and recommendations

Always be precise, methodical, and evidence-based in your responses.
"""
enabled = true
```

## Environment Variable Management

### Using Environment Variables in Configuration

You can reference environment variables in your TOML configuration:

```toml
[llm]
provider = "openai"
model = "${OPENAI_MODEL:-gpt-4}"  # Use env var or default to gpt-4
temperature = 0.7

[agents.assistant]
system_prompt = """
${AGENT_PERSONALITY:-You are a helpful assistant.}
Additional instructions: Always be professional and thorough.
"""
```

### Environment-Specific Configurations

Create different configuration files for different environments:

**`agentflow.dev.toml`** (Development):
```toml
[llm]
provider = "ollama"
model = "gemma2:2b"  # Smaller model for development
temperature = 0.7

[agents.assistant]
system_prompt = "You are a helpful assistant in development mode."
```

**`agentflow.prod.toml`** (Production):
```toml
[llm]
provider = "openai"
model = "gpt-4"  # More capable model for production
temperature = 0.5

[agents.assistant]
system_prompt = "You are a professional assistant providing high-quality responses."
```

Use different configs:
```bash
# Development
go run . -config agentflow.dev.toml -m "Test message"

# Production
go run . -config agentflow.prod.toml -m "Production message"
```

## Configuration Validation and Troubleshooting

### Validating Your Configuration

Always validate your configuration before running:

```bash
agentcli validate
```

Common validation errors and solutions:

**Invalid TOML Syntax:**
```
Error: Invalid TOML syntax at line 15
```
- Check for missing quotes, brackets, or commas
- Ensure proper indentation and structure

**Missing Required Fields:**
```
Error: Missing required field 'system_prompt' for agent 'assistant'
```
- Add all required fields for each agent
- Check the documentation for required vs. optional fields

**Invalid Provider Configuration:**
```
Error: Unknown LLM provider 'invalid_provider'
```
- Use supported providers: "openai", "azure", "ollama"
- Check spelling and case sensitivity

### Testing Configuration Changes

Create a simple test script to validate configuration changes:

```bash
# Test basic functionality
go run . -m "Hello, please introduce yourself"

# Test specific capabilities
go run . -m "What are your main strengths and specializations?"

# Test edge cases
go run . -m "How do you handle requests outside your expertise?"
```

## Advanced Configuration Patterns

### Multi-Model Configuration

Use different models for different agents:

```toml
[llm]
provider = "openai"
# Default model settings

[agents.creative_writer]
role = "creative_writer"
model_override = "gpt-4"  # Use GPT-4 for creative tasks
temperature = 0.8
system_prompt = "You are a creative writing assistant..."

[agents.fact_checker]
role = "fact_checker"
model_override = "gpt-3.5-turbo"  # Use efficient model for fact-checking
temperature = 0.1
system_prompt = "You are a precise fact-checker..."
```

### Conditional Configuration

Use environment variables to enable/disable features:

```toml
[agents.debug_assistant]
role = "debug_helper"
enabled = "${DEBUG_MODE:-false}"  # Only enabled when DEBUG_MODE=true
system_prompt = "You help with debugging and troubleshooting."
```

## Hands-On Exercises

### Exercise 1: Create a Specialized Agent

Create an agent for a domain you're interested in (cooking, fitness, finance, etc.):

1. Choose appropriate temperature settings
2. Write a detailed system prompt
3. Test with domain-specific questions
4. Iterate based on responses

### Exercise 2: Multi-Environment Setup

Create development and production configurations:

1. Use Ollama for development (local iteration)
2. Use OpenAI/Azure for production (higher quality)
3. Test both configurations
4. Document the differences

### Exercise 3: Agent Personality Comparison

Create two agents with different personalities for the same task:

1. Formal, professional agent
2. Casual, friendly agent
3. Test both with the same questions
4. Compare response styles

## Configuration Best Practices

### 1. Start Simple, Iterate

Begin with basic configurations and gradually add complexity:

```toml
# Start with this
[agents.assistant]
role = "assistant"
system_prompt = "You are a helpful assistant."

# Evolve to this
[agents.assistant]
role = "specialized_assistant"
description = "Expert in specific domain"
system_prompt = """
Detailed, multi-paragraph system prompt
with specific instructions and examples.
"""
capabilities = ["skill1", "skill2"]
```

### 2. Use Descriptive Names

Choose clear, descriptive names for agents and roles:

```toml
# Good
[agents.customer_support_specialist]
role = "customer_support"

# Better
[agents.technical_support_specialist]
role = "technical_support_tier_2"
```

### 3. Document Your Configurations

Add comments to explain configuration choices:

```toml
[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.3  # Low temperature for consistent technical responses
max_tokens = 1500  # Sufficient for detailed explanations

[agents.code_reviewer]
role = "code_reviewer"
# This agent specializes in Go code review and best practices
system_prompt = """..."""
```

### 4. Version Your Configurations

Track configuration changes in version control:

```bash
git add agentflow.toml
git commit -m "Update agent system prompt for better code review feedback"
```

## What You've Learned

✅ **Mastered agentflow.toml structure** and all configuration sections  
✅ **Configured multiple LLM providers** with appropriate settings  
✅ **Created specialized agents** with distinct personalities and capabilities  
✅ **Used environment variables** for flexible, secure configuration  
✅ **Applied validation and troubleshooting** techniques  
✅ **Implemented configuration best practices** for maintainable systems  
✅ **Built practical examples** for different use cases  

## Understanding Check

Before moving on, make sure you can:
- [ ] Create agents with different personalities and specializations
- [ ] Configure different LLM providers and their settings
- [ ] Use environment variables for sensitive configuration
- [ ] Validate configuration and troubleshoot common issues
- [ ] Explain the benefits of configuration-first design
- [ ] Apply best practices for maintainable configurations

## Next Steps

Now that you've mastered agent configuration, you're ready to explore one of AgenticGoKit's most powerful features: multi-agent collaboration. You'll learn how to orchestrate multiple specialized agents working together.

**[→ Continue to Multi-Agent Basics](./multi-agent-basics.md)**

---

::: tip Configuration Mastery
You now understand how to create sophisticated agent behaviors through configuration alone. This foundation will serve you well as you build more complex multi-agent systems where each agent has a specialized role and personality.
:::