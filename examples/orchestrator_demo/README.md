# Orchestrator Demo

A minimal, runnable example that demonstrates the public plugin-based architecture:

- Runner via `plugins/runner/default`
- Orchestrator via `plugins/orchestrator/default`
- Logging via `plugins/logging/zerolog`
- In-memory memory via `plugins/memory/memory`

## Run

Use PowerShell:

- go run ./examples/orchestrator_demo

Expected output (abridged):

- Runner started.
- RouteOrchestrator: Registered agent (echo, done, error-handler)
- Final state printed with fields: message, handled_by, timestamp

Notes:
- Routing is driven by `core.RouteMetadataKey` in the emitted event metadata (set to `echo`).
- The example registers a minimal `error-handler` agent that absorbs error events to avoid loops.
