---
title: Debugging and Monitoring
description: Comprehensive debugging and monitoring strategies for AgenticGoKit applications
sidebar: true
outline: deep
editLink: true
lastUpdated: true
prev:
  text: 'Core Concepts'
  link: '../core-concepts/'
next:
  text: 'Debugging Multi-Agent Systems'
  link: './debugging-multi-agent-systems'
tags:
  - debugging
  - monitoring
  - tracing
  - logging
  - troubleshooting
head:
  - - meta
    - name: keywords
      content: AgenticGoKit, debugging, monitoring, multi-agent systems, tracing, logging
  - - meta
    - property: og:title
      content: Debugging and Monitoring - AgenticGoKit
  - - meta
    - property: og:description
      content: Comprehensive debugging and monitoring strategies for AgenticGoKit applications
---

# Debugging and Monitoring in AgenticGoKit

## Overview

This section covers comprehensive debugging and monitoring strategies for AgenticGoKit applications. You'll learn how to troubleshoot multi-agent systems, implement effective logging, set up monitoring, and optimize performance using AgenticGoKit's built-in debugging capabilities.

Debugging multi-agent systems presents unique challenges due to their concurrent execution, asynchronous communication, and complex state management. This guide provides practical tools and techniques using AgenticGoKit's tracing system and structured logging to help you identify, diagnose, and resolve issues effectively.

## What You'll Learn

::: info Learning Objectives
Master the essential debugging skills for multi-agent systems development.
:::

- **[Debugging Multi-Agent Systems](./debugging-multi-agent-systems)**: Techniques for troubleshooting complex agent interactions using current API
- **[Logging and Tracing](./logging-and-tracing)**: Implementing structured logging and trace analysis with agentcli
- **[Practical Examples](./practical-examples)**: Complete, runnable debugging examples and use cases

## Prerequisites

::: tip Prerequisites
Ensure you understand these core concepts before diving into debugging techniques.
:::

Before diving into debugging and monitoring, you should be familiar with:

- [Agent Lifecycle](../core-concepts/agent-lifecycle) - Understanding agent creation and execution
- [Message Passing](../core-concepts/message-passing) - Event flow and callback systems  
- [Orchestration Patterns](../core-concepts/orchestration-patterns) - Multi-agent coordination
- [State Management](../core-concepts/state-management) - State interface and data flow

## Why Debugging Multi-Agent Systems is Different

::: info Complexity Challenge
Multi-agent systems are inherently more complex to debug due to their distributed and asynchronous nature.
:::

Multi-agent systems introduce several debugging challenges:

### 1. Concurrent Execution
- Agents run concurrently within the same process using goroutines
- State changes occur across multiple agents simultaneously
- Race conditions and timing issues require careful debugging with proper synchronization

### 2. Asynchronous Event Processing
- Events flow through the Runner's callback system asynchronously
- Cause and effect relationships may be separated in time and traced through session IDs
- Error propagation follows the callback chain and can be complex

### 3. Complex State Management
- State is shared and modified by multiple agents through the State interface
- State transformations follow the Clone/Merge pattern for thread safety
- Debugging requires understanding state flow through trace entries

### 4. Orchestration Complexity
- Different orchestration modes (route, collaborative, sequential, loop, mixed) have different failure patterns
- Agent interactions are coordinated through the Runner and callback system
- System behavior varies based on configuration and agent registration order

## Debugging Philosophy

Effective debugging of multi-agent systems requires a systematic approach:

### 1. Observability First
- Use AgenticGoKit's built-in tracing system to capture execution flow
- Implement structured logging with zerolog for consistent formats
- Include session IDs and correlation IDs to track requests across agents

### 2. Fail Fast and Clearly
- Validate inputs and state at agent boundaries using the State interface
- Use clear, actionable error messages with proper Go error wrapping
- Implement callback-based error handling to prevent cascade failures

### 3. Isolate and Reproduce
- Create minimal test cases using the Agent interface for reproduction
- Use deterministic testing with controlled State inputs
- Implement agent mocking using interface composition for isolated testing

### 4. Monitor Continuously
- Use callback registration for real-time monitoring of agent behavior
- Track performance trends through trace analysis with agentcli
- Implement health checks using the Agent interface for early issue detection

## Getting Started

::: tip Learning Path
Start with the fundamentals, then progress to advanced techniques as you become more comfortable with the debugging tools.
:::

**Recommended Learning Sequence:**

1. **[Debugging Multi-Agent Systems](./debugging-multi-agent-systems)** - Learn fundamental debugging techniques using the current AgenticGoKit API
2. **[Logging and Tracing](./logging-and-tracing)** - Explore advanced observability patterns with structured logging and trace analysis
3. **[Practical Examples](./practical-examples)** - Apply debugging concepts with complete, runnable examples

## Quick Reference

### Common Debugging Commands

::: code-group

```bash [Basic Commands]
# List available trace sessions
agentcli list

# View execution trace for a session
agentcli trace <session-id>
```

```bash [Advanced Commands]
# View only agent flow without state details
agentcli trace --flow-only <session-id>

# Filter trace to specific agent
agentcli trace --filter agent=<agent-name> <session-id>

# View verbose trace with full state details
agentcli trace --verbose <session-id>

# Debug trace structure and validate JSON
agentcli trace --debug <session-id>
```

:::

### Configuration in agentflow.toml

```toml
[logging]
level = "debug"
format = "json"

[agent_flow]
name = "my-agent-system"
version = "1.0.0"

[runtime]
max_concurrent_agents = 10

[agents.debug-agent]
role = "debugger"
description = "Debugging and monitoring agent"
system_prompt = "You are a debugging expert."
enabled = true
capabilities = ["debugging", "monitoring"]
timeout = 30
```

### Key Metrics to Monitor

- **Agent Execution Time**: Duration from BeforeAgentRun to AfterAgentRun hooks
- **Error Rate**: Percentage of failed agent executions tracked through callbacks
- **State Size**: Memory usage of State objects passed between agents
- **Callback Chain Length**: Number of registered callbacks per hook point
- **Trace File Size**: Size of generated .trace.json files for session analysis
