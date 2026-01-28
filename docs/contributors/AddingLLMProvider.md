# Adding a New LLM Provider

This guide outlines the process of adding a new Large Language Model (LLM) provider driver to AgenticGoKit. We use the **Anthropic** implementation as a reference.

## Architecture Overview

AgenticGoKit uses a layered architecture for LLM providers:

1.  **Core Interface (`internal/llm/types.go`)**: Defines the `ModelProvider` interface that all drivers must implement.
2.  **Internal Adapter (`internal/llm/<provider>_adapter.go`)**: The concrete implementation of the driver.
3.  **Factory (`internal/llm/factory.go`)**: Instantiates the provider based on configuration.
4.  **Wrappers (`internal/llm/wrappers.go`)**: Adapts the internal provider for public/plugin consumption.
5.  **Plugin (`plugins/llm/<provider>/<provider>.go`)**: Registers the provider with the core system so it can be used via configuration strings.
6.  **Configuration (`v1beta/config.go`)**: High-level configuration validation.

---

## Step-by-Step Guide

### Step 1: Create the Internal Adapter

Create a new file `internal/llm/<provider>_adapter.go`.

Implement the `ModelProvider` interface:

```go
type ModelProvider interface {
    Call(ctx context.Context, prompt Prompt) (Response, error)
    Stream(ctx context.Context, prompt Prompt) (<-chan Token, error)
    Embeddings(ctx context.Context, texts []string) ([][]float64, error)
}
```

**Key Responsibilities:**
- **Struct Definition**: Create a struct (e.g., `AnthropicAdapter`) to hold config (apiKey, model, etc.).
- **Constructor**: Create `New<Provider>Adapter` and `New<Provider>AdapterWithConfig`.
- **Call()**: Handle non-streaming requests. Map the generic `Prompt` to the provider's specific request format. Convert the response back to `Response`.
- **Stream()**: Handle streaming requests. Should return a channel of `Token`s.
- **Embeddings()**: Return embeddings if supported. If not, return a clear error.
- **Tool Calling**: If the provider supports tool/function calling, map `Prompt.Tools` to the provider's schema and parse tool calls in the response.

**Example Reference:** `internal/llm/anthropic_adapter.go`

### Step 2: Update the Internal Factory

Modify `internal/llm/factory.go`:

1.  **Add Constant**: Add a new `ProviderType` constant.
    ```go
    const (
        // ...
        ProviderTypeAnthropic ProviderType = "anthropic"
    )
    ```

2.  **Update Config**: Add provider-specific fields to `ProviderConfig` struct.
    ```go
    type ProviderConfig struct {
        // ...
        // Anthropic-specific fields
        AnthropicTopP       float32  `json:"anthropic_top_p,omitempty"`
        // ...
    }
    ```

3.  **Add Switch Case**: Add a case in `CreateProvider` switch statement.

4.  **Create Factory Method**: Implement `create<Provider>Provider` method to map `ProviderConfig` to your adapter's config.

### Step 3: Update Wrappers

Modify `internal/llm/wrappers.go`:

1.  **Add Public Config**: Add fields to `PublicLLMProviderConfig` (mirroring `ProviderConfig`).

2.  **Add Constructor Wrapper**: specific wrapper function for your provider.
    ```go
    func NewAnthropicAdapterWrapped(...) (PublicModelProvider, error) {
        adapter, err := NewAnthropicAdapter(...)
        // ...
        return NewModelProviderWrapper(adapter), nil
    }
    ```

3.  **Update Generic Wrapper**: Update `NewModelProviderFromConfigWrapped` to map the public config fields to the internal `ProviderConfig`.

### Step 4: Create the Plugin

Create a new directory `plugins/llm/<provider>/` and a file `<provider>.go`.

1.  **Adapter Struct**: Create a simple struct to adapt the public wrapper to `core.ModelProvider`.
2.  **Factory Function**: Create a function that initiates your wrapper from `core.LLMProviderConfig`.
3.  **Registration**: Use `init()` to register your factory.

```go
func init() {
    core.RegisterModelProviderFactory("anthropic", factory)
}
```

**Example Reference:** `plugins/llm/anthropic/anthropic.go`

### Step 5: Update Configuration Validation

Modify `v1beta/config.go`:

1.  Update `validateLLMProvider` function to include your new provider string in `validProviders`.

---

## Testing

### Unit Tests
Create `internal/llm/<provider>_adapter_test.go`.

- **Test Logic**: Verify request building and response parsing.
- **Test Streaming**: Verify the channel behavior.
- **Mocking vs Live**: For live tests requiring API keys, check for the environment variable and `t.Skip()` if missing.

```go
func TestAnthropicAdapter_Call(t *testing.T) {
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    if apiKey == "" {
        t.Skip("Skipping integration test")
    }
    // ... test code ...
}
```

### Verification
1.  Run `go build ./...` to ensure all interfaces match.
2.  Run `go test ./internal/llm/...` to run your tests.
