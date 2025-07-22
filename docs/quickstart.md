# 🚀 5-Minute Quickstart

Get your first AgenticGoKit multi-agent system running in 5 minutes. No complex setup, no configuration files—just working code.

## What You'll Build

A simple but powerful multi-agent system where:
- 🤖 **Agent 1** processes your request
- 🤖 **Agent 2** enhances the response  
- 🤖 **Agent 3** formats the final output

All working together automatically!

## Prerequisites

- **Go 1.21+** ([install here](https://golang.org/dl/))
- **OpenAI API Key** ([get one here](https://platform.openai.com/api-keys))

*That's it! No Docker, no databases, no complex setup.*

---

## Step 1: Create Your Project (30 seconds)

```bash
# Create a simple collaborative multi-agent project
agentcli create my-agents --agents 3 --orchestration-mode collaborative

# Enter the project directory
cd my-agents
```

## Step 2: Configure Your API Key (30 seconds)

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your-api-key-here

# Or create a .env file (recommended)
echo "OPENAI_API_KEY=your-api-key-here" > .env
```

## Step 3: Run Your Multi-Agent System! (30 seconds)

```bash
export OPENAI_API_KEY=your-api-key-here
go run main.go
```

You should see:
```
🤖 Starting multi-agent collaboration...

✅ Multi-Agent Processing Complete!
==================================================
📊 Execution Stats:
   • Agents involved: 3
   • Trace entries: 12
   • Session ID: quickstart-demo
   • First event: 14:32:15
   • Last event: 14:32:18
==================================================

🎉 Check the trace file: quickstart-demo.trace.json
💡 Run: agentcli trace quickstart-demo
```

---

## 🎉 Congratulations!

You just created a **multi-agent system** that:
- ✅ Runs three agents in parallel
- ✅ Combines their outputs intelligently  
- ✅ Handles errors gracefully
- ✅ Provides execution metrics

**And it took less than 5 minutes!**

---

## 🤔 What Just Happened?

### The Magic Behind the Scenes

1. **🏗️ Agent Creation**: Each agent has a specialized role and system prompt
2. **🤝 Collaborative Orchestration**: `CreateCollaborativeRunner` makes agents work together
3. **⚡ Parallel Processing**: All agents process your message simultaneously
4. **🧠 Intelligent Combination**: Results are automatically merged and enhanced
5. **📊 Built-in Monitoring**: You get metrics and error handling for free

### Key Concepts You Just Used

- **`core.AgentHandler`**: The interface for all agents
- **`core.CreateCollaborativeRunner`**: Orchestrates multiple agents in parallel
- **`runner.Emit(event)`**: Sends events to agents for processing
- **`core.NewEvent()`**: Creates events with data and metadata
- **System Prompts**: Give each agent a specialized role and personality
- **Tracing**: Built-in execution tracking with `agentcli trace`

---

## 🚀 Next Steps

Now that you have a working multi-agent system, here's what to explore next:

### 🎓 **15-Minute Tutorials** (Choose Your Path)

<table>
<tr>
<td width="50%">

**🤝 Multi-Agent Patterns**
Learn different orchestration modes:
- Collaborative (parallel)
- Sequential (pipeline)  
- Mixed (hybrid workflows)

[→ Multi-Agent Tutorial](tutorials/multi-agent.md)

</td>
<td width="50%">

**🧠 Memory & RAG**
Add persistent memory and knowledge:
- Vector databases
- Document ingestion
- Semantic search

[→ Memory Tutorial](tutorials/memory-rag.md)

</td>
</tr>
<tr>
<td width="50%">

**🔧 Tool Integration**
Connect to external tools:
- Web search
- File operations
- API integrations
- Custom tools

[→ Tools Tutorial](tutorials/tools.md)

</td>
<td width="50%">

**🏭 Production Ready**
Deploy and scale your agents:
- Docker deployment
- Monitoring setup
- Performance optimization

[→ Production Tutorial](tutorials/production.md)

</td>
</tr>
</table>

### 🎯 **Quick Wins** (5-10 minutes each)

- **[🔄 Try Sequential Processing](how-to/sequential-agents.md)** - Build a data processing pipeline
- **[🌐 Add Web Search](how-to/web-search.md)** - Give your agents internet access
- **[💾 Add Memory](how-to/add-memory.md)** - Make agents remember conversations
- **[📊 Add Monitoring](how-to/monitoring.md)** - See what your agents are doing

### 🏗️ **Build Something Cool**

Ready to build a real application? Try these examples:

```bash
# Research assistant with web search and analysis
agentcli create research-assistant --mcp-enabled --mcp-tools "web_search,summarize"

# Data processing pipeline with error handling  
agentcli create data-pipeline --orchestration-mode sequential --agents 4

# Chat system with persistent memory
agentcli create chat-system --memory-enabled --memory-provider pgvector

# Knowledge base with document ingestion and RAG
agentcli create knowledge-base --memory-enabled --memory-provider pgvector --rag-enabled
```

---

## 🆘 Need Help?

### Common Issues

**❌ "OpenAI API key not found"**
```bash
# Make sure your API key is set
export OPENAI_API_KEY=sk-your-key-here
echo $OPENAI_API_KEY  # Should show your key
```

**❌ "Module not found"**
```bash
# Make sure you're in the right directory and ran go mod init
go mod tidy
```

**❌ "Context deadline exceeded"**
```bash
# Increase the timeout if agents are taking too long
runner := core.CreateCollaborativeRunner(agents, 60*time.Second) // Increased to 60s
```

### Get Support

- **💬 [Discord Community](https://discord.gg/agenticgokit)** - Real-time help
- **💡 [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)** - Q&A
- **📖 [Troubleshooting Guide](troubleshooting.md)** - Common solutions

---

## 🎯 What's Next?

You've successfully created your first multi-agent system! Here are some paths to continue your AgenticGoKit journey:

<div align="center">

### [🎓 **Take the 15-Minute Tutorial**](tutorials/multi-agent.md)
*Learn advanced orchestration patterns*

### [🏗️ **Build a Real Application**](../examples/)
*Explore production-ready examples*

### [📖 **Read the Full Documentation**](../README.md)
*Dive deep into all features*

</div>

---

*⏱️ **Actual time**: Most developers complete this in 3-4 minutes. The extra minute is for reading and understanding!*