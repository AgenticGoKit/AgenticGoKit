# Design Document

## Overview

This design outlines a comprehensive approach to update all AgenticGoKit documentation to use the new consolidated CLI syntax consistently. The update will transform outdated multi-flag examples into modern template-based and consolidated flag approaches while maintaining the educational value and learning progression of the documentation.

## Architecture

### Current Documentation State Analysis

**Files Requiring Updates:**

1. **Tutorial Files (High Priority)**
   - `docs/tutorials/getting-started/quickstart.md`
   - `docs/tutorials/getting-started/your-first-agent.md`
   - `docs/tutorials/getting-started/multi-agent-collaboration.md`
   - `docs/tutorials/getting-started/memory-and-rag.md`
   - `docs/tutorials/getting-started/tool-integration.md`
   - `docs/tutorials/getting-started/production-deployment.md`

2. **Guide Files (Medium Priority)**
   - `docs/guides/setup/vector-databases.md`
   - `docs/guides/ScaffoldMemoryGuide.md`
   - `docs/guides/RAGConfiguration.md`
   - `docs/guides/EmbeddingModelGuide.md`
   - `docs/guides/MemoryTroubleshooting.md`

3. **Reference Files (Already Updated)**
   - `docs/reference/cli.md` ✅
   - `docs/reference/cli-quick-reference.md` ✅
   - `docs/guides/project-templates.md` ✅

### Update Strategy

#### 1. Template-First Approach

**Priority Order:**
1. **Templates for common patterns** - Use built-in templates where applicable
2. **Consolidated flags for customization** - Use new simplified flags
3. **Template overrides for advanced cases** - Show how to customize templates

**Example Transformation:**
```bash
# Old (complex multi-flag)
agentcli create research-bot --agents 3 --orchestration-mode collaborative \
  --collaborative-agents "researcher,analyzer,synthesizer" --mcp-enabled \
  --mcp-tools "web_search,summarize" --with-cache

# New (template-based)
agentcli create research-bot --template research-assistant

# New (template with customization)
agentcli create research-bot --template research-assistant --agents 4 --mcp production
```

#### 2. Learning Progression Design

**Beginner Level (Templates):**
```bash
# Start with simple templates
agentcli create my-project --template basic
agentcli create research-bot --template research-assistant
agentcli create knowledge-base --template rag-system
```

**Intermediate Level (Template + Flags):**
```bash
# Show customization with consolidated flags
agentcli create my-kb --template rag-system --memory weaviate
agentcli create my-research --template research-assistant --agents 5
```

**Advanced Level (Full Customization):**
```bash
# Show advanced flag combinations
agentcli create custom-system --memory pgvector --embedding openai --rag 2000 --mcp production
```

## Components and Interfaces

### 1. Documentation Update Patterns

#### Pattern 1: Simple Project Creation
```bash
# Old Pattern
agentcli create my-project --agents 2 --provider openai

# New Pattern
agentcli create my-project --template basic
```

#### Pattern 2: Memory-Enabled Projects
```bash
# Old Pattern
agentcli create my-project --memory-enabled --memory-provider pgvector \
  --embedding-provider openai --embedding-model text-embedding-ada-002

# New Pattern
agentcli create my-project --memory pgvector --embedding openai
```

#### Pattern 3: RAG Systems
```bash
# Old Pattern
agentcli create rag-system --memory-enabled --memory-provider pgvector \
  --rag-enabled --rag-chunk-size 1000 --rag-overlap 100 --rag-top-k 5 \
  --embedding-provider openai

# New Pattern
agentcli create rag-system --template rag-system
# or with customization
agentcli create rag-system --memory pgvector --embedding openai --rag 1000
```

#### Pattern 4: MCP Integration
```bash
# Old Pattern
agentcli create tool-agent --mcp-enabled --mcp-tools "web_search,summarize" \
  --with-cache --with-metrics

# New Pattern
agentcli create tool-agent --template research-assistant
# or
agentcli create tool-agent --mcp production
```

#### Pattern 5: Multi-Agent Orchestration
```bash
# Old Pattern
agentcli create collab-system --agents 3 --orchestration-mode collaborative \
  --collaborative-agents "agent1,agent2,agent3"

# New Pattern
agentcli create collab-system --template research-assistant
# or
agentcli create collab-system --orchestration collaborative --agents 3
```

### 2. File-Specific Update Strategies

#### Tutorial Files Strategy

**quickstart.md:**
- Replace complex first example with simple template usage
- Show progression from template to customization
- Maintain learning flow while using new syntax

**your-first-agent.md:**
- Start with `--template basic`
- Show single-flag customizations
- Explain template benefits

**multi-agent-collaboration.md:**
- Use `--template research-assistant` for collaborative examples
- Use `--template data-pipeline` for sequential examples
- Show orchestration flag for custom setups

**memory-and-rag.md:**
- Use `--template rag-system` for complete examples
- Show `--memory`, `--embedding`, `--rag` flags for customization
- Demonstrate template overrides

**tool-integration.md:**
- Use `--template research-assistant` for MCP examples
- Show `--mcp basic`, `--mcp production` levels
- Demonstrate custom tool configurations

#### Guide Files Strategy

**vector-databases.md:**
- Update all examples to use `--memory [provider]` syntax
- Show template usage for complete setups
- Maintain technical depth while simplifying commands

**ScaffoldMemoryGuide.md:**
- Update embedding examples to use `--embedding [provider:model]`
- Show intelligent defaults and template usage
- Maintain technical accuracy

### 3. Content Transformation Rules

#### Rule 1: Template Preference
- If a common pattern exists as a template, use the template first
- Show customization as template overrides
- Explain when templates are sufficient vs. when custom flags are needed

#### Rule 2: Flag Consolidation
- Replace memory flag combinations with single `--memory` flag
- Replace RAG flag combinations with single `--rag` flag
- Replace MCP flag combinations with single `--mcp` flag
- Replace orchestration flag combinations with single `--orchestration` flag

#### Rule 3: Learning Progression
- Start simple (templates)
- Add complexity gradually (template + flags)
- Show advanced usage (full customization)
- Explain the reasoning behind each approach

#### Rule 4: Context Preservation
- Maintain the educational context of each example
- Explain why certain approaches are recommended
- Show the relationship between old and new syntax where helpful

## Data Models

### Update Mapping Structure

```go
type DocumentationUpdate struct {
    FilePath        string
    Priority        Priority
    UpdatePatterns  []PatternUpdate
    LearningLevel   LearningLevel
    Dependencies    []string
}

type PatternUpdate struct {
    OldPattern      string
    NewPattern      string
    Context         string
    Explanation     string
    AlternativeApproaches []string
}

type LearningLevel int

const (
    Beginner LearningLevel = iota
    Intermediate
    Advanced
)

type Priority int

const (
    HighPriority Priority = iota
    MediumPriority
    LowPriority
)
```

### File Update Categories

```go
var DocumentationFiles = map[string]DocumentationUpdate{
    "docs/tutorials/getting-started/quickstart.md": {
        Priority: HighPriority,
        LearningLevel: Beginner,
        UpdatePatterns: []PatternUpdate{
            {
                OldPattern: "agentcli create my-agents --agents 3 --orchestration-mode collaborative",
                NewPattern: "agentcli create my-agents --template research-assistant",
                Context: "First example in quickstart",
                Explanation: "Templates provide the simplest way to get started",
            },
        },
    },
    // ... other files
}
```

## Error Handling

### Update Validation Strategy

1. **Syntax Validation**
   - Ensure all new CLI examples use valid syntax
   - Verify template names exist
   - Check flag combinations are valid

2. **Learning Flow Validation**
   - Ensure examples progress logically from simple to complex
   - Verify that advanced examples build on earlier concepts
   - Check that template usage is explained before customization

3. **Consistency Validation**
   - Cross-reference examples across files for consistency
   - Ensure similar use cases use similar CLI patterns
   - Verify that all files use the same template names and flag syntax

### Rollback Strategy

1. **Incremental Updates**
   - Update files in priority order
   - Test each file update independently
   - Maintain git history for easy rollback

2. **Validation Checkpoints**
   - Validate syntax after each file update
   - Check cross-references between updated files
   - Verify learning progression is maintained

## Testing Strategy

### 1. CLI Syntax Testing
- Test all new CLI examples to ensure they work
- Verify template names and flag combinations
- Check that examples produce expected results

### 2. Documentation Flow Testing
- Follow tutorial sequences using new syntax
- Verify that examples build on each other logically
- Test that learning progression is maintained

### 3. Cross-Reference Testing
- Check that examples are consistent across files
- Verify that similar use cases use similar approaches
- Ensure template usage is consistent

### 4. User Experience Testing
- Follow tutorials as a new user would
- Verify that examples are clear and easy to follow
- Check that the progression from simple to advanced is smooth

## Implementation Phases

### Phase 1: High-Priority Tutorial Files (Week 1)
- Update `quickstart.md` with template-first approach
- Update `your-first-agent.md` with basic template usage
- Update `multi-agent-collaboration.md` with orchestration templates
- Validate learning progression and syntax

### Phase 2: Memory and RAG Documentation (Week 2)
- Update `memory-and-rag.md` with consolidated memory/RAG flags
- Update `vector-databases.md` with simplified setup examples
- Update `ScaffoldMemoryGuide.md` with new embedding syntax
- Test all memory-related examples

### Phase 3: MCP and Tool Integration (Week 3)
- Update `tool-integration.md` with MCP level flags
- Update MCP-related guides with simplified syntax
- Update production deployment examples
- Validate MCP integration examples

### Phase 4: Advanced Guides and References (Week 4)
- Update remaining guide files
- Update troubleshooting documentation
- Cross-validate all documentation for consistency
- Final testing and validation

### Phase 5: Quality Assurance and Polish (Week 5)
- Comprehensive review of all updated documentation
- User experience testing with updated tutorials
- Final consistency checks across all files
- Documentation of changes and migration notes

## Migration Strategy

### Content Migration Approach

1. **Preserve Educational Value**
   - Maintain the learning objectives of each tutorial
   - Keep the same conceptual progression
   - Preserve important explanations and context

2. **Enhance with New Capabilities**
   - Show how templates simplify common tasks
   - Demonstrate the power of consolidated flags
   - Explain the benefits of the new approach

3. **Provide Migration Context**
   - Briefly explain the evolution from old to new syntax
   - Show how complex flag combinations are now simplified
   - Highlight the improved user experience

### Backward Compatibility Notes

1. **Acknowledge Previous Approach**
   - Mention that the CLI has been simplified
   - Explain that old syntax still works but is not recommended
   - Direct users to new approaches

2. **Migration Guidance**
   - Show equivalent new syntax for old examples
   - Explain the benefits of migrating to new syntax
   - Provide clear migration paths

This design ensures that all documentation will be updated systematically while maintaining educational value and providing a smooth learning experience for users at all levels.