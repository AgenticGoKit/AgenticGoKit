# Scope and boundaries

A decision record for the recurring question: *does X belong, and if so, where?* Written to stop the same debates being re-litigated per feature.

Companion to `docs/CONSOLIDATED_REPORT.md` (assessment) and `docs/DEVELOPER_FIRST_ROADMAP.md` (what to build). This document is about **where things live and what to refuse**.

---

## 1. The layering rule

Four questions, asked in order. The first "yes" decides the home.

| # | Question | Home |
|---|---|---|
| 1 | Does it change the *shape* of user code, or become impossible to add later without breaking everyone? | **`agentkit` core** — contracts only, zero dependencies |
| 2 | Does it need a third-party dependency, or speak an external protocol? | **Official module** (`agentkit/providers/*`, `agentkit/mcp`, `agentkit/a2a`, `agentkit/otel`) |
| 3 | Does it need a human at a terminal or browser? | **`agk` CLI** |
| 4 | Does Go or the existing ecosystem already solve it well? | **Don't build it** |

The fourth question is the one frameworks skip, and it is where most bloat comes from.

---

## 2. A2A: yes, but not in core, and not yet

**Verdict: build it, as `agentkit/a2a`, after durable execution lands (v0.2). Not before.**

### Why it fits the ecosystem

A2A solves agent↔agent interop across organizational boundaries: an Agent Card for discovery, a Task lifecycle (`submitted → working → input-required / auth-required ⇄ → completed / failed / canceled`), named artifacts, SSE streaming, and webhook push for long-running or disconnected work. That is a genuinely different problem from MCP's agent↔tools scope, and it is not expressible as a tool call: a tool call is request/response, while an A2A task can run for hours, come back asking for input, and notify by webhook.

### Why it must not go in the core

**A2A's task lifecycle is a projection of a durable run's state machine, not a separate concept.** A durable run already has: a stable ID, a status, suspension with a typed reason (needs input, needs approval, needs auth), resumption with a value, named outputs, and completion or failure.

Line those up and the mapping is nearly one-to-one:

| Durable run (v0.2) | A2A task |
|---|---|
| `runID` | `taskId` |
| running | `working` |
| `Interrupt` awaiting input | `input-required` |
| `Interrupt` awaiting approval/auth | `auth-required` |
| `Resume(runID, value)` | task update with the input |
| typed output / artifacts | `artifacts` |
| terminal error | `failed` |

Build the interior model first and A2A becomes a **serialization** of it — a small module that maps an existing state machine onto a wire format. Build A2A first and you implement a second state machine, then spend years reconciling the two whenever either changes. This ordering argument is the whole reason to wait.

### Why not immediately

- **MCP-server export covers most of "let another system use my agent" today**, with vastly more client support (Claude, VS Code, Cursor, and every MCP-aware tool). Ship that first — it is in the v0.5 milestone — and A2A becomes an addition rather than the only interop story.
- A2A adoption is real but much thinner than MCP's. Speculative protocol surface in a pre-1.0 core is a maintenance liability.
- The users who need A2A (multi-vendor, cross-org, long-running) are not the users who will judge the framework in its first year.

### When it does land

- Module `agentkit/a2a`, built on the `a2a-go` SDK. **Adopt, do not reimplement** — the same rule applied to MCP.
- Both directions: serve an agent as an A2A endpoint (Agent Card generated from agent metadata and tool definitions), and consume a remote A2A agent as an ordinary sub-agent so the calling code cannot tell local from remote.
- Push notifications map onto the durable journal, so a task that outlives a process still delivers.

---

## 2b. Agentic UI (AG-UI), and the unifying architecture

The same reasoning applied to A2A generalizes, and it resolves the whole interop question at once.

### One interior model, three exterior projections

An agent has exactly three edges, and the industry has standardized one protocol per edge:

| Edge | Protocol | What it carries |
|---|---|---|
| agent ↔ tools/data | **MCP** | capability invocation |
| agent ↔ agent | **A2A** | task delegation across boundaries |
| agent ↔ user interface | **AG-UI** | run events, shared state, user interaction |

These look like three large integrations. They are not: **each is an encoder over the same interior state.** If the runtime owns a run identity, a lifecycle, an event stream, addressable state, and durable suspension, then every protocol module is a projection of that — a few hundred lines, addable or droppable without touching the core. Get the interior wrong and each protocol grows its own state machine, and they diverge forever.

This is the single most consequential architectural decision in the project, and it is a v0.1 decision even though every protocol module ships later.

### The finding that changes the core

**Agentic UI needs something the current core does not have: addressable run state with deltas.**

AG-UI is an event protocol over SSE (~17 event types) whose distinguishing feature is state synchronization: `STATE_SNAPSHOT` carries the full state, `STATE_DELTA` carries incremental updates as **JSON Patch (RFC 6902)** operations. That is what enables the interaction pattern chat streaming cannot express — an agent drafting a document, filling a form, or building a plan that the UI renders live and the user edits in place.

Today `agentkit` has *messages*, not *state*. A message list is append-only prose; it cannot express "field 3 of the draft changed." Adding state later means either breaking the run API or bolting a parallel state channel beside it.

**Recommendation:** give the run a typed state object with delta emission, decided before v0.1 freezes. **Do not add a third type parameter to `Agent`** — `Agent[D, S, O]` would tax every call site for a feature most agents don't use. State belongs to the *run/session*, reachable through the run context and snapshotted by the checkpointer, so `Agent[D, O]` stays intact and stateful runs opt in.

### The second finding: four features are one primitive

These all reduce to *durable suspension with a typed reason*:

| Surface | Suspension reason |
|---|---|
| Human approval before a destructive tool | needs approval |
| **Frontend tools** (executed in the browser, not the server) | needs client execution |
| A2A `input-required` / `auth-required` | needs input |
| AG-UI human-in-the-loop | needs input |

A browser-executed tool is not a new concept — it is a tool call that suspends the run until an external party returns a result, which is exactly `Interrupt`/`Resume`. Build the primitive once and all four fall out.

**Consequence for the roadmap:** durable execution (v0.2) is not merely a reliability feature. It is the keystone the entire interop story rests on — A2A, agentic UI, HITL, and frontend tools are all blocked behind it. That justifies its position ahead of workflows and multi-agent primitives more strongly than the reliability argument alone did.

### Which wire format to emit

AI SDK v5 and AG-UI are not competitors at the same layer, and the choice is not exclusive — both are encoders over one internal event stream.

| | AI SDK v5 protocol | AG-UI |
|---|---|---|
| Reach | Very large — any `useChat` frontend | Growing, framework-neutral |
| Shape | Chat-shaped | Agent-shaped: shared state, frontend tools, HITL, generative UI |
| Go support | N/A (wire format only) | Go SDK exists but is **community-tier**, while TypeScript and Python are first-class |

**Ship the AI SDK v5 encoder first** (v0.5): the largest install base, the shortest path to "my Go backend has a modern chat UI," and it needs nothing beyond the event stream. **Add the AG-UI encoder once run state lands**, because its distinguishing feature is precisely the thing that depends on state.

### The Go opportunity

AG-UI's Go SDK sitting in the community tier while TypeScript and Python are first-class is the same gap as MCP-server export: a protocol where Go is underserved and a well-built Go implementation is cheap differentiation. Worth targeting deliberately — but as a module, on the official protocol, and only after the interior model can feed it.

---

## 3. What else the core should have (not yet in the roadmap)

Five additions, ordered by how often a developer will feel their absence.

### 3.1 `log/slog`, injected — never a global logger

Accept `*slog.Logger` through an option; default to `slog.Default()`. **No package-level `Logger()`, no bundled logging library.** The previous generation baked in zerolog behind a global, which means a library that cannot be silenced, cannot be routed, and cannot participate in a host application's structured logging. slog is standard library now; there is no excuse for a framework logging dependency.

### 3.2 Context compaction and history management

The most-felt real-world pain that no Go framework addresses: long-running agents blow the context window, and the failure mode is a hard error mid-run after the user has already paid for the conversation.

- A `Compactor` interface invoked when history approaches a budget: summarize older turns, offload oversized tool outputs to a store and replace them with truncated stubs plus a retrieval handle.
- A token budget derived from `Capabilities.ContextWindow`, with a diagnostic before the wall, not an error at it.
- Ship one default implementation (summarize-oldest) and keep the interface small.

This is arguably higher day-to-day value than A2A, workflows, or multi-agent primitives — it is the difference between an agent that works in a demo and one that survives an afternoon.

### 3.3 Record/replay testing ("cassettes")

`agenttest` covers scripted behavior. The complement is recording *real* provider interactions once and replaying them forever: a `Recorder` provider decorator writing request/response pairs to JSON, and a `Replayer` reading them back. Tests then exercise real model quirks with no key, no cost, and no flake, and a provider's behavioral change shows up as a diff. Cheap to build on top of an existing `Provider` interface, and it is what makes an eval suite runnable in a fork's CI.

### 3.4 Policy-based model routing

Distinct from failure fallback: routing chooses a model *by policy* before a call — cheap model for classification, expensive one for synthesis, long-context model when history exceeds a threshold. `router.New(rules...)` implementing `Provider`, so it composes with fallback and middleware. Cost control is the top production concern after correctness, and today every team hand-rolls this.

### 3.5 Prompt-cache control as a first-class concept

Cache breakpoints are the single largest cost lever on long system prompts and reused context, and they are provider-specific enough that users get them wrong. Model it in the request (mark a message or prefix as cacheable), let adapters translate or ignore, and report cache hits through `Usage.CachedInputTokens` — which the core type already carries.

---

## 4. What I would now cut

Being willing to cut one's own proposals is part of scope discipline.

- **Code-as-action sandbox (Starlark/wazero).** Listed earlier as a differentiator. It is a compelling demo, but it forks the tool story — every tool must now be callable two ways — and it carries a permanent security-review burden. Park it as an experiment, not a roadmap item.
- **A framework-owned HTTP eval server.** `agk` already has one and evals belong primarily in `go test`. Keep the HTTP mode only because the CLI already depends on it; do not grow it.
- **Memory extraction pipelines (mem0-style) in the core roadmap.** A product in its own right. Ship the storage and retrieval contracts; let the opinionated extraction loop be a separate module that can fail without taking the core with it.

---

## 5. What the ecosystem must not build

| Don't build | Because |
|---|---|
| A hosted platform / SaaS control plane | Premature by years; splits a small team's focus away from the library that has to be excellent first |
| A React/Svelte component library | Emit the AI SDK v5 protocol instead and inherit every existing chat UI |
| A model gateway or proxy | That is LiteLLM's and OpenRouter's turf; a first-class `openai-compatible` provider reaches all of it with no ongoing cost |
| A vector database | Wrap the ones that exist behind a small interface |
| A custom logger, config loader, or DI container | `log/slog`, plain structs, and constructors. Users pick their own config library |
| A bespoke job queue | River and Asynq exist. The durable journal is a different thing — it is execution state, not work distribution |
| A fork of MCP or A2A | Protocol churn is constant; adopt the official SDKs and let them absorb it |
| A plugin marketplace | The template registry is enough surface for the community size |

---

## 6. Ecosystem division of labor

| Repository | Owns | Must not own |
|---|---|---|
| **`agentkit`** | Contracts and runtime: message IR, provider/tool/embedder interfaces, agent loop, durability, workflows. Zero third-party dependencies, enforced in CI. | HTTP servers, CLIs, UIs, storage engines, protocol implementations |
| **`agentkit/providers/*`, `/mcp`, `/a2a`, `/otel`, `/stores/*`** | Everything with a dependency or a wire protocol. Multi-module in the same repository so a contract change and its adapters move in one PR, with one CI. | Opinions about project structure |
| **`AgenticGoKit`** | The brand, the docs site, the umbrella examples. `v1beta` frozen with a published sunset date and a migration guide. | New feature work — evolving two cores splits a small team and confuses adopters |
| **`agk`** | Everything needing a human: scaffolding, dev playground, trace explorer, eval runner, `mcp publish`. | Runtime logic. Its eval matchers (embedding, LLM-judge, hybrid) belong in a library so `go test` can use them too — the CLI should be one of two frontends, not the only home |
| **`agk-templates`** | Opinions about project structure, one per credible starting point. Versioned, pinned, conformance-tested in CI. | Workarounds for framework bugs. A scaffolder that papers over a defect ships that defect to every new user |

### The one structural change worth making now

**Move the eval scorers out of `agk` and into a library**, and **publish a versioned contract between the framework and the CLI** (a stable trace schema and an OpenAPI-described eval API in a small shared module, with nightly cross-repo CI). Today the CLI's two best features stand on undocumented internals across a repository boundary — the debug-format trace file and an eval server whose traces are empty. That is not a feature gap; it is a structural one, and it gets more expensive every release.

---

## 7. The test for anything new

Before adding a capability, answer:

1. **Does it change user code shape?** If no, it does not belong in core.
2. **Would its absence be felt in the first hour, the first week, or the first year?** First-hour and first-week items outrank first-year items regardless of how interesting they are.
3. **Does something in the Go ecosystem already do this well?** If yes, integrate rather than implement.
4. **Can it fail without taking the core down?** If not, it belongs behind an interface with a boring default.
5. **Is there a real user waiting for it, or a hypothetical one?** Speculative surface in a pre-1.0 core is a promise that has to be kept forever.
