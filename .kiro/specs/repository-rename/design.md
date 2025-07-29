# Design Document

## Overview

This design outlines the systematic approach to rename the repository from `agentflow` to `AgenticGoKit` throughout the entire codebase. The rename affects multiple layers including module declarations, import statements, template files, documentation, configuration references, and example code. The design ensures a comprehensive update that maintains functionality while establishing the new brand identity.

## Architecture

The repository rename affects several architectural layers:

1. **Module Layer**: Go module declaration and dependency management
2. **Import Layer**: All internal and external import statements
3. **Template Layer**: Code generation templates used by scaffolding
4. **Documentation Layer**: README files, examples, and inline documentation
5. **Configuration Layer**: Configuration file references and naming conventions
6. **Test Layer**: Test files and test data that reference the old names

## Components and Interfaces

### 1. Core Module System

**Current State:**
- Module name: `github.com/kunalkushwaha/agentflow`
- All internal imports use the old repository name
- External projects depend on the old module name

**Target State:**
- Module name: `github.com/kunalkushwaha/AgenticGoKit`
- All internal imports updated to new repository name
- Backward compatibility considerations for external projects

### 2. Scaffolding System

**Current State:**
- Template files contain hardcoded old repository imports
- Generated go.mod files reference old repository
- Generated documentation links point to old repository
- Version constant references old name

**Target State:**
- All template files use new repository name
- Generated projects work immediately with new module name
- Documentation links point to new repository
- Version constant updated appropriately

### 3. Configuration System

**Current State:**
- Configuration files named `agentflow.toml`
- References to "agentflow" in configuration contexts
- Database names and connection strings use old naming

**Target State:**
- Evaluate whether to rename configuration files to maintain consistency
- Update references in code and documentation
- Consider migration path for existing configurations

## Data Models

### File Categories for Update

```go
type FileUpdateCategory struct {
    Category     string
    FilePatterns []string
    UpdateType   UpdateType
}

type UpdateType int

const (
    ModuleDeclaration UpdateType = iota
    ImportStatement
    TemplateContent
    DocumentationLink
    ConfigurationReference
    VariableName
    CommentText
)
```

### Update Mapping

```go
type UpdateMapping struct {
    OldPattern    string
    NewPattern    string
    FileTypes     []string
    UpdateType    UpdateType
    RequiresTest  bool
}
```

Key mappings:
- `github.com/kunalkushwaha/agentflow` → `github.com/kunalkushwaha/AgenticGoKit`
- `agentflow "github.com/kunalkushwaha/agentflow/core"` → `agenticgokit "github.com/kunalkushwaha/Agenticgokit/core"`
- Repository URLs in documentation
- Configuration file references

## Error Handling

### Validation Strategy

1. **Pre-update Validation**
   - Verify all files compile before changes
   - Create backup of critical files
   - Validate that new repository name follows Go module conventions

2. **Update Validation**
   - After each category of updates, run `go mod tidy`
   - Execute test suite to ensure functionality is preserved
   - Validate that generated templates produce working code

3. **Post-update Validation**
   - Full build and test execution
   - Validate scaffolding generates working projects
   - Check that all documentation links are accessible

### Rollback Strategy

- Maintain list of all modified files
- Use version control to track changes
- Implement atomic updates where possible
- Provide rollback procedure if issues are discovered

## Testing Strategy

### Unit Testing
- Verify all existing tests pass after updates
- Add specific tests for scaffolding template generation
- Test configuration file loading with updated references

### Integration Testing
- Test full scaffolding workflow with new repository name
- Verify generated projects build and run correctly
- Test that examples work with updated imports

### Manual Testing
- Generate sample projects using scaffolding
- Verify all documentation links work
- Test configuration loading and error messages

### Regression Testing
- Ensure all existing functionality remains intact
- Verify performance is not impacted
- Test memory and MCP integrations still work

## Implementation Phases

### Phase 1: Core Module Updates
- Update go.mod module declaration
- Update all internal import statements
- Update internal package references

### Phase 2: Template System Updates
- Update all template files in `internal/scaffold/templates/`
- Update scaffold generators
- Update version constants and references

### Phase 3: Documentation Updates
- Update README files
- Update example code
- Update inline documentation and comments

### Phase 4: Configuration Updates
- Evaluate configuration file naming
- Update configuration references in code
- Update error messages and help text

### Phase 5: Testing and Validation
- Run comprehensive test suite
- Test scaffolding functionality
- Validate generated projects work correctly

## Configuration Considerations

### Configuration File Naming
The current system uses `agentflow.toml` as the configuration file name. Design considerations:

**Option 1: Keep existing name**
- Pros: No breaking changes for existing users
- Cons: Inconsistent with new branding

**Option 2: Rename to `agenticgokit.toml`**
- Pros: Consistent branding
- Cons: Breaking change for existing users

**Recommendation**: Keep `agentflow.toml` for backward compatibility, but update all references and documentation to reflect the new repository name. Consider adding support for both names with deprecation warnings.

### Database and Connection Naming
Current database names and connection strings reference "agentflow". These should be updated in:
- Example connection strings
- Documentation
- Default database names in examples
- Test configurations

## Migration Path

### For End Users
1. Update import statements in their projects
2. Update go.mod require statements
3. Update any hardcoded repository references
4. Configuration files can remain unchanged

### For Contributors
1. Update local development environment
2. Update any scripts or automation
3. Update documentation contributions
4. Update any external references

## Risk Assessment

### High Risk
- Breaking existing user projects if not handled carefully
- Template generation producing non-working code
- Missing references causing build failures

### Medium Risk
- Documentation links becoming stale
- Configuration inconsistencies
- Test failures due to missed references

### Low Risk
- Variable name inconsistencies
- Comment text inconsistencies
- Non-critical documentation updates

## Success Criteria

1. All Go code compiles successfully with new module name
2. All tests pass without modification
3. Scaffolding generates working projects with correct imports
4. Documentation links are accessible and correct
5. Examples run successfully with new repository name
6. No references to old repository name remain in code
7. Configuration system works with updated references