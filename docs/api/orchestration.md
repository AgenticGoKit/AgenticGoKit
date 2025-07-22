# Orchestration API

**Multi-agent coordination and workflow patterns**

This document covers AgenticGoKit's Orchestration API, which enables sophisticated coordination between multiple agents. The orchestration system provides various patterns for agent collaboration, from simple routing to complex hybrid workflows.

## 📋 Core Concepts

### Orchestration Modes

AgenticGoKit supports multiple orchestration patterns:

```go
type OrchestrationMode string

const (
    // OrchestrationRoute sends each event to a single agent based on routing metadata (default)
    OrchestrationRoute OrchestrationMode = "route"
    
    // OrchestrationCollaborate sends each event to ALL registered agents in parallel
    OrchestrationCollaborate OrchestrationMode = "collaborate"
    
    // OrchestrationSequential processes agents one after another
    OrchestrationSequential OrchestrationMode = "sequential"
    
    // OrchestrationParallel processes agents in parallel (similar to collaborate)
    OrchestrationParallel OrchestrationMode = "parallel"
    
    // OrchestrationLoop repeats processing with a single agent
    OrchestrationLoop OrchestrationMode = "loop"
    
    // OrchestrationMixed combines collaborative and sequential patterns
    OrchestrationMixed OrchestrationMode = "mixed"
)
```

### Core Interfaces

```go
// Orchestrator manages agent execution and coordination
type Orchestrator interface {
    RegisterAgent(name string, handler AgentHandler) error
    Dispatch(ctx context.Context, event Event) (AgentResult, error)
    GetCallbackRegistry() *CallbackRegistry
    Stop()
}

// Runner provides the main interface for processing events with orchestration
type Runner interface {
    RegisterAgent(name string, handler AgentHandler) error
    ProcessEvent(ctx context.Context, event Event) (map[string]AgentResult, error)
    GetCallbackRegistry() *CallbackRegistry
    Stop()
}
```

## 🚀 Basic Usage

### Route Orchestration (Default)

Route orchestration sends each event to a single agent based on routing metadata:

```go
package main

import (
    "context"
    "fmt"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func routeExample() {
    // Create agents
    agents := map[string]core.AgentHandler{
        "greeter": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            name := event.Data["name"].(string)
            return core.AgentResult{
                Data: map[string]interface{}{
                    "greeting": fmt.Sprintf("Hello, %s!", name),
                },
            }, nil
        }),
        
        "calculator": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            a := event.Data["a"].(float64)
            b := event.Data["b"].(float64)
            return core.AgentResult{
                Data: map[string]interface{}{
                    "result": a + b,
                },
            }, nil
        }),
    }
    
    // Create route runner (default behavior)
    runner := core.CreateRouteRunner(agents)
    
    // Route to greeter
    greetEvent := core.NewEvent("greeting", map[string]interface{}{
        "name": "Alice",
    })
    greetEvent.Metadata["route"] = "greeter"
    
    results, _ := runner.ProcessEvent(context.Background(), greetEvent)
    fmt.Printf("Greeting: %s\n", results["greeter"].Data["greeting"])
    
    // Route to calculator
    calcEvent := core.NewEvent("calculation", map[string]interface{}{
        "a": 5.0,
        "b": 3.0,
    })
    calcEvent.Metadata["route"] = "calculator"
    
    results, _ = runner.ProcessEvent(context.Background(), calcEvent)
    fmt.Printf("Result: %f\n", results["calculator"].Data["result"])
}
```

### Collaborative Orchestration

Collaborative orchestration sends events to all agents simultaneously:

```go
func collaborativeExample() {
    // Create analysis agents
    agents := map[string]core.AgentHandler{
        "sentiment": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            text := event.Data["text"].(string)
            // Simulate sentiment analysis
            sentiment := "positive"
            if strings.Contains(strings.ToLower(text), "bad") {
                sentiment = "negative"
            }
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "sentiment": sentiment,
                    "confidence": 0.85,
                },
            }, nil
        }),
        
        "keywords": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            text := event.Data["text"].(string)
            // Simulate keyword extraction
            words := strings.Fields(strings.ToLower(text))
            keywords := []string{}
            for _, word := range words {
                if len(word) > 4 { // Simple filter
                    keywords = append(keywords, word)
                }
            }
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "keywords": keywords,
                    "count": len(keywords),
                },
            }, nil
        }),
        
        "summary": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            text := event.Data["text"].(string)
            // Simulate summarization
            words := strings.Fields(text)
            summary := strings.Join(words[:min(10, len(words))], " ") + "..."
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "summary": summary,
                    "length": len(words),
                },
            }, nil
        }),
    }
    
    // Create collaborative runner
    runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
    
    // All agents process the same event in parallel
    event := core.NewEvent("analyze", map[string]interface{}{
        "text": "This is a great product with excellent features and outstanding quality.",
    })
    
    results, err := runner.ProcessEvent(context.Background(), event)
    if err != nil {
        panic(err)
    }
    
    // All agents have processed the event
    fmt.Printf("Sentiment: %s (%.2f confidence)\n", 
        results["sentiment"].Data["sentiment"], 
        results["sentiment"].Data["confidence"])
    fmt.Printf("Keywords: %v\n", results["keywords"].Data["keywords"])
    fmt.Printf("Summary: %s\n", results["summary"].Data["summary"])
}
```

### Sequential Orchestration

Sequential orchestration processes agents in a specific order, passing state between them:

```go
func sequentialExample() {
    // Create pipeline agents
    agents := map[string]core.AgentHandler{
        "extractor": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            text := event.Data["text"].(string)
            // Extract entities (simplified)
            entities := []string{"Apple", "iPhone", "technology"}
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "entities": entities,
                    "original_text": text,
                },
            }, nil
        }),
        
        "enricher": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            // Get entities from previous agent's output
            entities, ok := state.Data["entities"].([]string)
            if !ok {
                return core.AgentResult{}, fmt.Errorf("no entities found in state")
            }
            
            // Enrich entities with additional info
            enriched := make(map[string]interface{})
            for _, entity := range entities {
                enriched[entity] = map[string]interface{}{
                    "type": "company",
                    "confidence": 0.9,
                }
            }
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "enriched_entities": enriched,
                    "entity_count": len(entities),
                },
            }, nil
        }),
        
        "formatter": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            // Get enriched entities from previous agent
            enriched, ok := state.Data["enriched_entities"].(map[string]interface{})
            if !ok {
                return core.AgentResult{}, fmt.Errorf("no enriched entities found")
            }
            
            // Format final output
            var formatted []string
            for entity, info := range enriched {
                infoMap := info.(map[string]interface{})
                formatted = append(formatted, fmt.Sprintf("%s (%s)", entity, infoMap["type"]))
            }
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "formatted_output": strings.Join(formatted, ", "),
                    "processing_complete": true,
                },
            }, nil
        }),
    }
    
    // Create sequential runner with specific order
    agentOrder := []string{"extractor", "enricher", "formatter"}
    runner := core.NewOrchestrationBuilder(core.OrchestrationSequential).
        WithAgents(agents).
        WithTimeout(30 * time.Second).
        Build()
    
    // Configure sequential order (this would be done internally)
    // For now, we'll use the builder pattern
    
    event := core.NewEvent("process", map[string]interface{}{
        "text": "Apple released the new iPhone with advanced technology.",
    })
    
    results, err := runner.ProcessEvent(context.Background(), event)
    if err != nil {
        panic(err)
    }
    
    // Final result from the last agent in the sequence
    fmt.Printf("Processed entities: %s\n", results["formatter"].Data["formatted_output"])
}
```

## 🔧 Advanced Orchestration Patterns

### Mixed Orchestration

Mixed orchestration combines collaborative and sequential patterns:

```go
func mixedOrchestrationExample() {
    // Create agents for different phases
    agents := map[string]core.AgentHandler{
        // Collaborative phase - parallel analysis
        "sentiment_analyzer": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            text := event.Data["text"].(string)
            return core.AgentResult{
                Data: map[string]interface{}{
                    "sentiment": "positive",
                    "score": 0.8,
                },
            }, nil
        }),
        
        "topic_extractor": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            text := event.Data["text"].(string)
            return core.AgentResult{
                Data: map[string]interface{}{
                    "topics": []string{"technology", "product"},
                },
            }, nil
        }),
        
        // Sequential phase - ordered processing
        "synthesizer": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            // Combine results from collaborative phase
            sentiment := state.Data["sentiment"].(string)
            topics := state.Data["topics"].([]string)
            
            synthesis := fmt.Sprintf("Analysis shows %s sentiment about %s", 
                sentiment, strings.Join(topics, " and "))
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "synthesis": synthesis,
                },
            }, nil
        }),
        
        "reporter": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            synthesis := state.Data["synthesis"].(string)
            
            report := fmt.Sprintf("REPORT: %s\nGenerated at: %s", 
                synthesis, time.Now().Format(time.RFC3339))
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "final_report": report,
                },
            }, nil
        }),
    }
    
    // Create mixed orchestration
    collaborativeAgents := []string{"sentiment_analyzer", "topic_extractor"}
    sequentialAgents := []string{"synthesizer", "reporter"}
    
    runner := core.NewOrchestrationBuilder(core.OrchestrationMixed).
        WithAgents(agents).
        WithTimeout(45 * time.Second).
        Build()
    
    // Configure mixed mode (this would be done internally)
    // The mixed orchestrator would run collaborative agents first, then sequential
    
    event := core.NewEvent("analyze", map[string]interface{}{
        "text": "The new smartphone features are impressive and user-friendly.",
    })
    
    results, err := runner.ProcessEvent(context.Background(), event)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Final Report:\n%s\n", results["reporter"].Data["final_report"])
}
```

### Loop Orchestration

Loop orchestration repeats processing with a single agent until a condition is met:

```go
func loopOrchestrationExample() {
    // Create an iterative refinement agent
    refinerAgent := core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        // Get current iteration and text
        iteration := 0
        if iter, ok := state.Data["iteration"].(int); ok {
            iteration = iter
        }
        
        text := event.Data["text"].(string)
        if currentText, ok := state.Data["current_text"].(string); ok {
            text = currentText
        }
        
        // Simulate refinement process
        refined := strings.ReplaceAll(text, "  ", " ") // Remove double spaces
        refined = strings.TrimSpace(refined)
        
        iteration++
        
        // Check if we should continue (max 3 iterations or no changes)
        shouldContinue := iteration < 3 && refined != text
        
        result := core.AgentResult{
            Data: map[string]interface{}{
                "current_text": refined,
                "iteration": iteration,
                "changes_made": refined != text,
                "loop_completed": !shouldContinue,
            },
        }
        
        return result, nil
    })
    
    agents := map[string]core.AgentHandler{
        "refiner": refinerAgent,
    }
    
    // Create loop orchestration
    runner := core.NewOrchestrationBuilder(core.OrchestrationLoop).
        WithAgents(agents).
        WithTimeout(30 * time.Second).
        Build()
    
    event := core.NewEvent("refine", map[string]interface{}{
        "text": "This  is   a  text   with    irregular   spacing.",
    })
    
    results, err := runner.ProcessEvent(context.Background(), event)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Original: %s\n", event.Data["text"])
    fmt.Printf("Refined: %s\n", results["refiner"].Data["current_text"])
    fmt.Printf("Iterations: %d\n", results["refiner"].Data["iteration"])
}
```

## 🏗️ Orchestration Builder Pattern

The OrchestrationBuilder provides a fluent interface for creating complex orchestrations:

```go
func builderPatternExample() {
    // Create agents
    agents := map[string]core.AgentHandler{
        "validator": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            data := event.Data["data"].(string)
            valid := len(data) > 0
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "valid": valid,
                    "data": data,
                },
            }, nil
        }),
        
        "processor": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            data := state.Data["data"].(string)
            processed := strings.ToUpper(data)
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "processed_data": processed,
                },
            }, nil
        }),
    }
    
    // Build orchestration with custom configuration
    runner := core.NewOrchestrationBuilder(core.OrchestrationSequential).
        WithAgents(agents).
        WithTimeout(60 * time.Second).
        WithMaxConcurrency(5).
        WithFailureThreshold(0.8).
        WithRetryPolicy(&core.RetryPolicy{
            MaxRetries:    3,
            BackoffFactor: 1.5,
            MaxDelay:      10 * time.Second,
        }).
        Build()
    
    event := core.NewEvent("process", map[string]interface{}{
        "data": "hello world",
    })
    
    results, err := runner.ProcessEvent(context.Background(), event)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Processed: %s\n", results["processor"].Data["processed_data"])
}
```

## 🔧 Configuration and Customization

### Orchestration Configuration

```go
type OrchestrationConfig struct {
    Timeout          time.Duration // Overall timeout for orchestration operations
    MaxConcurrency   int           // Maximum number of concurrent agent executions
    FailureThreshold float64       // Percentage of failures before stopping (0.0-1.0)
    RetryPolicy      *RetryPolicy  // Policy for retrying failed operations
}

func configurationExample() {
    config := core.OrchestrationConfig{
        Timeout:          45 * time.Second,
        MaxConcurrency:   10,
        FailureThreshold: 0.7, // Stop if 70% of agents fail
        RetryPolicy: &core.RetryPolicy{
            MaxRetries:    5,
            BackoffFactor: 2.0,
            MaxDelay:      30 * time.Second,
        },
    }
    
    agents := map[string]core.AgentHandler{
        "agent1": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            // Agent implementation
            return core.AgentResult{
                Data: map[string]interface{}{
                    "result": "processed",
                },
            }, nil
        }),
    }
    
    runner := core.NewOrchestrationBuilder(core.OrchestrationCollaborate).
        WithAgents(agents).
        WithConfig(config).
        Build()
    
    // Use the configured runner
    event := core.NewEvent("test", nil)
    results, _ := runner.ProcessEvent(context.Background(), event)
    
    fmt.Printf("Results: %+v\n", results)
}
```

### Specialized Runners

AgenticGoKit provides several pre-configured runners for common use cases:

```go
func specializedRunnersExample() {
    agents := map[string]core.AgentHandler{
        "worker1": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            return core.AgentResult{Data: map[string]interface{}{"work": "done"}}, nil
        }),
        "worker2": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            return core.AgentResult{Data: map[string]interface{}{"work": "done"}}, nil
        }),
    }
    
    // High throughput runner - optimized for performance
    highThroughputRunner := core.CreateHighThroughputRunner(agents)
    
    // Fault tolerant runner - aggressive retry policies
    faultTolerantRunner := core.CreateFaultTolerantRunner(agents)
    
    // Load balanced runner - distributes load across agent instances
    loadBalancedRunner := core.CreateLoadBalancedRunner(agents, 20)
    
    event := core.NewEvent("work", nil)
    
    // Use different runners based on requirements
    results1, _ := highThroughputRunner.ProcessEvent(context.Background(), event)
    results2, _ := faultTolerantRunner.ProcessEvent(context.Background(), event)
    results3, _ := loadBalancedRunner.ProcessEvent(context.Background(), event)
    
    fmt.Printf("High throughput: %+v\n", results1)
    fmt.Printf("Fault tolerant: %+v\n", results2)
    fmt.Printf("Load balanced: %+v\n", results3)
}
```

## 🧪 Testing Orchestration

### Unit Testing Orchestration Patterns

```go
func TestCollaborativeOrchestration(t *testing.T) {
    // Create test agents
    agent1 := core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        return core.AgentResult{
            Data: map[string]interface{}{
                "agent1_result": "success",
            },
        }, nil
    })
    
    agent2 := core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        return core.AgentResult{
            Data: map[string]interface{}{
                "agent2_result": "success",
            },
        }, nil
    })
    
    agents := map[string]core.AgentHandler{
        "agent1": agent1,
        "agent2": agent2,
    }
    
    // Create collaborative runner
    runner := core.CreateCollaborativeRunner(agents, 10*time.Second)
    
    // Test event processing
    event := core.NewEvent("test", map[string]interface{}{
        "input": "test data",
    })
    
    results, err := runner.ProcessEvent(context.Background(), event)
    
    // Verify results
    require.NoError(t, err)
    assert.Len(t, results, 2)
    assert.Equal(t, "success", results["agent1"].Data["agent1_result"])
    assert.Equal(t, "success", results["agent2"].Data["agent2_result"])
}

func TestSequentialOrchestration(t *testing.T) {
    // Create agents that depend on each other
    agent1 := core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        input := event.Data["input"].(string)
        return core.AgentResult{
            Data: map[string]interface{}{
                "step1_output": input + "_step1",
            },
        }, nil
    })
    
    agent2 := core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        // Should receive output from agent1
        step1Output := state.Data["step1_output"].(string)
        return core.AgentResult{
            Data: map[string]interface{}{
                "final_output": step1Output + "_step2",
            },
        }, nil
    })
    
    agents := map[string]core.AgentHandler{
        "agent1": agent1,
        "agent2": agent2,
    }
    
    // Create sequential runner
    runner := core.NewOrchestrationBuilder(core.OrchestrationSequential).
        WithAgents(agents).
        Build()
    
    event := core.NewEvent("test", map[string]interface{}{
        "input": "start",
    })
    
    results, err := runner.ProcessEvent(context.Background(), event)
    
    require.NoError(t, err)
    assert.Equal(t, "start_step1_step2", results["agent2"].Data["final_output"])
}
```

### Integration Testing

```go
func TestOrchestrationIntegration(t *testing.T) {
    // Test a complete workflow
    agents := map[string]core.AgentHandler{
        "input_validator": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            data := event.Data["data"].(string)
            if len(data) == 0 {
                return core.AgentResult{}, fmt.Errorf("empty input")
            }
            return core.AgentResult{
                Data: map[string]interface{}{
                    "validated_data": data,
                    "valid": true,
                },
            }, nil
        }),
        
        "data_processor": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            data := state.Data["validated_data"].(string)
            processed := strings.ToUpper(data)
            return core.AgentResult{
                Data: map[string]interface{}{
                    "processed_data": processed,
                },
            }, nil
        }),
        
        "output_formatter": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            data := state.Data["processed_data"].(string)
            formatted := fmt.Sprintf("RESULT: %s", data)
            return core.AgentResult{
                Data: map[string]interface{}{
                    "final_output": formatted,
                },
            }, nil
        }),
    }
    
    // Test the complete pipeline
    runner := core.NewOrchestrationBuilder(core.OrchestrationSequential).
        WithAgents(agents).
        WithTimeout(30 * time.Second).
        Build()
    
    event := core.NewEvent("process", map[string]interface{}{
        "data": "hello world",
    })
    
    results, err := runner.ProcessEvent(context.Background(), event)
    
    require.NoError(t, err)
    assert.Equal(t, "RESULT: HELLO WORLD", results["output_formatter"].Data["final_output"])
}
```

## 📚 Best Practices

### 1. Choosing the Right Orchestration Pattern

- **Route**: Use for simple request-response patterns where each event needs specific handling
- **Collaborate**: Use when you need multiple perspectives or parallel analysis of the same data
- **Sequential**: Use for data pipelines where each step depends on the previous one
- **Mixed**: Use for complex workflows that need both parallel and sequential processing
- **Loop**: Use for iterative refinement or optimization processes

### 2. Performance Considerations

```go
// Good: Configure appropriate timeouts and concurrency
runner := core.NewOrchestrationBuilder(core.OrchestrationCollaborate).
    WithAgents(agents).
    WithTimeout(30 * time.Second).        // Reasonable timeout
    WithMaxConcurrency(10).               // Limit concurrent executions
    WithFailureThreshold(0.8).            // Allow some failures
    Build()

// Bad: No limits or unrealistic timeouts
runner := core.NewOrchestrationBuilder(core.OrchestrationCollaborate).
    WithAgents(agents).
    WithTimeout(5 * time.Minute).         // Too long
    WithMaxConcurrency(1000).             // Too many
    WithFailureThreshold(0.0).            // No tolerance for failures
    Build()
```

### 3. Error Handling in Orchestration

```go
// Good: Handle errors gracefully with fallbacks
runner := core.NewOrchestrationBuilder(core.OrchestrationCollaborate).
    WithAgents(agents).
    WithRetryPolicy(&core.RetryPolicy{
        MaxRetries:    3,
        BackoffFactor: 1.5,
        MaxDelay:      10 * time.Second,
    }).
    WithFailureThreshold(0.7).            // Continue if 30% succeed
    Build()
```

### 4. State Management in Sequential Orchestration

```go
// Good: Clear state passing between agents
func sequentialAgent1(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Process and set clear state keys
    result := processData(event.Data["input"])
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "step1_result": result,
            "step1_metadata": map[string]interface{}{
                "processed_at": time.Now(),
                "agent": "step1",
            },
        },
    }, nil
}

func sequentialAgent2(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Check for required state from previous agent
    step1Result, ok := state.Data["step1_result"]
    if !ok {
        return core.AgentResult{}, fmt.Errorf("missing step1_result from previous agent")
    }
    
    // Continue processing
    finalResult := processStep2(step1Result)
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "final_result": finalResult,
        },
    }, nil
}
```

## 🔗 Related APIs

- **[Agent API](agent.md)** - Building individual agents
- **[State & Event API](state-event.md)** - Data flow and communication
- **[Memory API](memory.md)** - Persistent storage and RAG
- **[Configuration API](configuration.md)** - System configuration

---

*This documentation covers the current Orchestration API in AgenticGoKit. The framework is actively developed, so some interfaces may evolve.*