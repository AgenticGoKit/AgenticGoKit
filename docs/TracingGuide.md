# AgentFlow Tracing Guide

## Overview

**AgentFlow** provides a robust tracing system aligned with **Azure observability best practices**, enabling deep monitoring, debugging, and understanding of multi-agent workflow execution.

---

## Tracing Architecture

```mermaid
graph TD
    A[Callback Registry] --> B[Trace Hooks]
    B --> C[Trace Logger]
    C --> D[Trace Entries]
    D --> E[Trace Storage]
    E --> F[AgentCLI Visualization]
```

### Core Components

* **Callback Registry**: Central registration point for trace hooks.
* **Trace Hooks**: Capture data at key execution points.
* **Trace Logger**: Interface for storing trace entries.
* **Trace Storage**: In-memory or file-based persistence.
* **AgentCLI**: Tool for trace visualization and analysis.

---

## Lifecycle Hooks

AgentFlow registers trace hooks at critical lifecycle stages:

| Hook Point            | Trigger Timing                        | Purpose                              |
| --------------------- | ------------------------------------- | ------------------------------------ |
| `BeforeEventHandling` | Before event processing begins        | Event validation and preparation     |
| `BeforeAgentRun`      | Twice: Orchestrator and Runner phases | Routing setup and execution prep     |
| `AfterAgentRun`       | After agent completes                 | State transition and result handling |
| `AfterEventHandling`  | After event is fully processed        | Cleanup and metrics collection       |
| `AgentError`          | On agent execution failure            | Error reporting and recovery         |

---

## Dual `BeforeAgentRun` Hooks Explained

AgentFlow intentionally records two `BeforeAgentRun` entries per execution:

### 1. **Orchestrator Phase**

* **When**: Before routing decision.
* **Purpose**: Shows how the orchestrator selects an agent.
* **State**: Pre-routing data.

### 2. **Runner Phase**

* **When**: Right before calling `Run()`.
* **Purpose**: Marks the actual execution start.
* **State**: May reflect updated routing decisions.
* **Use**: Final checkpoint for circuit breaking or cancellation.

> This dual-hook pattern enhances observability by capturing both decision-making and execution contexts.

---

## Setting Up Tracing

Add tracing in **3 simple steps**:

```go
// 1. Create a trace logger
traceLogger := agentflow.NewInMemoryTraceLogger()

// 2. Set the logger on the runner
runner.SetTraceLogger(traceLogger)

// 3. Register trace hooks
agentflow.RegisterTraceHooks(callbackRegistry, traceLogger)
```

---

## Trace Entry Structure

A `TraceEntry` provides detailed insight into agent execution:

```go
type TraceEntry struct {
    Timestamp     time.Time
    Type          string
    EventID       string
    SessionID     string
    AgentID       string       // optional
    State         State        // optional
    Error         string       // optional
    Hook          HookPoint    // optional
    TargetAgentID string       // optional
    SourceAgentID string       // optional
    AgentResult   *AgentResult // optional
}
```

---

## Analyzing Trace Output

Use the `agentcli` tool to inspect trace data:

```bash
# Full trace with all details
agentcli trace <session-id>

# Visualize agent flow only
agentcli trace --flow-only <session-id>

# Filter by specific agent
agentcli trace --filter <agent-name> <session-id>
```

### Trace Insights:

* Agent interaction sequences
* State transitions and routing logic
* Error handling patterns
* Performance bottlenecks

---

## Best Practices

### 1. Enable Full Tracing in Development

```go
traceLogger := agentflow.NewInMemoryTraceLogger()
runner.SetTraceLogger(traceLogger)
agentflow.RegisterTraceHooks(callbackRegistry, traceLogger)
```

### 2. Use File-Based Tracing in Production

```go
traceLogger := agentflow.NewFileTraceLogger("./traces")
runner.SetTraceLogger(traceLogger)
agentflow.RegisterTraceHooks(callbackRegistry, traceLogger)
```

### 3. Maintain Consistent Session IDs

```go
sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())

eventMetadata := map[string]string{
    agentflow.SessionIDKey: sessionID,
}

event := agentflow.NewEvent(targetAgent, payload, eventMetadata)
```

### 4. Provide Rich State Data

```go
state.Set("request_id", requestID)
state.Set("operation_context", context)
state.Set("input_parameters", params)
```

### 5. Analyze Dual `BeforeAgentRun` Hooks

Example:

```bash
16:27:31.956 | BeforeAgentRun | planner  | {user_request: "Research..."}
16:27:31.956 | BeforeAgentRun | planner  | {user_request: "Research...", route: "researcher"}
```

* First entry: Initial state before routing.
* Second entry: Updated state after routing decisions.

---

## Troubleshooting

### Missing Trace Entries?

* Confirm trace logger is initialized and set.
* Ensure `RegisterTraceHooks` was called.
* Check for consistent session IDs.

### Unexpected Routing?

* Compare state between the two `BeforeAgentRun` hooks.
* Look for `RouteMetadataKey` after `AfterAgentRun`.
* Inspect orchestrator decisions in the trace.

---

## Advanced Usage

### Custom Trace Loggers

Implement the `TraceLogger` interface:

```go
type TraceLogger interface {
    Log(entry TraceEntry) error
    GetTrace(sessionID string) ([]TraceEntry, error)
}
```

### Trace Analysis in CI/CD

```bash
agentcli trace --validate-flow <expected-pattern.json> <session-id>
```

### Distributed Trace Context Propagation

```go
event.SetMetadata("trace_id", span.SpanContext().TraceID().String())
event.SetMetadata("span_id", span.SpanContext().SpanID().String())
```

---

## Conclusion

Effective tracing in **AgentFlow** is essential for:

* Debugging workflows
* Understanding agent behavior
* Diagnosing performance issues

By leveraging the **dual-hook architecture** and best practices outlined above, you'll gain deep, actionable insight into your multi-agent systems.


