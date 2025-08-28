# Quickstart (CLI)

Goal: Create and run a project using agentcli with config-first runtime.

## Steps (pwsh)

1) Create project (uses OpenAI provider)
```pwsh
agentcli create my-agents --template research-assistant -p openai
```

2) Enter project
```pwsh
Set-Location my-agents
```

3) Validate configuration
```pwsh
agentcli validate
```

4) Run
```pwsh
go run .
```

Expected
- CLI validation shows Status: VALID
- App starts without errors and prints agent responses

## Notes
- Set your OpenAI API key: `$Env:OPENAI_API_KEY="your-api-key"` (PowerShell) or `export OPENAI_API_KEY="your-api-key"` (Bash)
- Alternative providers: `-p azure` (requires `AZURE_OPENAI_*` env vars) | `-p ollama` (requires local Ollama running)
- List config templates first if needed: `agentcli config template --list`

## Bash equivalents
```bash
agentcli create my-agents --template research-assistant -p openai
cd my-agents
agentcli validate
go run .
```

## Verification checklist
- [ ] agentcli create executed successfully
- [ ] agentcli validate reported VALID
- [ ] go run . produced agent output
