# agentkit — issue backlog (ready to file)

Each issue below is delimited by a line of `===`, with YAML front matter carrying the fields the GitHub API needs (`title`, `type`, `milestone`, `labels`) followed by the markdown body. This file is the source of record until the issues are filed; `scripts/file-issues` (or a single API pass) can consume it verbatim.

Ordering reflects the intended milestone sequence. Design references point at `docs/DESIGN.md` in this repo and the analysis set in [AgenticGoKit](https://github.com/AgenticGoKit/AgenticGoKit) (`docs/CONSOLIDATED_REPORT.md`, `docs/DEVELOPER_FIRST_ROADMAP.md`).

===
---
title: "Epic: v0.1 — lock the core contracts"
type: Task
milestone: "v0.1 Core contracts"
labels: [epic]
---
Umbrella for the work that makes agentkit usable end to end against real models, and freezes the contracts everything else builds on.

The contracts in `docs/DESIGN.md` (D1–D12) are deliberate and evidence-backed; they should not change casually after v0.1 because adapters, workflows, and durability all encode them.

**In scope**
- [ ] Streaming agent runs (`Agent.Stream`)
- [ ] Reference provider adapter: OpenAI (proves the contract end to end)
- [ ] Anthropic, Ollama, Gemini adapters
- [ ] Generic `openai-compatible` adapter
- [ ] Provider conformance test suite
- [ ] Structured output: native and prompted modes
- [ ] Parallel tool execution
- [ ] Provider middleware (retry / rate limit / usage accounting)
- [ ] OTel GenAI instrumentation

**Definition of done**
`go test -race ./...` green across core and every adapter module; a documented quickstart that runs against OpenAI, Anthropic, and Ollama unchanged except for two strings; zero third-party dependencies in the core module.

===
---
title: "Streaming agent runs: Agent.Stream with agent-level events"
type: Feature
milestone: "v0.1 Core contracts"
labels: [core, streaming]
---
`Provider.Stream` and the `Stream` cursor exist; the agent has no streaming entry point yet.

**Scope**
- `func (a *Agent[D, O]) Stream(ctx, input string, deps D) (RunStream, error)`.
- Agent-level events, not raw provider passthrough: `StepStart`, `TextDelta`, `ReasoningDelta`, `ToolCallStart/Delta/Done`, `ToolResult`, `StepEnd`, `Done{Output O, Usage}`.
- The tool loop runs inside the stream: tool calls execute and their results are emitted, then generation continues.
- Terminal `Done` carries the typed output for structured runs.

**Requirements that are non-negotiable** (each corresponds to a shipped bug in the previous generation, see `DEVELOPER_FIRST_ROADMAP.md` §1.2):
- Cancelling the returned stream must cancel the in-flight provider request. A cancel that keeps generating tokens costs real money.
- Every producer failure must be observable through `Err()`. A failed stream must never be indistinguishable from a successful one.
- Terminal usage must be populated, so a streaming run reports the same accounting as a non-streaming one.
- No protocol text leaks: if a fallback strategy encodes tool calls in text, that text must not reach the caller as content.

**Acceptance**
Tests using `agenttest` covering: mid-stream cancellation stops the provider; a provider error surfaces from `Err()` and not as a silent end; usage is present on `Done`; a tool call executes and its result appears before subsequent text.

===
---
title: "Provider conformance test suite"
type: Feature
milestone: "v0.1 Core contracts"
labels: [core, testing]
---
Small interfaces attract third-party implementations only when there is an objective definition of "correct". Ship one.

**Scope**
- New package `providertest` exporting `providertest.Run(t *testing.T, newProvider func() agentkit.Provider, opts ...)`.
- Verifies: multi-turn message round-tripping; system message handling; tool definitions reaching the model and tool results being accepted back; explicit temperature 0 is honored; usage is populated on both paths; streaming emits deltas then exactly one terminal event; errors wrap `ErrProvider`; `Capabilities` is self-consistent (claims native tools ⇒ tool calls actually return `ToolCall` parts).
- Opt-out flags for genuinely unsupported features, so an adapter declares gaps instead of failing.
- Runs against `agenttest` in CI and against live providers behind a build tag.

**Why**
LangGraph's checkpointer ecosystem exists because the contract was tiny *and* conformance-tested. The same lever applies here.

===
---
title: "OpenAI provider adapter (reference implementation)"
type: Feature
milestone: "v0.1 Core contracts"
labels: [provider]
---
First real adapter; doubles as the worked example for every other one.

**Scope**
- Separate module `github.com/agenticgokit/agentkit/providers/openai` (core stays dependency-free, D12).
- Full message IR translation including images, tool calls, and tool results.
- Native tool calling and parallel tool calls.
- Native structured output (`response_format` with JSON schema) wired to `OutputNative`.
- Streaming with tool-call argument deltas and `stream_options.include_usage` so terminal usage is real.
- `Capabilities` per model, including context window.
- Prompt-caching token reporting mapped onto `Usage.CachedInputTokens`.
- Passes `providertest`.

**Notes**
Base URL must be configurable — this adapter is also the substrate for the `openai-compatible` provider.

===
---
title: "Anthropic provider adapter"
type: Feature
milestone: "v0.1 Core contracts"
labels: [provider]
---
**Scope**
- Module `providers/anthropic`.
- System prompt as a top-level parameter, not a message.
- `tool_use` / `tool_result` block translation, including the requirement that assistant tool-use blocks are echoed back verbatim on subsequent turns.
- Extended thinking mapped to `Reasoning` parts, preserving `Signature` for turns that require it.
- Prompt caching (`cache_control`) surfaced through `Usage.CachedInputTokens`.
- No `Embedder` implementation — Anthropic has no embeddings API, and per D3 that is expressed by *not implementing the interface* rather than by a stub.
- Passes `providertest`.

===
---
title: "Ollama provider adapter"
type: Feature
milestone: "v0.1 Core contracts"
labels: [provider]
---
**Scope**
- Module `providers/ollama`.
- Chat API with native tool calling; streaming tool calls where the server supports them.
- `format` field wired to `OutputNative` for JSON-schema-constrained output.
- Implements `Embedder` (`/api/embed`) with dimensions derived from the model.
- Actionable errors: a missing model must tell the user to `ollama pull <model>`; an unreachable host must name the URL that was tried and suggest `ollama serve`.
- Passes `providertest`.

**Why the error requirement is in scope**
Local-first is the first-five-minutes experience for most Go developers; a cryptic connection error there is a lost adopter.

===
---
title: "Google Gemini provider adapter"
type: Feature
milestone: "v0.1 Core contracts"
labels: [provider]
---
**Scope**
- Module `providers/gemini`, speaking the native API (not an OpenAI compatibility shim).
- `contents`/`parts` translation, function calling, `responseSchema` for structured output, multimodal input, and `Embedder` via the embedding models.
- Long-context reporting in `Capabilities.ContextWindow`.
- Passes `providertest`.

**Why**
Gemini is the largest provider with no adapter in the ecosystem today, and its multimodal and context-window strengths are exactly what an agent framework should be able to reach without a proxy.

===
---
title: "Generic openai-compatible provider"
type: Feature
milestone: "v0.1 Core contracts"
labels: [provider]
---
A first-class `openai-compatible` provider that takes a required base URL and speaks the OpenAI wire format.

**Why this is the highest-leverage adapter**
Most serving stacks and aggregators (vLLM, Together, Groq, Mistral, OpenRouter, LM Studio, llama.cpp, BentoML, MLflow gateways) expose OpenAI-compatible endpoints. One adapter covers all of them with zero new code per vendor, and it means a new vendor never blocks a user waiting for framework support.

**Scope**
- Explicit capability declaration, since compatibility varies: the caller states whether the endpoint supports tools/JSON mode, with conservative defaults and a `Diagnostic` when a requested feature is not declared.
- Documented presets for common backends.
- Clear naming: `openai` means api.openai.com semantics; `openai-compatible` means bring-your-own endpoint.

===
---
title: "Structured output: native and prompted modes"
type: Feature
milestone: "v0.1 Core contracts"
labels: [core]
---
`OutputTool` is implemented. `OutputNative` and `OutputPrompted` are declared but not honored end to end.

**Scope**
- `OutputNative`: pass the schema through `Request.Output` and parse the model's JSON response directly.
- `OutputPrompted`: inject the schema into instructions, extract JSON from the reply defensively (fenced blocks, leading prose), and validate.
- Validation failures feed the existing self-correction loop (D8) with a bounded budget rather than failing the run.
- `WithOutputRetries(n)` to bound correction attempts.

**Acceptance**
Each mode covered by `agenttest` scripts, including a malformed-response path that recovers and one that exhausts the budget and returns `ErrNoOutput`.

===
---
title: "Parallel tool execution with bounded concurrency"
type: Feature
milestone: "v0.1 Core contracts"
labels: [core, performance]
---
When a model requests several independent tool calls in one turn, run them concurrently.

**Scope**
- `errgroup` fan-out bounded by a configurable limit (`WithMaxParallelTools(n)`, default e.g. 5).
- Results appended in the model's original call order regardless of completion order, so transcripts stay deterministic.
- First non-retryable tool error cancels siblings and aborts; retryable errors are collected per call.
- Respect `Capabilities.ParallelTools`: when the provider cannot express parallel calls, this is simply never exercised.

**Why**
Three independent fetches currently serialize, tripling latency. This is precisely where Go should beat Python frameworks, and it is ~30 lines behind a clean tool interface.

===
---
title: "Provider middleware: retry, rate limit, timeout, usage accounting"
type: Feature
milestone: "v0.1 Core contracts"
labels: [core, reliability]
---
Cross-cutting reliability written once as `Provider` decorators, never re-implemented per adapter.

**Scope**
- `middleware.Retry(policy)` — exponential backoff with jitter, honoring `Retry-After`, retrying only classified-retryable errors.
- `middleware.RateLimit(limiter)` — request and token budgets.
- `middleware.Timeout(d)` — per-call deadline.
- `middleware.Usage(recorder)` — per provider/model token and cost rollup, with user-supplied pricing (never hardcoded).
- `middleware.Chain(base, ...)`.
- Error classification (`rate_limited`, `overloaded`, `invalid_request`, `auth`, `timeout`) normalized across adapters, so retry policy is principled rather than string-matched.

**Why in the core milestone**
The previous generation shipped retry/rate-limit/circuit-breaker *configuration* that no code read. Building it as middleware from the start means the knobs exist only where behavior exists.

===
---
title: "Provider fallback chain"
type: Feature
milestone: "v0.1 Core contracts"
labels: [provider, reliability]
---
`fallback.New(primary, secondary, ...)` implementing `Provider`, trying each in order.

**Scope**
- Trigger on classified API errors by default; `FallbackOn(func(err error, resp *Response) bool)` to also trigger on response content (e.g. `StopFiltered`, `StopMaxTokens`).
- Per-provider model and params preserved.
- All failures joined with `errors.Join` so nothing is swallowed.
- A `Diagnostic` when a fallback is used, so silent degradation to a weaker model is visible.

**Effort/value**
Pure decorator over an existing interface; makes provider resilience a one-liner.

===
---
title: "OpenTelemetry instrumentation using GenAI semantic conventions"
type: Feature
milestone: "v0.1 Core contracts"
labels: [core, observability]
---
Instrument runs with OTel spans following the **GenAI semantic conventions** (`gen_ai.*`), in a separate module so the core stays dependency-free.

**Scope**
- One CLIENT span per model call (`gen_ai.operation.name`, `gen_ai.system`, `gen_ai.request.model`, `gen_ai.usage.input_tokens`, `gen_ai.usage.output_tokens`, `gen_ai.response.finish_reasons`), one span per tool call, one parent span per run.
- Opt-in content capture (prompts/responses) with a redaction hook.
- Standard `OTEL_*` environment variables honored; endpoint configured with `WithEndpointURL` (a scheme-carrying URL passed to `WithEndpoint` silently produces a malformed target — the exact bug in the previous generation).
- Header support so hosted backends are reachable.

**Why the convention matters**
Bespoke attribute names are invisible to every LLM observability backend. Speaking `gen_ai.*` means Langfuse, Braintrust, Datadog, Grafana, and Jaeger LLM views work with no adapter.

===
---
title: "Embedder implementations and a store-facing contract"
type: Feature
milestone: "v0.1 Core contracts"
labels: [provider, memory]
---
`Embedder` is defined (D3) but has no implementations.

**Scope**
- Implementations alongside their providers: OpenAI, Gemini, Ollama, plus a deterministic `agenttest` fake.
- Every implementation reports true `Dimensions()`.
- A helper that fails loudly when a configured store dimension disagrees with the embedder's — a mismatch silently corrupts a vector index, and detecting it at construction is cheap.
- Explicitly no dummy/zero-vector implementation outside `agenttest`.

===
---
title: "Epic: v0.2 — durable execution"
type: Task
milestone: "v0.2 Durable"
labels: [epic, durability]
---
The strategic differentiator: nothing in the Go agent ecosystem offers durable agent runs, and Go is the language of durable execution (Temporal, DBOS, Restate, River all ship first-class Go SDKs).

**Architecture decision to confirm first (see `docs/DESIGN.md` open question 1):** journal-skip, not event-history replay. On recovery the function is re-invoked from the top and completed steps return their journaled results. This imposes *no* determinism constraints on user code — plain goroutines, `time.Now`, and map iteration stay legal outside steps — which is why Temporal's model is explicitly not being copied.

**In scope**
- [ ] Journal + `Checkpointer` contract, SQLite-backed
- [ ] LLM and tool calls auto-wrapped as steps
- [ ] `Interrupt`/`Resume` for human-in-the-loop
- [ ] Durable sleep and idempotency keys
- [ ] Fork-from-step

===
---
title: "Journal-skip durable execution engine over an embedded SQL journal"
type: Feature
milestone: "v0.2 Durable"
labels: [durability]
---
**Scope**
- Two tables — run status and step journal — on SQLite by default (a separate module; core stays dependency-free), Postgres as an alternative backend.
- `Step[T](ctx, name string, fn func(context.Context) (T, error)) (T, error)`: the closure is where non-determinism lives; the result is journaled by `(runID, stepSeq)` and replayed on recovery. Three independent Go SDKs converged on this signature, which is strong evidence it is the right idiom.
- `Checkpointer` interface kept deliberately small (put / put-writes / get / list / delete-thread), addressed by `(threadID, namespace, checkpointID)`, with pending-writes so successful parallel siblings are not re-executed after a partial failure.
- Per-run durability mode: `Exit` (checkpoint at end), `Async` (background writer, default), `Sync` (write before proceeding).
- Recovery scans for orphaned runs and resumes them.

**Why a library, not a server**
DBOS demonstrated a complete durable-execution engine as a plain library over SQL. Zero-infrastructure durability is a headline differentiator; requiring an orchestrator server would forfeit it.

===
---
title: "Auto-wrap model and tool calls as durable steps"
type: Feature
milestone: "v0.2 Durable"
labels: [durability]
---
Durability must be a builder option, not a different agent type.

**Scope**
- `agentkit.WithDurability(store)` — every model call, tool call, and (later) MCP round-trip becomes a journaled step keyed by agent name and call sequence.
- The agent API is otherwise unchanged; the same code runs durably and non-durably.
- Only serializable values cross a step boundary; a clear error names the offending value rather than failing at deserialize time.

**Reference**
PydanticAI's Temporal integration proved the ergonomic: users wrap nothing by hand.

===
---
title: "Interrupt and Resume: durable human-in-the-loop"
type: Feature
milestone: "v0.2 Durable"
labels: [durability, hitl]
---
Approval gates are the most-requested production agent feature and cannot be bolted on after the fact.

**Scope**
- `Interrupt[T](ctx, payload any) (T, error)` — suspends the run durably; the payload surfaces to the caller; `Resume(ctx, runID, value)` re-invokes the run, and the interrupt returns the recorded value.
- Node re-execution semantics (the code before the interrupt runs again on resume) so no stack serialization is required — this is what makes the pattern implementable in Go at all.
- A tool-level convenience: `WithToolApproval(func(ctx, ToolCall) (Decision, error))` gating on `ToolAnnotations.Destructive`.
- Emit approval requests as ordinary events so any frontend can render them without framework-specific glue.

===
---
title: "Durable timers, idempotency keys, and honest exactly-once semantics"
type: Feature
milestone: "v0.2 Durable"
labels: [durability]
---
**Scope**
- `Sleep(ctx, d)` that journals a wake deadline and survives restarts; cron scheduling for long-horizon agents.
- A deterministic idempotency key per step attempt `(runID, stepSeq)`, surfaced to tools via context so they can pass it to APIs that deduplicate.
- Documentation stating the guarantee precisely: step *results* are recorded exactly once, but a crash mid-step re-executes the body, so external side effects are at-least-once unless the callee deduplicates.
- For database side effects, a transactional path where the write and the journal entry commit together.

**Why the honesty matters**
Every engine has this property; frameworks that imply otherwise cause data corruption in production. State it in the API docs, not a footnote.

===
---
title: "Fork-from-step and run inspection"
type: Feature
milestone: "v0.2 Durable"
labels: [durability, devex]
---
The journal is a debugger, not only a recovery mechanism.

**Scope**
- `ForkRun(ctx, runID string, fromStep int)` — start a copy of a run reusing journaled results before the fork point.
- `ListRuns` / `GetRunSteps` for inspection.
- Documented state/topology migration rules for runs that outlive a deploy, plus a boolean `Patch(ctx, "change-id")` for versioning in-flight runs (DBOS's distillation of Temporal's `GetVersion`, which is famously confusing).

**Why it is disproportionately valuable here**
Agent steps cost real money. "The agent derailed at tool call 7 — fork from 6 with a corrected prompt" avoids re-paying for steps 1–6, which no print-statement workflow can offer.

===
---
title: "Epic: v0.3 — typed workflows"
type: Task
milestone: "v0.3 Workflows"
labels: [epic, workflow]
---
Multi-step orchestration where step boundaries are type-checked at build time.

**In scope**
- [ ] `Step[In, Out]` with chainable composition
- [ ] Dynamic fan-out (map-reduce over runtime collections)
- [ ] Per-step policies: retry, timeout, cache, fallback
- [ ] Graph validation and export before execution

**Design constraint from prior art**
Eino proves build-time type alignment works in Go; its own documentation concedes the concept load (four streaming paradigms, boxing rules) is a cost. Keep the public surface to `Invoke` and `Stream`, and let composition be chainable rather than error-returning at every `AddNode`.

===
---
title: "Typed workflow steps with compile-time composition"
type: Feature
milestone: "v0.3 Workflows"
labels: [workflow]
---
**Scope**
- `Step[In, Out]` where an agent, a function, or a nested workflow can all be steps.
- Chainable builder: `.Then`, `.Parallel`, `.Branch`, `.Loop`, `.Map`, terminating in a compile step that surfaces topology errors as values.
- The output type of step N is checked against the input type of step N+1.
- No stringly-typed data flow: fan-in produces a typed struct with named fields per dependency, so a step always knows which upstream produced what — the concatenate-with-newlines approach loses that information irrecoverably.
- `Validate()` (cycles, unknown dependencies) and `Mermaid()`/`DOT()` export before any token is spent.
- Identical semantics between the streaming and non-streaming entry points, enforced by a shared executor and tests that assert equivalence.

===
---
title: "Dynamic fan-out: map-reduce over runtime collections"
type: Feature
milestone: "v0.3 Workflows"
labels: [workflow]
---
The orchestrator-worker pattern — step A produces N items, each processed in parallel, results reduced — is currently impossible with a statically declared step list.

**Scope**
- A `Send`-style primitive: a router returns `[]Send[W]`, each invoking the target step with its own typed state; a reducer merges results on fan-in.
- Bounded concurrency and per-branch error policy.
- Streaming passthrough with branch attribution on every event.

**Why Go wins here**
This is `errgroup` plus a semaphore behind a typed API — a marquee demonstration that Go's concurrency is a real advantage over Python frameworks emulating it.

===
---
title: "Per-step policies: retry, timeout, cache, fallback"
type: Feature
milestone: "v0.3 Workflows"
labels: [workflow, reliability]
---
**Scope**
- `Retry(policy)`, `Timeout(d)`, `OnError(Skip|Fail|Fallback(step))`, `Cache(ttl, keyFn)` attached per step.
- Parallel branches: fail-fast cancellation of siblings, and `errors.Join` so *every* failure is reported rather than the first one.
- Cached step results marked in metadata so a run is never mistaken for fresh.

**Why cache is in scope early**
The inner development loop is where a framework is judged: not re-paying for an expensive step while iterating on the one after it is a small feature with outsized delight.

===
---
title: "Epic: v0.4 — multi-agent primitives"
type: Task
milestone: "v0.4 Multi-agent"
labels: [epic, multi-agent]
---
The primitives the ecosystem has converged on, implemented once and composably.

**In scope**
- [ ] Agents as tools (primary composition currency)
- [ ] Handoffs with typed payloads (when control must transfer)
- [ ] Guardrails with tripwires, run concurrently
- [ ] Sessions and scoped shared state

===
---
title: "Agents as tools"
type: Feature
milestone: "v0.4 Multi-agent"
labels: [multi-agent]
---
**Scope**
- `func (a *Agent[D, O]) AsTool(name, description string, opts ...ToolOption) Tool` — the sub-agent's typed output becomes the tool result; its input schema is derived.
- Optional output extractor for post-processing.
- Sub-agent events pass through to the parent's stream with attribution.
- Usage from sub-runs rolls up into the parent's accounting.

**Why this over handoffs as the default**
Eino converged on agent-as-tool and explicitly deprecated transfer-style handoffs; the parent keeps control and the composition currency stays uniform.

===
---
title: "Handoffs with typed payloads and history filters"
type: Feature
milestone: "v0.4 Multi-agent"
labels: [multi-agent]
---
For the triage/specialist pattern where control genuinely transfers.

**Scope**
- `Handoff(target, opts...)` producing an auto-generated `transfer_to_<name>` tool the model can call, carrying a schema-validated typed payload.
- History filter controlling what the receiving agent sees.
- A `Command{Update, Goto}` return shape so state update and routing are one primitive rather than two mechanisms.
- Loop protection: a bounded transfer depth with a clear error.

===
---
title: "Guardrails with tripwires, executed concurrently"
type: Feature
milestone: "v0.4 Multi-agent"
labels: [multi-agent, safety]
---
**Scope**
- Input, output, and tool-level guardrails: functions (often a small fast model) returning `(info, tripwire bool)`.
- Input guardrails run **concurrently with** the main agent; a tripwire cancels the expensive run through context.
- Tripwires produce a typed error identifying which guardrail fired.

**Why Go's version is better than the original**
The Python implementation of "run the cheap screen alongside the expensive call, cancel on trip" is awkward; in Go it is `errgroup` plus context cancellation, and cancellation actually propagates to the HTTP request.

===
---
title: "Sessions and scoped shared state"
type: Feature
milestone: "v0.4 Multi-agent"
labels: [multi-agent, memory]
---
**Scope**
- A deliberately tiny `Session` interface — `Items(ctx, limit)`, `Append(ctx, items)`, `Pop(ctx)`, `Clear(ctx)` — so backends (SQLite, Postgres, Redis) are community-implementable. `Pop` exists to support undo/regenerate flows.
- The runner prepends stored history and persists new turns, removing the manual message-threading boilerplate.
- Scoped state with prefix semantics (`app:`, `user:`, session, `temp:`) so app config, user profile, and scratch space are one mechanism.
- Instruction templating that reads from state.

**Depends on** the message IR (D1); this is the feature that makes multi-turn worth having.

===
---
title: "Epic: v0.5 — dev loop and ecosystem"
type: Task
milestone: "v0.5 Dev loop"
labels: [epic, devex]
---
Adoption is won in the first hour. This milestone is about what a developer sees, runs, and debugs.

**In scope**
- [ ] Action registry (introspection substrate)
- [ ] AI SDK v5 SSE wire format
- [ ] Evals under `go test`
- [ ] MCP client on the official Go SDK
- [ ] Agent → MCP server export
- [ ] `.prompt` files
- [ ] `agk dev` playground integration

===
---
title: "Action registry: one introspectable unit for agents, tools, workflows, prompts"
type: Feature
milestone: "v0.5 Dev loop"
labels: [devex]
---
**Scope**
- A small `Action` abstraction — name, kind, input schema, output schema, run function with an optional stream callback — that agents, tools, workflows, prompts, and evaluators all register as.
- Registry lookup by `kind/name`.

**Why this first in the milestone**
Genkit's dev UI, HTTP serving, and eval runner are all cheap *because* everything is an Action: the UI is list-actions plus run-action. Building the registry before the playground makes the playground small; building it after makes it bespoke.

===
---
title: "uistream: serve agent runs over the AI SDK v5 SSE protocol"
type: Feature
milestone: "v0.5 Dev loop"
labels: [devex, streaming]
---
**Scope**
- A `uistream` module: an `http.Handler` that serves any agent or workflow stream as AI SDK v5 typed SSE parts (`text-start`/`text-delta`/`text-end`, reasoning, `tool-input-start`/`-delta`/`-available`, `tool-output-available`, `start-step`/`finish-step`, `error`, `finish`, plus `data-*`).
- Resumability: a pluggable `EventStore` with `Last-Event-ID` replay.
- Heartbeats and correct flushing.

**Why this format**
It is the de facto frontend contract — any React/Svelte `useChat` frontend works against any backend that emits it, and other frameworks already ship v5-compatible emitters. Shipping it means a Go backend gets a modern chat UI for free, and it retires the hand-rolled WebSocket servers that examples currently duplicate.

===
---
title: "Evals that run under go test"
type: Feature
milestone: "v0.5 Dev loop"
labels: [devex, eval]
---
**Scope**
- `agenteval` module: `Case[In, Out]`, `Dataset`, and `Run(t, dataset, task, scorers...)` executing as an ordinary Go test.
- Scorers: exact/contains, embedding similarity, LLM judge, and custom `Scorer[In, Out]` returning 0..1 with a reason.
- Trials (run each case N times, aggregate) and CI thresholds that fail the build.
- Span-based assertions: "was tool X called before tool Y" against the run transcript, not only the final answer.
- The same scorers usable against sampled production runs, not only offline datasets.

**Positioning**
Evals in the native test runner is the idiomatic Go differentiator — no separate harness, no YAML dialect to learn, `go test ./...` covers correctness *and* quality.

===
---
title: "MCP client on the official modelcontextprotocol/go-sdk"
type: Feature
milestone: "v0.5 Dev loop"
labels: [mcp, ecosystem]
---
**Scope**
- Module `mcpclient` wrapping the official SDK (stable v1.0 with a compatibility guarantee): stdio and streamable HTTP, OAuth, elicitation, sampling, roots, resumable event stores.
- MCP tools exposed as ordinary `Tool` values, so the agent loop is unaware of their origin.
- Per-server configuration including command **args**, env, cwd, and credentials — `npx -y @modelcontextprotocol/server-filesystem /path` must be expressible.
- Namespaced tool names (`server:tool`) with an explicit collision policy; never route an unknown tool to an arbitrary server.
- Persistent sessions per server, not connect-per-call.
- MCP tool annotations mapped onto `ToolAnnotations` so approval gating works on MCP tools.

**Why adopt rather than maintain**
Protocol churn is constant; the official SDK absorbs it, and vendoring a personal fork is a supply-chain and feature-lag risk.

===
---
title: "Export an agent as an MCP server"
type: Feature
milestone: "v0.5 Dev loop"
labels: [mcp, ecosystem]
---
**Scope**
- `agent.AsMCPServer()` returning a server (stdio and streamable HTTP) that exposes the agent — and optionally its tools — to any MCP client.
- Generate `server.json` for the official MCP registry.

**Why**
This is the distribution play: write a Go agent, one command makes it installable in Claude, VS Code, Cursor, and every other MCP-aware client. Combined with the client work, agentkit sits on both sides of the ecosystem's settled interop standard.

===
---
title: "Prompts as versioned artifacts (.prompt files)"
type: Feature
milestone: "v0.5 Dev loop"
labels: [devex]
---
**Scope**
- `.prompt` files: YAML front matter (model, params, typed input/output schema) plus a template body.
- Loaded from a directory or an `embed.FS` so single-binary deploys keep working.
- Variants (`name.variant.prompt`) for A/B testing.
- Typed access: `LookupPrompt[In, Out](name)` returning something executable with compile-time types.

**Why**
It answers "where do prompts live" with a versioned, reviewable artifact instead of string literals scattered through Go files — and community templates already keep prompts in separate files, so the convention is emerging organically.

===
---
title: "Migration guide and compatibility story for AgenticGoKit v1beta users"
type: Task
milestone: "v0.1 Core contracts"
labels: [docs, ecosystem]
---
**Scope**
- A concept-mapping table (v1beta `Builder`/`Config`/`Agent` → agentkit `New[D, O]`/options; `Tool` map-args → typed `NewTool`; `Prompt{System,User}` → `[]Message`).
- Honest statement of what agentkit does **not** yet cover, so nobody migrates into a gap.
- A support policy for AgenticGoKit `v1beta` during the transition.
- Cross-links in both repositories so the existing audience finds this one.

**Why it belongs in the first milestone**
A clean-slate core starts with zero users unless the existing ones are given a path. Writing the guide early also pressure-tests whether the new API genuinely covers the old use cases.
