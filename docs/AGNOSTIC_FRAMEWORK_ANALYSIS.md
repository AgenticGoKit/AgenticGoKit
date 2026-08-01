# AgenticGoKit — Path to a Truly LLM-Agnostic & Platform-Agnostic Agentic Framework

**Scope:** Full-codebase analysis of `github.com/agenticgokit/agenticgokit` (v0.x, `v1beta` API generation) assessing what is needed to make it a genuinely LLM-agnostic, platform-agnostic, developer-friendly agentic AI framework for Go.

**Method:** Direct inspection of `v1beta/` (~14.5k LOC), `core/` (~9.4k LOC), `core/vnext/` (~11k LOC), `internal/` (~35k LOC), `plugins/`, `examples/`, and `docs/`. Every claim below cites a file (and line where useful).

---

## 1. Executive Summary

AgenticGoKit has strong bones: a clean builder-based public API, streaming-first design, four workflow patterns plus subworkflows, MCP support, embedded-by-default memory (chromem), and genuinely good observability (OpenTelemetry + zerolog + `AGK_TRACE`). Ten LLM providers are wired up, and the docs/examples footprint is large.

However, it is **not yet LLM-agnostic in the ways that matter**. The abstraction is *transport-agnostic* (many HTTP backends) but *semantically OpenAI-shaped and single-turn*:

1. **The core `Prompt` has no message history.** `internal/llm/types.go:30-46` models a call as `{System string, User string}`. Multi-turn agent loops are emulated by string-concatenating previous output and tool results into the next `User` string (`v1beta/agent_impl.go:1623-1631`). This defeats provider-native conversation handling, prompt caching, and tool-call protocols — and is the single biggest blocker to agnosticism.
2. **Tool calling is primarily prompt-engineered text parsing** (`TOOL_CALL{...}` markers parsed in `v1beta/utils.go:494-539`). Native tool calling exists only in the Anthropic and Ollama adapters (`internal/llm/anthropic_adapter.go:300`, `internal/llm/ollama_adapter.go:190`); the OpenAI adapter has **no tool support at all** (`internal/llm/openai_adapter.go`). The team knows this — see `docs/NATIVE_TOOL_CALLING_ANALYSIS.md`.
3. **There are two disconnected provider systems.** `plugins/llm/*` register into the *deprecated* `core` registry (`core.RegisterModelProviderFactory`, `core/llm.go:113`), while the *recommended* `v1beta` bypasses plugins entirely with a hard-coded switch (`internal/llm/factory.go:137-162`). The plugin adapters silently drop tools, multimodal content, and `BaseURL` (`plugins/llm/openai/openai.go:16-73`). Blank plugin imports shown in examples are no-ops for v1beta code paths.
4. **Three API generations coexist** (`core`, `core/vnext`, `v1beta`), and `v1beta` *imports the deprecated `core` package* (`v1beta/provider_factory.go:9`, `v1beta/agent_impl.go:16`), so the promised "delete core in v1.0" is not currently executable without internalizing it first.
5. **Platform coupling is mostly dependency-graph coupling**: one root `go.mod` forces `pgx`, `azure-sdk azcore` (used only in one *test* file — `internal/llm/azure_adapter_test.go:10`), `prometheus`, `chromem`, and `gorilla/websocket` onto every consumer. MCP rides on a personal repo (`github.com/kunalkushwaha/mcp-navigator-go v0.0.2`) rather than the official MCP Go SDK. Workflows have **no checkpoint/resume** — nothing survives a process restart.

The good news: none of these are architectural dead-ends. The fixes are well-understood and sequenced in §6.

---

## 2. Current Architecture Snapshot

```
v1beta/            ← recommended public API (builder, agent, workflow, streaming, memory, tools, eval server)
  └─ imports → core (deprecated!) + internal/llm + internal/observability
core/              ← legacy API #1 (deprecated, still load-bearing: registries, Memory, MCP init)
core/vnext/        ← legacy API #2 (renamed to v1beta; still shipped)
internal/llm/      ← real provider adapters (10 providers, hardcoded factory switch)
internal/{memory,mcp,embedding,observability,orchestrator,...}
plugins/
  llm/*            ← init()-registration into core registry (NOT used by v1beta)
  memory/*         ← chromem / pgvector / weaviate (registered into core registry; v1beta uses these via core)
  mcp/*            ← stdio / tcp / http_sse / http_streaming transports
observability/     ← thin re-export of internal/observability
```

**LLM interface** (`internal/llm/types.go:133-148`):
```go
type ModelProvider interface {
    Call(ctx, Prompt) (Response, error)
    Stream(ctx, Prompt) (<-chan Token, error)
    Embeddings(ctx, texts []string) ([][]float64, error)
}
```

**Provider capability matrix** (as implemented in `internal/llm/`):

| Provider | Streaming | Native tool calls | Multimodal in | Embeddings | Notes |
|---|---|---|---|---|---|
| OpenAI | ✅ | ❌ (none) | ✅ (content parts) | ✅ | default model hardcoded `gpt-4o-mini` (`factory.go:174`) |
| Azure OpenAI | ✅ | ❌ | ✅ | ✅ | requires embedding deployment even if unused (`factory.go:212`) |
| Anthropic | ✅ | ✅ (`anthropic_adapter.go:300,486`) | ✅ | ❌ (no embeddings API) | default `claude-sonnet-4-20250514` |
| Ollama | ✅ | ✅ (`ollama_adapter.go:190-291`) | ✅ | ✅ | |
| OpenRouter | ✅ | ❌ | ✅ | varies | |
| HuggingFace | ✅ | ❌ | partial | varies | 4 API sub-modes; default model `gpt2` (`factory.go:291`) |
| vLLM | ✅ | ❌ | partial | ✅ | |
| MLFlow Gateway | ✅ | ❌ | ❌ | ✅ | own retry config |
| BentoML | ✅ | ❌ | ❌ | varies | own retry config |
| Foundry Local | ✅ | ❌ | ❌ | ❌ | |
| **Missing** | — | — | — | — | **Gemini/Vertex, AWS Bedrock, Mistral, Groq, Together, Cohere, xAI; no generic `openai-compatible` type** |

---

## 3. LLM-Agnosticism: Gaps & Evidence

### 3.1 Single-turn prompt model (critical)
- `Prompt{System, User}` with a `TODO: Add fields for message history` (`internal/llm/types.go:45`).
- Agent tool loops re-prompt by string-stuffing: *"Previous response:\n%s\n\nTool execution results:\n%s\n\n..."* (`v1beta/agent_impl.go:1623-1631` and `1740-1748`).
- Consequences: no assistant/tool role fidelity, no provider prompt caching, degraded quality on long multi-turn tasks, impossible to implement correct native tool-call loops (which require echoing `tool_use`/`tool_result` blocks), and conversation memory must be flattened to text.

### 3.2 OpenAI-shaped neutral types
- `ToolDefinition{Type, Function{...}}` mirrors OpenAI's wrapper exactly (`internal/llm/types.go:17-27`); the shared multimodal builder emits OpenAI content-part JSON (`internal/llm/types.go:154-240`). Anthropic/Gemini formats must be back-translated from OpenAI's shape rather than from a neutral IR.

### 3.3 Tool calling is not provider-native by default
- Primary path parses `TOOL_CALL{...}` text markers from completions (`v1beta/utils.go:494-539`, `v1beta/tools.go:354-496`); agent even instructs models to emit that syntax (`v1beta/tools.go:354-357`).
- Native path exists but only Anthropic + Ollama adapters honor `prompt.Tools`; OpenAI/Azure/OpenRouter/vLLM ignore them silently — a call that "works" on Ollama silently loses native tools on OpenAI.
- `Token` stream carries only `{Content, Error}` — no tool-call deltas, no usage, no finish reason (`internal/llm/types.go:121-129`), so streaming tool use can't be surfaced even where the provider supports it.

### 3.4 Closed, duplicated provider factories
- v1beta → `internal/llm` hardcoded `switch config.Type` (`internal/llm/factory.go:137-162`); adding a provider = editing framework internals.
- `ProviderConfig` is a god-struct with per-provider flattened fields (`HFTopP`, `VLLMBestOf`, `BentoMLRunners`, ... — `internal/llm/factory.go:30-105`); it grows monotonically with every provider.
- The open registry that *does* exist (`core.RegisterModelProviderFactory`, `core/llm.go:104-122`) belongs to the deprecated API, and its plugin adapters drop capabilities (tools/media/BaseURL) in translation (`plugins/llm/openai/openai.go:16-73`).
- Three memory registries similarly coexist: `core/memory.go:322`, `core/vnext/memory.go:159-168`, `v1beta/memory.go:200-209`.

### 3.5 Opinionated defaults leak through the abstraction
- `MaxTokens` silently defaults to **150** (`internal/llm/factory.go:127-129`) — a surprising truncation for agent workloads.
- `Temperature: 0` cannot be expressed (zero-value conflated with "unset", `factory.go:130-132`), even though `ModelParameters` correctly uses pointers (`types.go:10-14`).
- Default embedding dimensions hardcoded to OpenAI's 1536 regardless of provider (`v1beta/provider_factory.go:87,119-126`).

### 3.6 Missing cross-cutting LLM services
- No capability-detection API (`SupportsNativeTools()`, `Modalities()`, `MaxContextWindow()`); callers can't branch safely.
- No structured-output / JSON-schema mode; no `RunTyped[T]`-style API despite Go generics.
- No unified retry/rate-limit/backoff middleware — retries are re-implemented ad hoc per adapter (MLFlow, BentoML configs) and per agent (`RunOptions.MaxRetries`).
- No prompt-caching, reasoning-token, or cost-accounting abstractions; `UsageStats` exists (`types.go:49-53`) but streaming discards it.

---

## 4. Platform-Agnosticism: Gaps & Evidence

### 4.1 Dependency-graph coupling (single go.mod)
Every consumer of the framework pulls: `pgx/v5` + `pgvector-go` (Postgres), `chromem-go`, `prometheus/client_golang`, `gorilla/websocket`, `fsnotify`, and `Azure azcore` — the last used **only in a test helper** (`internal/llm/azure_adapter_test.go:10`) yet listed as a direct dependency in `go.mod:7`. A multi-module layout (core + per-plugin modules) is the standard fix (cf. LangChainGo, OTel).

### 4.2 MCP on a personal fork
`github.com/kunalkushwaha/mcp-navigator-go v0.0.2` (`go.mod:12`) instead of the official `modelcontextprotocol/go-sdk`. Supply-chain and protocol-drift risk; MCP spec is moving fast (auth, streamable HTTP, elicitation).

### 4.3 No durable execution
No checkpoint/resume/persist anywhere in the 2,269-line workflow engine (`v1beta/workflow.go`); state lives in process memory. A crash mid-DAG loses everything. No pluggable state store, no idempotency keys, no distributed runner story — the features that make Temporal-style or serverless deployment possible.

### 4.4 Framework/app boundary blur
- An HTTP eval server ships inside the library package (`v1beta/eval_server.go:38,89,123`).
- Trace files are written to a hardcoded `.agk/runs/<run-id>` relative path (`v1beta/builder.go:790`) — awkward in read-only container filesystems; needs an env override that is documented as primary.

### 4.5 Storage & embeddings
- Vector stores: chromem, pgvector, weaviate only (`plugins/memory/*`). No Qdrant, Milvus, Redis, sqlite-vec, or cloud-managed stores.
- Embeddings are welded to `ModelProvider` (every LLM adapter must implement `Embeddings()` even when the provider has none, e.g., Anthropic) *and* to a separate internal factory supporting only openai/azure/ollama/dummy (`internal/embedding/factory.go:17-42`) — where the "azure" case constructs the *OpenAI* service (`factory.go:35`).
- No document loaders/chunkers as first-class pluggable components; no reranking interface.

### 4.6 What's already good
- Observability is genuinely pluggable and cloud-neutral: OTel tracer with console/file/OTLP exporters selected by env (`v1beta/builder.go:780-806`), zerolog, run-ID context propagation (`observability/exports.go`).
- Config: TOML with `${ENV_VAR}` expansion (`v1beta/config.go:542`) — solid, though TOML-only.
- Pure-Go HTTP adapters (no cgo), so cross-compilation and containers work today.

---

## 5. Developer Experience: Gaps & Evidence

1. **Three coexisting API surfaces** (`core` 9.4k + `core/vnext` 11k + `v1beta` 14.5k LOC over `internal` 35k). New users meet all three in examples: `examples/vllm-quickstart`, `bentoml-quickstart`, `mlflow-gateway-demo` still use `core`; 16 example imports alias v1beta as `vnext`; several blank-import LLM plugins that v1beta never consults.
2. **v1beta cannot shed `core`** without work: it imports `core` in `provider_factory.go`, `agent_impl.go`, `core_wrappers.go`, `tool_discovery.go`, `utils.go` — mostly for Memory and MCP machinery. The v1.0 plan ("remove core entirely", `DEPRECATION.md:126-135`) requires migrating those internals first.
3. **API bloat / legacy compat in the flagship types**: `Result` carries duplicated legacy fields (`v1beta/agent.go:71-77`), `RunOptions` is a 15-field god-struct (`agent.go:424-463`), `ToolManager` mixes MCP-specific methods into the general interface (`agent.go:282-303`).
4. **Reliability of tool use** — because default tool invocation depends on the model emitting `TOOL_CALL{...}` text, results vary wildly by model; this is the #1 practical DX complaint this design will generate.
5. **Versioning confusion**: a directory named `v1beta` inside a v0 module, with a stated plan to become directory `v1` — colliding conceptually with Go's semantic-import-versioning (`/v2` module suffixes). Needs an explicit decision and doc (`docs/API_VERSIONING.md` exists but the module-major story should be spelled out).
6. **Testing kit**: an internal `mock` provider exists (`internal/llm/factory.go:26,477-484`) and is reachable via config `provider="mock"`, but there is no documented public fake LLM/memory/tool kit for user unit tests (compare Genkit's test helpers).
7. **Docs drift**: `docs/ROADMAP.md` calls CLI tooling "completed" in-repo while the CLI now lives in a separate `agk` repo; `docs/reference/api/vnext/` still documents the renamed API; badges point at `kunalkushwaha/agenticgokit` (`README.md:14-15`).
8. **Genuinely good DX already present**: fluent builder (`v1beta/builder.go`), rich stream chunk taxonomy (`v1beta/streaming.go:32-45`), structured error codes (`v1beta/agent.go:577-617`), Mermaid workflow visualization, `AGK_TRACE=true` one-liner observability.

---

## 6. Recommendations (Prioritized)

### P0 — The agnosticism foundation (breaking, do before v1.0)

| # | Recommendation | Addresses |
|---|---|---|
| 1 | **Message-based chat IR.** Replace `Prompt{System, User}` with `Messages []Message` (roles `system/user/assistant/tool`; multi-part content: text/image/audio/tool_use/tool_result). Provide `Prompt`→`Messages` shim for compat. | §3.1 |
| 2 | **One open provider registry.** Make `v1beta` resolve providers through a public registry (`v1beta.RegisterProvider(name, factory)`); reimplement `internal/llm` adapters as self-registering plugins; delete the hardcoded switch and the god-config (per-provider options become typed option structs or `map[string]any` validated by the plugin). | §3.4 |
| 3 | **Native tool calling everywhere, parsing as fallback.** Implement `tools` in the OpenAI/Azure/OpenRouter/vLLM adapters; translate a *neutral* ToolDefinition per provider; keep `TOOL_CALL{}` parsing only as an explicit opt-in fallback for non-tool models. Add tool-call deltas + usage + finish reason to the stream type. | §3.3, §3.2 |
| 4 | **Capability discovery.** `type Capabilities struct { NativeTools, JSONMode, Streaming bool; Modalities []Modality; ContextWindow int }` on every provider; agents branch on it instead of silently degrading. | §3.6 |
| 5 | **Fix parameter semantics.** Pointers (or `omitzero`-style options) for all sampling params; kill the 150-token default; allow `temperature=0`. | §3.5 |

### P1 — Platform agnosticism

| # | Recommendation | Addresses |
|---|---|---|
| 6 | **Multi-module split.** `agenticgokit` (zero heavy deps) + `plugins/llm/*`, `plugins/memory/pgvector`, `observability/prometheus`, etc. as separate Go modules. Remove `azcore` from root deps (test-only). | §4.1 |
| 7 | **Adopt the official MCP Go SDK** (`modelcontextprotocol/go-sdk`), keeping the transport plugins as thin wrappers. | §4.2 |
| 8 | **Separate `Embedder` from `ModelProvider`**; add embedding plugins (Gemini, Voyage, Cohere, local/ONNX); make vector-store an independent `VectorStore` interface; add Qdrant/Redis/sqlite-vec providers; per-provider default dimensions. | §4.5 |
| 9 | **Durable workflows.** `CheckpointStore` interface (memory/file/Postgres/S3 implementations), step-level checkpoints on the existing workflow engine, `workflow.Resume(runID)`. This is the wedge that enables serverless & distributed execution later. | §4.3 |
| 10 | **Move the eval HTTP server and websocket demos out of the library** (separate module or the `agk` repo); make `.agk` trace dir fully configurable with documented env/opt. | §4.4 |
| 11 | **Add the missing majors:** Gemini (native, not proxy), AWS Bedrock, Mistral, Groq — plus a generic `openai-compatible` provider type so any conforming endpoint works with zero code. | §2 matrix |

### P2 — Developer experience

| # | Recommendation |
|---|---|
| 12 | **Execute the deprecation:** migrate the `core` machinery v1beta needs into `internal/`, delete `core` + `core/vnext` from the public surface, convert all examples to plain `v1beta` imports (no `vnext` alias, no dead blank imports). |
| 13 | **Structured output:** `agent.RunTyped[T](ctx, input)` using provider JSON-mode where available, schema-guided retry elsewhere. |
| 14 | **Public testing kit:** `v1beta/agenttest` with scriptable fake provider (incl. tool-call scripts and stream scripts), fake memory, fake tools. |
| 15 | **Cross-provider middleware:** retry/backoff/rate-limit/cost-tracking as provider-wrapping middleware (one implementation, all providers). |
| 16 | **Slim the flagship types:** split `Result` legacy fields behind a compat layer; decompose `ToolManager` (core tools vs MCP extension interface); document the module-major versioning decision. |
| 17 | **Docs hygiene:** retire `docs/reference/api/vnext`, fix `kunalkushwaha/*` badges, align ROADMAP with the split CLI repo. |

### Suggested sequencing
1. **Milestone A (unblocks everything):** #1 messages IR + #3 native tools + #5 params — one coherent breaking change to `internal/llm`.
2. **Milestone B:** #2 registry + #4 capabilities + #11 new providers (community can now contribute providers without touching internals).
3. **Milestone C:** #6 module split + #8 embedder/vector split + #7 MCP SDK.
4. **Milestone D:** #9 durability, then the P2 DX items riding along each release.

---

## 7. How this compares to the Go ecosystem

| Dimension | AgenticGoKit today | LangChainGo | Firebase Genkit (Go) | CloudWeGo Eino |
|---|---|---|---|---|
| Chat IR (multi-turn messages) | ❌ single-turn | ✅ | ✅ | ✅ |
| Native tool calling | partial (2/10 providers) | ✅ | ✅ | ✅ |
| Open provider registry | ❌ (in deprecated API only) | ✅ | ✅ (plugins) | ✅ |
| Structured output | ❌ | partial | ✅ (typed, generics) | ✅ |
| Multi-agent workflows (seq/par/DAG/loop/sub) | ✅ strong | ❌ weak | partial (flows) | ✅ (graphs) |
| Streaming-first API | ✅ strong | partial | ✅ | ✅ |
| Built-in RAG w/ zero config | ✅ (chromem default) | ❌ | partial | ❌ |
| MCP | ✅ (unofficial SDK) | partial | ✅ | ✅ |
| Durable execution | ❌ | ❌ | partial (via Firebase) | ❌ |
| Observability out of the box | ✅ strong | ❌ | ✅ | ✅ |

**Positioning insight:** AgenticGoKit's differentiators are the workflow engine, streaming, batteries-included memory, and observability. Its deficits are exactly the two things every competitor got right first — a message-based LLM IR and native tool calling behind an open registry. Closing §6 P0 makes the framework competitive; P1 makes it the most deployment-flexible option in Go.

---

*Report generated from static analysis of the repository at commit `e4efbc3` ("Guard generated data before persistence (#135)").*
