# Adding Features to AgentFlow

This guide walks through the process of adding new features to AgentFlow, from design to implementation to testing and documentation.

## 🎯 Feature Development Philosophy

AgentFlow follows these principles for feature development:

- **User-Centric**: Features should solve real user problems
- **API-First**: Design public APIs before implementation
- **Backward Compatibility**: Maintain compatibility when possible
- **Performance-Aware**: Consider performance implications
- **Test-Driven**: Write tests alongside code
- **Documentation-Complete**: Include comprehensive documentation

## 📋 Feature Development Process

### 1. Feature Proposal Phase

#### Create Feature Request
Start with a GitHub issue using the feature request template:

```markdown
## Feature Request: Advanced Agent Chaining

### Problem Statement
Users need to chain multiple agents in complex workflows where the output of one agent becomes the input of the next, with conditional branching and error handling.

### Proposed Solution
Implement an `AgentChain` component that allows:
- Sequential agent execution
- Conditional branching based on agent results
- Error handling and fallback strategies
- State passing between agents

### Use Cases
1. Research workflow: Search → Analyze → Summarize
2. Content workflow: Generate → Review → Publish
3. Data workflow: Extract → Transform → Load

### Success Criteria
- Agents can be chained declaratively
- Conditional logic can be expressed clearly
- Error handling is robust
- Performance is comparable to manual orchestration
```

#### Initial Design Discussion
- Post in GitHub Discussions for community feedback
- Discuss with core maintainers
- Consider alternatives and trade-offs
- Define scope and non-goals

### 2. Design Phase

#### API Design Document
Create a detailed design document:

```markdown
# Agent Chaining Feature Design

## Public API

### AgentChain Interface
```go
type AgentChain interface {
    // AddStep adds an agent to the chain
    AddStep(step ChainStep) AgentChain
    
    // AddConditionalStep adds a conditional step
    AddConditionalStep(condition Condition, step ChainStep) AgentChain
    
    // Execute runs the entire chain
    Execute(ctx context.Context, input ChainInput) (ChainResult, error)
    
    // SetErrorHandler sets global error handling strategy
    SetErrorHandler(handler ErrorHandler) AgentChain
}

type ChainStep struct {
    Name        string
    Agent       AgentHandler
    InputMapper InputMapper
    Condition   Condition
    OnError     ErrorAction
}

type ChainInput struct {
    Event Event
    State State
}

type ChainResult struct {
    Steps   []StepResult
    FinalState State
    Success bool
}
```

### Core Implementation Structure
- `core/agent_chain.go` - Public interface
- `internal/chain/` - Implementation details
- `internal/chain/executor.go` - Chain execution logic
- `internal/chain/condition.go` - Conditional logic
- `internal/chain/mapper.go` - Input/output mapping

### Configuration Integration
```toml
[agent_chain]
max_steps = 50
step_timeout = "30s"
enable_parallel_execution = true
```
```

### 3. Implementation Phase

#### Step 1: Create Public Interface

Start with the public API in `core/`:

```go
// core/agent_chain.go
package core

import (
    "context"
    "time"
)

// AgentChain defines the interface for chaining multiple agents
type AgentChain interface {
    AddStep(step ChainStep) AgentChain
    AddConditionalStep(condition Condition, step ChainStep) AgentChain
    Execute(ctx context.Context, input ChainInput) (ChainResult, error)
    SetErrorHandler(handler ErrorHandler) AgentChain
}

// ChainStep represents a single step in an agent chain
type ChainStep struct {
    Name        string
    Agent       AgentHandler
    InputMapper InputMapper
    Condition   Condition
    OnError     ErrorAction
    Timeout     time.Duration
}

// NewAgentChain creates a new agent chain
func NewAgentChain(name string) AgentChain {
    return internal.NewAgentChain(name)
}
```

#### Step 2: Implement Core Logic

Create the implementation in `internal/`:

```go
// internal/chain/agent_chain.go
package chain

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/kunalkushwaha/agentflow/core"
)

type agentChain struct {
    name         string
    steps        []core.ChainStep
    errorHandler core.ErrorHandler
    config       ChainConfig
    mu           sync.RWMutex
}

type ChainConfig struct {
    MaxSteps              int
    StepTimeout           time.Duration
    EnableParallelExecution bool
}

func NewAgentChain(name string) core.AgentChain {
    return &agentChain{
        name:   name,
        steps:  make([]core.ChainStep, 0),
        config: DefaultChainConfig(),
    }
}

func (c *agentChain) AddStep(step core.ChainStep) core.AgentChain {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.steps = append(c.steps, step)
    return c
}

func (c *agentChain) Execute(ctx context.Context, input core.ChainInput) (core.ChainResult, error) {
    executor := NewChainExecutor(c.config)
    return executor.Execute(ctx, c.steps, input, c.errorHandler)
}
```

#### Step 3: Add Execution Logic

```go
// internal/chain/executor.go
package chain

import (
    "context"
    "fmt"
    "time"
    
    "github.com/kunalkushwaha/agentflow/core"
)

type ChainExecutor struct {
    config ChainConfig
}

func NewChainExecutor(config ChainConfig) *ChainExecutor {
    return &ChainExecutor{config: config}
}

func (e *ChainExecutor) Execute(ctx context.Context, steps []core.ChainStep, input core.ChainInput, errorHandler core.ErrorHandler) (core.ChainResult, error) {
    result := core.ChainResult{
        Steps:      make([]core.StepResult, 0, len(steps)),
        FinalState: input.State.Clone(),
        Success:    true,
    }
    
    currentState := input.State.Clone()
    
    for i, step := range steps {
        // Check context cancellation
        select {
        case <-ctx.Done():
            return result, ctx.Err()
        default:
        }
        
        // Evaluate condition if present
        if step.Condition != nil && !step.Condition.Evaluate(currentState) {
            continue
        }
        
        // Execute step with timeout
        stepCtx, cancel := context.WithTimeout(ctx, e.getStepTimeout(step))
        stepResult, err := e.executeStep(stepCtx, step, input.Event, currentState)
        cancel()
        
        // Handle step result
        result.Steps = append(result.Steps, stepResult)
        
        if err != nil {
            if errorHandler != nil {
                action := errorHandler.HandleError(err, step, currentState)
                if action == core.ErrorActionStop {
                    result.Success = false
                    return result, fmt.Errorf("chain stopped at step %d: %w", i, err)
                }
                // Continue with ErrorActionContinue
            } else {
                result.Success = false
                return result, fmt.Errorf("step %d failed: %w", i, err)
            }
        } else {
            // Update state with step result
            if stepResult.State != nil {
                currentState = stepResult.State
            }
        }
    }
    
    result.FinalState = currentState
    return result, nil
}

func (e *ChainExecutor) executeStep(ctx context.Context, step core.ChainStep, event core.Event, state core.State) (core.StepResult, error) {
    // Map input if mapper is provided
    stepEvent := event
    stepState := state
    
    if step.InputMapper != nil {
        var err error
        stepEvent, stepState, err = step.InputMapper.Map(event, state)
        if err != nil {
            return core.StepResult{}, fmt.Errorf("input mapping failed: %w", err)
        }
    }
    
    // Execute agent
    agentResult, err := step.Agent.Run(ctx, stepEvent, stepState)
    if err != nil {
        return core.StepResult{
            StepName: step.Name,
            Success:  false,
            Error:    err,
        }, err
    }
    
    return core.StepResult{
        StepName: step.Name,
        Success:  true,
        Data:     agentResult.Data,
        State:    agentResult.State,
    }, nil
}
```

#### Step 4: Add Builder Pattern Support

```go
// core/agent_chain_builder.go
package core

// AgentChainBuilder provides a fluent interface for building agent chains
type AgentChainBuilder struct {
    chain AgentChain
}

// NewAgentChainBuilder creates a new builder
func NewAgentChainBuilder(name string) *AgentChainBuilder {
    return &AgentChainBuilder{
        chain: NewAgentChain(name),
    }
}

// Step adds a simple step to the chain
func (b *AgentChainBuilder) Step(name string, agent AgentHandler) *AgentChainBuilder {
    b.chain.AddStep(ChainStep{
        Name:  name,
        Agent: agent,
    })
    return b
}

// ConditionalStep adds a conditional step
func (b *AgentChainBuilder) ConditionalStep(name string, agent AgentHandler, condition Condition) *AgentChainBuilder {
    b.chain.AddConditionalStep(condition, ChainStep{
        Name:  name,
        Agent: agent,
    })
    return b
}

// WithErrorHandler sets the error handling strategy
func (b *AgentChainBuilder) WithErrorHandler(handler ErrorHandler) *AgentChainBuilder {
    b.chain.SetErrorHandler(handler)
    return b
}

// Build returns the configured chain
func (b *AgentChainBuilder) Build() AgentChain {
    return b.chain
}
```

### 4. Testing Phase

#### Unit Tests

```go
// core/agent_chain_test.go
package core

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAgentChain_Execute(t *testing.T) {
    tests := []struct {
        name     string
        steps    []ChainStep
        input    ChainInput
        expected ChainResult
        wantErr  bool
    }{
        {
            name: "Simple two-step chain",
            steps: []ChainStep{
                {
                    Name:  "step1",
                    Agent: &mockAgent{response: "result1"},
                },
                {
                    Name:  "step2", 
                    Agent: &mockAgent{response: "result2"},
                },
            },
            input: ChainInput{
                Event: NewEvent("test", map[string]interface{}{"query": "test"}),
                State: NewState(),
            },
            expected: ChainResult{
                Success: true,
            },
            wantErr: false,
        },
        {
            name: "Chain with error in middle step",
            steps: []ChainStep{
                {
                    Name:  "step1",
                    Agent: &mockAgent{response: "result1"},
                },
                {
                    Name:  "step2",
                    Agent: &mockAgent{err: fmt.Errorf("step error")},
                },
            },
            input: ChainInput{
                Event: NewEvent("test", map[string]interface{}{"query": "test"}),
                State: NewState(),
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            chain := NewAgentChain("test-chain")
            for _, step := range tt.steps {
                chain.AddStep(step)
            }
            
            result, err := chain.Execute(context.Background(), tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.expected.Success, result.Success)
        })
    }
}

func TestAgentChainBuilder(t *testing.T) {
    agent1 := &mockAgent{response: "result1"}
    agent2 := &mockAgent{response: "result2"}
    
    chain := NewAgentChainBuilder("test-chain").
        Step("step1", agent1).
        Step("step2", agent2).
        WithErrorHandler(&mockErrorHandler{}).
        Build()
    
    assert.NotNil(t, chain)
    
    input := ChainInput{
        Event: NewEvent("test", map[string]interface{}{"query": "test"}),
        State: NewState(),
    }
    
    result, err := chain.Execute(context.Background(), input)
    require.NoError(t, err)
    assert.True(t, result.Success)
    assert.Len(t, result.Steps, 2)
}

// Mock implementations for testing
type mockAgent struct {
    response string
    err      error
}

func (m *mockAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    if m.err != nil {
        return AgentResult{}, m.err
    }
    
    return AgentResult{
        Data: map[string]interface{}{
            "result": m.response,
        },
        Success: true,
        State:   state,
    }, nil
}
```

#### Integration Tests

```go
// integration/agent_chain_integration_test.go
//go:build integration
// +build integration

package integration

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/kunalkushwaha/agentflow/core"
)

func TestAgentChain_RealWorkflow(t *testing.T) {
    // Setup real agents
    searchAgent := &SearchAgent{
        provider: newTestSearchProvider(),
    }
    
    analysisAgent := &AnalysisAgent{
        llm: newTestLLMProvider(),
    }
    
    summaryAgent := &SummaryAgent{
        llm: newTestLLMProvider(),
    }
    
    // Create chain
    chain := core.NewAgentChainBuilder("research-workflow").
        Step("search", searchAgent).
        Step("analyze", analysisAgent).
        Step("summarize", summaryAgent).
        Build()
    
    // Execute chain
    input := core.ChainInput{
        Event: core.NewEvent("research", map[string]interface{}{
            "topic": "artificial intelligence trends 2024",
        }),
        State: core.NewState(),
    }
    
    result, err := chain.Execute(context.Background(), input)
    require.NoError(t, err)
    
    assert.True(t, result.Success)
    assert.Len(t, result.Steps, 3)
    assert.Contains(t, result.FinalState.GetString("summary"), "artificial intelligence")
}
```

### 5. Documentation Phase

#### API Documentation

```markdown
# Agent Chaining

Agent chaining allows you to create sophisticated workflows by connecting multiple agents in sequence, with conditional logic and error handling.

## Basic Usage

```go
import "github.com/kunalkushwaha/agentflow/core"

// Create agents
searchAgent := &SearchAgent{}
analysisAgent := &AnalysisAgent{}
summaryAgent := &SummaryAgent{}

// Build chain
chain := core.NewAgentChainBuilder("research-workflow").
    Step("search", searchAgent).
    Step("analyze", analysisAgent).
    Step("summarize", summaryAgent).
    Build()

// Execute chain
input := core.ChainInput{
    Event: core.NewEvent("research", map[string]interface{}{
        "topic": "AI trends",
    }),
    State: core.NewState(),
}

result, err := chain.Execute(context.Background(), input)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Final result: %v\n", result.FinalState)
```

## Advanced Features

### Conditional Steps

```go
chain := core.NewAgentChainBuilder("conditional-workflow").
    Step("initial", initialAgent).
    ConditionalStep("optional", optionalAgent, 
        core.StateCondition("need_extra_processing", true)).
    Step("final", finalAgent).
    Build()
```

### Error Handling

```go
errorHandler := &core.ContinueOnErrorHandler{
    MaxErrors: 2,
    LogErrors: true,
}

chain := core.NewAgentChainBuilder("resilient-workflow").
    Step("step1", agent1).
    Step("step2", agent2).
    WithErrorHandler(errorHandler).
    Build()
```
```

#### User Guide Update

Add section to `docs/guides/AgentBasics.md`:

```markdown
## Agent Chaining

For complex workflows requiring multiple agents, use AgentChain:

### Creating a Chain

```go
chain := core.NewAgentChainBuilder("my-workflow").
    Step("search", searchAgent).
    Step("analyze", analysisAgent).
    Step("summarize", summaryAgent).
    Build()
```

### Executing the Chain

```go
result, err := chain.Execute(ctx, core.ChainInput{
    Event: event,
    State: state,
})
```

The chain will execute each step in sequence, passing state between steps.
```

### 6. Configuration Integration

#### Add Configuration Support

```go
// core/config.go - Add to existing config

type Config struct {
    // ... existing fields ...
    
    AgentChain AgentChainConfig `toml:"agent_chain"`
}

type AgentChainConfig struct {
    MaxSteps              int           `toml:"max_steps"`
    StepTimeout           time.Duration `toml:"step_timeout"`
    EnableParallelSteps   bool          `toml:"enable_parallel_steps"`
    DefaultErrorStrategy  string        `toml:"default_error_strategy"`
}

func (c AgentChainConfig) Validate() error {
    if c.MaxSteps < 1 {
        return fmt.Errorf("max_steps must be at least 1")
    }
    if c.StepTimeout < time.Second {
        return fmt.Errorf("step_timeout must be at least 1 second")
    }
    return nil
}
```

#### Update Default Configuration

```toml
# Default agentflow.toml additions
[agent_chain]
max_steps = 50
step_timeout = "30s"
enable_parallel_steps = false
default_error_strategy = "stop"
```

### 7. CLI Integration

#### Add CLI Commands

```go
// cmd/agentcli/cmd/chain.go
package cmd

import (
    "github.com/spf13/cobra"
)

var chainCmd = &cobra.Command{
    Use:   "chain",
    Short: "Manage agent chains",
    Long:  "Commands for creating and managing agent chains",
}

var chainCreateCmd = &cobra.Command{
    Use:   "create <name>",
    Short: "Create a new agent chain",
    Args:  cobra.ExactArgs(1),
    RunE:  runChainCreate,
}

var chainRunCmd = &cobra.Command{
    Use:   "run <chain-name>",
    Short: "Execute an agent chain",
    Args:  cobra.ExactArgs(1),
    RunE:  runChainRun,
}

func init() {
    chainCmd.AddCommand(chainCreateCmd)
    chainCmd.AddCommand(chainRunCmd)
    rootCmd.AddCommand(chainCmd)
}
```

## 🔄 Feature Integration Checklist

### Pre-Implementation
- [ ] Feature request created and discussed
- [ ] Design document written and reviewed
- [ ] API design approved by maintainers
- [ ] Breaking changes identified and documented

### Implementation
- [ ] Public API implemented in `core/`
- [ ] Implementation details in `internal/`
- [ ] Configuration integration added
- [ ] Error handling implemented
- [ ] Performance considerations addressed

### Testing
- [ ] Unit tests written and passing
- [ ] Integration tests written and passing
- [ ] Benchmarks created for performance-critical paths
- [ ] Mock implementations created for testing

### Documentation
- [ ] API documentation written
- [ ] User guide updated
- [ ] Examples created and tested
- [ ] Migration guide written (if breaking changes)

### CLI Integration
- [ ] CLI commands added (if applicable)
- [ ] Help text and examples provided
- [ ] Shell completion updated

### Quality Assurance
- [ ] Code review completed
- [ ] Security review completed (if applicable)
- [ ] Performance testing completed
- [ ] Manual testing completed

### Release Preparation
- [ ] CHANGELOG.md updated
- [ ] Version compatibility documented
- [ ] Release notes drafted
- [ ] Deprecation notices added (if applicable)

## 📚 Feature Examples

### Simple Feature: Add Timeout to Agent Execution

```go
// 1. Extend AgentHandler interface
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
    GetTimeout() time.Duration // New method
}

// 2. Update Runner to use timeout
func (r *Runner) executeAgent(ctx context.Context, agent AgentHandler, event Event, state State) (AgentResult, error) {
    timeout := agent.GetTimeout()
    if timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, timeout)
        defer cancel()
    }
    
    return agent.Run(ctx, event, state)
}

// 3. Provide default implementation
type BaseAgent struct {
    timeout time.Duration
}

func (a *BaseAgent) GetTimeout() time.Duration {
    if a.timeout > 0 {
        return a.timeout
    }
    return 30 * time.Second // default
}
```

### Complex Feature: Agent Middleware System

```go
// 1. Define middleware interface
type Middleware interface {
    Handle(next AgentHandler) AgentHandler
}

// 2. Create middleware chain
type MiddlewareChain struct {
    middlewares []Middleware
}

func (c *MiddlewareChain) Then(handler AgentHandler) AgentHandler {
    for i := len(c.middlewares) - 1; i >= 0; i-- {
        handler = c.middlewares[i].Handle(handler)
    }
    return handler
}

// 3. Integrate with Runner
func (r *Runner) RegisterAgentWithMiddleware(name string, handler AgentHandler, middlewares ...Middleware) {
    chain := &MiddlewareChain{middlewares: middlewares}
    wrappedHandler := chain.Then(handler)
    r.RegisterAgent(name, wrappedHandler)
}
```

This comprehensive guide provides the framework for adding any feature to AgentFlow while maintaining code quality, performance, and user experience standards.
