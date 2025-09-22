# Prompt Templates

This directory contains template files for system prompts used by the AgenticGoKit scaffold system. These templates have been extracted from the main code to improve readability and maintainability.

## Structure

### RAG-Specific Templates
- `rag_document_ingester.txt` - Specialized prompt for document ingestion agents
- `rag_query_processor.txt` - Specialized prompt for query processing agents  
- `rag_response_generator.txt` - Specialized prompt for response generation agents
- `rag_retrieval_agent.txt` - Specialized prompt for information retrieval agents

### Sequential Mode Templates
- `sequential_first_agent.txt` - Instructions for the first agent in sequential workflows
- `sequential_final_agent.txt` - Instructions for the final agent in sequential workflows
- `sequential_middle_agent.txt` - Instructions for middle agents in sequential workflows

### Universal Guidelines
- `tool_usage_guidelines.txt` - Common tool usage strategies for all agents
- `response_quality_standards.txt` - Quality standards applied to all responses

## Usage

These templates are embedded into the Go binary using `//go:embed` directives and loaded at compile time. They are used by the prompt generation functions in `prompts.go` to create comprehensive system prompts for different agent types and orchestration modes.

## Benefits

- **Improved Readability**: Large multi-line strings are moved out of the source code
- **Better Maintainability**: Prompts can be edited without touching Go code
- **Modularity**: Each prompt type is in its own file for easy management
- **Version Control**: Changes to prompts are clearly visible in diffs
- **Reusability**: Common guidelines are shared across multiple prompt types