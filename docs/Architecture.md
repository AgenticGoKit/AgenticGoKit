# AgentFlow Architecture

This Go-based AI Agent Framework provides production-ready, multi-modal agent capabilities with Model Context Protocol (MCP) integration. It combines ultra-fast agent instantiation, dynamic tool discovery, advanced multi-agent coordination patterns, and comprehensive developer tooling for building scalable agent workflows.

## 1. Core Execution & Orchestration

- **Runner**: A central execution engine manages `Events`, drives agent workflows, and coordinates calls to LLM providers and MCP tools.  
- **Orchestrator Modes**: Supports high-performance multi-agent patterns—**route** (single dispatch), **collaborate** (parallel cooperation), and **coordinate** (hierarchical delegation)—for flexible agent teamwork.

## 2. Agents & Workflow Primitives

- **AgentHandler Interface**: Primary interface for implementing event-driven agent logic that processes user queries and manages state.
- **Multi-Agent Systems**: Compose specialized agents into teams with intelligent handoffs and collaborative processing pipelines.
- **Universal Agent Pattern**: Agents analyze queries, choose appropriate tools dynamically, and provide comprehensive answers without counter-questions.

## 3. MCP Integration & Tool Ecosystem

- **Model Context Protocol**: Dynamic tool discovery and schema-based tool integration enables agents to find and use any MCP server tools.
- **Tool Discovery**: Automatic detection of available tools via `core.FormatToolsForPrompt()` for intelligent tool selection by LLMs.
- **Tool Execution**: Centralized tool call parsing and execution via `core.ParseAndExecuteToolCalls()` with comprehensive error handling.
- **MCP Servers**: Built-in support for Docker, web search, databases, file systems, and custom MCP servers.

## 4. Model & Provider Integration

- **ModelProvider Adapters**: Abstract any LLM backend—Azure OpenAI (default), OpenAI, Ollama, or custom providers—via a unified interface.
- **Multi-Provider Architecture**: Switch between providers without changing agent code; configure via `agentflow.toml`.
- **Provider-Agnostic Development**: Write agents once, deploy with any supported LLM provider.

## 5. Memory & State Management

- **Session Service**: Manages short-term conversation state (`State`) and event histories (`Session`) per user interaction.
- **State Cloning**: Thread-safe state management with cloning support for concurrent agent processing.
- **Event Tracking**: Comprehensive event correlation and session tracking for debugging and analytics.

## 6. Configuration & Project Management

- **TOML Configuration**: Centralized configuration via `agentflow.toml` for providers, MCP servers, and orchestration settings.
- **CLI Scaffolding**: `agentcli create` generates production-ready projects with current best practices and patterns.
- **Project Templates**: Support for basic agents, multi-agent workflows, and production systems with monitoring.

## 7. Debugging, Callbacks & Monitoring

- **Callbacks**: Register hooks at key lifecycle points (before/after agent runs, tool calls, state changes) for logging and validation.
- **Debug & Trace**: Built-in support for step-by-step execution tracing and session correlation for diagnosing agent decisions.
- **Prometheus Metrics**: Optional metrics integration for production monitoring of agent performance and tool usage.

## 8. Developer Experience & Deployment

- **CLI Tools**: Comprehensive `agentcli` for project creation, scaffolding, and development workflow management.
- **Production Patterns**: Built-in error handling, retry logic, circuit breakers, and input validation.
- **Containerized Deployment**: Docker-ready with support for cloud deployment and scaling patterns.

## 9. Performance & Extensibility

- **Go-Native Efficiency**: Lightweight agent implementation with minimal memory footprint and fast initialization.
- **High-Throughput**: Engineered for concurrency with event-driven architecture and non-blocking tool execution.
- **Plugin Architecture**: Easily extend with new ModelAdapters, MCP servers, custom tools, or orchestration patterns.

## 10. Production Features

- **Error Resilience**: Specialized error handlers for validation, timeout, and critical failures with automatic routing.
- **Session Management**: Built-in session correlation and state management for multi-turn conversations.
- **Observability**: Comprehensive logging, tracing, and metrics for production monitoring and debugging.
- **Responsible AI**: Built-in content safety and validation patterns for production deployments.
