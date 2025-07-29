# Requirements Document

## Introduction

This specification addresses the need to update all tutorial and guide documentation to use the new consolidated CLI syntax consistently. While the main reference documentation has been updated to reflect the new simplified `agentcli create` command with templates and consolidated flags, many tutorial files and guides still use the old complex multi-flag syntax. This creates inconsistency and confusion for users following the documentation.

## Requirements

### Requirement 1: Update Tutorial Documentation to Use New CLI Syntax

**User Story:** As a user following AgenticGoKit tutorials, I want all examples to use the current CLI syntax so that I can successfully follow along without encountering deprecated or inconsistent commands.

#### Acceptance Criteria

1. WHEN a user follows any tutorial THEN all `agentcli create` commands SHALL use the new consolidated flag syntax or template-based approach
2. WHEN tutorials show project creation THEN they SHALL prioritize template usage over complex flag combinations
3. WHEN tutorials need custom configuration THEN they SHALL use the consolidated flags (--memory, --embedding, --mcp, --rag) instead of multiple separate flags
4. WHEN tutorials reference CLI commands THEN they SHALL be consistent with the current CLI reference documentation

### Requirement 2: Update Guide Documentation for CLI Consistency

**User Story:** As a developer reading AgenticGoKit guides, I want all CLI examples to reflect the current best practices so that I can implement solutions using the most efficient and up-to-date approaches.

#### Acceptance Criteria

1. WHEN guides show project creation examples THEN they SHALL use template-based creation where appropriate
2. WHEN guides demonstrate specific features THEN they SHALL use consolidated flags instead of deprecated multi-flag syntax
3. WHEN guides reference memory, RAG, or MCP setup THEN they SHALL use the simplified flag structure
4. WHEN guides show advanced configurations THEN they SHALL demonstrate template overrides and consolidated flag usage

### Requirement 3: Maintain Tutorial Learning Progression

**User Story:** As a new user learning AgenticGoKit, I want tutorials to progress logically from simple template usage to more advanced custom configurations so that I can build understanding incrementally.

#### Acceptance Criteria

1. WHEN tutorials introduce CLI usage THEN they SHALL start with simple template-based examples
2. WHEN tutorials progress to advanced topics THEN they SHALL show how to customize templates with additional flags
3. WHEN tutorials demonstrate complex setups THEN they SHALL explain the relationship between templates and custom flags
4. WHEN tutorials show multiple approaches THEN they SHALL clearly indicate which approach is recommended

### Requirement 4: Update Memory and RAG Documentation

**User Story:** As a developer implementing memory and RAG features, I want documentation examples to use the simplified memory and RAG flags so that I can quickly set up these features without complex flag combinations.

#### Acceptance Criteria

1. WHEN documentation shows memory setup THEN it SHALL use `--memory [provider]` instead of `--memory-enabled --memory-provider [provider]`
2. WHEN documentation shows RAG configuration THEN it SHALL use `--rag [chunk-size]` instead of multiple RAG-specific flags
3. WHEN documentation shows embedding setup THEN it SHALL use `--embedding [provider:model]` format
4. WHEN documentation shows complete RAG systems THEN it SHALL recommend the `rag-system` template with overrides

### Requirement 5: Update MCP Integration Documentation

**User Story:** As a developer integrating MCP tools, I want documentation to show the simplified MCP flag usage so that I can quickly enable MCP features without understanding complex flag dependencies.

#### Acceptance Criteria

1. WHEN documentation shows MCP setup THEN it SHALL use `--mcp [level]` instead of multiple MCP flags
2. WHEN documentation shows basic MCP usage THEN it SHALL use `--mcp basic` instead of `--mcp-enabled --mcp-tools [tools]`
3. WHEN documentation shows production MCP THEN it SHALL use `--mcp production` instead of multiple production flags
4. WHEN documentation shows research assistants THEN it SHALL recommend the `research-assistant` template

### Requirement 6: Update Orchestration Examples

**User Story:** As a developer building multi-agent systems, I want documentation to show the simplified orchestration syntax so that I can quickly set up different agent coordination patterns.

#### Acceptance Criteria

1. WHEN documentation shows collaborative agents THEN it SHALL use `--orchestration collaborative` instead of `--orchestration-mode collaborative --collaborative-agents [list]`
2. WHEN documentation shows sequential pipelines THEN it SHALL use `--orchestration sequential` or the `data-pipeline` template
3. WHEN documentation shows complex orchestration THEN it SHALL demonstrate template usage with orchestration overrides
4. WHEN documentation shows agent naming THEN it SHALL show how templates provide sensible agent names automatically

### Requirement 7: Maintain Backward Compatibility References

**User Story:** As a user with existing projects created with old CLI syntax, I want documentation to acknowledge the old syntax while clearly directing me to the new approach.

#### Acceptance Criteria

1. WHEN documentation updates CLI examples THEN it SHALL not completely remove context about the old approach where relevant
2. WHEN documentation shows new syntax THEN it SHALL briefly explain the benefits over the old multi-flag approach
3. WHEN documentation introduces templates THEN it SHALL explain how they simplify common flag combinations
4. WHEN documentation shows advanced usage THEN it SHALL explain when to use templates vs. custom flags

### Requirement 8: Ensure Documentation Consistency Across All Files

**User Story:** As a user navigating different parts of the documentation, I want consistent CLI syntax and examples so that I don't encounter conflicting information.

#### Acceptance Criteria

1. WHEN any documentation file shows CLI usage THEN it SHALL be consistent with the CLI reference documentation
2. WHEN multiple files show similar examples THEN they SHALL use the same CLI syntax patterns
3. WHEN documentation shows project creation THEN it SHALL consistently prioritize template usage
4. WHEN documentation is updated THEN all related files SHALL be updated to maintain consistency