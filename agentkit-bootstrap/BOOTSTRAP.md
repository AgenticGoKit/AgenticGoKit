# Bootstrap notes (delete this file after transplanting)

This directory is the initial content for `github.com/agenticgokit/agentkit`, staged inside the AgenticGoKit repository because the session that produced it could not push to the new repository (its GitHub access was scoped to `agenticgokit/agenticgokit` only).

It is a complete, self-contained Go module: it builds, vets clean, is `gofmt`-clean, and its tests pass under `-race` with no third-party dependencies.

## Transplant

```bash
# from a clone of the (empty) agentkit repository
git clone git@github.com:agenticgokit/agentkit.git
cd agentkit

# copy everything except this file
rsync -a --exclude BOOTSTRAP.md /path/to/AgenticGoKit/agentkit-bootstrap/ .

go test -race ./...     # expect: ok github.com/agenticgokit/agentkit

git add -A
git commit -m "feat: initial core — message IR, provider contract, typed tools and agents

Establishes the contracts documented in docs/DESIGN.md: a multi-turn
message IR, a four-method Provider interface with capability discovery,
an Embedder kept separate from Provider, typed tools whose JSON schemas
are derived from Go types, and a generic Agent[D, O] with a tool loop,
structured output, and self-correcting retries.

Includes agenttest for LLM-free testing, sentinel errors throughout, and
build-time diagnostics as values rather than log lines."
git push -u origin main
```

Then remove the staging copy from the AgenticGoKit repository:

```bash
cd /path/to/AgenticGoKit
git rm -r agentkit-bootstrap
git commit -m "chore: remove agentkit bootstrap staging directory (transplanted)"
```

## File the issues

`ISSUES.md` holds 37 ready-to-file issues with front matter (title, type, milestone, labels). Either:

```bash
# with the gh CLI, from the agentkit repo root
./scripts/file-issues.sh
```

or grant a Claude Code session access to `agenticgokit/agentkit` and ask it to file them from `ISSUES.md` — it will create the milestones and issues, and link the epics to their children.

Create these milestones first (the script does it for you):

| Milestone | Theme |
|---|---|
| v0.1 Core contracts | Real providers, streaming, conformance |
| v0.2 Durable | Journal, interrupt/resume, fork |
| v0.3 Workflows | Typed steps, fan-out, per-step policies |
| v0.4 Multi-agent | Agents-as-tools, handoffs, guardrails, sessions |
| v0.5 Dev loop | Registry, wire format, evals, MCP, prompts |

## Repository settings worth setting on day one

- Default branch `main`; require the CI check before merge.
- Enable Issues and Discussions; disable Wiki and Projects until needed.
- Description: *The typed, durable core of AgenticGoKit — build AI agents in Go.*
- Topics: `go`, `golang`, `ai-agents`, `llm`, `agentic-ai`, `agents`, `mcp`.
- Add a link from the AgenticGoKit README so the existing audience finds this repo.
