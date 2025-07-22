# Common Orchestration Patterns

## Overview

This guide explores common orchestration patterns that solve real-world problems in multi-agent systems. These patterns combine the basic orchestration modes (route, collaborative, sequential, loop, mixed) into proven solutions for specific use cases.

Understanding these patterns helps you design effective agent workflows without reinventing the wheel. Each pattern includes implementation examples, configuration templates, and guidance on when to use them.

## Prerequisites

- Understanding of all basic orchestration modes
- Familiarity with [State Management](../core-concepts/state-management.md)
- Knowledge of [Error Handling](../core-concepts/error-handling.md)
- Experience with AgenticGoKit configuration

## Pattern Categories

### 1. Information Processing Patterns
- Research and Analysis Pipeline
- Content Creation Workflow
- Data Processing Pipeline
- Knowledge Extraction Pipeline

### 2. Decision Making Patterns
- Multi-Criteria Decision Making
- Consensus Building
- Expert Panel Review
- Risk Assessment Pipeline

### 3. Problem Solving Patterns
- Iterative Refinement
- Divide and Conquer
- Brainstorm and Filter
- Hypothesis Testing

### 4. Quality Assurance Patterns
- Review and Approval
- Multi-Stage Validation
- Continuous Improvement
- Error Detection and Correction

## Information Processing Patterns

### Research and Analysis Pipeline

**Problem**: Need to gather information from multiple sources, analyze it systematically, and produce comprehensive reports.

**Solution**: Collaborative research → Sequential analysis → Loop refinement → Route delivery

```go
// Research and Analysis Pipeline Pattern
func CreateResearchAnalysisPipeline() *MixedOrchestrator {
    return &MixedOrchestrator{
        stages: []OrchestrationStage{
            // Stage 1: Collaborative Research
            {
                Name:    "research",
                Pattern: PatternCollaborative,
                Agents:  []string{"web-researcher", "academic-researcher", "expert-researcher"},
                Config: StageConfig{
                    CollaborativeTimeout: 90 * time.Second,
                    FailureThreshold:     0.6, // Continue if 60% succeed
                },
                Timeout: 120 * time.Second,
            },
            // Stage 2: Sequential Analysis
            {
                Name:    "analysis",
                Pattern: PatternSequential,
                Agents:  []string{"data-analyzer", "trend-analyzer", "insight-generator"},
                Config: StageConfig{
                    SequentialOrder: []string{"data-analyzer", "trend-analyzer", "insight-generator"},
                },
                Timeout: 180 * time.Second,
            },
            // Stage 3: Loop Refinement
            {
                Name:    "refinement",
                Pattern: PatternLoop,
                Agents:  []string{"content-refiner"},
                Config: StageConfig{
                    LoopAgent:     "content-refiner",
                    MaxIterations: 3,
                    LoopCondition: func(state core.State) bool {
                        if quality, ok := state.Get("quality_score"); ok {
                            return quality.(float64) >= 0.85
                        }
                        return false
                    },
                },
                Timeout: 90 * time.Second,
            },
            // Stage 4: Route Delivery
            {
                Name:    "delivery",
                Pattern: PatternRoute,
                Agents:  []string{"report-formatter"},
                Config: StageConfig{
                    RouteTarget: "report-formatter",
                },
                Timeout: 30 * time.Second,
            },
        },
    }
}
```

**Configuration**:
```toml
[orchestration]
mode = "mixed"

[[orchestration.mixed.stages]]
name = "research"
pattern = "collaborative"
agents = ["web-researcher", "academic-researcher", "expert-researcher"]
timeout = "120s"

[orchestration.mixed.stages.config]
failure_threshold = 0.6

[[orchestration.mixed.stages]]
name = "analysis"
pattern = "sequential"
agents = ["data-analyzer", "trend-analyzer", "insight-generator"]
timeout = "180s"

[[orchestration.mixed.stages]]
name = "refinement"
pattern = "loop"
agents = ["content-refiner"]
timeout = "90s"

[orchestration.mixed.stages.config]
max_iterations = 3
loop_condition = "quality_score >= 0.85"

[[orchestration.mixed.stages]]
name = "delivery"
pattern = "route"
agents = ["report-formatter"]
timeout = "30s"
```

**When to Use**:
- Academic research projects
- Market analysis reports
- Technical documentation
- Competitive intelligence
- Due diligence processes

### Content Creation Workflow

**Problem**: Create high-quality content through collaborative ideation, structured writing, and iterative improvement.

**Solution**: Collaborative brainstorming → Sequential creation → Loop review → Route publishing

```go
func CreateContentCreationWorkflow() *MixedOrchestrator {
    return &MixedOrchestrator{
        stages: []OrchestrationStage{
            // Stage 1: Collaborative Brainstorming
            {
                Name:    "brainstorm",
                Pattern: PatternCollaborative,
                Agents:  []string{"creative-writer", "technical-writer", "editor"},
                Config: StageConfig{
                    CollaborativeTimeout: 60 * time.Second,
                    FailureThreshold:     0.5,
                },
                Timeout: 90 * time.Second,
            },
            // Stage 2: Sequential Creation
            {
                Name:    "creation",
                Pattern: PatternSequential,
                Agents:  []string{"outline-creator", "content-writer", "fact-checker"},
                Config: StageConfig{
                    SequentialOrder: []string{"outline-creator", "content-writer", "fact-checker"},
                },
                Timeout: 240 * time.Second,
            },
            // Stage 3: Loop Review and Revision
            {
                Name:    "review",
                Pattern: PatternLoop,
                Agents:  []string{"content-reviewer"},
                Config: StageConfig{
                    LoopAgent:     "content-reviewer",
                    MaxIterations: 5,
                    LoopCondition: func(state core.State) bool {
                        if approved, ok := state.Get("review_approved"); ok {
                            return approved.(bool)
                        }
                        return false
                    },
                },
                Timeout: 150 * time.Second,
            },
            // Stage 4: Route Publishing
            {
                Name:    "publish",
                Pattern: PatternRoute,
                Agents:  []string{"content-publisher"},
                Config: StageConfig{
                    RouteTarget: "content-publisher",
                },
                Timeout: 30 * time.Second,
            },
        },
    }
}
```

**When to Use**:
- Blog post creation
- Marketing content development
- Technical documentation
- Social media campaigns
- Educational materials

### Data Processing Pipeline

**Problem**: Process large datasets through multiple transformation and validation stages.

**Solution**: Route classification → Collaborative processing → Sequential validation → Loop error correction

```go
func CreateDataProcessingPipeline() *MixedOrchestrator {
    return &MixedOrchestrator{
        stages: []OrchestrationStage{
            // Stage 1: Route Classification
            {
                Name:    "classify",
                Pattern: PatternRoute,
                Agents:  []string{"data-classifier"},
                Config: StageConfig{
                    RouteTarget: "data-classifier",
                },
                Timeout: 30 * time.Second,
            },
            // Stage 2: Collaborative Processing
            {
                Name:    "process",
                Pattern: PatternCollaborative,
                Agents:  []string{"data-cleaner", "data-transformer", "data-enricher"},
                Config: StageConfig{
                    CollaborativeTimeout: 120 * time.Second,
                    FailureThreshold:     0.7,
                },
                Timeout: 180 * time.Second,
            },
            // Stage 3: Sequential Validation
            {
                Name:    "validate",
                Pattern: PatternSequential,
                Agents:  []string{"schema-validator", "quality-checker", "completeness-validator"},
                Config: StageConfig{
                    SequentialOrder: []string{"schema-validator", "quality-checker", "completeness-validator"},
                },
                Timeout: 90 * time.Second,
            },
            // Stage 4: Loop Error Correction
            {
                Name:    "correct",
                Pattern: PatternLoop,
                Agents:  []string{"error-corrector"},
                Config: StageConfig{
                    LoopAgent:     "error-corrector",
                    MaxIterations: 3,
                    LoopCondition: func(state core.State) bool {
                        if errorCount, ok := state.Get("error_count"); ok {
                            return errorCount.(int) == 0
                        }
                        return true
                    },
                },
                Timeout: 120 * time.Second,
            },
        },
    }
}
```

**When to Use**:
- ETL processes
- Data migration projects
- Real-time data processing
- Data quality improvement
- Batch processing workflows

## Decision Making Patterns

### Multi-Criteria Decision Making

**Problem**: Make complex decisions by evaluating multiple criteria and perspectives.

**Solution**: Collaborative evaluation → Sequential scoring → Route decision → Loop validation

```go
func CreateDecisionMakingPipeline() *MixedOrchestrator {
    return &MixedOrchestrator{
        stages: []OrchestrationStage{
            // Stage 1: Collaborative Evaluation
            {
                Name:    "evaluate",
                Pattern: PatternCollaborative,
                Agents:  []string{"cost-evaluator", "risk-evaluator", "benefit-evaluator"},
                Config: StageConfig{
                    CollaborativeTimeout: 90 * time.Second,
                    FailureThreshold:     0.8, // Need high success rate for decisions
                },
                Timeout: 120 * time.Second,
            },
            // Stage 2: Sequential Scoring
            {
                Name:    "score",
                Pattern: PatternSequential,
                Agents:  []string{"criteria-scorer", "weight-calculator", "final-scorer"},
                Config: StageConfig{
                    SequentialOrder: []string{"criteria-scorer", "weight-calculator", "final-scorer"},
                },
                Timeout: 60 * time.Second,
            },
            // Stage 3: Route Decision
            {
                Name:    "decide",
                Pattern: PatternRoute,
                Agents:  []string{"decision-maker"},
                Config: StageConfig{
                    RouteTarget: "decision-maker",
                },
                Timeout: 30 * time.Second,
            },
            // Stage 4: Loop Validation
            {
                Name:    "validate",
                Pattern: PatternLoop,
                Agents:  []string{"decision-validator"},
                Config: StageConfig{
                    LoopAgent:     "decision-validator",
                    MaxIterations: 2,
                    LoopCondition: func(state core.State) bool {
                        if confidence, ok := state.Get("decision_confidence"); ok {
                            return confidence.(float64) >= 0.8
                        }
                        return false
                    },
                },
                Timeout: 60 * time.Second,
            },
        },
    }
}
```

**When to Use**:
- Investment decisions
- Vendor selection
- Strategic planning
- Resource allocation
- Policy decisions

### Consensus Building

**Problem**: Build consensus among multiple stakeholders with different perspectives.

**Solution**: Collaborative discussion → Loop negotiation → Route final decision

```go
func CreateConsensusBuildingPipeline() *MixedOrchestrator {
    return &MixedOrchestrator{
        stages: []OrchestrationStage{
            // Stage 1: Collaborative Discussion
            {
                Name:    "discuss",
                Pattern: PatternCollaborative,
                Agents:  []string{"stakeholder-1", "stakeholder-2", "stakeholder-3", "mediator"},
                Config: StageConfig{
                    CollaborativeTimeout: 180 * time.Second,
                    FailureThreshold:     0.75,
                },
                Timeout: 240 * time.Second,
            },
            // Stage 2: Loop Negotiation
            {
                Name:    "negotiate",
                Pattern: PatternLoop,
                Agents:  []string{"negotiator"},
                Config: StageConfig{
                    LoopAgent:     "negotiator",
                    MaxIterations: 5,
                    LoopCondition: func(state core.State) bool {
                        if consensus, ok := state.Get("consensus_reached"); ok {
                            return consensus.(bool)
                        }
                        return false
                    },
                },
                Timeout: 300 * time.Second,
            },
            // Stage 3: Route Final Decision
            {
                Name:    "finalize",
                Pattern: PatternRoute,
                Agents:  []string{"decision-finalizer"},
                Config: StageConfig{
                    RouteTarget: "decision-finalizer",
                },
                Timeout: 30 * time.Second,
            },
        },
    }
}
```

**When to Use**:
- Team decision making
- Committee decisions
- Conflict resolution
- Policy development
- Collaborative planning

## Problem Solving Patterns

### Iterative Refinement

**Problem**: Solve complex problems through iterative improvement and refinement.

**Solution**: Route problem analysis → Loop solution generation → Collaborative evaluation → Route implementation

```go
func CreateIterativeRefinementPipeline() *MixedOrchestrator {
    return &MixedOrchestrator{
        stages: []OrchestrationStage{
            // Stage 1: Route Problem Analysis
            {
                Name:    "analyze",
                Pattern: PatternRoute,
                Agents:  []string{"problem-analyzer"},
                Config: StageConfig{
                    RouteTarget: "problem-analyzer",
                },
                Timeout: 60 * time.Second,
            },
            // Stage 2: Loop Solution Generation
            {
                Name:    "generate",
                Pattern: PatternLoop,
                Agents:  []string{"solution-generator"},
                Config: StageConfig{
                    LoopAgent:     "solution-generator",
                    MaxIterations: 5,
                    LoopCondition: func(state core.State) bool {
                        if solutionQuality, ok := state.Get("solution_quality"); ok {
                            return solutionQuality.(float64) >= 0.8
                        }
                        return false
                    },
                },
                Timeout: 240 * time.Second,
            },
            // Stage 3: Collaborative Evaluation
            {
                Name:    "evaluate",
                Pattern: PatternCollaborative,
                Agents:  []string{"feasibility-evaluator", "cost-evaluator", "risk-evaluator"},
                Config: StageConfig{
                    CollaborativeTimeout: 90 * time.Second,
                    FailureThreshold:     0.7,
                },
                Timeout: 120 * time.Second,
            },
            // Stage 4: Route Implementation
            {
                Name:    "implement",
                Pattern: PatternRoute,
                Agents:  []string{"solution-implementer"},
                Config: StageConfig{
                    RouteTarget: "solution-implementer",
                },
                Timeout: 60 * time.Second,
            },
        },
    }
}
```

**When to Use**:
- Software debugging
- Process optimization
- Creative problem solving
- Research and development
- Continuous improvement

### Divide and Conquer

**Problem**: Solve large, complex problems by breaking them into smaller, manageable parts.

**Solution**: Route decomposition → Collaborative parallel solving → Sequential integration → Route validation

```go
func CreateDivideAndConquerPipeline() *MixedOrchestrator {
    return &MixedOrchestrator{
        stages: []OrchestrationStage{
            // Stage 1: Route Decomposition
            {
                Name:    "decompose",
                Pattern: PatternRoute,
                Agents:  []string{"problem-decomposer"},
                Config: StageConfig{
                    RouteTarget: "problem-decomposer",
                },
                Timeout: 60 * time.Second,
            },
            // Stage 2: Collaborative Parallel Solving
            {
                Name:    "solve",
                Pattern: PatternCollaborative,
                Agents:  []string{"sub-solver-1", "sub-solver-2", "sub-solver-3"},
                Config: StageConfig{
                    CollaborativeTimeout: 180 * time.Second,
                    FailureThreshold:     0.8, // Need most sub-problems solved
                },
                Timeout: 240 * time.Second,
            },
            // Stage 3: Sequential Integration
            {
                Name:    "integrate",
                Pattern: PatternSequential,
                Agents:  []string{"solution-combiner", "consistency-checker", "optimizer"},
                Config: StageConfig{
                    SequentialOrder: []string{"solution-combiner", "consistency-checker", "optimizer"},
                },
                Timeout: 120 * time.Second,
            },
            // Stage 4: Route Validation
            {
                Name:    "validate",
                Pattern: PatternRoute,
                Agents:  []string{"solution-validator"},
                Config: StageConfig{
                    RouteTarget: "solution-validator",
                },
                Timeout: 60 * time.Second,
            },
        },
    }
}
```

**When to Use**:
- Large-scale system design
- Complex algorithm development
- Project management
- Distributed computing problems
- Modular development

## Quality Assurance Patterns

### Review and Approval

**Problem**: Ensure quality through systematic review and approval processes.

**Solution**: Route initial review → Collaborative multi-reviewer → Loop revision → Route final approval

```go
func CreateReviewApprovalPipeline() *MixedOrchestrator {
    return &MixedOrchestrator{
        stages: []OrchestrationStage{
            // Stage 1: Route Initial Review
            {
                Name:    "initial-review",
                Pattern: PatternRoute,
                Agents:  []string{"initial-reviewer"},
                Config: StageConfig{
                    RouteTarget: "initial-reviewer",
                },
                Timeout: 60 * time.Second,
            },
            // Stage 2: Collaborative Multi-Reviewer
            {
                Name:    "multi-review",
                Pattern: PatternCollaborative,
                Agents:  []string{"technical-reviewer", "content-reviewer", "compliance-reviewer"},
                Config: StageConfig{
                    CollaborativeTimeout: 120 * time.Second,
                    FailureThreshold:     0.8,
                },
                Timeout: 180 * time.Second,
            },
            // Stage 3: Loop Revision
            {
                Name:    "revise",
                Pattern: PatternLoop,
                Agents:  []string{"content-reviser"},
                Config: StageConfig{
                    LoopAgent:     "content-reviser",
                    MaxIterations: 3,
                    LoopCondition: func(state core.State) bool {
                        if allApproved, ok := state.Get("all_reviews_approved"); ok {
                            return allApproved.(bool)
                        }
                        return false
                    },
                },
                Timeout: 180 * time.Second,
            },
            // Stage 4: Route Final Approval
            {
                Name:    "final-approval",
                Pattern: PatternRoute,
                Agents:  []string{"final-approver"},
                Config: StageConfig{
                    RouteTarget: "final-approver",
                },
                Timeout: 30 * time.Second,
            },
        },
    }
}
```

**When to Use**:
- Code review processes
- Document approval workflows
- Compliance checking
- Quality assurance
- Publication workflows

### Multi-Stage Validation

**Problem**: Validate complex systems through multiple independent validation stages.

**Solution**: Sequential validation stages → Collaborative cross-validation → Loop error correction → Route certification

```go
func CreateMultiStageValidationPipeline() *MixedOrchestrator {
    return &MixedOrchestrator{
        stages: []OrchestrationStage{
            // Stage 1: Sequential Validation Stages
            {
                Name:    "validate",
                Pattern: PatternSequential,
                Agents:  []string{"syntax-validator", "logic-validator", "performance-validator"},
                Config: StageConfig{
                    SequentialOrder: []string{"syntax-validator", "logic-validator", "performance-validator"},
                },
                Timeout: 180 * time.Second,
            },
            // Stage 2: Collaborative Cross-Validation
            {
                Name:    "cross-validate",
                Pattern: PatternCollaborative,
                Agents:  []string{"validator-1", "validator-2", "validator-3"},
                Config: StageConfig{
                    CollaborativeTimeout: 120 * time.Second,
                    FailureThreshold:     0.8,
                },
                Timeout: 150 * time.Second,
            },
            // Stage 3: Loop Error Correction
            {
                Name:    "correct",
                Pattern: PatternLoop,
                Agents:  []string{"error-corrector"},
                Config: StageConfig{
                    LoopAgent:     "error-corrector",
                    MaxIterations: 3,
                    LoopCondition: func(state core.State) bool {
                        if validationPassed, ok := state.Get("validation_passed"); ok {
                            return validationPassed.(bool)
                        }
                        return false
                    },
                },
                Timeout: 120 * time.Second,
            },
            // Stage 4: Route Certification
            {
                Name:    "certify",
                Pattern: PatternRoute,
                Agents:  []string{"certifier"},
                Config: StageConfig{
                    RouteTarget: "certifier",
                },
                Timeout: 30 * time.Second,
            },
        },
    }
}
```

**When to Use**:
- Software testing
- System validation
- Compliance verification
- Security auditing
- Quality certification

## Pattern Selection Guide

### Choosing the Right Pattern

| Use Case | Recommended Pattern | Key Characteristics |
|----------|-------------------|-------------------|
| Research Projects | Research and Analysis | Multiple sources, systematic analysis |
| Content Creation | Content Creation Workflow | Creative collaboration, iterative improvement |
| Data Processing | Data Processing Pipeline | Transformation, validation, error correction |
| Decision Making | Multi-Criteria Decision | Multiple perspectives, systematic evaluation |
| Problem Solving | Iterative Refinement | Continuous improvement, quality focus |
| Quality Assurance | Review and Approval | Multiple reviewers, systematic validation |

### Pattern Customization

Most patterns can be customized by:

1. **Adjusting Stage Configuration**:
   - Timeout values
   - Failure thresholds
   - Loop conditions
   - Agent assignments

2. **Adding/Removing Stages**:
   - Skip unnecessary stages
   - Add domain-specific stages
   - Reorder stages as needed

3. **Changing Orchestration Patterns**:
   - Switch collaborative to sequential
   - Add loop stages for refinement
   - Use route for simple dispatch

### Performance Considerations

- **Collaborative stages**: Higher resource usage, better quality
- **Sequential stages**: Lower resource usage, predictable flow
- **Loop stages**: Variable duration, quality improvement
- **Route stages**: Minimal overhead, simple dispatch

## Implementation Best Practices

### 1. Configuration Management

```toml
# Use environment-specific configurations
[orchestration]
mode = "mixed"
timeout = "${WORKFLOW_TIMEOUT:300s}"

# Define reusable stage templates
[templates.research_stage]
pattern = "collaborative"
timeout = "120s"
failure_threshold = 0.6

# Reference templates in stages
[[orchestration.mixed.stages]]
name = "research"
template = "research_stage"
agents = ["web-researcher", "academic-researcher"]
```

### 2. Monitoring and Metrics

```go
// Add comprehensive monitoring
func setupPatternMonitoring(runner *core.Runner, patternName string) {
    runner.RegisterCallback(core.HookBeforeAgentRun, "pattern-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            metrics.RecordStageStart(patternName, args.AgentID)
            return args.State, nil
        },
    )
    
    runner.RegisterCallback(core.HookAfterAgentRun, "pattern-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            metrics.RecordStageCompletion(patternName, args.AgentID, args.Duration)
            return args.State, nil
        },
    )
}
```

### 3. Error Handling

```go
// Implement pattern-specific error handling
func createPatternErrorHandler(patternName string) func(context.Context, core.CallbackArgs) (core.State, error) {
    return func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        switch patternName {
        case "research-analysis":
            return handleResearchError(ctx, args)
        case "content-creation":
            return handleContentError(ctx, args)
        default:
            return handleGenericError(ctx, args)
        }
    }
}
```

### 4. Testing Patterns

```go
func TestOrchestrationPattern(t *testing.T) {
    // Create test orchestrator
    orchestrator := CreateResearchAnalysisPipeline()
    
    // Create test agents
    testAgents := createTestAgents()
    for name, agent := range testAgents {
        orchestrator.RegisterAgent(name, agent)
    }
    
    // Create test event
    event := core.NewEvent(
        "test-research",
        core.EventData{"topic": "test topic"},
        map[string]string{"session_id": "test-123"},
    )
    
    // Execute pattern
    result, err := orchestrator.Dispatch(context.Background(), event)
    
    // Verify results
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    // Check pattern-specific outcomes
    assert.Equal(t, "true", result.OutputState.GetMeta("research_completed"))
    assert.Equal(t, "true", result.OutputState.GetMeta("analysis_completed"))
}
```

## Conclusion

Orchestration patterns provide proven solutions for common multi-agent system challenges. By understanding and applying these patterns, you can build sophisticated agent workflows that are reliable, maintainable, and effective.

Key takeaways:
- Choose patterns based on your specific use case requirements
- Customize patterns to fit your domain and constraints
- Implement proper monitoring and error handling
- Test patterns thoroughly before production deployment
- Consider performance implications of different orchestration modes

These patterns serve as starting points that you can adapt and extend for your specific needs. As you gain experience, you'll develop your own patterns that solve domain-specific problems.

## Next Steps

- [Mixed Orchestration](mixed-mode.md) - Learn to implement complex patterns
- [State Management](../core-concepts/state-management.md) - Master data flow between stages
- [Error Handling](../core-concepts/error-handling.md) - Implement robust error management
- [Performance Optimization](../advanced-patterns/load-balancing.md) - Scale your orchestrated systems

## Further Reading

- [API Reference: Orchestration](../../api/core.md#orchestration)
- [Examples: Orchestration Patterns](../../examples/)
- [Configuration Guide: Advanced Orchestration](../../configuration/orchestration.md)