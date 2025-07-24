# Production Deployment Tutorial (15 minutes)

## Overview

Learn how to deploy and scale your AgenticGoKit agents for production use. You'll containerize your application, set up monitoring, implement fault tolerance, and configure for high availability.

## Prerequisites

- Complete the [Tool Integration](tool-integration.md) tutorial
- Docker installed
- Basic understanding of containerization and deployment

## Learning Objectives

By the end of this tutorial, you'll understand:
- How to containerize AgenticGoKit applications
- Production configuration management
- Monitoring and observability setup
- Fault tolerance and error handling
- Scaling strategies for multi-agent systems

## What You'll Build

A production-ready agent system with:
1. **Docker containerization** for consistent deployment
2. **Configuration management** for different environments
3. **Monitoring and logging** for observability
4. **Fault tolerance** with circuit breakers and retries
5. **Health checks** and graceful shutdown

---

## Part 1: Containerization (5 minutes)

Package your agents for consistent deployment across environments.

### Create a Production-Ready Project

```bash
# Create production-ready project with all features
agentcli create production-system --memory-enabled --memory-provider pgvector \
  --mcp-production --with-cache --with-metrics --rag-enabled --agents 3
cd production-system
```

### Understanding the Generated Dockerfile

The project includes a multi-stage Dockerfile:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/agentflow.toml .
EXPOSE 8080
CMD ["./main"]
```

### Build and Test Locally

```bash
# Build the Docker image
docker build -t production-system:latest .

# Test the container
docker run --rm -e OPENAI_API_KEY=your-key production-system:latest
```

### Understanding Docker Compose Setup

The generated `docker-compose.yml` includes all services:

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - DATABASE_URL=postgres://agentflow:password@postgres:5432/agentflow
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  postgres:
    image: pgvector/pgvector:pg15
    environment:
      POSTGRES_DB: agentflow
      POSTGRES_USER: agentflow
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

---

## Part 2: Configuration Management (5 minutes)

Set up proper configuration for different environments.

### Environment-Specific Configuration

The project includes multiple configuration files:

**`agentflow.toml`** (Development):
```toml
[agent_flow]
name = "production-system"
version = "1.0.0"
provider = "openai"

[logging]
level = "debug"
format = "json"

[runtime]
max_concurrent_agents = 5
timeout_seconds = 30

[agent_memory]
provider = "pgvector"
connection = "postgres://agentflow:password@localhost:5432/agentflow"
```

**`agentflow.prod.toml`** (Production):
```toml
[agent_flow]
name = "production-system"
version = "1.0.0"
provider = "openai"

[logging]
level = "info"
format = "json"
file = "/var/log/agentflow.log"

[runtime]
max_concurrent_agents = 20
timeout_seconds = 60

[agent_memory]
provider = "pgvector"
connection = "${DATABASE_URL}"

[error_routing]
enabled = true
max_retries = 3
enable_circuit_breaker = true

[mcp.production]
enabled = true
metrics_enabled = true
circuit_breaker_enabled = true
```

### Environment Variable Management

Create `.env.example`:
```bash
# LLM Provider
OPENAI_API_KEY=your-openai-api-key-here

# Database
DATABASE_URL=postgres://agentflow:password@localhost:5432/agentflow

# External APIs
WEATHER_API_KEY=your-weather-api-key
WEB_SEARCH_API_KEY=your-search-api-key

# Application
LOG_LEVEL=info
MAX_CONCURRENT_AGENTS=20
```

### Configuration Loading

The generated code includes environment-aware configuration:

```go
func loadConfig() (*core.Config, error) {
    // Determine environment
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }
    
    // Load appropriate config file
    var configPath string
    switch env {
    case "production":
        configPath = "agentflow.prod.toml"
    case "staging":
        configPath = "agentflow.staging.toml"
    default:
        configPath = "agentflow.toml"
    }
    
    return core.LoadConfig(configPath)
}
```

---

## Part 3: Monitoring and Observability (5 minutes)

Set up comprehensive monitoring for production systems.

### Metrics and Health Checks

The production configuration includes metrics:

```toml
[mcp.metrics]
enabled = true
port = 8080
path = "/metrics"

[monitoring]
enable_health_checks = true
health_check_port = 8081
```

### Health Check Endpoint

The generated code includes health checks:

```go
func setupHealthChecks(runner core.Runner) {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        // Check runner status
        if runner == nil {
            http.Error(w, "Runner not initialized", http.StatusServiceUnavailable)
            return
        }
        
        // Check database connectivity
        if err := checkDatabase(); err != nil {
            http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
            return
        }
        
        // Check MCP tools
        if err := checkMCPTools(); err != nil {
            http.Error(w, "Tools unavailable", http.StatusServiceUnavailable)
            return
        }
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
            "timestamp": time.Now().Format(time.RFC3339),
        })
    })
    
    log.Println("Health checks available at :8081/health")
    go http.ListenAndServe(":8081", nil)
}
```

### Structured Logging

Production logging configuration:

```go
func setupLogging(config *core.Config) {
    // Configure structured logging
    logger := zerolog.New(os.Stdout).With().
        Timestamp().
        Str("service", "production-system").
        Str("version", "1.0.0").
        Logger()
    
    // Set global logger
    core.SetLogger(logger)
    
    // Log startup information
    logger.Info().
        Str("environment", os.Getenv("ENVIRONMENT")).
        Str("log_level", config.Logging.Level).
        Msg("Application starting")
}
```

### Metrics Collection

```go
func setupMetrics() {
    // Agent execution metrics
    agentExecutionTime := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "agent_execution_duration_seconds",
            Help: "Time spent executing agents",
        },
        []string{"agent_name", "status"},
    )
    
    // Tool execution metrics
    toolExecutionTime := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "tool_execution_duration_seconds",
            Help: "Time spent executing tools",
        },
        []string{"tool_name", "status"},
    )
    
    // Register metrics
    prometheus.MustRegister(agentExecutionTime, toolExecutionTime)
    
    // Expose metrics endpoint
    http.Handle("/metrics", promhttp.Handler())
}
```

## Production Deployment Strategies

### Docker Deployment

```bash
# Build for production
docker build -t production-system:v1.0.0 .

# Run with production config
docker run -d \
  --name production-system \
  -p 8080:8080 \
  -p 8081:8081 \
  -e ENVIRONMENT=production \
  -e OPENAI_API_KEY=${OPENAI_API_KEY} \
  -e DATABASE_URL=${DATABASE_URL} \
  --restart unless-stopped \
  production-system:v1.0.0
```

### Docker Compose Production

```bash
# Start all services
docker-compose -f docker-compose.prod.yml up -d

# Check service health
docker-compose ps
curl http://localhost:8081/health

# View logs
docker-compose logs -f app
```

### Kubernetes Deployment

Create `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: production-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: production-system
  template:
    metadata:
      labels:
        app: production-system
    spec:
      containers:
      - name: app
        image: production-system:v1.0.0
        ports:
        - containerPort: 8080
        - containerPort: 8081
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: openai-key
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: production-system-service
spec:
  selector:
    app: production-system
  ports:
  - name: app
    port: 8080
    targetPort: 8080
  - name: health
    port: 8081
    targetPort: 8081
```

## Fault Tolerance Configuration

### Circuit Breaker Setup

```toml
[error_routing.circuit_breaker]
failure_threshold = 10
success_threshold = 5
timeout_ms = 60000
max_concurrent_calls = 3

[mcp.production.circuit_breaker]
failure_threshold = 5
success_threshold = 3
timeout_ms = 30000
```

### Retry Policies

```toml
[error_routing.retry]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 30000
backoff_factor = 2.0
enable_jitter = true
```

### Graceful Shutdown

```go
func main() {
    // ... setup code ...
    
    // Setup graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        log.Println("Shutting down gracefully...")
        
        // Stop runner
        runner.Stop()
        
        // Close database connections
        if db != nil {
            db.Close()
        }
        
        // Close trace logger
        if traceLogger != nil {
            traceLogger.Close()
        }
        
        os.Exit(0)
    }()
    
    // ... start services ...
}
```

## Monitoring and Alerting

### Prometheus Metrics

Key metrics to monitor:
- Agent execution time and success rate
- Tool execution time and error rate
- Memory usage and garbage collection
- Database connection pool status
- Circuit breaker state changes

### Grafana Dashboard

Create dashboards for:
- System overview (CPU, memory, requests)
- Agent performance (execution time, error rate)
- Tool usage (calls per minute, cache hit rate)
- Database metrics (connections, query time)

### Alerting Rules

```yaml
# prometheus-alerts.yml
groups:
- name: agenticgokit
  rules:
  - alert: HighErrorRate
    expr: rate(agent_execution_errors_total[5m]) > 0.1
    for: 2m
    annotations:
      summary: "High error rate detected"
      
  - alert: DatabaseDown
    expr: up{job="postgres"} == 0
    for: 1m
    annotations:
      summary: "Database is down"
```

## Performance Optimization

### Resource Limits

```yaml
# In Kubernetes deployment
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Connection Pooling

```toml
[agent_memory]
provider = "pgvector"
connection = "${DATABASE_URL}"
max_connections = 20
idle_connections = 5
connection_lifetime = "1h"
```

### Caching Strategy

```toml
[mcp.cache]
backend = "redis"
redis_url = "${REDIS_URL}"
default_timeout_ms = 300000
max_size = 10000
```

## Troubleshooting Production Issues

### Common Issues

**High memory usage:**
```bash
# Check memory metrics
curl http://localhost:8080/metrics | grep memory

# Adjust configuration
[runtime]
max_concurrent_agents = 10  # Reduce if needed
```

**Database connection issues:**
```bash
# Check database connectivity
docker-compose exec postgres psql -U agentflow -d agentflow -c "SELECT 1;"

# Check connection pool
curl http://localhost:8081/health
```

**Tool execution failures:**
```bash
# Check MCP server status
agentcli mcp servers

# Check circuit breaker status
curl http://localhost:8080/metrics | grep circuit_breaker
```

### Debugging Commands

```bash
# View application logs
docker-compose logs -f app

# Check system resources
docker stats

# Monitor database
docker-compose exec postgres pg_stat_activity

# Test health endpoints
curl http://localhost:8081/health
curl http://localhost:8080/metrics
```

## Security Best Practices

### Environment Variables

```bash
# Use secrets management
export OPENAI_API_KEY=$(vault kv get -field=key secret/openai)

# Rotate keys regularly
# Use least-privilege access
# Monitor API usage
```

### Network Security

```yaml
# In docker-compose.yml
networks:
  internal:
    driver: bridge
    internal: true
  external:
    driver: bridge

services:
  app:
    networks:
      - internal
      - external
  postgres:
    networks:
      - internal  # Database not exposed externally
```

## Next Steps

Your agents are now production-ready! Consider:

1. **Advanced Monitoring**: Set up distributed tracing with Jaeger
2. **Auto-scaling**: Implement horizontal pod autoscaling in Kubernetes
3. **Multi-region**: Deploy across multiple regions for high availability
4. **CI/CD**: Set up automated deployment pipelines

## Key Takeaways

- **Containerization**: Docker ensures consistent deployment across environments
- **Configuration**: Environment-specific configs for different deployment stages
- **Monitoring**: Comprehensive observability with metrics, logs, and health checks
- **Fault Tolerance**: Circuit breakers and retries for resilient systems
- **Security**: Proper secret management and network isolation

## Further Reading

- [Advanced Patterns](../advanced/README.md) - Production patterns and best practices
- [Monitoring Guide](../debugging/performance-monitoring.md) - Deep dive into monitoring
- [Security Best Practices](../../guides/security.md) - Comprehensive security guide