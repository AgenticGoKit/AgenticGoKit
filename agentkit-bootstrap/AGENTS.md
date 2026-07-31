# AGENTS.md

Guidance for AI coding agents working in this repository. Humans may find it useful too.

## What this project is

`agentkit` is the typed, durable core of AgenticGoKit: a Go library for building AI agents. Module path `github.com/agenticgokit/agentkit`, conventionally imported as `agk`.

## Commands

```bash
go build ./...
go test -race ./...      # -race is mandatory; the CI gate enforces it
go vet ./...
gofmt -l .               # must print nothing
```

There is no code generation step and no external service needed to run the tests.

## Project layout

| Path | Contents |
|---|---|
| `message.go` | Message IR: roles, content parts |
| `provider.go` | `Provider`, `Embedder`, request/response types, capabilities, stream events |
| `tool.go` | `Tool` interface, typed `NewTool`, tool set |
| `schema.go` | Reflection-based JSON Schema generation |
| `agent.go` | `Agent[D, O]`, options, run loop |
| `errors.go` | Sentinel errors and error wrappers |
| `diagnostics.go` | Build-time diagnostic values |
| `agenttest/` | LLM-free test doubles |
| `docs/DESIGN.md` | Every design decision with its reasoning |

## House rules

These are not stylistic preferences; they are the reasons this core exists. Read `docs/DESIGN.md` before changing a contract.

1. **No `map[string]any` in public APIs.** If it can be typed, type it. Tool arguments cross the wire as `json.RawMessage` and are unmarshaled into a user type immediately.
2. **No silent fallbacks.** If something cannot work as asked, return an error or emit a `Diagnostic`. Never substitute a degraded implementation quietly — that is the specific failure this project was created to avoid.
3. **No decorative configuration.** Do not add an option, config field, or struct tag unless code reads it and behavior changes. A knob that does nothing is worse than no knob.
4. **Errors wrap sentinels.** Every error path wraps one of the `Err*` values in `errors.go`, using `%w`. Never `fmt.Errorf("...: %s", err)`.
5. **Optional scalars are pointers.** Especially sampling parameters: an explicit `0` must be distinguishable from unset.
6. **Zero third-party dependencies in the core module.** Integrations go in their own modules. Do not add anything to `go.mod`.
7. **Race-free by construction.** Shared state gets a mutex or is not shared. Every test must pass `go test -race`.
8. **No fabricated data.** Never return a placeholder result that looks like a real one — not in a stub, not in a "TODO" path. Return `ErrUnsupported` instead.

## Testing conventions

- Tests live in package `agentkit_test` (black box), so they exercise the public API a user sees.
- Use `agenttest.NewScript` for model behavior; never call a real provider in a test.
- Assert on `errors.Is` / `errors.As`, not on error strings.
- A test for a bug fix must fail without the fix.

## Commit conventions

Conventional-commit prefixes (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`). Explain *why* in the body; the diff already shows what. Reference the issue number.
