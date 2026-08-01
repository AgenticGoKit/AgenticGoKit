# AgenticGoKit — Roadmap to the Best Developer-Focused Agentic Framework in Go

**Method:** 10-agent deep research pass — 4 agents re-reading this codebase subsystem by subsystem (workflow engine, streaming/handlers, tools/MCP, config/observability/eval), 6 agents studying the current framework landscape from primary sources (LangGraph, PydanticAI, OpenAI Agents SDK, Genkit Go, Google ADK-go, CloudWeGo Eino, Temporal/Restate/DBOS/Inngest/River, MCP go-sdk, Vercel AI SDK v5, Mastra, mem0/Letta, smolagents). 146 findings synthesized here. Companion documents: `docs/AGNOSTIC_FRAMEWORK_ANALYSIS.md` (LLM/platform agnosticism) and issues #139–#155.

---

## 0. Thesis: what "best developer-focused" means in Go

Every successful 2025–26 framework won on a different axis: LangGraph on **durable state**, PydanticAI on **type safety**, Genkit/Mastra on the **dev loop**, Eino on **typed streaming orchestration**, the OpenAI SDK on **small proven primitives** (handoffs, guardrails, sessions). Nobody in Go has combined them, and the incumbent Go options are weak: LangChainGo is a pre-generics port stuck at v0.1.x openly seeking maintainers; Eino is powerful but carries a steep concept load (four streaming paradigms, boxing, pregel-vs-dag) that its own docs apologize for.

Go's unfair advantages line up exactly with what agent developers are missing:

1. **Compile-time safety** → typed agents, tools, and workflow steps that fail at build, not on request N in production.
2. **Single static binary** → dev UI, API server, console REPL, and A2A endpoint from *one* compiled artifact (embed.FS) — no npm, no sidecar.
3. **Real concurrency** → parallel tool calls, guardrails racing the main agent, streaming fan-in — things Python frameworks emulate.
4. **`go test` culture** → evals and LLM-free agent tests as ordinary tests, race-detector-clean as a brand promise.
5. **Temporal-class durability heritage** → Go is *the* language of durable execution (Temporal, DBOS, Restate, River all ship first-class Go SDKs); an embedded journal-skip engine is a natural fit.

The strategy: **fix the trust layer, then bet on three differentiators — typed everything, durable by default, and the best dev loop in the ecosystem.**

---

## 1. Part I — The trust layer: correctness debt that blocks "best DX" claims

Deep re-reading found bugs that would end an evaluation by a serious Go team. These come **before** any feature work; several fail `go test -race`.

### 1.1 Workflow engine (v1beta/workflow.go)

| # | Finding | Evidence |
|---|---------|----------|
| W1 | **DAG + RunStream silently executes as Sequential** — `executeDAGStreaming` is a stub delegating to sequential; dependencies ignored, no validation. Same object, different semantics per entry point. | workflow.go:796-801 |
| W2 | **`Step.Transform` applied twice in every streaming path** ("Summarize: Summarize: ..."). | workflow.go:599-614, 712-727 vs 928-931 |
| W3 | **Workflows are single-use and racy**: one shared `WorkflowContext` across runs (stale StepResults on second Run), unlocked writes to `CurrentStep`/`IterationNum` under Parallel mode. Needs per-run state; Workflow should be a reusable template like `http.Handler`. | workflow.go:248-256, 271-272, 1694, 1537 |
| W4 | **No per-step retry/timeout/failure strategy**; parallel mode returns `errors[0]` and discards the rest (no `errors.Join`), no fail-fast cancellation. Ironically tools config *has* retry/circuit-breaker knobs. | workflow.go:60-67, 1324-1327; config.go:158-169 |
| W5 | **DAG execution is serial** (ready steps run one-at-a-time), O(n²), and cycles/typo'd dependencies surface only mid-run as "deadlock" after burning tokens. | workflow.go:1375-1497 |
| W6 | **Stringly-typed data flow**: fan-in is `"\n"`-join of dependency outputs — a step can't tell which dep produced what; failures silently fall back to original input. | workflow.go:1737-1757 |
| W7 | Streaming discards the whole `WorkflowResult` (no StepResults/IterationInfo) and **emits the final output twice**. | workflow.go:493-517, 472-476 |
| W8 | Step errors flattened to strings with `%s` — `errors.Is/As` impossible; the framework's own typed-error package is unusable against its own engine. | workflow.go:1715-1726, 1174-1181 |
| W9 | Loop condition builders are stateful one-shot closures (reuse silently mis-converges); `Convergence` "edit distance" is a positional byte compare. | workflow.go:1914-1961 |
| W10 | Library **prints to stdout and `os.Setenv`s AGK_RUN_ID/TRACE vars as IPC** — races across concurrent workflows; unacceptable in a server. | workflow.go:332-388 |
| W11 | SubWorkflow recursion guard inert (depth never propagates); `RunStreamWithOptions` drops options; nil-deref on error path. | subworkflow.go:117, 229-232, 154-160 |

### 1.2 Streaming (v1beta/streaming.go, agent_impl.go)

| # | Finding | Evidence |
|---|---------|----------|
| S1 | **`Stream.Cancel()` never cancels the LLM request** — provider gets the parent ctx; tokens (and money) keep flowing after cancel. | streaming.go:277-280 vs agent_impl.go:1003 |
| S2 | **Producer errors never reach `Wait()`** — failed streams return `(nil, nil)`; workflow steps mark failed agents `Success=true` and feed empty output onward. | agent_impl.go:947-949, 1004-1013; workflow.go:1017-1029 |
| S3 | **`RunStreamWithOptions` mutates the live agent** (config swap restored before the goroutine reads it — a straight race; `a.tools` filtered with *no restore at all* — one `ToolMode:"none"` call permanently strips tools). | agent_impl.go:1151-1223 |
| S4 | **Parallel workflow streaming is a data race** on `basicStream.Write` (unlocked `chunkIdx`, concurrent user handler). | workflow.go:688-729; streaming.go:346-393 |
| S5 | **`Capabilities.Tools`/`Memory` are never wired** — the entire handlers.go augmentation library (`WithToolAugmentation` etc.) silently no-ops; three incompatible tool-call prompt protocols coexist. | agent_impl.go:573-585; handlers.go:77-86, 100-131 |
| S6 | `StreamChunk.Error` is type `error` → JSON-marshals as `{}`; **no SSE/WebSocket bridge exists**, so examples hand-roll three ~236-line websocket servers with a parallel message taxonomy. | streaming.go:60; examples/*/websocket_server.go |
| S7 | `FlushInterval`, `Config.Streaming` are **dead config** — docs recommend tuning knobs nothing reads. | streaming.go:174-250; docs/v1beta/performance.md:81-87 |
| S8 | Backpressure undefined: blocking Write + hidden always-on 5-minute timeout; sync handler in the hot path; handler `false` doesn't stop generation. | streaming.go:378-392, 187 |
| S9 | Tool calls parsed only after stream ends — raw `TOOL_CALL{...}` text streams to the user; no tool-call deltas; no post-tool synthesis in stream path; ToolRes chunks lack ToolID. | agent_impl.go:1068-1087, 2094-2145 |

### 1.3 Tools & MCP

| # | Finding | Evidence |
|---|---------|----------|
| T1 | **Reliability config is decorative**: `Timeout`, `MaxRetries`, `RateLimit`, `MaxConcurrent`, `CircuitBreaker` are validated, documented — and never enforced anywhere in the execution path. A hung tool blocks forever. | tools.go:214-239 vs agent_impl.go:1515-1638 |
| T2 | **`MCPAwareAgent` fabricates successful tool results** when caching is off (`executeToolDirect` returns hardcoded "Result from %s" — never contacts a server). Plausible wrong answers, silently. | internal/mcp/mcp_agent.go:277-310 |
| T3 | **Unified MCP transport dials + handshakes + disconnects per tool call**; the 1,800 lines of pooling/load-balancing/retry code in internal/mcp are never instantiated. `Connect()` is a no-op bool-flip, so health checks lie. | plugins/mcp/unified/unified.go:817-831, 602-624 |
| T4 | Fallback tool schema is `{"input": string}` — teaches models to double-wrap args; the framework ships a `normalizeToolArgs` hack to undo its own fallback. | agent_impl.go:2077-2092; unified.go:33-70 |
| T5 | Only tool registration is a **global unlocked map**; every agent inherits every tool; no `WithTool(t Tool)` builder option. | tool_discovery.go:25-31; builder.go:226-287 |
| T6 | stdio MCP servers can't receive **args/env** (`npx -y server-filesystem /path` unconfigurable); the stdio plugin the error messages recommend is an unimplemented stub. | core/mcp.go:212-220; plugins/mcp/stdio/stdio.go:28-53 |
| T7 | Transport selection is last-`init()`-wins on an unsynchronized global; unknown tools **route to an arbitrary server**; names collide first-wins. | plugins/mcp/*/…; unified.go:783-800 |
| T8 | No parallel tool execution despite `ParallelExecution: true` config; no human-approval hooks or dangerous-tool gating anywhere; MCP auth is one global env token for all servers. | agent_impl.go:1691-1702; unified.go:962-998 |
| T9 | Weather-demo heuristics (`"sf"→"san francisco"`) hardcoded in the generic pipeline. | agent_impl.go:1893-1905 |
| T10 | Cache config accepts redis/file backends and lfu/ttl eviction — only in-memory LRU exists; size caps not enforced. | internal/mcp/cache_manager.go:57 |

### 1.4 Config, observability, eval

| # | Finding | Evidence |
|---|---------|----------|
| C1 | **Unknown TOML keys silently ignored** (BurntSushi `md.Undecoded()` makes strict mode a ~10-line fix). | config.go:348, 379 |
| C2 | **`timeout = 30` parses as 30 nanoseconds** and passes validation → instant, undebuggable deadline errors. | config.go:24; builder.go:767 |
| C3 | TOML provider whitelist **rejects five providers the factory supports** (openrouter, huggingface, vllm, mlflow, bentoml). | config.go:650 vs internal/llm/factory.go:15-26 |
| C4 | Env expansion is a 12-field whitelist (misses `Memory.Options` where `embedding_api_key` lives); **unset `${VAR}` passes through as the literal string** and becomes a 401 downstream. | config.go:433-457, 556-566 |
| C5 | Non-critical validation errors are computed, then thrown away. | config.go:640-645 |
| C6 | Span attributes are bespoke `agk.*`, not **OTel GenAI semconv (`gen_ai.*`)** — invisible to Langfuse/Datadog/Braintrust LLM views; no per-LLM-call span; cost attribute defined, never emitted. | internal/observability/attributes.go:12-23 |
| C7 | OTLP export broken as documented (scheme passed to `WithEndpoint`; needs `WithEndpointURL`); no headers → no SaaS backend reachable; std `OTEL_*` env vars ignored. | builder.go:839-841; tracer.go:78 |
| C8 | `.agk/runs/trace.jsonl` is OTel's *debug* stdouttrace format — explicitly unstable; manifest parser pokes at Go field names; 64KB Scanner limit truncates detailed spans. | internal/observability/exporters.go:42-68 |
| C9 | `manifest.json` fabricates stats: `llm_calls` always 0 (name-match on "llm" that no span satisfies), tokens/cost never computed (already tracked as #138). | agent_impl.go:1436-1463 |
| C10 | Eval server: traces are empty shells (no spans/IO, never persisted), **sessions don't feed history to the agent**, workflow success hardcoded `true`, binds `:8787` all-interfaces, no auth, no HTTP timeouts. | eval_handlers.go:140-307; eval_server.go:122-126 |
| C11 | Agent metrics unsynchronized (race), private, write-only; Prometheus dep serves only the legacy MCP path. | agent_impl.go:142-148, 716-728 |

**Deliverable for Part I: a "Trustworthy" release whose banner claims are `go test -race` clean, `go vet` clean, every documented config knob enforced or deleted, `errors.Is/As` works end-to-end, and no fabricated data anywhere (results, manifests, health checks).** File issues per row; most are individually small.

---

## 2. Part II — The differentiating feature set

Six pillars, each grounded in proven patterns from the landscape research, each with a Go-idiomatic design sketch.

### Pillar A — Typed everything (Go's headline advantage)

**A1. Generic agents — `Agent[Deps, Output]`** *(PydanticAI's headline feature; maps perfectly to Go generics)*
```go
agent := agk.New[Deps, Verdict](model,
    agk.Output(agk.ToolMode),        // structured-output strategy: tool | native | prompted
    agk.Retries(agk.RetryBudget{Tools: 2, Output: 1}),
)
res, err := agent.Run(ctx, prompt, deps)   // res.Output is Verdict — no casting
```
Typed dependency injection via `RunContext[Deps]` (deps, retry count, usage-so-far) as the second arg to tools/validators; `agent.Override(deps|model)` for tests.

**A2. Typed tools with inferred schemas — kill `map[string]interface{}`** *(convergent across Eino InferTool, adk-go functiontool, MCP go-sdk, AI SDK v5)*
```go
tool := tools.New[SearchArgs, []Result]("search", "Search the web",
    func(ctx context.Context, rc *agk.RunContext[Deps], in SearchArgs) ([]Result, error) { ... })
```
Input *and output* JSON schemas derived from struct tags (`jsonschema:"required,enum=..."`), schema-aware arg coercion, output validated before returning to the model. Fixes T4/T5 as a side effect (per-agent `WithTool`, no global map, no `{"input":...}` fallback).

**A3. Structured output with validation-driven self-correction** — declare once (`Output[T]`), choose mode independently (tool-call schema / provider-native JSON / prompted), and wire `agk.Retry("use a full name")` errors back to the model with per-tool/output budgets (PydanticAI `ModelRetry` — the proven reliability loop).

**A4. Typed workflow steps** *(Mastra `.then/.branch/.parallel`, Eino compile-time edge checks — fixes W6)*
```go
wf := flow.New[Query, Report]().
    Then(classify).                       // Step[Query, Class]
    Branch(flow.Case(isRefund, refundFlow), flow.Default(generalFlow)).
    Parallel(fetchDocs, fetchHistory).    // fan-out, typed merge
    Then(compose)                         // compile error if types don't line up
```
Keep the existing string-based API as a compatibility layer over the typed core. Per-step `Retry`, `Timeout`, `OnError(Skip|Fail|Fallback)`, `Cache(ttl, keyFn)`; `Validate()` (cycles, unknown deps) and `Mermaid()/DOT()` export before any token is spent.

**A5. Model fallback & flow control as decorators** — `llm.NewFallback(primary, secondary, llm.FallbackOn(...))` over the existing `ModelProvider` interface; declarative per-agent throttle/debounce/rate-limit/priority (Inngest-style) instead of hand-rolled `x/time/rate`.

### Pillar B — Durable by default (the strategic wedge)

Nothing in the Go agent space has this; Go's durable-execution pedigree makes it the highest-leverage bet. Design synthesis from the research:

**B1. Journal-skip, not event-replay.** Re-invoke the function from the top on recovery; completed steps return checkpointed results (DBOS/Restate/Inngest model). *No* Temporal-style determinism constraints on user code — plain goroutines and `time.Now` stay legal outside steps.

**B2. Embedded SQL journal, zero infrastructure.** Two tables (run status + step journal) on SQLite by default, Postgres for scale (DBOS proved this is a library, not a server). The journal doubles as the debugger: standard SQL to inspect any run.

**B3. One tiny contract, community backends.** LangGraph won persistence via a 5-method `Checkpointer` interface + conformance test suite. Copy that discipline: `Put/PutWrites/Get/List/DeleteThread`, addressed by `(threadID, namespace, checkpointID)` with pending-writes so successful parallel siblings aren't re-executed.

**B4. Auto-wrap LLM/tool/MCP calls as durable steps** (PydanticAI-Temporal pattern) — durability is a builder option, not a different agent type: `agk.Durable(store)`. Per-run knob `Durability(Exit|Async|Sync)`.

**B5. `Interrupt`/`Resume` for human-in-the-loop** — the #1 production ask. LangGraph semantics (node re-executes to the interrupt point; recorded value returned on resume) are implementable in Go *without* stack serialization:
```go
approval, err := flow.Interrupt[Approval](ctx, payload)  // suspends durably
// later: runner.Resume(ctx, runID, approval)
```
Plus awakeables (hand an ID to an external system; run suspends until resolved), durable `Sleep`/cron, and ADK's wire-level convention (a reserved `request_confirmation` event any frontend can render).

**B6. Time travel & fork-from-step** — `ForkRun(runID, fromStep)` reusing checkpointed results: "the agent went off the rails at tool call 7 — fork from 6 with a corrected prompt" *without re-paying for steps 1-6*. Uniquely valuable when steps cost real money.

**B7. Exactly-once honesty** — deterministic idempotency keys `(runID, stepSeq)` surfaced to tools via context; transactional checkpoint+side-effect for DB writes (River's insight); DBOS-style boolean `Patch("change-id")` for code versioning of in-flight runs; a `workflowcheck`-style `go vet` analyzer flagging raw I/O outside steps — a distinctly Go advantage.

### Pillar C — Streaming that's a product feature, not a liability

**C1. Fix the semantics first** (Part I S1-S9): `Wait()` as the single source of truth for errors, real cancellation, race-free writer, usage/finish-reason on the final token, incremental tool-call deltas with stable IDs, suppression of protocol text.

**C2. Speak the Vercel AI SDK v5 UIMessage SSE protocol natively.** It is the de facto frontend contract (PydanticAI and Mastra both emit it). A `uistream` package — `http.Handler` that serves any agent/workflow stream as v5-typed SSE parts (text-delta, reasoning, tool-input-delta, tool-output-available, data-*) — makes every React/Svelte `useChat` frontend work against AgenticGoKit out of the box, and deletes the three hand-rolled websocket servers in examples.

**C3. Resumable streams** — pluggable `EventStore` (memory/Redis) with `Last-Event-ID` replay; both AI SDK v5 and MCP streamable-HTTP converged on this design.

**C4. Stream-aware routing** (Eino's trick): decide "tool call or final answer?" from the first frame with a pluggable checker — preserve first-token latency in ReAct loops. Expose only `Invoke`/`Stream` publicly; keep Eino's Collect/Transform machinery internal (their concept-load mistake is our lesson).

**C5. Custom progress events** — LangGraph's `get_stream_writer()` equivalent: any tool/step can emit typed progress ("fetched 3/10 pages") into the client stream.

### Pillar D — Multi-agent primitives the industry converged on

**D1. Agents-as-tools as the primary composition currency** (Eino explicitly deprecated transfer-style handoffs; OpenAI ships both). `subAgent.AsTool(name, desc, WithOutputExtractor(...))` — parent keeps control; sub-agent events pass through to the parent stream.

**D2. Handoffs where control must move** — auto-generated `transfer_to_X` tools with typed, schema-validated payloads and history filters (OpenAI SDK), implemented as `Command{Update, Goto}` returns (LangGraph) so state-update+routing is one primitive.

**D3. Dynamic fan-out (`Send`)** — orchestrator-worker/map-reduce over collections not known at build time; with Go this is `errgroup` + semaphore under a typed API, a marquee "Go does this better" feature.

**D4. Guardrails with tripwires, raced in parallel** — cheap model screens while the expensive agent runs; tripwire cancels via context. Input/output/tool-level. Go's cancellation makes this *genuinely better* than the Python original.

**D5. Sessions & shared state** — OpenAI's 4-method `Session` interface (Items/Append/PopItem/Clear — the pop enables undo) + ADK's scope-prefixed state (`app:`/`user:`/session/`temp:`) with `OutputKey` auto-write and `{key}` instruction templating. Depends on the message-IR work (#139).

**D6. Usage limits as hard brakes** — `UsageLimits{Requests, ToolCalls, InputTokens, OutputTokens}` per run (PydanticAI); `tool_choice` forcing + `stop_on_first_tool` (skip a whole LLM round-trip when the tool result *is* the answer).

### Pillar E — The dev loop (adoption is won here)

**E1. `agk dev` — a single-binary playground.** Genkit's Dev UI and Mastra's playground are their most-loved features; Eino built an IDE plugin for the same reason. Go can do it best: embed the UI with `embed.FS`, expose a small reflection protocol (list actions, run action, stream trace), and ship chat-with-any-agent, workflow graph visualization, run/suspend/resume, trace inspection with per-step payloads, and an eval tab — from the compiled binary, zero npm. The ADK-go **launcher pattern** completes it: the *same* binary runs `console` | `dev` | `api` | `a2a`.

**E2. An `Action` registry as the substrate** (Genkit's architectural insight): agents, workflows, tools, prompts, evaluators all register as one introspectable unit (name, kind, in/out schema, run fn). The Dev UI, HTTP serving, eval runner, and MCP-server export all become one lookup away.

**E3. `.prompt` files** (dotprompt): prompts as versioned artifacts — YAML frontmatter (model, config, typed input/output schema) + template body, loaded from a directory or `embed.FS`, with `name.variant.prompt` for A/B tests and typed wrappers (`LookupPrompt[In, Out]`).

**E4. Evals under `go test`.** The industry converged on Eval = data + task + scorers (Braintrust), native-test-runner integration (DeepEval), CI thresholds (promptfoo), and scorers that also run against sampled production traces (Mastra). Go shape: `agkeval.Run(t, dataset, task, scorers...)` with trials, LLM-judge + heuristic scorers, span-based assertions ("was tool X called before Y?"), and — ADK's differentiator — **capture a live session in the dev UI → save as an eval case**. Rebuild the eval server on this (fixes C10).
Prereq: honest cost/token accounting per run (fixes C9, #138/#153).

**E5. LLM-free testing kit** (extends #152 with proven designs): `TestModel` (auto-calls every tool with schema-valid fabricated args — exercises plumbing with zero LLM), `ScriptModel` (scripted tool-call/text turns), `agk.BlockRealRequests()` global guard for CI, `CaptureMessages()` transcript assertions.

**E6. Meta-DX for the AI era**: scaffold `AGENTS.md` in generated projects and ship one in-repo (framework users build *with* coding agents; 60k+ repos, Linux Foundation spec); publish `llms.txt`/`llms-full.txt` + per-page markdown docs — the dominant consumer of framework docs is now an assistant in Cursor/Claude Code.

### Pillar F — Ecosystem & interop

**F1. Official MCP go-sdk, both directions** (#145, expanded): the SDK is v1.0-stable with streamable HTTP, OAuth, elicitation, sampling, resumable event stores. Client side replaces the personal fork *and* fixes T3/T6/T7; server side is one line — `agent.AsMCPServer()` — making any AgenticGoKit agent installable in Claude/VS Code/Cursor; `agk mcp publish` generates `server.json` for the official MCP Registry.

**F2. OTel GenAI semantic conventions** (fixes C6-C8): `gen_ai.*` attributes, per-LLM-call CLIENT spans, `WithEndpointURL` + headers + standard `OTEL_*` env vars, a stable documented trace schema, real per-run manifest rollups. Every LLM observability vendor reads this dialect.

**F3. Memory with extraction semantics** (builds on #146): keep the storage layer, add the two proven architectures — a mem0-style background pipeline (extract → consolidate → ADD/UPDATE/DELETE facts, invisible to the agent loop) and Letta-style tiers (core/recall/archival) with agent-editable core memory as tools; thread/resource scoping aligned with D5 sessions.

**F4. A2A via `a2a-go`** when agents cross team/vendor boundaries: serve an Agent Card + task lifecycle from the launcher (`agk a2a`), and consume remote agents as sub-agents. Adopt the SDK; don't reimplement.

**F5. Code-as-action mode (opt-in, differentiating):** smolagents showed ~30% fewer steps when the model writes code instead of JSON tool calls; Python needs external sandboxes — Go can embed one natively (Starlark via starlark-go, or WASM via wazero) with registered typed tools injected as functions. No other Go framework can demo this safely today.

---

## 3. Part III — Sequenced roadmap

| Release | Theme | Contents |
|---|---|---|
| **v0.6 "Trustworthy"** | Zero known lies | All Part I fixes (W/S/T/C tables). Brand promise: `-race` clean, every config knob real, `errors.Is` end-to-end, honest manifests/health/results. Strict TOML + duration parsing (C1-C4). Ship `AGENTS.md` + `llms.txt` (E6 — near-zero cost). |
| **v0.7 "Typed"** | Compile-time safety story | Message IR (#139) → native tools everywhere (#140) → open provider registry (#141-142) → `Agent[D,O]` (A1) + `tools.New[In,Out]` (A2) + structured output w/ retry loop (A3) + typed workflow steps (A4). Testing kit E5. Gemini/Bedrock/`openai-compatible` (#149). |
| **v0.8 "Durable"** | The wedge | Journal-skip engine on SQLite (B1-B3), auto-wrapped steps (B4), Interrupt/Resume + approval events (B5), sessions & scoped state (D5), durable sleep/cron. Multi-module split lands here (#144) so the journal store is a plugin. |
| **v0.9 "Delightful"** | Dev loop | Action registry (E2) → `agk dev` playground + launcher (E1) → AI SDK v5 wire format + resumable SSE (C2-C3) → `.prompt` files (E3) → evals under `go test` + capture-loop (E4) → OTel GenAI (F2). |
| **v1.0 "Complete"** | Ecosystem | Multi-agent primitives (D1-D4, D6), MCP server export + registry publish (F1), memory extraction tiers (F3), A2A (F4), fork-from-step time travel (B6), guardrails, fallback/flow-control decorators (A5). Delete `core`/`core/vnext` (#150). |

**Dependency spine:** message IR → typed agents/tools → sessions → durability → dev UI → evals. Everything else hangs off that spine.

---

## 4. The checklist: "best developer-focused Go agentic framework"

| Capability | AGK today | Target | Best-in-class reference |
|---|---|---|---|
| Multi-turn message IR | ❌ single-turn | ✅ | everyone |
| Typed agents `Agent[D,O]` | ❌ | ✅ | PydanticAI |
| Typed tools, inferred schemas (in+out) | ❌ map[string]any | ✅ | Eino / adk-go / MCP SDK |
| Structured output + self-correction | ❌ | ✅ | PydanticAI |
| Typed workflow steps, build-time checks | ❌ strings | ✅ | Eino / Mastra |
| Per-step retry/timeout/fallback/cache | ❌ | ✅ | LangGraph / Temporal |
| Dynamic fan-out (Send/map-reduce) | ❌ | ✅ | LangGraph |
| Durable checkpoint/resume, zero infra | ❌ | ✅ SQLite journal | DBOS (design), LangGraph (contract) |
| Interrupt/HITL approval, resumable | ❌ | ✅ | LangGraph / Mastra / ADK |
| Time travel / fork-from-step | ❌ | ✅ | DBOS / LangGraph |
| Streaming: usage, tool deltas, cancel | ❌ broken | ✅ | AI SDK v5 / Eino |
| Frontend wire protocol (useChat-compatible) | ❌ hand-rolled WS | ✅ v5 SSE | Vercel AI SDK |
| Guardrails (parallel, tripwire) | ❌ | ✅ | OpenAI SDK |
| Handoffs / agents-as-tools | ❌ | ✅ | OpenAI SDK / Eino |
| Sessions + scoped shared state | ❌ | ✅ | OpenAI SDK / ADK |
| Usage limits / cost brakes | ❌ | ✅ | PydanticAI |
| Dev UI playground (single binary) | ❌ | ✅ embed.FS | Genkit / Mastra / EinoDev |
| Prompts as artifacts (.prompt) | ❌ | ✅ | Genkit dotprompt |
| Evals in native test runner + CI + live traces | ❌ shell | ✅ | Braintrust/DeepEval/Mastra |
| LLM-free TestModel/ScriptModel | partial (mock) | ✅ | PydanticAI |
| OTel GenAI semconv | ❌ agk.* | ✅ | PydanticAI / OpenAI SDK |
| MCP client (official SDK, OAuth, streamable) | ❌ fork | ✅ | MCP go-sdk |
| Agent → MCP server export + registry | ❌ | ✅ | (open field) |
| A2A | ❌ | ✅ via a2a-go | ADK |
| Memory extraction pipeline + tiers | ❌ storage only | ✅ | mem0 / Letta |
| Code-as-action sandbox (native) | ❌ | ✅ Starlark/wazero | smolagents (idea) |
| Race-clean guarantee | ❌ fails | ✅ CI-enforced | (Go table stakes) |

---

## 5. Part IV — The ecosystem: agk CLI + template registry

The framework does not ship alone: the `AgenticGoKit` org carries **agk** (developer CLI, ~78 files) and **agentic-examples**, and the **agk-templates** org carries a template registry plus community templates. Evaluated from the public repos (July 2026):

### 5.1 What exists — and it's more than the framework docs let on

| Asset | State |
|---|---|
| `agk init` scaffolding | Embedded quickstart + workflow templates (Go text/template + Sprig), variables for provider/model, post-create hooks |
| Template registry | `agk template add/remove/list`, go-git fetcher, local cache, `agk-template.toml` manifest contract (variables, include/exclude, `min_agk_version`, hooks); registry = `agk-templates/registry/index.json` |
| `agk trace list/show/view/mermaid` | A substantial **bubbletea TUI trace explorer** (52KB viewer + span tree) over `.agk/runs`, plus Mermaid export — a genuine differentiator no other Go framework ships |
| `agk eval` | YAML test files → HTTP calls against the v1beta eval server (`AGK_EVAL_MODE`, `:8787`, `/invoke`, `/traces/{id}`), with **embedding, LLM-judge, and hybrid matchers**, thresholded confidence scores, markdown reports with progress bars and trace links, documented CI recipe |
| Community templates | `test-agent`, `translate-workflow`, `flight-search-assistant`; templates already keep prompts as separate files (`prompts/*.system.txt`) — convergent with the dotprompt direction (E3) |

This resolves an earlier finding: the eval server has no in-repo consumers (C10) because **its consumer lives in the agk repo**. The two are halves of one product, split across repositories.

### 5.2 Ecosystem gaps and risks

1. **Version skew by construction.** agk pins `agenticgokit v0.5.5`; every framework fix (including PR #156) reaches CLI users only after a framework release *and* an agk bump. There is no compatibility contract, no cross-repo CI, and `min_agk_version` exists in template manifests but nothing comparable binds agk↔framework.
2. **The CLI is built on the framework's two least stable contracts.** `agk trace` parses `.agk/runs/trace.jsonl` — the OTel stdouttrace *debug* format (C8) that can break on any otel-go upgrade — and inherits the fabricated manifest stats (C9: 0 tokens, $0.00, 0 llm_calls in `agk trace show`). `agk eval`'s trace links point at eval-server traces that are empty shells, its multi-turn sessions don't actually feed history to the agent, and workflow evals report hardcoded success (C10). The CLI's flagship features silently degrade because their upstream contracts were never designed as contracts.
3. **Templates scaffold the framework's known bugs into every new project.** Generated code blank-imports `plugins/llm/<provider>` — a **no-op for v1beta** (the dual-registry split, #141); `translate-workflow` imports three provider plugins it doesn't need. No template imports `plugins/embedding`, so any memory-enabled scaffold reproduced #137 pre-fix. The scaffolder is a bug amplifier: whatever the templates get wrong, every new user starts wrong.
4. **The registry index is a bare name→repo map.** No versions, no checksums/signatures, no descriptions/categories, no enforcement of `min_agk_version` at resolve time; `agk template add` fetches a mutable default branch — unpinned, unreproducible, and a soft supply-chain risk the moment templates execute `post_create` hooks.
5. **Three homes for examples** (framework `examples/`, `agentic-examples`, templates) with no shared CI — drift among them is already visible in the framework's own examples.
6. **Residual fork confusion:** `AgenticGoKit/mcp-navigator-go` exists in the org, but the framework's go.mod still depends on the personal `kunalkushwaha/mcp-navigator-go` (superseded by #145 either way).

### 5.3 Ecosystem roadmap items

| # | Recommendation | Ties to |
|---|---|---|
| G1 | **A versioned contract between framework and CLI**: a stable, documented trace schema (fixes C8) and an OpenAPI-specified eval-server API, in a small shared module both repos import; nightly cross-repo CI (agk against framework master). | v0.6 |
| G2 | **Fix the scaffolds**: remove no-op plugin imports (or make them real via #141), single provider per scaffold, embedding config for memory templates, generated `AGENTS.md` (E6) + a starter eval YAML in every template. | v0.6 |
| G3 | **Registry v2**: index entries carry version, commit pin/checksum, description, `min_agk_version` (enforced at `agk init`); a published **template conformance action** (render × providers → `go build` → `go vet`) required for registry PRs. | v0.7 |
| G4 | **`agk dev` absorbs the dev-loop pillar** (E1): the playground/launcher work should land in agk, upgrading the existing TUI investment — TUI for terminals, embedded web UI for browsers, one reflection protocol underneath. | v0.9 |
| G5 | **Unify the eval stack**: the matchers (embedding/LLM-judge/hybrid) move into a framework `agkeval` package consumed by both `go test` (E4) and `agk eval` HTTP mode — one scorer implementation, two frontends; wire real trace IDs and honest workflow results (C10) underneath. | v0.9 |
| G6 | **Distribution polish**: released binaries via the existing goreleaser (brew tap/scoop), `agk version --check` compatibility warnings against the project's framework version. | v0.7 |

**Strategic read:** the ecosystem's *shape* is exactly right — CLI + registry + templates is what Genkit/Mastra/ADK all converged on — and the TUI trace viewer plus LLM-judge eval matchers are ahead of most Go competition. What's missing is *contracts*: the CLI consumes undocumented internals across a repo boundary, and the scaffolder amplifies framework bugs. Contract-first is cheaper than any new feature and multiplies the value of everything in Parts I–III.

---

## 6. What NOT to build (negative space from the research)

1. **Don't port Python literally** — LangChainGo's `Chain.Call(ctx, map[string]any)` shape is the cautionary tale: pre-generics design, huge in-tree integration surface, single-maintainer burnout at v0.1.x. Keep interfaces tiny and contracts conformance-tested so the community owns backends.
2. **Don't adopt Temporal's programming model** — event-history replay imposes determinism constraints (no goroutines/`time.Now`/map iteration) that would poison Go DX. Journal-skip gives 90% of the value with none of the constraints; offer a Temporal *adapter* for enterprises instead.
3. **Don't expose four streaming paradigms** — Eino's own docs concede the concept load. Public API: `Invoke` and `Stream`. Period.
4. **Don't build a proprietary wire format** — speak AI SDK v5 SSE and MCP; the frontends and clients already exist.
5. **Don't ship features as config-only surface** — the current codebase's core credibility problem (decorative retry/rate-limit/cache/metrics knobs). A knob that does nothing is worse than no knob: it's a lie a customer discovers in production. Enforce or delete.
6. **Don't hide failures to look friendly** — the zero-vector embedding incident (#137) generalizes: every silent fallback found in Part I converts a user's debugging session into distrust. Fail loudly, return diagnostics as values (the `Diagnostic`/sentinel-error pattern now in v1beta), and make "no fabricated data" a review gate.

---

*Sources: 10-agent research run (4 codebase, 6 landscape, 936k tokens, all primary sources cited inline in the underlying findings). Repository findings reference the PR #156 branch (`claude/go-agnostic-framework-analysis-ekyl3a`).*
