# ParallelAgent Benchmarks

This document records the benchmark results for the `ParallelAgent`.

## Benchmark: Run Latency (50 Concurrent No-Op Agents)

Measures the end-to-end latency and overhead of the `ParallelAgent` when running 50 concurrent sub-agents that do minimal work (no-op). This highlights the cost of goroutine scheduling, channel communication, state cloning, and result aggregation.

**Command:**

```bash
go test -C "C:\Users\kkushwaha\work\agentflow" -bench=BenchmarkParallelAgent_Run -benchmem -run=^$ ./internal/agents -benchtime=3s
```
*(Adjust benchtime as needed for stable results)*

**Results:**

```
goos: windows
goarch: amd64
pkg: kunalkushwaha/agentflow/internal/agents
cpu: AMD Ryzen 7 PRO 7840U w/ Radeon 780M Graphics
BenchmarkParallelAgent_Run-16    	   35434	     32987 ns/op	   26587 B/op	     415 allocs/op
PASS
ok  	kunalkushwaha/agentflow/internal/agents	8.224s
```

**Analysis:**

*   Running 50 concurrent no-op agents takes approximately 33,000 ns (33 microseconds) end-to-end on this machine.
*   This includes launching 50 goroutines, cloning the initial state 50 times, channel communication for results, wait group synchronization, and final state/error aggregation.
*   The memory allocation is significant (26.5 KB and 415 allocations per operation), primarily driven by:
    *   Cloning the `initialState` 50 times within the goroutines (`input.Clone()`).
    *   Creating the `subResult` struct for the channel.
    *   Allocations related to goroutine stacks and channel operations.
    *   Cloning the `initialState` once more for the `finalState` aggregation.
    *   Allocating the `allErrors` slice.
*   The overhead per agent is roughly 660 ns (33000 ns / 50 agents) and ~8 allocations (415 / 50), though this is an average and doesn't account for the fixed setup/teardown cost of the `ParallelAgent` itself.
*   Compared to the `SequentialAgent` (550 ns for 5 agents, ~110 ns/agent), the per-agent overhead appears higher here, likely due to the added costs of concurrency management (goroutines, channels, waitgroups). However, the total execution time for parallel agents doing real work should be significantly lower than sequential execution if the work can truly happen in parallel.