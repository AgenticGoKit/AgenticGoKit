# 15-Minute Tutorial Series

## Overview

This series of focused tutorials takes you from the basics to advanced AgenticGoKit concepts in just 15 minutes each. Each tutorial builds on the previous ones, providing a structured learning path for mastering multi-agent systems.

## Prerequisites

- Complete the [5-Minute Quickstart](../../quickstart.md)
- Go 1.21+ installed
- OpenAI API key (or other LLM provider)

## Tutorial Series

### 🤝 [Multi-Agent Collaboration](multi-agent-collaboration.md) (15 minutes)
Learn different orchestration patterns and how agents work together:
- Collaborative orchestration (parallel processing)
- Sequential orchestration (pipeline processing)
- Mixed orchestration (hybrid workflows)
- When to use each pattern

### 🧠 [Memory and RAG](memory-and-rag.md) (15 minutes)
Add persistent memory and knowledge systems to your agents:
- Setting up vector databases (pgvector, Weaviate)
- Document ingestion and chunking
- RAG (Retrieval-Augmented Generation) implementation
- Hybrid search with semantic and keyword matching

### 🔧 [Tool Integration](tool-integration.md) (15 minutes)
Connect your agents to external tools and APIs:
- MCP (Model Context Protocol) setup
- Using built-in tools (web search, file operations)
- Creating custom tools
- Tool caching and performance optimization

### 🏭 [Production Deployment](production-deployment.md) (15 minutes)
Deploy and scale your agents for production use:
- Docker containerization
- Configuration management
- Monitoring and logging
- Error handling and fault tolerance

## Learning Path

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Quickstart    │───▶│ Multi-Agent     │───▶│ Memory & RAG    │───▶│ Tool Integration│
│   (5 minutes)   │    │ Collaboration   │    │  (15 minutes)   │    │  (15 minutes)   │
└─────────────────┘    │  (15 minutes)   │    └─────────────────┘    └─────────────────┘
                       └─────────────────┘                                      │
                                                                                 ▼
                                                                       ┌─────────────────┐
                                                                       │ Production      │
                                                                       │ Deployment      │
                                                                       │ (15 minutes)    │
                                                                       └─────────────────┘
```

## What You'll Build

By the end of this series, you'll have built:

- **Multi-agent systems** with different orchestration patterns
- **Knowledge-aware agents** with RAG and vector search
- **Tool-enabled agents** that can interact with external services
- **Production-ready systems** with monitoring and fault tolerance

## Getting Started

Start with [Multi-Agent Collaboration](multi-agent-collaboration.md) to learn how agents work together, then progress through the series based on your interests and needs.

Each tutorial is designed to be completed in 15 minutes and includes:
- Clear learning objectives
- Step-by-step instructions
- Working code examples
- Troubleshooting tips
- Next steps and further reading