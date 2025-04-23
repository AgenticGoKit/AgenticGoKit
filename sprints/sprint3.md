## Summary

Sprint 3 delivers the **ModelProvider & Tool Integration** epic by defining a pluggable LLM abstraction (`ModelProvider`), implementing adapters for **OpenAI**, **Azure OpenAI**, and **Ollama**, and building a generic **FunctionTool** registry. These tasks ensure that any LLM backend can be swapped in, core tools can be invoked as functions by agents, and integration points are fully tested, documented, and benchmarked.  

---

## 1. Define `ModelProvider` Interface

1. **Design `ModelProvider` interface**  
   - Specify methods:  
     ```go
     type ModelProvider interface {
         Call(ctx context.Context, prompt Prompt) (Response, error)
         Stream(ctx context.Context, prompt Prompt) (<-chan Token, error)
         Embeddings(ctx context.Context, texts []string) ([][]float64, error)
     }
     ```  
   - Ensure thread-safety and clear error propagation following Go idioms (avoid panics; return `error`) citeturn0search0turn0search4.  
2. **Write unit tests for interface behavior**  
   - Use a mock provider to verify `Call`, `Stream`, and `Embeddings` signatures and error paths citeturn0search3.  

---

## 2. Implement `OpenAIAdapter`

1. **Add dependency on the official OpenAI Go library**  
   - Import `"github.com/openai/openai-go"` and configure via `option.WithAPIKey` or `os.LookupEnv("OPENAI_API_KEY")` citeturn1search0.  
2. **Implement adapter methods**  
   - Map `ModelProvider.Call` to `client.Chat.Completions.New` and handle streaming variants via `client.Chat.Completions.Stream` citeturn1search0.  
   - For `Embeddings`, call `client.Embeddings.Create` and convert to `[][]float64` citeturn1search1.  
3. **Unit tests & mocking**  
   - Mock `openai.Client` to simulate API responses and errors, verifying adapter wraps errors and contexts correctly citeturn1search2.  

---

## 3. Implement `AzureOpenAIAdapter`

1. **Add Azure SDK dependency**  
   - Import `"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"` and `"github.com/Azure/azure-sdk-for-go/sdk/azidentity"` citeturn1search9.  
2. **Implement `NewAzureAdapter(…)` constructor**  
   - Use `azopenai.NewClientWithKeyCredential` for Azure endpoints and `NewClientForOpenAI` for public endpoints citeturn1search9.  
3. **Adapter mapping**  
   - Implement `Call`, `Stream`, and `Embeddings` by forwarding to `client.GetCompletions`, `client.GetCompletionsStream`, and `client.GetEmbeddings` respectively citeturn1search9.  
4. **Integration tests**  
   - Run against a test Azure OpenAI instance (or mock `azopenai.Client`) to validate parameter and error handling citeturn1search6.  

---

## 4. Implement `OllamaAdapter`

1. **Add Ollama API client**  
   - Vendor or import `"github.com/ollama/ollama/api"` (core REST client) and/or `"github.com/xyproto/ollamaclient/v2"` for richer utilities citeturn2search0turn2search3.  
2. **Construct `OllamaAdapter`**  
   - Read `OLLAMA_HOST` from env, initialize `api.ClientFromEnvironment`, wrap calls to `/api/generate` and `/api/embeddings` endpoints citeturn2search0turn2search5.  
3. **Streaming & error handling**  
   - Support streaming by consuming chunked HTTP responses; wrap errors into adapter error types citeturn2search2.  
4. **Local integration tests**  
   - Spin up a local Ollama server in CI to run end-to-end calls, asserting expected outputs and timeouts citeturn2search6.  

---

## 5. Build Generic `FunctionTool` Framework

1. **Define `FunctionTool` interface**  
   - 
     ```go
     type FunctionTool interface {
         Name() string
         Call(ctx context.Context, args map[string]any) (map[string]any, error)
     }
     ```  
   - Follow ADK’s best practices: JSON-serializable params, no defaults in signatures citeturn3search2.  
2. **Implement a `ToolRegistry`**  
   - Register tools by name; agents lookup and invoke via registry citeturn3search0.  
3. **Sample tools**  
   - Create `WebSearchTool` stub, and `ComputeMetricTool` that runs simple math or API calls, demonstrating invocation patterns citeturn3search4.  
4. **Unit tests for tools**  
   - Validate correct argument parsing, error wrapping, and output structure for each sample tool citeturn3search3.  

---

## 6. Testing, Benchmarking & Documentation

1. **Integration tests**  
   - Compose a `Runner` pipeline with each adapter and a `FunctionTool`, assert end-to-end LLM calls and tool invocations work together citeturn3search6.  
2. **Benchmarks**  
   - Write Go benchmarks for each adapter’s `Call` and `Embeddings`, measuring median latency under 50 ms for remote calls and under 5 ms for local ones (Ollama) citeturn0search1.  
3. **GoDoc & examples**  
   - Document all public types and methods; add examples to `examples/` folder showing how to swap providers and invoke tools citeturn0search8.  
4. **Sprint demo**  
   - Prepare a demo: an agent that queries the web, calls a custom math tool, and summarizes via the local LLM; capture trace logs to showcase `FunctionTool` hooks citeturn3search8.  

---

By completing these tasks, Sprint 3 will equip our framework with a **flexible LLM abstraction layer**, **multiple production-ready adapters**, and a **tool invocation mechanism**, setting the stage for advanced multi-agent workflows in Sprint 4.