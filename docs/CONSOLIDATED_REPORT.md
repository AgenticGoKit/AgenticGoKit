# AgenticGoKit — Consolidated Assessment & Roadmap

**Date:** 2026-07-31 · **Branch:** `claude/go-agnostic-framework-analysis-ekyl3a` (PR #156)
**Scope of work behind this report:** full static analysis of the repository; a 10-agent deep-research pass (4 codebase subsystem deep-dives + 6 competitive-landscape studies from primary sources, 146 findings); evaluation of the org ecosystem (`agk` CLI, `agk-templates` registry); 17 GitHub issues filed; 2 critical bugs fixed, tested, and shipped on this branch.

**Companion documents (full detail):**
- `docs/AGNOSTIC_FRAMEWORK_ANALYSIS.md` — LLM/platform-agnosticism analysis with file:line evidence
- `docs/DEVELOPER_FIRST_ROADMAP.md` — full research synthesis: trust-layer bug tables, feature pillars, ecosystem evaluation
- Issues #139–#155 (umbrella: #155) · PR #156 (fixes for #137/#143)

---

## 1. Executive summary

AgenticGoKit has the right ambitions and several genuinely strong assets: a builder-based v1beta API, streaming-first design, four workflow modes plus subworkflows, batteries-included memory, OTel-based observability, and an ecosystem shape (CLI + template registry) that matches what Genkit/Mastra/ADK converged on. The `agk` TUI trace explorer and LLM-judge eval matchers are ahead of most Go competition.

Against the goal — *the best developer-focused, LLM-agnostic, platform-agnostic agentic framework in Go* — there are four gaps, in order of severity:

1. **The LLM abstraction is single-turn and OpenAI-shaped.** No message history (`Prompt{System, User}`); tool calling is mostly prompt-parsed `TOOL_CALL{...}` text; native tool calls exist in only 2 of 10 adapters; the plugin registry serves only the deprecated `core` API while v1beta uses a hardcoded switch.
2. **A trust layer of correctness debt.** Deep re-reading found data races (`go test -race` fails), silent semantic divergences (DAG streams as Sequential; `Transform` applied twice), broken cancellation, error-swallowing streams, fabricated tool results, and an extensive surface of *decorative configuration* — documented knobs (retry, rate-limit, circuit breaker, cache backends, flush intervals) that no code reads.
3. **No durability.** Workflows are in-process only; nothing survives a restart; no checkpoint/resume, no human-in-the-loop suspension.
4. **Ecosystem without contracts.** `agk` pins framework v0.5.5 and consumes undocumented internals (debug-format trace files, an eval server whose traces are empty); scaffolded templates bake known framework bugs into every new project.

None are dead ends. The strategy: **fix the trust layer, then bet on Go's unfair advantages — typed everything, durable by default, and the best dev loop in the ecosystem.**

---

## 2. What was already delivered on this branch (PR #156)

| Item | Status |
|---|---|
| Zero-vector/dummy embedding silent degradation (#137) — factories now always registered in v1beta; fail-loud resolver (`core.NewEmbeddingServiceForConfig`); real embedding models derived from LLM provider (never the chat model); dimensions auto-derived; `NewMemory` no longer swallows errors | ✅ fixed + tested |
| `WithConfig` discarding `NewBuilder` name; `Memory.Enabled=false` ignored (#137 A/B) | ✅ fixed + tested |
| `MaxTokens` silent 150 default; temperature 0 unexpressible; per-run sampling overrides never reaching the provider (#143 quick wins) | ✅ fixed + tested |
| Actionable errors (`ollama pull <model>`, `ollama serve`, API-key hints); dimension-mismatch warning at construction | ✅ shipped |
| Consumer-facing diagnostics: typed `Diagnostic` values via `WithDiagnosticHandler`/`DiagnosticsOf`, typed sentinel errors (`errors.Is`-able) — findings as values, not black-hole logs | ✅ shipped |
| 25 new tests across 7 test files; all touched suites green; no new failures vs master | ✅ |

**Issues filed:** #139–#154 (16 findings) + #155 (umbrella tracker), cross-referenced with pre-existing #121, #137, #138.

---

## 3. LLM- & platform-agnosticism assessment (condensed)

*(Full evidence: `docs/AGNOSTIC_FRAMEWORK_ANALYSIS.md`)*

### Blockers to true LLM-agnosticism
- **Single-turn prompt model** — multi-turn is emulated by string-stuffing prior output into the next `User` string; breaks provider conversations, prompt caching, and correct tool loops. → **#139** (message-based chat IR).
- **Tool calling** — `TOOL_CALL{}` text parsing by default; OpenAI adapter has zero native tool support; streams can't carry tool-call deltas or usage. → **#140**.
- **Two disconnected provider systems** — `plugins/llm/*` register into deprecated `core` (and drop tools/multimodal/BaseURL in translation); v1beta uses a hardcoded factory switch with a per-provider god-config. → **#141, #142** (open registry + capability discovery).
- **Missing providers** — no Gemini, Bedrock, Mistral, Groq; no generic `openai-compatible` type. → **#149**.
- **Parameter semantics** — fixed in part (PR #156); full pointer-based optionals remain in **#143**.

### Blockers to platform-agnosticism
- **One fat go.mod** — every consumer pulls pgx, chromem, Prometheus, websocket, and an Azure SDK used only in a test. → **#144** (multi-module split).
- **MCP on a personal fork** (`kunalkushwaha/mcp-navigator-go v0.0.2`) instead of the official v1.0-stable `modelcontextprotocol/go-sdk`. → **#145**.
- **No durable execution** — no checkpoint/resume anywhere in the 2,270-line workflow engine. → **#147**.
- **Embedded HTTP eval server in the library package**; hardcoded `.agk/runs` relative path. → **#148**.
- **Embeddings welded to `ModelProvider`**; three parallel memory registries; vector stores limited to chromem/pgvector/weaviate. → **#146**.

### API-surface debt
Three coexisting generations (`core` ~9.4k LOC, `core/vnext` ~11k, `v1beta` ~14.5k over `internal` ~35k); **v1beta imports the deprecated `core`**, so "delete core at v1.0" requires internalizing it first (**#150**). `Result`/`RunOptions`/`ToolManager` bloat → **#154**.

---

## 4. The trust layer: new correctness findings (deep-research pass)

*(Full tables with file:line evidence: `docs/DEVELOPER_FIRST_ROADMAP.md` §1. These are new findings beyond §3, discovered by subsystem re-reads; none yet filed as issues.)*

### Workflow engine
- DAG + `RunStream` **silently executes as Sequential** (stub delegates; dependencies ignored).
- `Step.Transform` **applied twice** in every streaming path.
- Workflows are **single-use and racy**: shared mutable `WorkflowContext` across runs; unlocked writes under Parallel mode.
- **No per-step retry/timeout/failure strategy**; parallel errors collapse to `errors[0]`; DAG runs ready steps *serially* and reports cycles only mid-run as "deadlock".
- Stringly-typed data flow: fan-in is a lossy `"\n"`-join; step errors flattened to strings (breaks `errors.Is/As`).
- Library **prints to stdout and uses `os.Setenv` as IPC** between workflow and agents (races across concurrent workflows).

### Streaming
- **`Stream.Cancel()` never cancels the LLM request** (tokens and money keep flowing).
- **Producer errors never reach `Wait()`** — failed agent steps are marked `Success=true` and empty output feeds the next step.
- `RunStreamWithOptions` mutates the live agent (races; `ToolMode:"none"` permanently strips tools).
- Parallel workflow streaming is a **data race** on the shared stream writer.
- `Capabilities.Tools`/`Memory` never wired — the entire handler-augmentation library silently no-ops; three incompatible tool-call prompt protocols coexist.
- `StreamChunk.Error` doesn't JSON-serialize; no SSE/WS bridge (examples hand-roll three ~236-line servers); `FlushInterval`/`Config.Streaming` are dead config that docs tell users to tune.

### Tools & MCP
- **Reliability config is decorative**: tool `Timeout`/`MaxRetries`/`RateLimit`/`MaxConcurrent`/`CircuitBreaker` validated, documented, never enforced.
- **`MCPAwareAgent` fabricates successful tool results** when caching is off (never contacts a server).
- Unified MCP transport **dials + handshakes + disconnects per tool call**; 1,800 lines of pooling/LB/retry code never instantiated; health checks pass via a no-op `Connect()`.
- Fallback tool schema `{"input": string}` corrupts native tool calls (framework ships a hack to undo its own fallback).
- stdio MCP servers can't receive args/env; the stdio plugin error messages recommend is an unimplemented stub; unknown tools route to an arbitrary server.
- No parallel tool execution, no approval hooks/dangerous-tool gating, one global auth token for all MCP servers.

### Config, observability, eval
- **Unknown TOML keys silently ignored**; `timeout = 30` parses as **30 nanoseconds** and passes validation.
- TOML provider whitelist **rejects five providers the factory supports**; env expansion is a 12-field whitelist and unset `${VAR}` passes through as a literal.
- Spans use bespoke `agk.*` attributes instead of **OTel GenAI semconv (`gen_ai.*`)** — invisible to LLM observability tooling; OTLP export broken as documented (scheme bug, no headers); trace.jsonl is OTel's *unstable debug format*; `manifest.json` fabricates stats (0 tokens, $0 — #138).
- Eval server: traces are empty shells, sessions never feed history to the agent, workflow success hardcoded `true`, binds all interfaces with no auth.

**Verdict:** the "Trustworthy" release (§7, v0.6) exists to zero this list. Brand promise: `-race` clean, every documented knob enforced or deleted, `errors.Is` end-to-end, no fabricated data anywhere.

---

## 5. Feature roadmap: six pillars (landscape-informed)

*(Full design sketches and sources: `docs/DEVELOPER_FIRST_ROADMAP.md` §2. Landscape studied: LangGraph, PydanticAI, OpenAI Agents SDK, Genkit Go, ADK-go, Eino, Temporal/Restate/DBOS/Inngest/River, MCP go-sdk, Vercel AI SDK v5, Mastra, mem0/Letta, smolagents.)*

**A. Typed everything** — Go's headline advantage:
`Agent[Deps, Output]` generics with typed DI (`RunContext[D]`) and test overrides (PydanticAI); `tools.New[In, Out]` with schemas inferred from struct tags, input *and* output validated (Eino/ADK-go/MCP-SDK convergence); structured output declared once with pluggable mode (tool/native/prompted) and a validation-driven retry loop (`ModelRetry`); typed workflow steps with compile-time edge checking, per-step retry/timeout/fallback/cache, `Validate()` + Mermaid export (Mastra/Eino); provider fallback and declarative throttle/debounce as decorators.

**B. Durable by default** — the strategic wedge, nothing in Go agent-land has it:
journal-skip execution (DBOS/Restate model — *no* Temporal determinism constraints) over an embedded **SQLite journal** (two tables, zero infrastructure; Postgres for scale); a tiny 5-method `Checkpointer` contract with a conformance suite (LangGraph's winning move); LLM/tool/MCP calls auto-wrapped as durable steps (`agk.Durable(store)`, per-run `Durability(Exit|Async|Sync)`); `Interrupt`/`Resume` for human approval; awakeables, durable sleep/cron; **fork-from-step time travel** ("agent derailed at tool call 7 — fork from 6 without re-paying for 1–6"); deterministic idempotency keys surfaced to tools; a `workflowcheck`-style `go vet` analyzer.

**C. Streaming as a product**:
fix semantics first (Part 4); then speak the **Vercel AI SDK v5 UIMessage SSE protocol** natively (`uistream` http.Handler — every `useChat` frontend works instantly; PydanticAI and Mastra already emit it); resumable streams via a pluggable `EventStore`; stream-aware routing (decide tool-vs-answer from the first frame); typed custom progress events from any tool/step.

**D. Multi-agent primitives the industry converged on**:
agents-as-tools as the primary currency (Eino deprecated transfer-handoffs); handoffs with typed payloads + history filters where control must move (OpenAI); dynamic fan-out `Send` for map-reduce (LangGraph — Go's errgroup makes it a marquee feature); guardrails with tripwires raced in parallel via context cancellation; 4-method `Session` interface + ADK scope-prefixed shared state (`app:`/`user:`/`temp:`) with `OutputKey`; hard `UsageLimits` and `stop_on_first_tool`.

**E. The dev loop** — adoption is won here:
`agk dev` single-binary playground (embed.FS web UI + the existing TUI, one reflection protocol; Genkit/Mastra's most-loved feature); an `Action` registry as the substrate (agents/workflows/tools/prompts/evaluators all introspectable — Genkit's architectural insight); `.prompt` files (dotprompt) for versioned prompts with typed schemas; **evals under `go test`** (data + task + scorers, trials, LLM-judge, span-based assertions, capture-live-session→eval-case loop); LLM-free `TestModel`/`ScriptModel` + `BlockRealRequests()` CI guard; scaffold `AGENTS.md` and publish `llms.txt` (your users' coding agents are your docs' main readers).

**F. Ecosystem & interop**:
official MCP go-sdk both directions — client (replaces fork, fixes transports/auth) and `agent.AsMCPServer()` + `agk mcp publish` to the official registry; OTel GenAI semconv + honest cost rollups; memory with extraction semantics (mem0-style pipeline + Letta-style tiers) on the existing storage layer; A2A via `a2a-go`; opt-in **code-as-action** sandbox (Starlark/wazero — natively safe in Go, impossible for Python without external sandboxes).

---

## 6. Ecosystem: agk CLI + template registry

*(Full detail: `docs/DEVELOPER_FIRST_ROADMAP.md` §5.)*

**Assets** — stronger than the framework docs advertise: `agk init` scaffolding with a real template manifest contract (`agk-template.toml`); template registry with fetcher/cache; a substantial **bubbletea TUI trace explorer** + Mermaid export; `agk eval` with embedding/LLM-judge/hybrid matchers, markdown reports, CI recipe. The eval server's missing consumer turned out to live here.

**Risks:**
1. **Version skew by construction** — agk pins framework v0.5.5; no compatibility contract; no cross-repo CI.
2. **CLI stands on the two least-stable framework contracts** — the debug-format trace.jsonl and the empty-shell eval traces; `agk trace show` inherits fabricated 0-token/$0 stats.
3. **Scaffolder amplifies framework bugs** — generated projects carry no-op plugin imports (v1beta ignores the plugin registry) and no embedding config (pre-fix, that reproduced #137 in every memory scaffold).
4. **Registry index is a bare name→repo map** — no versions/pins/checksums, `min_agk_version` unenforced, templates fetch mutable branches and can run `post_create` hooks (soft supply-chain risk).

**Recommendations (G1–G6):** a versioned framework↔CLI contract (stable trace schema + OpenAPI eval spec in a shared module, nightly cross-repo CI); fix the scaffolds (real imports, embedding config, generated `AGENTS.md`, starter eval file); registry v2 (versions, pins, enforced `min_agk_version`, published template-conformance CI action); `agk dev` absorbs the dev-loop pillar; unify eval matchers into a framework package with two frontends; distribution polish (released binaries, compat check).

---

## 7. Sequenced release plan

| Release | Theme | Contents |
|---|---|---|
| **v0.6 "Trustworthy"** | Zero known lies | All §4 fixes (`-race` clean CI gate); strict TOML + duration parsing; enforce-or-delete every config knob; honest manifests/health/results; framework↔CLI contract (G1) + scaffold fixes (G2); ship `AGENTS.md`/`llms.txt`. |
| **v0.7 "Typed"** | Compile-time safety | Message IR (#139) → native tools everywhere (#140) → open provider registry + capabilities (#141/#142) → `Agent[D,O]`, `tools.New[In,Out]`, structured output w/ retry, typed steps; testing kit (#152); Gemini/Bedrock/`openai-compatible` (#149); registry v2 (G3). |
| **v0.8 "Durable"** | The wedge | SQLite journal-skip engine, `Checkpointer` contract, auto-wrapped steps, `Interrupt`/`Resume` + approval events, sessions & scoped state, durable sleep/cron; multi-module split (#144). |
| **v0.9 "Delightful"** | Dev loop | Action registry → `agk dev` playground (G4) → AI SDK v5 wire format + resumable SSE → `.prompt` files → evals under `go test` + capture loop (G5) → OTel GenAI semconv. |
| **v1.0 "Complete"** | Ecosystem | Multi-agent primitives (agents-as-tools, handoffs, Send, guardrails, usage limits), MCP server export + registry publish, memory tiers, A2A, fork-from-step; delete `core`/`core/vnext` (#150); API-surface cleanup (#154). |

**Dependency spine:** message IR → typed agents/tools → sessions → durability → dev UI → evals.

---

## 8. Capability checklist (today → target)

| Capability | Today | Reference |
|---|---|---|
| Multi-turn message IR | ❌ | everyone |
| Typed agents / tools / steps | ❌ | PydanticAI / Eino / Mastra |
| Structured output + self-correction | ❌ | PydanticAI |
| Per-step retry/timeout/fallback/cache | ❌ | LangGraph |
| Durable checkpoint/resume (zero infra) | ❌ | DBOS design, LangGraph contract |
| Interrupt / HITL / time travel | ❌ | LangGraph / Mastra / DBOS |
| Streaming: cancel, usage, tool deltas | ❌ broken | AI SDK v5 / Eino |
| useChat-compatible wire protocol | ❌ | Vercel AI SDK v5 |
| Guardrails / handoffs / Send / sessions / usage limits | ❌ | OpenAI SDK / LangGraph / ADK |
| Dev UI playground (single binary) | ❌ (TUI trace viewer ✅ in agk) | Genkit / Mastra |
| Prompts as artifacts | partial (`prompts/*.txt` in templates) | Genkit dotprompt |
| Evals in `go test` + CI + live traces | partial (`agk eval` HTTP-only) | Braintrust / DeepEval / Mastra |
| LLM-free TestModel/ScriptModel | partial (mock provider) | PydanticAI |
| OTel GenAI semconv | ❌ `agk.*` | PydanticAI / OpenAI SDK |
| MCP official SDK client + server export | ❌ fork, client-only | MCP go-sdk |
| A2A | ❌ | ADK |
| Memory extraction pipeline + tiers | ❌ storage only | mem0 / Letta |
| Code-as-action sandbox (native) | ❌ | smolagents (idea), Go-unique execution |
| Template registry + scaffolding | ✅ shape right, contracts missing | (ahead of Go competition) |
| Race-clean guarantee | ❌ fails | Go table stakes |

---

## 9. What NOT to build

1. **No literal Python ports** — LangChainGo's `map[string]any` chains + single-maintainer stall is the cautionary tale. Tiny interfaces, conformance tests, community-owned backends.
2. **No Temporal programming model** — event-replay determinism constraints would poison Go DX; journal-skip gives ~90% of the value. Offer a Temporal *adapter* instead.
3. **No four streaming paradigms** — Eino's concept load is its own documented regret. Public API: `Invoke` and `Stream`.
4. **No proprietary wire formats** — speak AI SDK v5 SSE and MCP; the clients already exist.
5. **No decorative config** — a knob that does nothing is a lie a customer discovers in production. Enforce or delete (CI-checkable).
6. **No silent fallbacks** — #137 generalizes. Fail loudly; return diagnostics as values (the `Diagnostic`/sentinel pattern now in v1beta); make "no fabricated data" a review gate.

---

## 10. Recommended immediate next steps

1. **Merge PR #156** (bug fixes + diagnostics + this report set).
2. **File the §4 trust-layer findings as issues** (~35, mostly small; several are one-line fixes with outsized credibility value: double-Transform, DAG-streams-as-sequential, Cancel, Wait errors).
3. **Start v0.6** with the race fixes and enforce-or-delete pass — it requires no design debate and rebuilds the foundation everything else stands on.
4. **Decide the two strategic bets** (typed core in v0.7, embedded durability in v0.8) — both need early API-shape decisions (`Agent[D,O]` signature; `Checkpointer` contract) since the community will build against them.
