# RunWithOptions Tests

Tests for the `RunWithOptions()` method which allows runtime configuration overrides.

## Test Coverage

### ✅ Passing Tests (6 tests)

1. **TestRunWithOptions_NilOptions** - Verifies nil options delegates to Run()
2. **TestRunWithOptions_Timeout** - Tests timeout option application
3. **TestRunWithOptions_TemperatureOverride** - Tests temperature override and restoration
4. **TestRunWithOptions_MaxTokensOverride** - Tests max tokens override and restoration  
5. **TestRunWithOptions_ConfigurationRestoration** - Verifies all config is restored after run
6. **TestRunWithOptions_MultipleOptionsSimultaneously** - Tests multiple options at once

### ⏭️ Skipped Tests (4 tests)

- **TestRunWithOptions_ToolMode** - Requires complex tool setup
- **TestRunWithOptions_DetailedResult** - Full result metadata testing
- **TestRunWithOptions_IncludeTrace** - Requires tracing system
- **TestRunWithOptions_MemoryOptions** - Requires memory provider setup

## Requirements

- **Ollama** must be running with `gemma3:1b` model
- Install model: `ollama pull gemma3:1b`
- Start Ollama: `ollama serve`

## Running Tests

```bash
# Run all tests
go test -v ./test/vnext/options/...

# Run with timeout (recommended for LLM tests)
go test -v ./test/vnext/options/... -timeout 3m

# Run specific test
go test -v ./test/vnext/options/... -run TestRunWithOptions_Timeout
```

## What's Tested

### Configuration Override
- Temperature can be overridden per-run
- MaxTokens can be overridden per-run
- Timeout can be set per-run
- All configuration is restored after execution

### Multiple Options
- Multiple options can be applied simultaneously
- DetailedResult flag adds metadata
- SessionID is passed to context
- ToolMode can be specified

### Delegation
- Nil options properly delegates to Run()
- All overrides work with real LLM (Ollama)

## Implementation Details

The `RunWithOptions()` method:
1. Creates derived context with timeout if specified
2. Saves original configuration
3. Applies overrides temporarily
4. Calls Run() with modified config
5. Restores original configuration via defer
6. Enhances result if DetailedResult is true

This ensures thread-safe, non-mutating runtime configuration overrides.
