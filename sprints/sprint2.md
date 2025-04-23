## Summary

Sprint 2 focuses on building and integrating our deterministic workflow agents—`SequentialAgent`, `ParallelAgent`, and `LoopAgent`—so that non-LLM pipelines can be composed, tested, and benchmarked. The tasks below break each user story into granular, actionable developer tasks with clear deliverables and owner estimates.

---

## 1. Agent Interface & Common Types

**Story:** Define the core `Agent` interface and shared models (`State`, `AgentResult`).  
**Tasks:**
1. **Design `Agent` interface**  
   - Method signature: `Run(ctx context.Context, in State) (out State, err error)`  
   - Add GoDoc comments and usage examples.  
2. **Create `State` and `AgentResult` types**  
   - `State` carries arbitrary data (map or struct) and metadata.  
   - `AgentResult` encapsulates output data, errors, and timing.  
3. **Write unit tests**  
   - Validate that `AgentResult` fields marshal/unmarshal correctly.  
   - Test `State` mutation and cloning behavior.  

---

## 2. Implement `SequentialAgent`

**Story:** Develop an agent that executes a slice of sub-agents in order, short-circuits on error.  
**Tasks:**
1. **Implement core logic**  
   - Loop through `[]Agent`, passing the evolving `State` from one to the next.  
   - Return immediately if any sub-agent returns an error.  
2. **Configuration struct**  
   - Define `SequentialAgentConfig` (e.g., sub-agents list).  
3. **Unit tests**  
   - Success path: all sub-agents run and produce final state.  
   - Failure path: early exit when a sub-agent errors.  
   - Edge cases: zero sub-agents, nil agent entries.  
4. **Benchmark**  
   - Measure `Run` latency for a chain of 5 no-op agents.  
   - Record results in `benchmarks/sequential_agent.md`.  

---

## 3. Implement `ParallelAgent`

**Story:** Build an agent that runs sub-agents concurrently, aggregates results and errors.  
**Tasks:**
1. **Core fan-out/fan-in logic**  
   - Launch each sub-agent in its own goroutine with context for cancellation.  
   - Collect outputs and errors via channels.  
2. **Timeout and cancellation**  
   - Accept a `timeout` parameter in `ParallelAgentConfig`.  
   - Cancel remaining goroutines on overall timeout or context cancellation.  
3. **Error aggregation**  
   - Define a `MultiError` type that aggregates multiple errors.  
   - Ensure at least partial results are returned if some agents succeed.  
4. **Unit tests**  
   - All success: verify all sub-agent outputs.  
   - Partial failure: inspect `MultiError` and successful results.  
   - Timeout behavior: one sub-agent intentionally sleeps beyond the timeout.  
5. **Benchmark**  
   - Run 50 concurrent no-op sub-agents, measure end-to-end latency.  
   - Document in `benchmarks/parallel_agent.md`.  

---

## 4. Implement `LoopAgent`

**Story:** Create an agent that invokes a sub-agent repeatedly until a condition is met or max iterations reached.  
**Tasks:**
1. **Loop control logic**  
   - Accept a `ConditionFunc(State) bool` and `MaxIterations` in `LoopAgentConfig`.  
   - Iterate: run sub-agent, evaluate condition, break if true or iteration cap hit.  
2. **Safety guard**  
   - Default `MaxIterations = 10`; override via config.  
   - Return a specific `ErrMaxIterationsReached` if cap is hit.  
3. **Unit tests**  
   - Loop ends when condition returns true.  
   - Error propagation if sub-agent errors mid-loop.  
   - Verify error on max iterations.  
4. **Benchmark**  
   - Test n = 10 loops of a trivial sub-agent; measure total time.  
   - Include results in `benchmarks/loop_agent.md`.  

---

## 5. Integration & Runner Wiring

**Story:** Wire the new workflow agents into the `Runner` so they can be dispatched like any other agent.  
**Tasks:**
1. **Runner registration support**  
   - Enable registering workflow agents by name in `Runner`.  
2. **Dispatch tests**  
   - Emit events that target workflow agents; assert correct outputs.  
   - Nested workflows: e.g., a `ParallelAgent` inside a `LoopAgent`.  
3. **End-to-end sample**  
   - Build a sample Go program in `examples/` demonstrating all three agents in a composite pipeline.  
   - Include README usage and expected output.  

---

## 6. Code Review, Documentation & Clean-up

**Tasks:**
1. **Peer reviews**  
   - Schedule review sessions for each agent’s implementation.  
2. **GoDoc updates**  
   - Ensure all public types and methods have comprehensive GoDoc.  
3. **Linting & Formatting**  
   - Run `golangci-lint` and `go fmt` across newly added code.  
4. **Sprint Demo Prep**  
   - Prepare a short demo script showcasing each agent in action.  

---

These tasks ensure that by the end of Sprint 2, our deterministic workflow agents will be fully implemented, tested, benchmarked, documented, and integrated—providing a solid foundation for Sprint 3’s LLM-driven orchestration work.