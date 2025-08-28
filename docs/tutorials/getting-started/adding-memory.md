---
title: "Adding Memory"
description: "Give your agents persistent memory and knowledge base capabilities"
prev:
  text: "Multi-Agent Basics"
  link: "./multi-agent-basics"
next:
  text: "Tool Integration"
  link: "./tool-integration"
---

# Adding Memory

Your agents are now working together effectively, but they forget everything between conversations. Let's change that by adding persistent memory and knowledge base capabilities to create truly intelligent, learning agents.

## Learning Objectives

By the end of this section, you'll be able to:
- Understand different types of agent memory (short-term, long-term, shared)
- Configure memory providers (in-memory, PostgreSQL, Weaviate)
- Implement RAG (Retrieval Augmented Generation) for knowledge-enhanced responses
- Upload and manage documents in knowledge bases
- Create agents that remember conversations and learn from interactions
- Troubleshoot memory-related issues

## Prerequisites

Before starting, make sure you've completed:
- ✅ [Multi-Agent Basics](./multi-agent-basics.md) - Understanding of multi-agent orchestration

## Why Agents Need Memory

Without memory, agents are like having a conversation with someone who has amnesia - they can't:
- Remember previous conversations
- Learn from past interactions
- Access specific knowledge bases
- Build context over time
- Provide personalized responses

With memory, agents become much more powerful:
- **Conversational Context**: Remember what you discussed earlier
- **Learning**: Improve responses based on feedback
- **Knowledge Access**: Query specific documents and data
- **Personalization**: Adapt to individual user preferences
- **Consistency**: Maintain coherent behavior across sessions

## Types of Memory in AgenticGoKit

### Short-Term Memory
Information that persists during a single conversation session.

**Use cases:**
- Maintaining conversation context
- Remembering user preferences for the current session
- Tracking the flow of a multi-step process

### Long-Term Memory
Information that persists across multiple sessions and conversations.

**Use cases:**
- User profiles and preferences
- Historical conversation summaries
- Learned patterns and insights
- System configuration and settings

### Shared Memory
Information that multiple agents can access and modify.

**Use cases:**
- Collaborative agent workflows
- Shared knowledge bases
- Team coordination and state
- Cross-agent learning

### Knowledge Base (RAG)
Structured document storage with semantic search capabilities.

**Use cases:**
- Company documentation
- Product manuals and guides
- Research papers and articles
- FAQ databases

## Memory Providers

AgenticGoKit supports several memory providers:

### In-Memory Provider
**Best for**: Development, testing, simple applications
**Pros**: Efficient, no setup required
**Cons**: Data lost when application stops

```toml
[agent_memory]
provider = "memory"
enable_rag = false
```

### PostgreSQL with pgvector
**Best for**: Production applications, scalable systems
**Pros**: Persistent, scalable, full SQL capabilities
**Cons**: Requires PostgreSQL setup

```toml
[agent_memory]
provider = "pgvector"
enable_rag = true
enable_knowledge_base = true
chunk_size = 1000
overlap_size = 200

[agent_memory.pgvector]
connection_string = "postgresql://user:password@localhost/agentdb"
```

### Weaviate
**Best for**: Large-scale vector operations, advanced search
**Pros**: Purpose-built for vector search, advanced features
**Cons**: More complex setup

```toml
[agent_memory]
provider = "weaviate"
enable_rag = true
enable_knowledge_base = true

[agent_memory.weaviate]
host = "http://localhost:8080"
```

## Setting Up Memory: Step by Step

Let's create a memory-enabled agent system that can remember conversations and access a knowledge base.

### Step 1: Create a Memory-Enabled Project

```bash
agentcli create smart-assistant --template memory-enabled
cd smart-assistant
```

### Step 2: Configure PostgreSQL (Recommended for Learning)

If you don't have PostgreSQL with pgvector, you can use Docker:

```bash
# Start PostgreSQL with pgvector extension
docker run -d \
  --name agenticgokit-postgres \
  -e POSTGRES_DB=agentdb \
  -e POSTGRES_USER=agent \
  -e POSTGRES_PASSWORD=agentpass \
  -p 5432:5432 \
  pgvector/pgvector:pg16
```

### Step 3: Configure Memory in agentflow.toml

```toml
[agent_flow]
name = "smart-assistant"
version = "1.0.0"
description = "An intelligent assistant with memory and knowledge base"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 1500

[agent_memory]
provider = "pgvector"
enable_rag = true
enable_knowledge_base = true
chunk_size = 1000        # Size of document chunks for RAG
overlap_size = 200       # Overlap between chunks
max_results = 5          # Maximum RAG results to include

[agent_memory.pgvector]
connection_string = "postgresql://agent:agentpass@localhost:5432/agentdb"

[orchestration]
mode = "collaborative"
collaborative_agents = ["assistant", "knowledge_expert"]

[agents.assistant]
role = "memory_enabled_assistant"
description = "A helpful assistant with persistent memory"
system_prompt = """
You are an intelligent assistant with access to persistent memory and knowledge bases.

Your capabilities include:
- Remembering previous conversations and user preferences
- Accessing relevant information from knowledge bases
- Learning from interactions to provide better responses
- Maintaining context across multiple sessions

When responding:
1. Check if you have relevant memories or knowledge about the topic
2. Reference previous conversations when appropriate
3. Use knowledge base information to enhance your responses
4. Remember important details for future conversations
5. Ask clarifying questions to build better understanding

Always be helpful, accurate, and personalize your responses based on what you know about the user.
"""
enabled = true
memory_enabled = true

[agents.knowledge_expert]
role = "knowledge_specialist"
description = "Specializes in finding and using knowledge base information"
system_prompt = """
You are a knowledge specialist who excels at finding and utilizing information from knowledge bases.

Your responsibilities:
1. Search knowledge bases for relevant information
2. Synthesize information from multiple sources
3. Provide accurate, well-sourced responses
4. Identify when information might be outdated or incomplete
5. Suggest areas where additional knowledge might be helpful

Always cite your sources and indicate confidence levels in your responses.
"""
enabled = true
memory_enabled = true
```

### Step 4: Test Basic Memory Functionality

```bash
# First conversation
go run . -m "Hello, my name is Alex and I'm a software developer working on Go projects"

# Second conversation (in a new session)
go run . -m "What do you remember about me?"
```

The agent should remember your name and profession from the previous conversation!

## Working with Knowledge Bases

### Adding Documents to Your Knowledge Base

Create some sample documents to add to your knowledge base:

**`docs/go-best-practices.md`:**
```markdown
# Go Best Practices

## Code Organization
- Use meaningful package names
- Keep packages focused and cohesive
- Avoid circular dependencies

## Error Handling
- Always handle errors explicitly
- Use custom error types when appropriate
- Wrap errors with context using fmt.Errorf

## Concurrency
- Use goroutines for concurrent operations
- Communicate through channels, not shared memory
- Always close channels when done
```

**`docs/agenticgokit-guide.md`:**
```markdown
# AgenticGoKit Guide

## Overview
AgenticGoKit is a Go framework for building intelligent multi-agent AI systems.

## Key Features
- Multi-agent orchestration
- Persistent memory and RAG
- Tool integration via MCP
- Configuration-first approach

## Getting Started
1. Install the CLI: go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest
2. Create a project: agentcli create my-project
3. Configure agents in agentflow.toml
4. Run: go run .
```

### Upload Documents Using the CLI

```bash
# Upload individual documents
agentcli knowledge upload docs/go-best-practices.md
agentcli knowledge upload docs/agenticgokit-guide.md

# Upload entire directories
agentcli knowledge upload docs/

# List uploaded documents
agentcli knowledge list
```

### Test Knowledge Base Integration

```bash
go run . -m "What are the best practices for error handling in Go?"
```

The agent should now provide information from your uploaded documents!

## Advanced Memory Configuration

### Customizing RAG Behavior

```toml
[agent_memory]
provider = "pgvector"
enable_rag = true
enable_knowledge_base = true

# RAG Configuration
chunk_size = 1500           # Larger chunks for more context
overlap_size = 300          # More overlap for better continuity
max_results = 8             # More results for comprehensive answers
similarity_threshold = 0.7   # Minimum similarity for inclusion
rerank_results = true       # Re-rank results for relevance

# Memory Configuration
conversation_memory_limit = 50    # Remember last 50 exchanges
summary_trigger_length = 20       # Summarize after 20 exchanges
enable_user_profiles = true       # Track user preferences
```

### Memory-Specific Agent Configuration

```toml
[agents.personal_assistant]
role = "personal_assistant"
memory_enabled = true
memory_scope = "user"              # user, session, or global
knowledge_base_access = true       # Can access knowledge base
conversation_history_limit = 100   # Remember last 100 messages
system_prompt = """
You are a personal assistant with access to the user's conversation history
and relevant knowledge bases. Use this information to provide personalized,
contextual responses.

Memory Guidelines:
- Reference previous conversations naturally
- Learn user preferences and adapt accordingly
- Use knowledge base information to enhance responses
- Remember important dates, preferences, and context
"""
```

## Hands-On Exercises

### Exercise 1: Build a Learning Assistant

Create an assistant that learns about your interests and provides increasingly personalized responses:

1. Configure memory with user profiles enabled
2. Have several conversations about your interests
3. Test how the agent's responses become more personalized
4. Upload documents related to your interests
5. See how the agent combines memory and knowledge

### Exercise 2: Create a Company Knowledge Assistant

Build an assistant for a fictional company:

1. Create documents about company policies, procedures, and FAQ
2. Upload them to the knowledge base
3. Configure an agent specialized in company information
4. Test with various employee questions
5. Add conversation memory to remember employee preferences

### Exercise 3: Multi-Agent Memory Sharing

Create a system where multiple agents share memory:

1. Configure shared memory between agents
2. Have one agent learn information about the user
3. Test that other agents can access this information
4. Create workflows where agents build on each other's memories

## Memory Management and Optimization

### Monitoring Memory Usage

```bash
# Check memory statistics
agentcli memory stats

# View recent conversations
agentcli memory conversations --limit 10

# Search knowledge base
agentcli knowledge search "error handling"

# Clean up old memories
agentcli memory cleanup --older-than 30d
```

### Performance Optimization

**For Large Knowledge Bases:**
```toml
[agent_memory]
# Optimize for large document collections
chunk_size = 800           # Smaller chunks for better precision
max_results = 3            # Fewer results for more efficient processing
enable_caching = true      # Cache frequent queries
batch_size = 100          # Process documents in batches
```

**For High-Frequency Conversations:**
```toml
[agent_memory]
# Optimize for frequent interactions
conversation_memory_limit = 20    # Limit memory to recent exchanges
enable_compression = true         # Compress older memories
summary_frequency = "daily"       # Summarize conversations daily
```

## Troubleshooting Memory Issues

### Common Problems and Solutions

**Memory Not Persisting:**
- Check database connection string
- Verify database permissions
- Ensure memory_enabled = true for agents

**RAG Not Finding Relevant Information:**
- Check chunk_size and overlap_size settings
- Verify documents were uploaded successfully
- Adjust similarity_threshold
- Try different search queries

**Performance Issues:**
- Reduce max_results for more efficient queries
- Optimize chunk_size for your use case
- Enable caching for frequent queries
- Consider using a more powerful vector database

**Memory Growing Too Large:**
- Set conversation_memory_limit
- Enable automatic summarization
- Implement regular cleanup procedures
- Use memory scoping appropriately

### Debugging Memory Operations

```bash
# Enable memory debugging
export AGENTICGOKIT_MEMORY_DEBUG=true
go run . -m "Test message"

# Check memory provider status
agentcli memory health

# Validate knowledge base integrity
agentcli knowledge validate
```

## Best Practices for Memory-Enabled Agents

### 1. Design Memory Scope Appropriately

```toml
# User-specific memory
[agents.personal_assistant]
memory_scope = "user"
memory_enabled = true

# Session-specific memory
[agents.task_helper]
memory_scope = "session"
memory_enabled = true

# Global shared memory
[agents.knowledge_base]
memory_scope = "global"
memory_enabled = true
```

### 2. Optimize Knowledge Base Content

**Good Document Structure:**
```markdown
# Clear Title

## Section Headers
Use clear, descriptive headers that agents can understand.

### Subsections
Break information into logical chunks.

**Key Points:**
- Use bullet points for important information
- Include examples and use cases
- Provide context and background
```

### 3. Handle Memory Gracefully

```toml
[agents.robust_assistant]
system_prompt = """
You have access to memory and knowledge bases, but handle cases gracefully when:
- Memory is unavailable or incomplete
- Knowledge base searches return no results
- Previous conversations are unclear or contradictory

Always acknowledge limitations and ask for clarification when needed.
"""
```

### 4. Privacy and Security Considerations

```toml
[agent_memory]
# Security settings
encrypt_memories = true
anonymize_sensitive_data = true
retention_policy = "90d"
gdpr_compliant = true

# Privacy settings
user_data_isolation = true
memory_access_logging = true
```

## What You've Learned

✅ **Understanding of agent memory types** and their use cases  
✅ **Configuration of memory providers** (in-memory, PostgreSQL, Weaviate)  
✅ **Implementation of RAG capabilities** for knowledge-enhanced responses  
✅ **Knowledge base management** including document upload and search  
✅ **Advanced memory configuration** for different scenarios  
✅ **Troubleshooting and optimization** of memory systems  
✅ **Best practices for memory-enabled agents** including privacy and security  

## Understanding Check

Before moving on, make sure you can:
- [ ] Configure different memory providers for various use cases
- [ ] Upload and manage documents in knowledge bases
- [ ] Create agents that remember conversations across sessions
- [ ] Implement RAG for knowledge-enhanced responses
- [ ] Troubleshoot common memory-related issues
- [ ] Apply memory optimization techniques
- [ ] Design memory systems with appropriate scope and privacy

## Next Steps

Your agents now have persistent memory and can access knowledge bases! Next, we'll expand their capabilities even further by integrating external tools through the MCP (Model Context Protocol) system.

**[→ Continue to Tool Integration](./tool-integration.md)**

---

::: tip Memory-Powered Intelligence
You've just transformed your agents from stateless responders into intelligent systems that learn, remember, and access vast knowledge bases. This is a crucial step toward building truly sophisticated AI systems.
:::