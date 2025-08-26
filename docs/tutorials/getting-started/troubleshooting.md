---
title: "Troubleshooting"
description: "Debug and solve common issues in AgenticGoKit systems"
prev:
  text: "Building Workflows"
  link: "./building-workflows"
next:
  text: "Next Steps"
  link: "./next-steps"
---

# Troubleshooting

Even well-designed agent systems can encounter issues. This comprehensive troubleshooting guide will help you diagnose and solve common problems, debug complex workflows, and maintain healthy agent systems.

## Learning Objectives

By the end of this section, you'll be able to:
- Diagnose common AgenticGoKit issues systematically
- Use debugging tools and techniques effectively
- Troubleshoot configuration, connectivity, and performance problems
- Debug multi-agent workflows and orchestration issues
- Resolve memory and tool integration problems
- Implement monitoring and alerting for proactive issue detection
- Apply best practices for maintaining healthy agent systems

## Prerequisites

Before starting, make sure you've completed:
- ✅ [Building Workflows](./building-workflows.md) - Understanding of complex agent systems

## Systematic Troubleshooting Approach

When facing issues with AgenticGoKit systems, follow this systematic approach:

### 1. Identify the Problem
- **What** is happening vs. what should happen?
- **When** does the problem occur?
- **Where** in the system does it manifest?
- **How** consistently does it reproduce?

### 2. Gather Information
- Check logs and error messages
- Review configuration files
- Test individual components
- Monitor system resources

### 3. Form Hypotheses
- Based on symptoms, what could be causing the issue?
- What are the most likely root causes?
- How can you test each hypothesis?

### 4. Test and Validate
- Test hypotheses systematically
- Make one change at a time
- Document what works and what doesn't

### 5. Implement and Monitor
- Apply the solution
- Monitor to ensure the fix works
- Document the solution for future reference

## Common Issues and Solutions

### Installation and Setup Issues

#### "agentcli: command not found"

**Symptoms:**
```bash
$ agentcli version
bash: agentcli: command not found
```

**Diagnosis:**
```bash
# Check if Go is installed
go version

# Check GOPATH and GOBIN
go env GOPATH
go env GOBIN

# Check if GOPATH/bin is in PATH
echo $PATH
```

**Solutions:**
```bash
# Reinstall agentcli
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest

# Add GOPATH/bin to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:$(go env GOPATH)/bin

# Or use full path temporarily
$(go env GOPATH)/bin/agentcli version
```

#### "provider not registered" Error

**Symptoms:**
```
Error: LLM provider "openai" not registered
```

**Diagnosis:**
Check if plugin imports are missing in your `main.go`:

```bash
grep -n "import" main.go
```

**Solution:**
Add missing plugin imports to your `main.go`:

```go
import (
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/openai"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/azure"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/orchestrator/default"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/runner/default"
)
```

### Configuration Issues

#### Configuration Validation Failures

**Symptoms:**
```bash
$ agentcli validate
Error: Invalid configuration at line 15: missing required field 'system_prompt'
```

**Diagnosis Tools:**
```bash
# Validate TOML syntax
agentcli validate --verbose

# Check specific sections
agentcli config check --section agents

# Show configuration with resolved environment variables
agentcli config show --resolved
```

**Common Configuration Fixes:**

**Missing Required Fields:**
```toml
# Bad: Missing system_prompt
[agents.assistant]
role = "helper"

# Good: All required fields present
[agents.assistant]
role = "helper"
description = "A helpful assistant"
system_prompt = "You are a helpful assistant."
enabled = true
```

**Invalid TOML Syntax:**
```toml
# Bad: Missing quotes
[agents.assistant]
system_prompt = You are a helpful assistant.

# Good: Proper quoting
[agents.assistant]
system_prompt = "You are a helpful assistant."
```

**Environment Variable Issues:**
```bash
# Check if environment variables are set
env | grep -E "(OPENAI|AZURE|OLLAMA)"

# Test with explicit values
export OPENAI_API_KEY="your-actual-key"
agentcli validate
```

### LLM Provider Issues

#### OpenAI Connection Problems

**Symptoms:**
- "Invalid API key" errors
- "Rate limit exceeded" errors
- "Model not found" errors

**Diagnosis:**
```bash
# Test API key directly
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/models

# Check rate limits and usage
agentcli llm status --provider openai

# Verify model availability
agentcli llm models --provider openai
```

**Solutions:**
```bash
# Verify API key is correct
echo $OPENAI_API_KEY

# Check account status and billing
# Visit https://platform.openai.com/account/usage

# Use different model if current one is unavailable
# Edit agentflow.toml:
[llm]
model = "gpt-3.5-turbo"  # Instead of "gpt-4"
```

#### Azure OpenAI Issues

**Symptoms:**
- "Resource not found" errors
- "Deployment not found" errors
- Authentication failures

**Diagnosis:**
```bash
# Check all required environment variables
echo "Endpoint: $AZURE_OPENAI_ENDPOINT"
echo "Key: $AZURE_OPENAI_API_KEY"
echo "Deployment: $AZURE_OPENAI_DEPLOYMENT"

# Test Azure connection
curl -H "api-key: $AZURE_OPENAI_API_KEY" \
     "$AZURE_OPENAI_ENDPOINT/openai/deployments?api-version=2024-02-15-preview"
```

**Solutions:**
```bash
# Verify endpoint format (must include https://)
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"

# Check deployment name matches Azure portal
export AZURE_OPENAI_DEPLOYMENT="your-exact-deployment-name"

# Verify API version compatibility
export AZURE_OPENAI_API_VERSION="2024-02-15-preview"
```

#### Ollama Connection Issues

**Symptoms:**
- "Connection refused" errors
- "Model not found" errors
- Slow response times

**Diagnosis:**
```bash
# Check if Ollama is running
curl http://localhost:11434/api/version

# List available models
ollama list

# Check Ollama logs
ollama logs

# Test model directly
ollama run llama3.1:8b "Hello, world!"
```

**Solutions:**
```bash
# Start Ollama service
ollama serve

# Pull required model
ollama pull llama3.1:8b

# Check available system resources
free -h  # Memory
df -h    # Disk space

# Use smaller model if resources are limited
ollama pull gemma2:2b
```

### Multi-Agent Orchestration Issues

#### Agents Not Collaborating

**Symptoms:**
- Only one agent responds in collaborative mode
- Agents seem to ignore each other's work
- Inconsistent results from multi-agent workflows

**Diagnosis:**
```bash
# Enable debug logging
export AGENTICGOKIT_LOG_LEVEL=debug
go run . -m "Test message"

# Check agent registration
agentcli agents list

# Verify orchestration configuration
agentcli config show --section orchestration
```

**Solutions:**

**Fix Collaborative Agent List:**
```toml
[orchestration]
mode = "collaborative"
# Ensure all agents are listed
collaborative_agents = ["researcher", "analyzer", "writer"]

# Verify agent names match exactly
[agents.researcher]  # Must match name in collaborative_agents
role = "researcher"
# ...
```

**Check Agent System Prompts:**
```toml
[agents.researcher]
system_prompt = """
You are part of a research team. Your role is to gather information.
Work with the analyzer and writer to create comprehensive reports.
Share your findings clearly so other agents can build on your work.
"""
```

#### Sequential Pipeline Breaks

**Symptoms:**
- Pipeline stops at a specific agent
- Agents can't process previous agent's output
- Data loss between pipeline stages

**Diagnosis:**
```bash
# Test each agent individually
agentcli test-agent researcher "Test input"
agentcli test-agent analyzer "Test input"

# Check pipeline configuration
agentcli config show --section orchestration

# Monitor pipeline execution
export AGENTICGOKIT_PIPELINE_DEBUG=true
go run . -m "Test message"
```

**Solutions:**

**Fix Sequential Agent Order:**
```toml
[orchestration]
mode = "sequential"
# Ensure logical order
sequential_agents = ["collector", "validator", "processor", "analyzer", "reporter"]
```

**Improve Agent Handoffs:**
```toml
[agents.validator]
system_prompt = """
Process the data from the collector and prepare it for the processor.
Always output your results in a structured format that the processor can understand:

VALIDATED_DATA:
- Field1: value
- Field2: value
- Status: VALID/INVALID
- Notes: any important observations
"""
```

### Memory System Issues

#### Memory Not Persisting

**Symptoms:**
- Agents don't remember previous conversations
- Knowledge base searches return no results
- Memory-related errors in logs

**Diagnosis:**
```bash
# Check memory provider status
agentcli memory status

# Test database connection
agentcli memory test-connection

# Check memory configuration
agentcli config show --section agent_memory

# List stored memories
agentcli memory list --limit 10
```

**Solutions:**

**Database Connection Issues:**
```bash
# Test PostgreSQL connection
psql "postgresql://agent:agentpass@localhost:5432/agentdb" -c "SELECT 1;"

# Check if pgvector extension is installed
psql "postgresql://agent:agentpass@localhost:5432/agentdb" -c "SELECT * FROM pg_extension WHERE extname = 'vector';"

# Install pgvector if missing
psql "postgresql://agent:agentpass@localhost:5432/agentdb" -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

**Memory Configuration Fixes:**
```toml
[agent_memory]
provider = "pgvector"
enable_rag = true
enable_knowledge_base = true

[agent_memory.pgvector]
# Ensure connection string is correct
connection_string = "postgresql://agent:agentpass@localhost:5432/agentdb"

[agents.memory_agent]
# Ensure memory is enabled for agents
memory_enabled = true
```

#### RAG Not Finding Relevant Information

**Symptoms:**
- Knowledge base searches return empty results
- Agents claim no relevant information exists
- Poor quality search results

**Diagnosis:**
```bash
# Check if documents are uploaded
agentcli knowledge list

# Test search directly
agentcli knowledge search "test query"

# Check chunk configuration
agentcli config show --section agent_memory

# Verify vector embeddings
agentcli memory debug --type embeddings
```

**Solutions:**

**Upload Documents:**
```bash
# Upload documents to knowledge base
agentcli knowledge upload ./docs/

# Verify upload
agentcli knowledge list

# Test search
agentcli knowledge search "your search term"
```

**Optimize RAG Configuration:**
```toml
[agent_memory]
chunk_size = 800           # Smaller chunks for better precision
overlap_size = 200         # Overlap for context continuity
max_results = 8            # More results for comprehensive answers
similarity_threshold = 0.6  # Lower threshold for more results
```

### Tool Integration Issues

#### MCP Tools Not Available

**Symptoms:**
- "Tool not found" errors
- Agents claim they can't perform actions
- MCP server connection failures

**Diagnosis:**
```bash
# Check MCP server status
agentcli mcp health

# List available tools
agentcli mcp tools

# Test specific server
agentcli mcp test web-search

# Check server logs
agentcli mcp logs filesystem
```

**Solutions:**

**Install Missing Dependencies:**
```bash
# Install uv/uvx for MCP servers
curl -LsSf https://astral.sh/uv/install.sh | sh

# Test uvx installation
uvx --version

# Install specific MCP server
uvx mcp-server-web-search --help
```

**Fix MCP Configuration:**
```toml
[[mcp.servers]]
name = "web-search"
command = "uvx"
args = ["mcp-server-web-search"]
# Add required environment variables
env = { "SEARCH_ENGINE" = "duckduckgo" }
```

**Check Tool Permissions:**
```toml
[[mcp.servers]]
name = "filesystem"
command = "uvx"
args = ["mcp-server-filesystem"]
env = { 
    "ALLOWED_DIRECTORIES" = "./workspace,./data",
    "MAX_FILE_SIZE" = "10MB"
}
```

#### Tool Permission Errors

**Symptoms:**
- "Permission denied" errors
- "Access restricted" messages
- Tools fail silently

**Diagnosis:**
```bash
# Check file permissions
ls -la ./workspace/

# Verify environment variables
env | grep -E "(ALLOWED|DENIED)"

# Test tool access directly
agentcli mcp test filesystem --operation read --path ./workspace/test.txt
```

**Solutions:**
```bash
# Fix directory permissions
chmod 755 ./workspace/
chmod 644 ./workspace/*

# Update MCP server configuration
# Edit agentflow.toml:
[[mcp.servers]]
name = "filesystem"
env = { 
    "ALLOWED_DIRECTORIES" = "./workspace,./reports",
    "ALLOWED_EXTENSIONS" = ".txt,.md,.json,.csv"
}
```

### Performance Issues

#### Slow Agent Responses

**Symptoms:**
- Long wait times for agent responses
- Timeouts in multi-agent workflows
- High resource usage

**Diagnosis:**
```bash
# Monitor system resources
top
htop
free -h

# Check agent performance
agentcli monitor --agent researcher --duration 60s

# Profile memory usage
agentcli profile --type memory

# Check network latency (for cloud LLMs)
ping api.openai.com
```

**Solutions:**

**Optimize LLM Settings:**
```toml
[llm]
model = "gpt-3.5-turbo"  # More efficient than gpt-4
max_tokens = 1000        # Reduce for quicker responses
temperature = 0.7        # Doesn't affect speed significantly
```

**Optimize Agent Configuration:**
```toml
[orchestration]
timeout_seconds = 120    # Reasonable timeout
max_concurrent_agents = 3 # Limit concurrent agents

[agent_memory]
max_results = 3          # Fewer RAG results for speed
chunk_size = 500         # Smaller chunks process more efficiently
```

**Use Local Models for Development:**
```toml
[llm]
provider = "ollama"
model = "gemma2:2b"      # Efficient, lightweight model
host = "http://localhost:11434"
```

#### Memory Usage Issues

**Symptoms:**
- Out of memory errors
- System becomes unresponsive
- Gradual memory leaks

**Diagnosis:**
```bash
# Monitor memory usage over time
watch -n 5 'free -h'

# Check Go memory stats
agentcli debug --type memory

# Profile memory allocation
go tool pprof http://localhost:6060/debug/pprof/heap
```

**Solutions:**
```toml
# Limit memory usage
[agent_memory]
conversation_memory_limit = 20  # Limit conversation history
enable_compression = true       # Compress old memories
cleanup_interval = "1h"         # Regular cleanup

[orchestration]
max_concurrent_agents = 2       # Reduce concurrent agents
```

## Debugging Tools and Techniques

### Enable Debug Logging

```bash
# Enable comprehensive debugging
export AGENTICGOKIT_LOG_LEVEL=debug
export AGENTICGOKIT_MEMORY_DEBUG=true
export AGENTICGOKIT_MCP_DEBUG=true
export AGENTICGOKIT_TOOL_DEBUG=true

# Run with debugging enabled
go run . -m "Debug test message"
```

### Use CLI Debugging Commands

```bash
# Test individual components
agentcli test-agent researcher "Test query"
agentcli test-memory "Test memory operation"
agentcli test-tool web-search "Test search"

# Health checks
agentcli health-check
agentcli mcp health
agentcli memory health

# Configuration validation
agentcli validate --verbose
agentcli config check --all

# Performance monitoring
agentcli monitor --duration 300s
agentcli metrics --timeframe 1h
```

### Log Analysis

```bash
# Filter logs by component
grep "ERROR" agenticgokit.log
grep "memory" agenticgokit.log
grep "mcp" agenticgokit.log

# Analyze performance
grep "duration" agenticgokit.log | sort -n

# Find error patterns
grep -E "(failed|error|timeout)" agenticgokit.log | sort | uniq -c
```

## Monitoring and Alerting

### Basic Monitoring Setup

```toml
[monitoring]
enabled = true
log_level = "info"
metrics_collection = true
health_check_interval = "30s"

[monitoring.alerts]
max_execution_time = 300
memory_threshold = "1GB"
error_rate_threshold = 0.1
```

### Health Check Endpoints

```bash
# Check overall system health
curl http://localhost:8080/health

# Check specific components
curl http://localhost:8080/health/memory
curl http://localhost:8080/health/mcp
curl http://localhost:8080/health/agents
```

### Automated Monitoring

```bash
#!/bin/bash
# health-check.sh - Simple monitoring script

check_health() {
    if ! agentcli health-check > /dev/null 2>&1; then
        echo "ALERT: AgenticGoKit health check failed"
        # Send notification (email, Slack, etc.)
    fi
}

# Run periodically
while true; do
    check_health
    sleep 300
done
```

## Best Practices for Troubleshooting

### 1. Implement Comprehensive Logging

```toml
[agents.well_logged_agent]
system_prompt = """
Always log your decision-making process:
- What information you received
- How you interpreted it
- What actions you decided to take
- Why you made those decisions
- Any issues or limitations encountered

This helps with debugging and improvement.
"""
```

### 2. Use Structured Error Handling

```go
// Example error handling pattern
func (a *Agent) Run(ctx context.Context, state core.State) (core.State, error) {
    log.Info("Agent starting", "agent", a.Name(), "input_keys", state.Keys())
    
    result, err := a.processInput(ctx, state)
    if err != nil {
        log.Error("Agent processing failed", 
            "agent", a.Name(), 
            "error", err,
            "input_size", len(state.Keys()))
        return state, fmt.Errorf("agent %s failed: %w", a.Name(), err)
    }
    
    log.Info("Agent completed", "agent", a.Name(), "output_keys", result.Keys())
    return result, nil
}
```

### 3. Create Reproducible Test Cases

```bash
# Create test cases for common scenarios
mkdir -p tests/scenarios/

# Test case for memory issues
cat > tests/scenarios/memory-test.sh << 'EOF'
#!/bin/bash
echo "Testing memory persistence..."
go run . -m "Remember that my name is Alice"
go run . -m "What is my name?"
EOF

# Test case for tool integration
cat > tests/scenarios/tool-test.sh << 'EOF'
#!/bin/bash
echo "Testing web search tool..."
go run . -m "Search for the latest news about AI"
EOF
```

### 4. Document Known Issues

```markdown
# Known Issues and Workarounds

## Issue: Memory not persisting after restart
**Symptoms**: Agents forget previous conversations
**Cause**: Database connection not properly configured
**Workaround**: Verify connection string and restart database
**Fix**: Update connection string in agentflow.toml

## Issue: Slow responses with GPT-4
**Symptoms**: Long wait times for responses
**Cause**: GPT-4 is slower than GPT-3.5-turbo
**Workaround**: Use GPT-3.5-turbo for development
**Fix**: Optimize prompts and reduce max_tokens
```

## What You've Learned

✅ **Systematic troubleshooting approach** for diagnosing issues  
✅ **Common problem patterns** and their solutions  
✅ **Debugging tools and techniques** for complex systems  
✅ **Performance optimization** strategies  
✅ **Monitoring and alerting** setup for proactive issue detection  
✅ **Best practices** for maintainable, debuggable systems  
✅ **Documentation and knowledge sharing** for team environments  

## Understanding Check

Before moving on, make sure you can:
- [ ] Diagnose common AgenticGoKit issues systematically
- [ ] Use CLI tools and logging for debugging
- [ ] Troubleshoot configuration and connectivity problems
- [ ] Debug multi-agent workflows and orchestration issues
- [ ] Resolve memory and tool integration problems
- [ ] Set up monitoring and alerting for production systems
- [ ] Document and share troubleshooting knowledge

## Next Steps

Congratulations! You've completed the comprehensive AgenticGoKit getting-started tutorial. You now have the knowledge and skills to build sophisticated, production-ready agent systems. Let's explore what comes next in your AgenticGoKit journey.

**[→ Continue to Next Steps](./next-steps.md)**

---

::: details Quick Navigation

**Previous:** [Building Workflows](./building-workflows.md) - Complex system integration  
**Next:** [Next Steps](./next-steps.md) - Continue your AgenticGoKit journey  
**Jump to:** [Installation](./installation.md) - If you're having setup issues  

:::

::: details Advanced Troubleshooting Resources

**For Production Systems:**
- [Deployment Guide](../../guides/deployment/README.md) - Production deployment patterns
- [Monitoring Guide](../../guides/monitoring/README.md) - System health and observability
- [Performance Optimization](../../guides/performance/README.md) - Scaling and optimization

**For Development:**
- [Debugging Guide](../debugging/README.md) - Advanced debugging techniques
- [Testing Strategies](../../guides/testing/README.md) - Comprehensive testing approaches
- [Development Best Practices](../../guides/development/README.md) - Code quality and patterns

**Community Support:**
- [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions) - Get help from the community
- [Known Issues](https://github.com/kunalkushwaha/agenticgokit/issues?q=is%3Aissue+label%3Abug) - Check for known problems
- [FAQ](../../guides/faq.md) - Frequently asked questions

:::

::: tip Troubleshooting Mastery
You now have the skills to diagnose and solve issues in complex agent systems. These troubleshooting techniques will serve you well as you build and maintain sophisticated AgenticGoKit applications in production environments.
:::