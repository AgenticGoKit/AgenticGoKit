# Deploy Basics

Goal: Deploy AgenticGoKit applications to production environments with proper configuration and monitoring.

## Prerequisites
Complete [tools-basics.md](tools-basics.md) to understand MCP tool integration.

## What is Deployment?
Deployment involves:
- Building production-ready binaries
- Configuring environment variables
- Containerizing applications
- Managing secrets securely
- Setting up monitoring

## 1) Validate configuration for production
```pwsh
agentcli validate
```

Expected: All configuration validates without errors before deployment.

## 2) Build production binary
```pwsh
# Build optimized binary
go build -ldflags="-s -w" -o agentapp .
```

This creates a production binary with debug information removed for smaller size.

## 3) Create production Dockerfile
```dockerfile
FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o agentapp .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/agentapp .
COPY --from=build /app/agentflow.toml .
EXPOSE 8080
CMD ["./agentapp"]
```

## 4) Build Docker image
```pwsh
docker build -t agenticgokit-app:v1.0.0 .
```

Expected: Image builds successfully with optimized layers.

## 5) Configure environment variables
Create production environment setup:
```pwsh
# Required API keys
$env:OPENAI_API_KEY = "your-production-api-key"

# Optional: Custom configuration file
$env:AGENTFLOW_CONFIG = "agentflow.prod.toml"

# Optional: Database connection for memory
$env:DATABASE_URL = "postgres://user:pass@db:5432/agents"
```

## 6) Run containerized application
```pwsh
docker run --rm `
  -e OPENAI_API_KEY=$env:OPENAI_API_KEY `
  -e DATABASE_URL=$env:DATABASE_URL `
  -p 8080:8080 `
  agenticgokit-app:v1.0.0
```

Expected: Application starts and responds on port 8080.

## 7) Production configuration example
Create `agentflow.prod.toml`:
```toml
[agent_llm]
provider = "openai"
model = "gpt-4o"
max_tokens = 4000
temperature = 0.7

[agent_memory]
provider = "pgvector"
connection = "${DATABASE_URL}"

[agent_logging]
level = "info"
format = "json"
output = "stdout"

[agent_tracing]
enabled = true
```

## 8) Health monitoring
Add health check endpoint to your application:
```go
// Add to main.go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

Test health endpoint:
```pwsh
curl http://localhost:8080/health
```

## 9) Common deployment patterns
### Cloud platforms
- **AWS**: Use ECS, EKS, or Lambda
- **Google Cloud**: Use Cloud Run, GKE
- **Azure**: Use Container Instances, AKS

### Container orchestration
- **Docker Compose**: Multi-service local deployment
- **Kubernetes**: Production-grade orchestration
- **Docker Swarm**: Simplified container clustering

## Security considerations
- Store API keys in secure secret management
- Use TLS for all external communications
- Implement proper access controls
- Regular security updates

## Next Steps
- Explore advanced deployment patterns in [Examples.md](../../guides/Examples.md)
- Learn about production monitoring in [ErrorHandling.md](../../guides/ErrorHandling.md)
- Review security best practices in [Providers.md](../../guides/Providers.md)

## Verification checklist
- [ ] agentcli validate succeeded
- [ ] Production binary built successfully
- [ ] Docker image built without errors
- [ ] Environment variables configured
- [ ] Container runs and responds to requests
- [ ] Health check endpoint accessible
- [ ] Production configuration file created
