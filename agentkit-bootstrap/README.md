# agentkit

**The typed, durable core of AgenticGoKit.**

```go
import agk "github.com/agenticgokit/agentkit"
```

agentkit is a ground-up core for building AI agents in Go: typed at compile time, honest at runtime, and durable by design. It is the foundation the [AgenticGoKit](https://github.com/AgenticGoKit/AgenticGoKit) ecosystem is converging on, and the module the [`agk`](https://github.com/AgenticGoKit/agk) CLI and [templates](https://github.com/agk-templates) will target.

> **Status: pre-alpha.** The contracts in this repository (message IR, `Provider`, `Tool`, `Agent[D, O]`) are being locked deliberately, because everything else builds on them. Expect breaking changes until v0.1; expect no silent breakage after it.

---

## Why another core

Go deserves an agent framework that plays to Go's strengths rather than transcribing a Python one. Four design commitments:

**1. Typed end to end.** Agents are generic over their dependencies and output; tools are plain typed Go functions whose JSON schemas are *derived* from their signatures, so a schema can never drift from the code that runs.

```go
type SearchArgs struct {
    Query string `json:"query" jsonschema:"description=Search terms"`
    Limit int    `json:"limit,omitempty"`
}

search := agk.NewTool("search", "Search the web",
    func(ctx context.Context, in SearchArgs) ([]Hit, error) {
        db, _ := agk.DepsFrom[Deps](ctx)   // deps, not globals
        return db.Search(ctx, in.Query, in.Limit)
    }, agk.ReadOnly())

agent, err := agk.New[Deps, Verdict](provider, "gpt-4o",
    agk.WithSystem("You review pull requests."),
    agk.WithTools(search),
    agk.WithUsageLimits(agk.UsageLimits{MaxTotalTokens: 50_000}),
)

res, err := agent.Run(ctx, "Review PR 42", deps)
fmt.Println(res.Output.Score)   // Verdict — no casting, no map[string]any
```

**2. Multi-turn from the first commit.** Conversations are `[]Message` with typed content parts (text, image, audio, reasoning, tool calls, tool results). Nothing is flattened into a string, so provider-native conversation handling, prompt caching, and correct tool-call protocols all remain available.

**3. Fail loudly, in values.** Every error wraps a sentinel you can `errors.Is`. Non-fatal findings are `Diagnostic` values delivered to your code — not log lines you might never wire up. Nothing silently degrades: if a model can't call tools natively, you get a diagnostic, not a mystery.

```go
if errors.Is(err, agk.ErrMaxSteps) { ... }

for _, d := range agent.Diagnostics() {
    if d.Code == agk.DiagNoNativeTools { /* fail the deploy */ }
}
```

**4. Testable without a network.** `agenttest` ships scripted providers so agents, tools, and loops are exercised in ordinary `go test` runs — no keys, no cost, no flakes.

```go
p := agenttest.NewScript(
    agenttest.Calls(agenttest.Call("c1", "search", SearchArgs{Query: "go generics"})),
    agenttest.Text("Here's what I found."),
)
agent, _ := agk.New[Deps, string](p, "test-model", agk.WithTools(search))
res, _ := agent.Run(ctx, "search for me", deps)
```

---

## What's here today

| Component | State |
|---|---|
| Message IR (roles, typed content parts) | ✅ |
| `Provider` interface + `Capabilities` discovery | ✅ contract |
| `Embedder` (separate from `Provider` by design) | ✅ contract |
| Typed tools + reflection-derived JSON Schema | ✅ |
| `Agent[D, O]` with tool loop, structured output, self-correcting retries | ✅ |
| Usage limits, step budgets | ✅ |
| Sentinel errors, `StepError`/`ToolError` unwrapping | ✅ |
| Build-time diagnostics | ✅ |
| `agenttest` scripted provider | ✅ |
| Provider adapters (OpenAI, Anthropic, Gemini, Ollama, Bedrock, …) | 🚧 |
| Streaming agent runs | 🚧 contract defined |
| Durable execution (journal, interrupt/resume, fork) | 🚧 |
| Typed workflows | 🚧 |
| Sessions, memory | 🚧 |

See [`docs/DESIGN.md`](docs/DESIGN.md) for the decisions behind each contract, and the issue tracker for the sequenced plan.

## Non-goals

- **No `map[string]any` APIs.** If it can be typed, it is typed.
- **No decorative configuration.** Every option this library accepts changes behavior, or it doesn't ship.
- **No silent fallbacks.** Degradation is reported as a `Diagnostic` or an error, never absorbed.
- **No fat dependency graph.** The core is standard library only; every integration is a separate module.

## Relationship to AgenticGoKit

[AgenticGoKit](https://github.com/AgenticGoKit/AgenticGoKit) is the current framework (`v1beta`); it remains supported while agentkit matures. agentkit is the clean-slate core its roadmap describes — same ecosystem, same CLI, no legacy surface. Migration guidance will land before v0.1.

## License

Apache 2.0
