# Your First Scaffold (AgentCLI Anatomy)

Goal: Understand the generated project and where to customize.

## Prerequisites
Complete [install.md](install.md) first to set up your environment.

## 1) Create a scaffold
```pwsh
agentcli create my-first-agent --template basic
Set-Location my-first-agent
```

## 2) What you get
- **main.go**: app entry with runner from config
- **agentflow.toml**: configuration (providers, orchestration)
- **go.mod / go.sum**: module files
- **agents/**: individual agent implementations
- **README.md**: quick usage notes

Tip: Some templates include multiple agents and helper wrappers (e.g., result collector).

## 3) Configuration-first approach
Open `agentflow.toml` and review the key sections:
```toml
[llm]
provider = "openai"  # or "ollama", "azure"
model = "gpt-4"

[orchestration]
mode = "collaborative" # route|collaborative|sequential|loop|mixed

[agents.agent1]
role = "researcher"
description = "Gathers and analyzes information"
system_prompt = "You are a research assistant..."
```

Validate configuration:
```pwsh
agentcli validate
```

Expected: "Status: VALID"

## 4) Run the generated project
```pwsh
# Set OpenAI API key if using OpenAI provider
$Env:OPENAI_API_KEY="your-api-key"

go run . -m "Tell me about artificial intelligence"
```

Expected:
- Program starts, processes the message
- Agent responds with relevant information
- Clean shutdown

## 5) Customize your agents
To modify agent behavior:

1. **Change system prompts** in `agentflow.toml`:
   ```toml
   [agents.agent1]
   system_prompt = "You are an expert software developer..."
   ```

2. **Edit agent logic** in `agents/agent1.go`:
   - Modify the `Run()` method
   - Add custom processing logic
   - Handle different input types

3. **Switch orchestration** in `agentflow.toml`:
   ```toml
   [orchestration]
   mode = "sequential"  # Changes how agents interact
   ```

## 6) Next steps
- Try different templates: `agentcli config template --list`
- Learn orchestration: [orchestration-basics.md](orchestration-basics.md)
- Add memory: [memory-basics.md](memory-basics.md)

## Verification checklist
- [ ] agentcli create executed successfully
- [ ] agentcli validate returned "Status: VALID"
- [ ] go run . produced agent output
- [ ] System prompt modifications work
