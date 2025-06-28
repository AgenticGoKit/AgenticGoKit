# Documentation Standards

This guide outlines the standards and best practices for writing and maintaining AgentFlow documentation.

## Table of Contents

- [Documentation Philosophy](#documentation-philosophy)
- [Structure and Organization](#structure-and-organization)
- [Writing Guidelines](#writing-guidelines)
- [Code Examples](#code-examples)
- [File Naming and Organization](#file-naming-and-organization)
- [Maintenance and Updates](#maintenance-and-updates)

## Documentation Philosophy

AgentFlow documentation follows these core principles:

### 1. User-Centric Approach
- **Start with the user's goal** - What are they trying to accomplish?
- **Provide immediate value** - Get users to success quickly
- **Explain the "why"** not just the "how"
- **Include real-world context** and use cases

### 2. Clarity and Accessibility
- **Write for beginners** while providing depth for experts
- **Use clear, concise language** - avoid jargon when possible
- **Structure content logically** with clear headings and flow
- **Make content scannable** with bullet points, code blocks, and visual breaks

### 3. Accuracy and Completeness
- **Keep examples working** - test all code samples
- **Stay current** with the codebase - documentation should never lag behind features
- **Provide complete context** - don't assume prior knowledge
- **Link between related concepts** to build understanding

### 4. Separation of Concerns
- **User docs** focus on building with AgentFlow
- **Contributor docs** focus on extending AgentFlow
- **API reference** provides comprehensive technical details
- **Examples** demonstrate practical applications

## Structure and Organization

### Documentation Hierarchy

```
docs/
├── README.md                    # Main documentation index
├── Architecture.md              # High-level system overview
├── ROADMAP.md                   # Project roadmap (maintained separately)
├── guides/                      # User-focused tutorials and guides
│   ├── AgentBasics.md          # Getting started with agents
│   ├── Examples.md             # Practical code examples
│   ├── ToolIntegration.md      # MCP and tool usage
│   ├── Providers.md            # LLM provider setup
│   ├── Configuration.md        # Project configuration
│   ├── Production.md           # Deployment and scaling
│   ├── ErrorHandling.md        # Error handling patterns
│   ├── CustomTools.md          # Building MCP servers
│   └── Performance.md          # Optimization guide
├── api/                        # Technical API reference
│   ├── core.md                 # Core package API
│   ├── agents.md              # Agent interfaces
│   ├── mcp.md                 # MCP integration API
│   └── cli.md                 # CLI command reference
├── contributors/               # Contributor-focused documentation
│   ├── ContributorGuide.md    # Getting started contributing
│   ├── CoreVsInternal.md      # Codebase architecture
│   ├── Testing.md             # Testing strategy
│   ├── ReleaseProcess.md      # Release management
│   ├── AddingFeatures.md      # Feature development
│   ├── CodeStyle.md           # Code standards
│   └── DocsStandards.md       # This document
└── archive/                   # Archived/outdated documents
    └── ...                    # Migration docs, old plans, etc.
```

### Cross-References

Always provide clear navigation paths:
- **Forward references**: Link to related concepts users will need
- **Backward references**: Link to prerequisite knowledge
- **Lateral references**: Link to alternative approaches or related topics

Example:
```markdown
## Configuration

Before configuring agents, make sure you've completed the [basic setup](AgentBasics.md#setup).

For production deployments, see the [Production Guide](Production.md) for advanced configuration options.

Related: [LLM Providers](Providers.md) | [Error Handling](ErrorHandling.md)
```

## Writing Guidelines

### Voice and Tone

**For User Documentation:**
- **Encouraging and supportive** - "You can easily..."
- **Direct and action-oriented** - Use imperative mood ("Create an agent...")
- **Confident but not arrogant** - "This approach works well" vs "This is the only way"

**For Contributor Documentation:**
- **Technical but approachable** - Assume programming knowledge but explain AgentFlow-specific concepts
- **Collaborative** - "We use this pattern because..."
- **Detailed and precise** - Include implementation details and reasoning

### Structure Templates

**Guide Template:**
```markdown
# Guide Title

Brief description of what this guide covers and who it's for.

## Table of Contents
- [Section 1](#section-1)
- [Section 2](#section-2)

## Prerequisites
- What users need to know/have before starting
- Links to required setup

## Main Content
### Step-by-step sections with:
- Clear headings
- Code examples
- Expected output
- Common pitfalls

## Next Steps
- Where to go from here
- Related guides
```

**API Reference Template:**
```markdown
# Package/Interface Name

Overview of the package/interface purpose.

## Types

### TypeName
Description of the type and its purpose.

```go
type TypeName struct {
    Field1 string // Description
    Field2 int    // Description
}
```

**Fields:**
- `Field1`: Detailed description, constraints, examples
- `Field2`: Detailed description, constraints, examples

## Functions

### FunctionName
Brief description.

```go
func FunctionName(param1 Type1, param2 Type2) (ReturnType, error)
```

**Parameters:**
- `param1`: Description and constraints
- `param2`: Description and constraints

**Returns:**
- `ReturnType`: Description of return value
- `error`: When and why errors occur

**Example:**
```go
// Working example with context
```
```

### Headings and Structure

- **Use consistent heading levels**
  - H1 (`#`) for document title
  - H2 (`##`) for major sections
  - H3 (`###`) for subsections
  - H4+ for detailed breakdowns if needed

- **Make headings descriptive**
  - Good: "Creating Your First Agent"
  - Bad: "Getting Started"

- **Use parallel structure** in lists and headings
  - "Creating agents", "Configuring tools", "Running workflows"
  - Not: "Create agents", "Tool configuration", "How to run workflows"

### Language and Style

**Do:**
- Use active voice: "The agent processes the request"
- Use present tense: "The function returns a result"
- Use specific verbs: "configure", "initialize", "execute"
- Define acronyms on first use: "Model Context Protocol (MCP)"
- Include units and constraints: "timeout (30 seconds max)", "1-100 agents"

**Don't:**
- Use passive voice unnecessarily: "The request is processed by the agent"
- Use future tense unless discussing roadmap: "The function will return"
- Use vague language: "some", "various", "might"
- Assume knowledge of other systems without context

## Code Examples

### Principles

1. **All examples must work** - Test every code sample
2. **Show complete context** - Include imports, setup, error handling
3. **Focus on the concept** - Don't include unrelated complexity
4. **Include expected output** when relevant

### Code Block Standards

**Complete, runnable examples:**
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Initialize MCP for tool discovery
    core.QuickStartMCP()
    
    // Create agent with Azure OpenAI
    config := core.LLMConfig{
        Provider: "azure-openai",
        APIKey:   "your-api-key",
        BaseURL:  "https://your-resource.openai.azure.com",
    }
    
    llm := core.NewAzureOpenAIAdapter(config)
    agent, err := core.NewMCPAgent("helper", llm)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create state and run agent
    state := core.NewState()
    state.Set("query", "What is the capital of France?")
    
    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Response:", result.GetResult())
}

// Expected output:
// Response: The capital of France is Paris.
```

**Code snippets (when full context isn't needed):**
```go
// Create an agent
agent, err := core.NewMCPAgent("helper", llmProvider)
if err != nil {
    return err
}
```

### Error Handling in Examples

Always show proper error handling:

**Good:**
```go
result, err := agent.Run(ctx, state)
if err != nil {
    log.Printf("Agent failed: %v", err)
    return err
}
```

**Bad:**
```go
result, _ := agent.Run(ctx, state) // Don't ignore errors in docs
```

### Configuration Examples

Use realistic but safe configuration values:

```go
// Good - Shows structure with placeholder values
config := core.LLMConfig{
    Provider: "azure-openai",
    APIKey:   "your-api-key",                    // Clear placeholder
    BaseURL:  "https://your-resource.openai.azure.com",
    Model:    "gpt-4",
    Timeout:  30 * time.Second,
}

// Bad - Fake values that might be confusing
config := core.LLMConfig{
    APIKey:  "sk-abc123",                        // Looks real but fake
    BaseURL: "https://api.openai.com",           // Wrong for Azure
}
```

## File Naming and Organization

### File Naming Conventions

- **Use PascalCase** for multi-word concepts: `ErrorHandling.md`, `CustomTools.md`
- **Be descriptive** but concise: `AgentBasics.md` not `Agents.md`
- **Avoid abbreviations** unless they're widely known: `Configuration.md` not `Config.md`
- **Group related files** in subdirectories

### Cross-File Dependencies

**Minimize dependencies** between guides:
- Each guide should be mostly self-contained
- Link to prerequisites clearly
- Don't duplicate content - reference other guides instead

**Good dependency pattern:**
```
AgentBasics.md (foundation)
    ↓
Examples.md (builds on basics)
    ↓
Production.md (builds on examples)
```

**Bad dependency pattern:**
```
AgentBasics.md ←→ ToolIntegration.md ←→ Configuration.md
(circular dependencies make docs hard to follow)
```

## Maintenance and Updates

### Keeping Documentation Current

1. **Update docs with code changes**
   - Documentation PRs should accompany feature PRs
   - Breaking changes require documentation updates
   - Deprecations need clear migration paths

2. **Regular audits**
   - Quarterly review of all user guides
   - Annual review of contributor documentation
   - Continuous monitoring of code examples

3. **Version compatibility**
   - Specify which version examples target
   - Archive old documentation when appropriate
   - Maintain migration guides for major changes

### Documentation Review Process

**For new documentation:**
1. Technical accuracy review (by code owner)
2. Clarity review (by someone unfamiliar with the feature)
3. Structure review (by documentation maintainer)

**For updates:**
1. Verify all code examples still work
2. Check links are still valid
3. Ensure consistent voice and style

### Measuring Documentation Quality

**Quantitative metrics:**
- User completion rates on tutorials
- Time spent on documentation pages
- Support questions that are answered in docs

**Qualitative feedback:**
- User feedback on clarity and usefulness
- Contributor feedback on development documentation
- Regular surveys and interviews

### Common Maintenance Tasks

**Monthly:**
- Test all code examples in user guides
- Check for broken links
- Update any references to specific versions

**Quarterly:**
- Review and update API documentation
- Audit cross-references and navigation
- Update screenshots and visual elements

**Annually:**
- Major reorganization if needed
- Archive outdated documentation
- Review and update writing standards

## Style Reference

### Formatting Standards

**Code references in text:**
- Use backticks for: `function names`, `package names`, `file names`
- Use code blocks for: multi-line code, configuration files

**Emphasis:**
- **Bold** for UI elements, important concepts, section headers in lists
- *Italic* for emphasis, new terms on first use
- `Code formatting` for technical terms, values, commands

**Lists:**
- Use bullet points for unordered concepts
- Use numbers for sequential steps
- Use consistent parallel structure

**Links:**
- Use descriptive link text: "[Agent Basics guide](AgentBasics.md)"
- Not: "Click [here](AgentBasics.md) for agent basics"

### Common Terminology

**Consistent terms:**
- "AgentFlow" (not "agentflow" or "Agent Flow") - the framework
- "agent" (lowercase) - an instance of an agent
- "MCP" - Model Context Protocol (define on first use)
- "LLM" - Large Language Model (define on first use)

**Preferred usage:**
- "configuration" not "config" (in formal documentation)
- "initialize" not "init" (in explanatory text)
- "function" not "func" (except in code)

## Templates and Tools

### Documentation Templates

See the structure templates in [Structure Templates](#structure-templates) above.

### Useful Tools

**For writing:**
- Use VS Code with Markdown preview
- Check spelling and grammar with tools like Grammarly
- Validate links with link checkers

**For code examples:**
- Test all Go code with `go run` or `go test`
- Use `gofmt` to ensure consistent formatting
- Validate JSON and YAML configuration examples

**For maintenance:**
- Use GitHub Issues to track documentation debt
- Create checklists for common update tasks
- Use automation for link checking and code validation

## Contributing to Documentation

### Getting Started

1. Read this standards guide
2. Look at existing documentation for examples
3. Start with small improvements to get familiar with the style
4. Ask questions in GitHub Discussions if unsure

### Making Changes

1. **Small fixes** (typos, broken links): Submit PR directly
2. **Content updates**: Open issue first to discuss approach
3. **Major reorganization**: Discuss in GitHub Discussions

### Review Process

All documentation changes go through the same review process as code:
- Technical accuracy
- Adherence to these standards  
- User experience and clarity

---

*This standards guide is a living document. Updates and improvements are welcome through the standard PR process.*
