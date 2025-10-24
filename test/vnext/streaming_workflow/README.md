# Streaming Workflow Integration Tests

This directory contains comprehensive integration tests for vnext workflow streaming functionality.

## Tests Included

### `main.go` - Integration Test Suite
- **Parallel Workflow Streaming**: Tests concurrent agent execution with real-time streaming
- **Sequential vs Parallel Comparison**: Performance comparison between execution modes
- **Extended Timeout Testing**: Uses 800-second timeouts for thorough validation
- **Error Handling**: Tests error conditions and recovery scenarios

## Running Tests

```bash
# From this directory
go run main.go

# Or from project root
go run test/vnext/streaming_workflow/main.go
```

## Prerequisites
- Ollama running on `localhost:11434`
- `gemma3:1b` model available (`ollama pull gemma3:1b`)

## Test Scenarios

### 1. Parallel Workflow Streaming
- Creates two agents with different system prompts
- Runs them in parallel with streaming output
- Validates concurrent execution and streaming integrity
- Measures total execution time vs sequential execution

### 2. Sequential Workflow Comparison
- Tests sequential execution for comparison
- Measures performance differences
- Validates step-by-step execution order
- Compares timing with parallel execution

## Expected Behavior
- ✅ All workflows should complete without "context canceled" errors
- ✅ Real-time streaming should display tokens as they're generated
- ✅ Parallel execution should be faster than sequential for independent tasks
- ✅ Error handling should provide clear diagnostics

## Integration with Benchmarks
These tests complement the performance benchmarks in `../benchmarks/` by providing:
- Functional validation of different workflow modes
- Extended testing scenarios
- Error condition testing
- Real-world usage patterns

## Troubleshooting
- If tests fail with timeouts, ensure Ollama is running and responsive
- Check that the gemma3:1b model is properly loaded
- Verify network connectivity to Ollama service
- Review timeout settings if tests consistently fail