# Design Document

## Overview

This design outlines comprehensive improvements to the AgenticGoKit CLI (agentcli) to address current issues and enhance user experience. The improvements focus on fixing the broken version command, reorganizing command structure, simplifying the complex create command, removing emojis for better compatibility, and adding missing functionality referenced in documentation.

## Architecture

### Current CLI Structure Analysis

The current CLI has several architectural issues:

1. **Version Command Registration**: The version command exists but isn't properly registered in the root command
2. **Command Organization**: Commands are not well-categorized in help output
3. **Create Command Complexity**: Over 30 flags with overlapping functionality
4. **Inconsistent Output**: Mix of emojis and text, inconsistent formatting
5. **Missing Commands**: Several commands mentioned in help text don't exist

### Proposed CLI Architecture

```
agentcli/
├── cmd/
│   ├── root.go           # Root command with improved help
│   ├── version.go        # Fixed version command registration
│   ├── create/           # Redesigned create command structure
│   │   ├── create.go     # Main create command
│   │   ├── templates.go  # Project templates
│   │   └── flags.go      # Consolidated flag definitions
│   ├── project/          # Project management commands
│   │   ├── info.go       # Project information
│   │   ├── health.go     # Health checks
│   │   └── validate.go   # Enhanced validation
│   ├── debug/            # Development and debugging
│   │   ├── trace.go      # Existing trace command
│   │   ├── memory.go     # Existing memory command
│   │   └── logs.go       # New logs command
│   ├── mcp/              # MCP-related commands
│   │   ├── mcp.go        # Existing MCP command
│   │   └── cache.go      # Existing cache command
│   └── utils/            # Utility commands
│       ├── list.go       # Existing list command
│       └── completion.go # Shell completion
└── internal/
    ├── output/           # Standardized output formatting
    ├── context/          # Project context detection
    └── templates/        # Project templates and presets
```

## Components and Interfaces

### 1. Fixed Version Command Registration

**Current Issue:**
```go
// version.go exists but isn't registered in root.go init()
```

**Solution:**
```go
// In root.go init()
func init() {
    rootCmd.AddCommand(versionCmd)  // Add this line
    // ... other command registrations
}
```

### 2. Redesigned Create Command Structure

**Current Problems:**
- 30+ flags with overlapping functionality
- Complex validation logic
- Overwhelming help output
- Redundant options

**Proposed Flag Consolidation:**

#### Before (Current):
```bash
# Memory/RAG flags (8 separate flags)
--memory-enabled
--memory-provider
--embedding-provider  
--embedding-model
--rag-enabled
--rag-chunk-size
--rag-overlap
--rag-top-k
--rag-score-threshold
--hybrid-search
--session-memory

# MCP flags (12 separate flags)
--mcp-enabled
--mcp-production
--with-cache
--with-metrics
--mcp-tools
--mcp-servers
--cache-backend
--metrics-port
--with-load-balancer
--connection-pool-size
--retry-policy
```

#### After (Proposed):
```bash
# Consolidated memory flag with intelligent defaults
--memory [provider]              # memory, pgvector, weaviate
--embedding [provider[:model]]   # openai, ollama:nomic-embed-text
--rag [chunk-size]              # enables RAG with optional chunk size

# Consolidated MCP flag
--mcp [level]                   # basic, production, full
--tools [tool1,tool2]           # MCP tools to include

# Preset templates
--template [name]               # research-assistant, data-pipeline, chat-system
```

### 3. Project Templates and Presets

**Template System:**
```go
type ProjectTemplate struct {
    Name        string
    Description string
    Config      ProjectConfig
    Features    []string
}

var ProjectTemplates = map[string]ProjectTemplate{
    "basic": {
        Name:        "Basic Multi-Agent System",
        Description: "Simple multi-agent project with 2 agents",
        Config:      ProjectConfig{NumAgents: 2, Provider: "openai"},
        Features:    []string{"sequential-orchestration"},
    },
    "research-assistant": {
        Name:        "Research Assistant",
        Description: "Web search, analysis, and synthesis agents",
        Config:      ProjectConfig{
            NumAgents: 3,
            OrchestrationMode: "collaborative",
            MCPEnabled: true,
            MCPTools: []string{"web_search", "summarize"},
        },
        Features: []string{"mcp-tools", "collaborative-agents"},
    },
    "rag-system": {
        Name:        "RAG Knowledge Base",
        Description: "Document ingestion and Q&A system",
        Config:      ProjectConfig{
            MemoryEnabled: true,
            MemoryProvider: "pgvector",
            RAGEnabled: true,
            EmbeddingProvider: "openai",
        },
        Features: []string{"memory", "rag", "vector-search"},
    },
}
```

### 4. Standardized Output System

**Output Interface:**
```go
type OutputFormatter interface {
    Format(data interface{}) ([]byte, error)
    ContentType() string
}

type OutputManager struct {
    format    string  // table, json, yaml
    verbose   bool
    quiet     bool
    noEmojis  bool    // Always true for CLI
}

// Replace emoji-based output
func (o *OutputManager) Success(msg string) {
    if !o.quiet {
        fmt.Printf("[SUCCESS] %s\n", msg)  // Instead of ✓
    }
}

func (o *OutputManager) Warning(msg string) {
    fmt.Printf("[WARNING] %s\n", msg)  // Instead of ⚠️
}

func (o *OutputManager) Info(msg string) {
    if o.verbose {
        fmt.Printf("[INFO] %s\n", msg)  // Instead of ℹ️
    }
}
```

### 5. Enhanced Command Categories

**Improved Help Organization:**
```go
type CommandCategory struct {
    Name        string
    Description string
    Prefix      string  // Text prefix instead of emoji
    Commands    []*cobra.Command
}

var CommandCategories = []CommandCategory{
    {
        Name:        "Project Management",
        Description: "Commands for creating and managing projects",
        Prefix:      "[PROJECT]",  // Instead of 🚀
        Commands:    []*cobra.Command{createCmd, infoCmd, validateCmd},
    },
    {
        Name:        "Development & Debug",
        Description: "Commands for debugging and development",
        Prefix:      "[DEBUG]",    // Instead of 🔧
        Commands:    []*cobra.Command{traceCmd, memoryCmd, logsCmd},
    },
    // ... other categories
}
```

### 6. Project Context Detection

**Context System:**
```go
type ProjectContext struct {
    IsProject     bool
    ProjectRoot   string
    ConfigPath    string
    Config        *ProjectConfig
    HasAgents     bool
    HasMemory     bool
    HasMCP        bool
}

func DetectProjectContext() (*ProjectContext, error) {
    // Look for agentflow.toml, go.mod, agents/ directory
    // Parse configuration and detect features
}

func (ctx *ProjectContext) ValidateStructure() []ValidationError {
    // Check project structure consistency
    // Validate configuration against actual files
}
```

## Data Models

### Consolidated Flag Structure

```go
type CreateFlags struct {
    // Basic project settings
    Name         string
    Template     string
    Interactive  bool
    
    // Agent configuration
    Agents       int
    Provider     string
    Orchestration string
    
    // Feature flags (simplified)
    Memory       string  // "", "memory", "pgvector", "weaviate"
    Embedding    string  // "provider:model" format
    RAG          int     // 0=disabled, >0=chunk size
    MCP          string  // "", "basic", "production", "full"
    Tools        []string
    
    // Output options
    Visualize    bool
    OutputDir    string
    
    // Global flags
    Verbose      bool
    Quiet        bool
    Format       string
}
```

### Template Configuration

```go
type TemplateConfig struct {
    Metadata     TemplateMetadata
    ProjectConfig ProjectConfig
    Customizations []Customization
}

type TemplateMetadata struct {
    Name         string
    Description  string
    Category     string
    Difficulty   string  // beginner, intermediate, advanced
    Features     []string
    Requirements []string
}
```

## Error Handling

### Improved Error Messages

**Current Issues:**
- Generic error messages
- No suggestions for fixes
- Inconsistent formatting

**Proposed Error System:**
```go
type CLIError struct {
    Code        string
    Message     string
    Suggestions []string
    Context     map[string]interface{}
}

func (e *CLIError) Error() string {
    var b strings.Builder
    b.WriteString(fmt.Sprintf("[ERROR] %s\n", e.Message))
    
    if len(e.Suggestions) > 0 {
        b.WriteString("\nSuggestions:\n")
        for _, suggestion := range e.Suggestions {
            b.WriteString(fmt.Sprintf("  - %s\n", suggestion))
        }
    }
    
    return b.String()
}

// Usage examples
func validateCreateFlags(flags *CreateFlags) error {
    if flags.Memory == "pgvector" && flags.Embedding == "" {
        return &CLIError{
            Code:    "MISSING_EMBEDDING",
            Message: "Memory provider 'pgvector' requires an embedding provider",
            Suggestions: []string{
                "Add --embedding openai for OpenAI embeddings",
                "Add --embedding ollama:nomic-embed-text for local embeddings",
                "Use --template rag-system for a complete setup",
            },
        }
    }
    return nil
}
```

### Flag Validation

```go
type FlagValidator struct {
    validators map[string]func(interface{}) error
    dependencies map[string][]string
}

func (v *FlagValidator) ValidateFlag(name string, value interface{}) error {
    // Individual flag validation
}

func (v *FlagValidator) ValidateDependencies(flags map[string]interface{}) error {
    // Cross-flag validation
}
```

## Testing Strategy

### Unit Tests

1. **Command Registration Tests**
   - Verify all commands are properly registered
   - Test command hierarchy and help output
   - Validate flag definitions

2. **Flag Consolidation Tests**
   - Test flag parsing and validation
   - Verify backward compatibility where possible
   - Test error messages and suggestions

3. **Template System Tests**
   - Test template loading and application
   - Verify generated project structure
   - Test customization points

4. **Output Formatting Tests**
   - Test all output formats (table, json, yaml)
   - Verify emoji removal
   - Test quiet and verbose modes

### Integration Tests

1. **End-to-End Command Tests**
   - Test complete command workflows
   - Verify project creation with various options
   - Test project context detection

2. **Template Generation Tests**
   - Generate projects from each template
   - Verify projects compile and run
   - Test feature combinations

### Compatibility Tests

1. **Backward Compatibility**
   - Test existing flag combinations still work
   - Verify migration path for deprecated flags
   - Test existing project compatibility

2. **Cross-Platform Tests**
   - Test on Windows, macOS, Linux
   - Verify shell completion works
   - Test emoji removal across terminals

## Implementation Phases

### Phase 1: Core Fixes (Week 1)
- Fix version command registration
- Remove emojis from all output
- Standardize error messages
- Add missing command stubs (info, health, logs)

### Phase 2: Create Command Redesign (Week 2)
- Consolidate flags into logical groups
- Implement template system
- Add flag validation and suggestions
- Update help text and examples

### Phase 3: Enhanced Commands (Week 3)
- Implement info, health, logs commands
- Add project context detection
- Enhance existing commands with new output formats
- Add shell completion support

### Phase 4: Testing and Polish (Week 4)
- Comprehensive testing suite
- Documentation updates
- Performance optimization
- User acceptance testing

## Migration Strategy

### Backward Compatibility

1. **Deprecated Flag Support**
   - Keep old flags working with deprecation warnings
   - Provide migration suggestions in warnings
   - Document migration path

2. **Gradual Migration**
   - Phase out old flags over multiple releases
   - Provide clear migration timeline
   - Support both old and new syntax during transition

### User Communication

1. **Migration Guide**
   - Document all flag changes
   - Provide before/after examples
   - Include automation scripts for bulk updates

2. **Release Notes**
   - Clear changelog with migration instructions
   - Highlight breaking changes
   - Provide upgrade recommendations

This design ensures a more maintainable, user-friendly CLI while preserving existing functionality and providing clear migration paths for users.