# agentkit design decisions

Every entry records a decision, the reasoning, and the alternative rejected. These are the contracts everything else builds on, so they are worth arguing about now and not later.

Evidence for the "learned from" notes lives in the AgenticGoKit analysis set: `docs/CONSOLIDATED_REPORT.md`, `docs/AGNOSTIC_FRAMEWORK_ANALYSIS.md`, and `docs/DEVELOPER_FIRST_ROADMAP.md` in that repository.

---

## D1. Messages, not a system/user pair

**Decision.** A call takes `[]Message`, each with a `Role` and typed `[]Part` (text, image, audio, reasoning, tool call, tool result).

**Why.** Anything less makes multi-turn a lie. The previous generation modeled a call as `Prompt{System, User string}` and simulated conversation by concatenating prior output and tool results into the next user string. That defeats provider-side conversation handling and prompt caching, degrades long-run quality, and makes correct native tool loops impossible — providers require the assistant's `tool_use` and the matching `tool_result` echoed back in structure, not prose.

**Rejected.** "Add a `History []string` field later." Retrofitting message structure through a stringly-typed core touches every adapter anyway; the cost is the same now and the damage is smaller.

## D2. Provider is small; capabilities are discoverable

**Decision.** `Provider` has four methods: `Name`, `Generate`, `Stream`, `Capabilities`. `Capabilities(model)` reports native tools, native JSON, streaming, modalities, context window.

**Why.** A small interface is implementable by third parties without depending on internals, and lets cross-cutting concerns (retry, rate limit, fallback, cost accounting, caching) be written *once* as decorators rather than N times inside adapters. Capability discovery exists so the runtime branches deliberately: when tools are registered on a model that can't call them, the developer gets a `Diagnostic`, not a mysteriously toolless agent.

**Rejected.** Capability flags inferred from model-name string matching. Model names change faster than any table can track, and the failure mode is silent.

## D3. Embedder is not Provider

**Decision.** Embeddings live behind a separate `Embedder` interface with an explicit `Dimensions()`.

**Why.** Learned the hard way: when `Embeddings()` was part of the model interface, providers without an embeddings API (Anthropic) had to stub it, and one stub returned zero vectors — semantic search silently returned meaningless results with no error. Separating the interfaces makes "this provider cannot embed" a compile-time fact instead of a runtime lie. `Dimensions()` exists so a store/model mismatch is caught before the first write, not after a corrupted index.

## D4. Sampling parameters are pointers

**Decision.** `Params` fields are `*float32` / `*int`.

**Why.** Temperature 0 is a legitimate, common setting (determinism, evals). With value types, "unset" and "zero" are the same bit pattern, so every layer guesses — and the previous generation guessed wrong in three separate places, silently rewriting explicit zeros and defaulting `MaxTokens` to 150. Pointers make intent unambiguous end to end.

## D5. Tools are typed functions; schemas are derived

**Decision.** `NewTool[In, Out](name, desc, fn)` derives both input and output JSON Schema from the Go types via reflection.

**Why.** A hand-written schema beside a Go function is drift waiting to happen, and the fallback for tools without schemas — a single string field named `input` — actively teaches models to double-wrap their arguments (the previous generation shipped a normalization hack to undo its own fallback). Deriving from the signature makes the two impossible to disagree. Output schemas are derived too, so contract drift is catchable at the boundary.

**Rejected.** Requiring an external schema library in the core. Reflection over `json` + `jsonschema` tags covers the common cases with zero dependencies; a richer generator can be layered in a separate module.

## D6. Dependencies travel in context, tools stay non-generic

**Decision.** `Run(ctx, input, deps D)` stores deps in the context; tools retrieve them with `DepsFrom[D](ctx)`.

**Why.** Typed dependency injection without globals, and without forcing every tool constructor and option to carry `[D]` type parameters — which would make every call site verbose (`WithTool[Deps, Out](...)`). The trade-off is that deps retrieval is checked at runtime, not compile time; it is contained to one call, returns `(D, bool)`, and keeps the common path clean.

## D7. Structured output is declared once, obtained three ways

**Decision.** `Agent[D, O]` where `O` is non-string implies structured output. *How* it's obtained is separate: `OutputTool` (schema as a tool call — the default, works nearly everywhere), `OutputNative` (provider JSON schema), `OutputPrompted` (schema in the prompt).

**Why.** What you want and what the model can do are independent concerns. Coupling them makes switching providers a rewrite. When native mode is requested on a model that lacks it, the agent falls back to tool mode *and says so* with a diagnostic.

## D8. Model mistakes are correctable; system faults are not

**Decision.** Invalid tool arguments, unknown tool names, and `Retry(...)` errors from tools are returned to the model as error tool-results so it can self-correct. Genuine failures (provider errors, tool panics, context cancellation) abort with a wrapped error.

**Why.** Validation-driven self-correction measurably improves structured-output reliability, and a hallucinated tool name is not a reason to discard a run. But the distinction has to be explicit: absorbing real failures into the loop is how a framework ends up reporting success for a broken run.

## D9. Streams are cursors, and errors cannot be lost

**Decision.** `Stream` is `Next() bool / Event() Event / Err() error / Close() error`.

**Why.** The single most damaging streaming bug in the previous generation was that producer errors never reached the terminal check: failed streams looked successful, and downstream workflow steps consumed empty output as if it were an answer. A cursor makes `Err()` the one place a caller must look, and `Next()` returning false is deliberately ambiguous between "done" and "failed" so that consulting `Err()` is mandatory. `Close()` cancels the underlying request — cancellation that doesn't stop token generation is cancellation that costs money.

**Rejected.** A bare `<-chan Event`. Channels have no error channel by construction, which is exactly how errors got dropped before.

## D10. Non-fatal findings are values

**Decision.** Degraded configurations produce `Diagnostic` values, available from `agent.Diagnostics()` and pushed to `WithDiagnosticHandler`.

**Why.** A framework that only logs is invisible to any consumer that wires its own logger — and the consumers most likely to have their own logger are exactly the production users who most need the warning. Diagnostics carry a stable `Code` so callers branch programmatically ("fail the deploy if embeddings are degraded") instead of matching message text. They are also logged; ignoring the API costs nothing.

## D11. Errors wrap sentinels

**Decision.** Every returned error wraps one of `ErrInvalidConfig`, `ErrMaxSteps`, `ErrUsageLimit`, `ErrToolNotFound`, `ErrNoOutput`, `ErrUnsupported`, `ErrProvider`; `StepError` and `ToolError` add location while preserving the chain via `Unwrap`.

**Why.** `errors.Is/As` is the Go contract for programmatic error handling. The previous generation flattened step errors to strings with `%s`, which destroyed the chain and made its *own* typed-error package unusable against its own engine.

## D12. Standard library only in the core

**Decision.** The core module has zero third-party dependencies. Adapters, stores, and telemetry integrations are separate modules.

**Why.** A single fat `go.mod` forced every consumer to pull Postgres drivers, a vector database, a metrics client, and a cloud SDK — one of which was used only by a test. Anyone deploying a small agent paid for all of it. Keeping the core dependency-free makes the cost of each integration explicit and opt-in.

---

## Decisions still open

These are deliberately unresolved and tracked as issues. They are the ones worth settling before v0.1, because they shape public API.

1. **Durable execution shape.** Journal-skip (re-invoke and replay completed steps from a journal) versus event-history replay. Leaning journal-skip: it imposes no determinism constraints on user code, so plain goroutines and `time.Now` stay legal outside steps.
2. **Checkpoint store contract.** Target a five-method interface addressable by `(threadID, namespace, checkpointID)`, with pending-writes so successful parallel siblings aren't re-executed on resume. Small contracts attract community backends; large ones don't.
3. **Interrupt/resume ergonomics.** `Interrupt[T](ctx, payload)` that suspends durably and returns the recorded value on resume, with the node re-executing to that point — no stack serialization required.
4. **Typed workflow steps.** How far to push compile-time checking between step boundaries without making construction verbose.
5. **Streaming agent runs.** `Agent.Stream` returning agent-level events (step start, tool call, text delta) versus provider events passed through.
6. **Session/state contract.** Interface shape and scope prefixes for cross-run state.
