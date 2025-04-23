# SequentialAgent Benchmarks

This document records the benchmark results for the `SequentialAgent`.

## Benchmark: Run Latency (5 No-Op Agents)

Measures the overhead of the `SequentialAgent` itself when running a chain of 5 sub-agents that do minimal work.

**Command:**

```bash
go test -bench=. -benchmem -run=^$ kunalkushwaha/agentflow/internal/agents -benchtime=1s
```
*(Note: The actual command run might vary slightly depending on the execution method, like the one VS Code used)*

**Results:**

```
goos: windows
goarch: amd64
pkg: kunalkushwaha/agentflow/internal/agents
cpu: AMD Ryzen 7 PRO 7840U w/ Radeon 780M Graphics
BenchmarkSequentialAgent_Run-16    	 2052136	       550.2 ns/op	     624 B/op	      13 allocs/op
PASS
ok  	kunalkushwaha/agentflow/internal/agents	4.297s
```

**Analysis:**

*   The overhead for running 5 no-op agents is approximately 550 ns.
*   This includes the loop, context checks, and state cloning within the `SequentialAgent`.
*   Each iteration allocates memory primarily for the state cloning (`initialState.Clone()`), resulting in 13 allocations total for the 5 agents plus potentially the loop overhead. The 624 bytes likely correspond to the cloned state maps and internal struct overhead.
*   The overhead per agent seems reasonable (around 110 ns and ~2-3 allocations per agent, considering the cloning).