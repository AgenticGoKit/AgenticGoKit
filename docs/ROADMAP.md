## Summary

AgentFlow has evolved significantly with many core features already implemented, including multi-agent orchestration, workflow visualization, and comprehensive CLI tooling. Below is the current implementation status and future roadmap.

---

## Implementation Status

| Component | Status | Notes |
|-----------|--------|-------|
| **Core Runner & Orchestration** | ✅ **COMPLETED** | Event system, Runner core, Route/Collaborate modes |
| **Multi-Agent Orchestration** | ✅ **COMPLETED** | Collaborative, Sequential, Loop, and Mixed patterns |
| **Workflow Visualization** | ✅ **COMPLETED** | Automatic Mermaid diagram generation |
| **CLI Tooling** | ✅ **COMPLETED** | Project scaffolding, orchestration modes, visualization |
| **ModelProvider Interface** | ✅ **COMPLETED** | OpenAI, Azure OpenAI, Ollama adapters |
| **MCP Integration** | ✅ **COMPLETED** | Tool discovery, execution, server management |
| **Tool Integration** | ✅ **COMPLETED** | Dynamic tool discovery via MCP protocol |
| **Memory & State Management** | ✅ **COMPLETED** | Session storage, state management |
| **Error Handling & Resilience** | ✅ **COMPLETED** | Retry policies, circuit breakers, fault tolerance |
| **Developer Tooling** | ✅ **COMPLETED** | Comprehensive CLI, project templates |
| **REST API** | 🔄 **IN PROGRESS** | HTTP endpoints for programmatic access |
| **Developer UI** | 📋 **PLANNED** | Dashboard for workflow visualization and debugging |
| **Advanced RAG** | 📋 **PLANNED** | Vector database integration, retrieval workflows |

---

## Recently Completed Features

### ✅ Multi-Agent Orchestration (Completed)
- **Collaborative Mode**: All agents process events in parallel with result aggregation
- **Sequential Mode**: Agents process events in pipeline order with state passing
- **Loop Mode**: Single agent repeats execution until conditions are met
- **Mixed Mode**: Combine collaborative and sequential patterns in one workflow

### ✅ Workflow Visualization (Completed)
- **Automatic Mermaid Generation**: CLI generates workflow diagrams automatically
- **Multiple Diagram Types**: Support for all orchestration patterns
- **Custom Configuration**: Configurable diagram styles, directions, and metadata
- **File Export**: Save diagrams as `.mmd` files for documentation

### ✅ Enhanced CLI Tooling (Completed)
- **Orchestration Mode Flags**: `--orchestration-mode`, `--collaborative-agents`, `--sequential-agents`, `--loop-agent`
- **Visualization Flags**: `--visualize`, `--visualize-output`
- **Configuration Options**: `--failure-threshold`, `--max-concurrency`, `--orchestration-timeout`
- **Project Templates**: Complete multi-agent project scaffolding

### ✅ Production-Ready Features (Completed)
- **Fault Tolerance**: Configurable failure thresholds and error handling
- **Retry Policies**: Exponential backoff, circuit breakers, jitter
- **State Management**: Comprehensive state passing between agents
- **Callback System**: Pre/post execution hooks for monitoring
- **Observability**: Structured logging and tracing throughout

---

## Original Roadmap (Historical)

## Sprint 1 (Weeks 1–2): Core Runner & Orchestration ✅ COMPLETED

**Epic**: Establish the heart of the framework—the Runner—that ingests events and dispatches to agents in different orchestration modes.

1. **User Story**: As a developer, I want an `Event` interface so that all agent workflows can receive a uniform input structure.  
   **Acceptance Criteria**:  
   - Define Go interface `Event` with common fields (ID, payload, metadata).  
   - Unit tests for JSON (un)marshalling of `Event`.  

2. **User Story**: As a developer, I want a `Runner` core service to register and emit events so workflows can start.  
   **Acceptance Criteria**:  
   - Implement `Runner.RegisterAgent(name string, agent Agent)`.  
   - Implement `Runner.Emit(event Event)`.  
   - In-memory event queue with FIFO ordering.  

3. **User Story**: As a developer, I want a “route” mode so a single event is forwarded to exactly one agent.  
   **Acceptance Criteria**:  
   - `RouteOrchestrator` picks agent by round-robin or weighted rules.  
   - Unit tests covering routing logic.  

4. **User Story**: As a developer, I want a “collaborate” mode so multiple agents can process the same event in parallel.  
   **Acceptance Criteria**:  
   - `CollaborateOrchestrator` invokes all registered agents concurrently.  
   - Proper error aggregation and timeout handling.  

---

## Sprint 2 (Weeks 3–4): Deterministic Workflow Agents

**Epic**: Build non-LLM workflow primitives for structured pipelines, enabling hybrid flows without external model calls.

1. **User Story**: As a developer, I want a `SequentialAgent` that runs a list of sub-agents in order.  
   **Acceptance Criteria**:  
   - Accepts `[]Agent`, executes one after another, passing state forward.  
   - Error short-circuit: stops on first failure.  

2. **User Story**: As a developer, I want a `ParallelAgent` that runs sub-agents concurrently and aggregates results.  
   **Acceptance Criteria**:  
   - Executes `[]Agent` with Go routines and channels.  
   - Collects outputs and errors into a slice.  

3. **User Story**: As a developer, I want a `LoopAgent` that repeats a sub-agent until a condition is met.  
   **Acceptance Criteria**:  
   - Accepts a `ConditionFunc(State) bool`.  
   - Safety cap on max iterations.  

4. **User Story**: As a developer, I want the workflow agents wired into the Runner so they can be used in orchestrator modes.  
   **Acceptance Criteria**:  
   - Runner can dispatch to any Agent (including workflow agents).  
   - Integration tests for nested workflows.  

---

## Sprint 3 (Weeks 5–6): ModelProvider & Tool Integration

**Epic**: Abstract LLM backends and build in essential tools for agents to call.

1. **User Story**: As a developer, I want a `ModelProvider` interface so any LLM can be plugged in.  
   **Acceptance Criteria**:  
   - Define methods `Call(ctx, Prompt) (Response, error)`.  
   - Mock implementation for tests.  

2. **User Story**: As a developer, I want an `OpenAIAdapter` that implements `ModelProvider` for OpenAI’s APIs.  
   **Acceptance Criteria**:  
   - Supports streaming and non-streaming calls.  
   - Configurable via environment variables.  

3. **User Story**: As a developer, I want an `OllamaAdapter` for local LLM hosting (Ollama).  
   **Acceptance Criteria**:  
   - HTTP client to local Ollama server.  
   - Fallback to file-based prompts if server unavailable.  

4. **User Story**: As a developer, I want a generic `FunctionTool` so agents can execute external functions.  
   **Acceptance Criteria**:  
   - Register named functions with signature `func(Context) (Output, error)`.  
   - Example tools: web search stub, code executor.  

---

## Sprint 4 (Weeks 7–8): Memory & Artifact Management

**Epic**: Build short-term session storage, long-term memory via vector DB, and file artifact handling.

1. **User Story**: As a developer, I want an in-memory `SessionStore` to keep per-user conversation state.  
   **Acceptance Criteria**:  
   - CRUD operations for `Session` and `State` objects.  
   - Concurrency-safe with mutexes or sync.Map.  

2. **User Story**: As a developer, I want a `VectorMemory` interface with a Pinecone driver.  
   **Acceptance Criteria**:  
   - Methods `Store(embedding, metadata)`, `Query(embedding, k)`.  
   - Integration tests with Pinecone sandbox.  

3. **User Story**: As a developer, I want a file `ArtifactService` to save and version artifacts like logs or images.  
   **Acceptance Criteria**:  
   - Local FS backend storing files under `/artifacts/{sessionID}/`.  
   - Metadata JSON manifest for each artifact.  

4. **User Story**: As a developer, I want the memory and artifact services injected into Runner so agents can access them.  
   **Acceptance Criteria**:  
   - Dependency injection via `RunnerConfig`.  
   - Sample agent that writes a debug log artifact.  

---

## Sprint 5 (Weeks 9–10): Debugging, Callbacks & CLI

**Epic**: Enable deep introspection with hooks, traces, and a command-line inspector.

1. **User Story**: As a developer, I want pre- and post-call callbacks around model and tool invocations.  
   **Acceptance Criteria**:  
   - Register callbacks `OnBeforeCall`, `OnAfterCall`.  
   - Callback context includes input, output, timestamps.  

2. **User Story**: As a developer, I want a `TraceLogger` that records each execution step into an in-memory log.  
   **Acceptance Criteria**:  
   - Structured log entries for each agent invocation.  
   - APIs to stream or retrieve trace logs by session ID.  

3. **User Story**: As a developer, I want a CLI command `agentcli trace <sessionID>` to replay an execution trace.  
   **Acceptance Criteria**:  
   - Tabular or tree-view output in the terminal.  
   - Flags for filtering by agent name or time range.  

4. **User Story**: As a developer, I want tests covering callback registration and trace retrieval.  
   **Acceptance Criteria**:  
   - End-to-end test: run a sample workflow, assert trace contains expected entries.  

---

## Sprint 6 (Weeks 11–12): REST API, Dev UI & Deployment

**Epic**: Expose framework functionality via HTTP, build a minimal dashboard, and prepare containerization.

1. **User Story**: As a developer, I want REST endpoints to submit events and retrieve session state.  
   **Acceptance Criteria**:  
   - POST `/v1/events`, GET `/v1/sessions/{id}`.  
   - OpenAPI spec generated.  

2. **User Story**: As a developer, I want a simple React-based Developer UI to visualize trace logs and session memory.  
   **Acceptance Criteria**:  
   - Dashboard page listing active sessions.  
   - On-click drill-down to trace view and memory snapshots.  

3. **User Story**: As a developer, I want Dockerfiles and Helm charts so the framework can be deployed to Kubernetes.  
   **Acceptance Criteria**:  
   - Multi-stage Docker build with Go binary and static assets.  
   - Helm chart values for customizable replicas, resource limits.  

4. **User Story**: As a DevOps engineer, I want a GitHub Actions pipeline that lints code, runs tests, builds Docker image, and publishes to registry.  
   **Acceptance Criteria**:  
   - Workflow triggers on push to `main`.  
   - Artifacts include test coverage report and Docker image tag.  

---

## MCP Integration (Proposed Sprint)

**Epic**: Add support for Model Context Protocol (MCP) tools to enhance the framework's tool ecosystem and context management capabilities.

1. **User Story**: As a developer, I want to define an `MCPTool` interface to standardize MCP tool integration.
   **Acceptance Criteria**:
   - Define `MCPTool` interface with methods for tool invocation and configuration.
   - Unit tests for the interface implementation.

2. **User Story**: As a developer, I want to extend the `ToolRegistry` to support MCP tools.
   **Acceptance Criteria**:
   - Add methods to register and retrieve MCP tools.
   - Ensure compatibility with existing tools.

3. **User Story**: As a developer, I want an adapter to wrap MCP tools for seamless integration.
   **Acceptance Criteria**:
   - Implement an `MCPToolAdapter` to bridge MCP tools with the framework.
   - Unit tests for adapter functionality.

4. **User Story**: As a developer, I want to manage MCP-specific context using the `State` interface.
   **Acceptance Criteria**:
   - Add methods to get and set MCP context in `State`.
   - Ensure context is passed correctly during tool invocation.

5. **User Story**: As a developer, I want tracing hooks for MCP tool invocations to ensure observability.
   **Acceptance Criteria**:
   - Add tracing hooks for MCP tool calls.
   - Log inputs, outputs, and errors for each invocation.

6. **User Story**: As a developer, I want an example workflow demonstrating MCP tool usage.
   **Acceptance Criteria**:
   - Create a new example in the `examples/mcp_tool/` folder.
   - Document the workflow in the `examples README`.

7. **User Story**: As a developer, I want to discover MCP servers dynamically to register their tools.
   **Acceptance Criteria**:
   - Implement an `MCPServerDiscovery` interface for discovering MCP servers.
   - Provide implementations for mDNS-based discovery and static configuration.
   - Unit tests for discovery mechanisms.

8. **User Story**: As a developer, I want to query discovered MCP servers for available tools.
   **Acceptance Criteria**:
   - Implement an `MCPClient` to query MCP servers for tool information.
   - Ensure compatibility with the `ToolRegistry` for dynamic registration.
   - Integration tests for querying and registering tools.

9. **User Story**: As a developer, I want to dynamically register tools from discovered MCP servers.
   **Acceptance Criteria**:
   - Extend the `ToolRegistry` to support dynamic registration of MCP tools.
   - Add an adapter for invoking tools from MCP servers.
   - Unit tests for dynamic registration and invocation.

10. **User Story**: As a developer, I want tracing hooks for MCP server discovery and tool registration.
    **Acceptance Criteria**:
    - Add tracing hooks to log discovered servers and registered tools.
    - Integration tests for tracing functionality.

11. **User Story**: As a developer, I want an example workflow demonstrating MCP server discovery and tool usage.
    **Acceptance Criteria**:
    - Create a new example in the `examples/mcp_discovery/` folder.
    - Document the workflow in the `examples README`.

---

## RAG Integration (Proposed Sprint)

**Epic**: Add support for Retrieval-Augmented Generation (RAG) to enhance the framework's ability to combine retrieval mechanisms with generative AI models.

1. **User Story**: As a developer, I want to define a `Retriever` interface to standardize retrieval operations.
   **Acceptance Criteria**:
   - Define `Retriever` interface with methods for querying and retrieving documents.
   - Unit tests for the interface implementation.

2. **User Story**: As a developer, I want to implement retriever backends for popular systems like Pinecone and Elasticsearch.
   **Acceptance Criteria**:
   - Implement `PineconeRetriever` and `ElasticsearchRetriever`.
   - Integration tests for each backend.

3. **User Story**: As a developer, I want a `RAGAgent` that combines retrieval and generation workflows.
   **Acceptance Criteria**:
   - Implement `RAGAgent` to query retrievers and pass retrieved context to LLMs.
   - Unit tests for the agent's logic.

4. **User Story**: As a developer, I want to extend the `State` interface to manage retrieved context.
   **Acceptance Criteria**:
   - Add methods to store and retrieve RAG-specific context in `State`.
   - Ensure compatibility with existing agents.

5. **User Story**: As a developer, I want tracing hooks for RAG workflows to ensure observability.
   **Acceptance Criteria**:
   - Add tracing hooks to log retrieval queries, results, and LLM outputs.
   - Integration tests for tracing functionality.

6. **User Story**: As a developer, I want an example workflow demonstrating RAG usage.
   **Acceptance Criteria**:
   - Create a new example in the `examples/rag_workflow/` folder.
   - Document the workflow in the `examples README`.