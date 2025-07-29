# Design Document

## Overview

This design outlines the improvements to the `agentcli create` command to generate projects with a well-organized directory structure, clear developer placeholders, and better separation of concerns. The current implementation places all generated files in the project root, creating a cluttered structure that doesn't follow Go best practices.

The improved structure will organize code into logical subdirectories, provide clear customization points for developers, and maintain backward compatibility while following standard Go project layout conventions.

## Architecture

### Current Structure Analysis

The current scaffold system in `internal/scaffold/` uses:
- `scaffold.go`: Main project creation logic
- `templates/`: Go templates for agent and main files
- `config.go`: Configuration structures
- Template-based generation with all files in project root

### Proposed Directory Structure

The new project structure will follow Go project layout standards:

```
project-name/
├── README.md                    # Comprehensive project documentation
├── go.mod                       # Go module definition
├── agentflow.toml              # Agent configuration (root for easy access)
├── main.go                     # Main entry point with clear structure
├── agents/                     # Agent implementations directory
│   ├── agent1.go              # Individual agent files
│   ├── agent2.go              # Additional agents
│   └── README.md              # Agent-specific documentation
├── internal/                   # Internal packages (if needed)
│   ├── config/                # Configuration handling
│   └── handlers/              # Shared handler utilities
├── cmd/                       # Command-line interfaces (if multiple)
└── docs/                      # Additional documentation
    └── CUSTOMIZATION.md       # Developer customization guide
```

### Key Design Principles

1. **Separation of Concerns**: Agents in dedicated directory, configuration at root
2. **Go Conventions**: Follow standard Go project layout
3. **Clear Customization Points**: Explicit TODO comments and placeholders
4. **Immediate Runnability**: Generated project works out of the box
5. **Progressive Enhancement**: Easy to extend and customize

## Components and Interfaces

### 1. Enhanced Scaffold Generator

**Location**: `internal/scaffold/scaffold.go`

**Modifications**:
- Update `CreateAgentProjectModular()` to create directory structure
- Add directory creation functions
- Modify file paths to use new structure

**New Functions**:
```go
func createProjectDirectories(config ProjectConfig) error
func createAgentsDirectory(config ProjectConfig) error
func createInternalDirectory(config ProjectConfig) error
func createDocsDirectory(config ProjectConfig) error
```

### 2. Updated Templates

**Agent Template** (`internal/scaffold/templates/agent.go`):
- Add clear TODO comments for customization points
- Include package documentation
- Add example implementations with placeholders
- Maintain current functionality while improving clarity

**Main Template** (`internal/scaffold/templates/main.go`):
- Update import paths for new structure
- Add comprehensive comments explaining flow
- Include customization guidance
- Maintain all current features

**New Templates**:
- `agents_readme.go`: Documentation for agents directory
- `customization_guide.go`: Developer customization instructions
- `project_readme.go`: Enhanced project README

### 3. Configuration Updates

**ProjectConfig Structure**:
- No changes to existing fields (backward compatibility)
- Add optional fields for structure customization
- Maintain all current functionality

### 4. File Path Resolution

**Template Data Enhancement**:
- Update import paths in templates
- Adjust file creation paths in scaffold functions
- Ensure all references use new structure

## Data Models

### Enhanced TemplateData Structure

```go
type TemplateData struct {
    Config         ProjectConfig
    Agent          AgentInfo
    Agents         []AgentInfo
    AgentIndex     int
    TotalAgents    int
    NextAgent      string
    PrevAgent      string
    IsFirstAgent   bool
    IsLastAgent    bool
    SystemPrompt   string
    RoutingComment string
    
    // New fields for improved structure
    ProjectStructure ProjectStructureInfo
    CustomizationPoints []CustomizationPoint
    ImportPaths     ImportPathInfo
}

type ProjectStructureInfo struct {
    AgentsPackage   string
    InternalPackage string
    MainPackage     string
}

type CustomizationPoint struct {
    Location    string
    Description string
    Example     string
    Required    bool
}

type ImportPathInfo struct {
    AgentsImport   string
    InternalImport string
    ModuleName     string
}
```

### Directory Structure Configuration

```go
type DirectoryStructure struct {
    AgentsDir    string  // Default: "agents"
    InternalDir  string  // Default: "internal" 
    DocsDir      string  // Default: "docs"
    CmdDir       string  // Default: "cmd" (optional)
}
```

## Error Handling

### Directory Creation Errors
- Graceful handling of permission issues
- Clear error messages for common problems
- Rollback mechanism for partial failures

### Template Processing Errors
- Validation of template data before processing
- Clear error messages for template failures
- Fallback to basic structure if advanced features fail

### Import Path Resolution
- Validation of Go module names
- Handling of special characters in project names
- Automatic correction of invalid package names

## Testing Strategy

### Unit Tests
1. **Directory Creation Tests**
   - Verify correct directory structure creation
   - Test permission handling
   - Validate cleanup on failures

2. **Template Processing Tests**
   - Test all template variations
   - Verify correct import path generation
   - Validate customization point insertion

3. **Integration Tests**
   - End-to-end project generation
   - Compilation verification
   - Runtime execution tests

### Test Structure
```go
func TestCreateProjectDirectories(t *testing.T)
func TestAgentTemplateGeneration(t *testing.T)
func TestMainTemplateGeneration(t *testing.T)
func TestImportPathResolution(t *testing.T)
func TestCustomizationPointInsertion(t *testing.T)
func TestGeneratedProjectCompilation(t *testing.T)
func TestGeneratedProjectExecution(t *testing.T)
```

### Validation Tests
- Generated projects must compile without errors
- Generated projects must run successfully
- All import paths must be valid
- All customization points must be clearly marked

## Implementation Phases

### Phase 1: Core Structure
- Create directory structure functions
- Update basic templates with new paths
- Ensure backward compatibility

### Phase 2: Enhanced Templates
- Add comprehensive customization points
- Improve documentation and comments
- Add example implementations

### Phase 3: Documentation and Guides
- Create comprehensive README templates
- Add customization guides
- Include best practices documentation

### Phase 4: Testing and Validation
- Comprehensive test suite
- Integration testing
- Performance validation

## Backward Compatibility

### Existing Projects
- No changes to existing generated projects
- New structure only applies to newly created projects
- All existing configuration options remain functional

### Configuration Compatibility
- All existing `ProjectConfig` fields maintained
- New fields are optional with sensible defaults
- Existing templates continue to work during transition

## Migration Strategy

### Gradual Rollout
1. Implement new structure alongside existing
2. Add feature flag for new structure (optional)
3. Make new structure default after testing
4. Maintain old structure support for transition period

### Developer Communication
- Clear documentation of changes
- Migration guide for existing users
- Examples showing before/after structure