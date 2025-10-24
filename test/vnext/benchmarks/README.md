# vnext Streaming Benchmarks

This directory contains performance benchmarks for the vnext workflow streaming implementation.

## Benchmarks

### `streaming_overhead.go`
Compares the performance characteristics of different execution modes:

1. **Direct Agent Streaming** - Single agent with streaming output
2. **Direct Agent Non-Streaming** - Single agent with batch output  
3. **Workflow Streaming** - Multi-agent workflow with streaming output
4. **Workflow Non-Streaming** - Multi-agent workflow with batch output

## Running Benchmarks

### Prerequisites
- Ollama running on `localhost:11434`
- `gemma3:1b` model available (`ollama pull gemma3:1b`)

### Execution
```bash
# From this directory
go run streaming_overhead.go

# Or use the convenience scripts in examples/vnext/streaming_workflow/
./run_benchmark.sh    # Linux/Mac
./run_benchmark.ps1   # Windows
```

## Metrics Measured

- **Duration**: Total execution time
- **Tokens/second**: Processing rate (estimated)
- **Chunks/second**: Stream chunks per second
- **MB/second**: Data throughput
- **Overhead percentages**: Comparative performance impact

## Understanding Results

- **Streaming Overhead**: Cost of real-time token streaming vs batch processing
- **Workflow Overhead**: Cost of multi-agent coordination vs direct execution  
- **Total Overhead**: Combined cost of workflow + streaming

### Expected Results
- Streaming adds 5-15% overhead for real-time UX benefits
- Workflow adds 10-25% overhead for multi-agent orchestration
- Combined overhead should be acceptable (<40%) for most use cases

Results vary based on:
- LLM response speed and latency
- Network conditions  
- System resources and concurrent load
- Agent complexity and workflow size