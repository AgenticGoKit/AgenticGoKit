# LoopAgent Benchmarks

This document records the benchmark results for the `LoopAgent`.

## Benchmark: Run Latency (10 Iterations, Trivial Sub-Agent)

Measures the end-to-end latency and overhead of the `LoopAgent` when running 10 iterations of a sub-agent that does minimal work (increments a counter in the state). This highlights the cost of the loop control, condition checking, and state cloning per iteration.

**Command:**

```bash
go test -C "C:\Users\kkushwaha\work\agentflow" -bench=BenchmarkLoopAgent_Run -benchmem -run=^$ ./internal/agents -benchtime=1s
```

**Results:**

```
goos: windows
goarch: amd64
pkg: kunalkushwaha/agentflow/internal/agents
cpu: AMD Ryzen 7 PRO 7840U w/ Radeon 780M Graphics
BenchmarkLoopAgent_Run-16         215462              5174 ns/op            9552 B/op         94 allocs/op
PASS
ok      kunalkushwaha/agentflow/internal/agents 3.791s
```

**Analysis:**

*   Running 10 iterations of a trivial sub-agent takes approximately 5,174 ns (5.2 microseconds) end-to-end.
*   This includes the loop setup, 10 calls to the sub-agent's `Run` method, 10 state clones (`currentState.Clone()`), 10 condition function evaluations, and context checks.
*   The average overhead per iteration is roughly 517 ns (5174 ns / 10 iterations).
*   Memory allocation is 9.5 KB and 94 allocations per 10 iterations (roughly 955 bytes and 9.4 allocations per iteration). This is primarily driven by the state cloning (`currentState.Clone()`) within the loop and potentially allocations within the `CounterAgent` and condition function.
*   The per-iteration overhead (517 ns, 9.4 allocs) is higher than the `SequentialAgent`'s per-agent overhead (~110 ns, ~2-3 allocs) due to the added logic for loop control (iteration check, condition function call) within each cycle.