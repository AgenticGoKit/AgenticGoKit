# Load Balancing and Scaling Patterns

**Building horizontally scalable agent systems with intelligent load distribution**

As your agent systems grow, you need strategies to distribute load effectively and scale horizontally. This guide covers load balancing patterns, horizontal scaling strategies, and performance optimization techniques for AgenticGoKit applications.

## 🏗️ Scaling Architecture Overview

```mermaid
flowchart TB
    subgraph "Load Balancer Layer"
        LB[Load Balancer]
        HEALTH[Health Checker]
    end
    
    subgraph "Agent Instances"
        AGENT1[Agent Instance 1]
        AGENT2[Agent Instance 2]
        AGENT3[Agent Instance 3]
        AGENTN[Agent Instance N]
    end
    
    subgraph "Shared Resources"
        REDIS[Redis Cache]
        POSTGRES[PostgreSQL]
        VECTOR[Vector DB]
        MCP[MCP Servers]
    end
    
    subgraph "Monitoring"
        METRICS[Metrics Collector]
        ALERTS[Alert Manager]
    end
    
    CLIENT[Client Requests] --> LB
    LB --> AGENT1
    LB --> AGENT2
    LB --> AGENT3
    LB --> AGENTN
    
    HEALTH --> AGENT1
    HEALTH --> AGENT2
    HEALTH --> AGENT3
    HEALTH --> AGENTN
    
    AGENT1 --> REDIS
    AGENT2 --> REDIS
    AGENT3 --> REDIS
    AGENTN --> REDIS
    
    AGENT1 --> POSTGRES
    AGENT2 --> POSTGRES
    AGENT3 --> POSTGRES
    AGENTN --> POSTGRES
    
    AGENT1 --> VECTOR
    AGENT2 --> VECTOR
    AGENT3 --> VECTOR
    AGENTN --> VECTOR
    
    AGENT1 --> MCP
    AGENT2 --> MCP
    AGENT3 --> MCP
    AGENTN --> MCP
    
    AGENT1 --> METRICS
    AGENT2 --> METRICS
    AGENT3 --> METRICS
    AGENTN --> METRICS
    
    METRICS --> ALERTS
    
    classDef lb fill:#e3f2fd,stroke:#1976d2,stroke-width:2px
    classDef agent fill:#e8f5e8,stroke:#388e3c,stroke-width:2px
    classDef shared fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    classDef monitor fill:#fce4ec,stroke:#c2185b,stroke-width:2px
    
    class LB,HEALTH lb
    class AGENT1,AGENT2,AGENT3,AGENTN agent
    class REDIS,POSTGRES,VECTOR,MCP shared
    class METRICS,ALERTS monitor
```

## ⚖️ Load Balancing Strategies

### **1. Round Robin Load Balancing**

```go
package balancer

import (
    "context"
    "sync"
    "sync/atomic"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

type RoundRobinBalancer struct {
    agents  []core.AgentHandler
    counter int64
    mu      sync.RWMutex
}

func NewRoundRobinBalancer(agents []core.AgentHandler) *RoundRobinBalancer {
    return &RoundRobinBalancer{
        agents: agents,
    }
}

func (rb *RoundRobinBalancer) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    if len(rb.agents) == 0 {
        return nil, errors.New("no agents available")
    }
    
    // Get next agent in round-robin fashion
    index := atomic.AddInt64(&rb.counter, 1) % int64(len(rb.agents))
    agent := rb.agents[index]
    
    return agent.Execute(ctx, event, state)
}

func (rb *RoundRobinBalancer) AddAgent(agent core.AgentHandler) {
    rb.mu.Lock()
    defer rb.mu.Unlock()
    
    rb.agents = append(rb.agents, agent)
}

func (rb *RoundRobinBalancer) RemoveAgent(index int) {
    rb.mu.Lock()
    defer rb.mu.Unlock()
    
    if index >= 0 && index < len(rb.agents) {
        rb.agents = append(rb.agents[:index], rb.agents[index+1:]...)
    }
}
```

### **2. Weighted Load Balancing**

```go
type WeightedBalancer struct {
    agents  []WeightedAgent
    totalWeight int
    mu      sync.RWMutex
}

type WeightedAgent struct {
    Agent  core.AgentHandler
    Weight int
    CurrentWeight int
}

func NewWeightedBalancer() *WeightedBalancer {
    return &WeightedBalancer{
        agents: make([]WeightedAgent, 0),
    }
}

func (wb *WeightedBalancer) AddAgent(agent core.AgentHandler, weight int) {
    wb.mu.Lock()
    defer wb.mu.Unlock()
    
    wb.agents = append(wb.agents, WeightedAgent{
        Agent:  agent,
        Weight: weight,
        CurrentWeight: 0,
    })
    wb.totalWeight += weight
}

func (wb *WeightedBalancer) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    wb.mu.Lock()
    defer wb.mu.Unlock()
    
    if len(wb.agents) == 0 {
        return nil, errors.New("no agents available")
    }
    
    // Weighted round-robin algorithm
    selected := wb.selectAgent()
    if selected == nil {
        return nil, errors.New("no agent selected")
    }
    
    return selected.Execute(ctx, event, state)
}

func (wb *WeightedBalancer) selectAgent() core.AgentHandler {
    var selected *WeightedAgent
    
    for i := range wb.agents {
        agent := &wb.agents[i]
        agent.CurrentWeight += agent.Weight
        
        if selected == nil || agent.CurrentWeight > selected.CurrentWeight {
            selected = agent
        }
    }
    
    if selected != nil {
        selected.CurrentWeight -= wb.totalWeight
        return selected.Agent
    }
    
    return nil
}
```

### **3. Least Connections Load Balancing**

```go
type LeastConnectionsBalancer struct {
    agents []ConnectionTrackingAgent
    mu     sync.RWMutex
}

type ConnectionTrackingAgent struct {
    Agent       core.AgentHandler
    Connections int64
    mu          sync.RWMutex
}

func NewLeastConnectionsBalancer(agents []core.AgentHandler) *LeastConnectionsBalancer {
    trackingAgents := make([]ConnectionTrackingAgent, len(agents))
    for i, agent := range agents {
        trackingAgents[i] = ConnectionTrackingAgent{
            Agent: agent,
        }
    }
    
    return &LeastConnectionsBalancer{
        agents: trackingAgents,
    }
}

func (lcb *LeastConnectionsBalancer) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    agent := lcb.selectLeastConnectedAgent()
    if agent == nil {
        return nil, errors.New("no agents available")
    }
    
    // Increment connection count
    atomic.AddInt64(&agent.Connections, 1)
    defer atomic.AddInt64(&agent.Connections, -1)
    
    return agent.Agent.Execute(ctx, event, state)
}

func (lcb *LeastConnectionsBalancer) selectLeastConnectedAgent() *ConnectionTrackingAgent {
    lcb.mu.RLock()
    defer lcb.mu.RUnlock()
    
    if len(lcb.agents) == 0 {
        return nil
    }
    
    var selected *ConnectionTrackingAgent
    minConnections := int64(^uint64(0) >> 1) // Max int64
    
    for i := range lcb.agents {
        agent := &lcb.agents[i]
        connections := atomic.LoadInt64(&agent.Connections)
        
        if connections < minConnections {
            minConnections = connections
            selected = agent
        }
    }
    
    return selected
}
```

### **4. Performance-Based Load Balancing**

```go
type PerformanceBasedBalancer struct {
    agents []PerformanceTrackingAgent
    mu     sync.RWMutex
}

type PerformanceTrackingAgent struct {
    Agent           core.AgentHandler
    ResponseTimes   []time.Duration
    SuccessRate     float64
    LastUpdate      time.Time
    mu              sync.RWMutex
}

func NewPerformanceBasedBalancer(agents []core.AgentHandler) *PerformanceBasedBalancer {
    trackingAgents := make([]PerformanceTrackingAgent, len(agents))
    for i, agent := range agents {
        trackingAgents[i] = PerformanceTrackingAgent{
            Agent:         agent,
            ResponseTimes: make([]time.Duration, 0, 100),
            SuccessRate:   1.0,
            LastUpdate:    time.Now(),
        }
    }
    
    return &PerformanceBasedBalancer{
        agents: trackingAgents,
    }
}

func (pbb *PerformanceBasedBalancer) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    agent := pbb.selectBestPerformingAgent()
    if agent == nil {
        return nil, errors.New("no agents available")
    }
    
    start := time.Now()
    result, err := agent.Agent.Execute(ctx, event, state)
    duration := time.Since(start)
    
    // Record performance metrics
    pbb.recordPerformance(agent, duration, err == nil)
    
    return result, err
}

func (pbb *PerformanceBasedBalancer) selectBestPerformingAgent() *PerformanceTrackingAgent {
    pbb.mu.RLock()
    defer pbb.mu.RUnlock()
    
    if len(pbb.agents) == 0 {
        return nil
    }
    
    var selected *PerformanceTrackingAgent
    bestScore := -1.0
    
    for i := range pbb.agents {
        agent := &pbb.agents[i]
        score := pbb.calculatePerformanceScore(agent)
        
        if score > bestScore {
            bestScore = score
            selected = agent
        }
    }
    
    return selected
}

func (pbb *PerformanceBasedBalancer) calculatePerformanceScore(agent *PerformanceTrackingAgent) float64 {
    agent.mu.RLock()
    defer agent.mu.RUnlock()
    
    // Calculate average response time
    avgResponseTime := time.Duration(0)
    if len(agent.ResponseTimes) > 0 {
        total := time.Duration(0)
        for _, rt := range agent.ResponseTimes {
            total += rt
        }
        avgResponseTime = total / time.Duration(len(agent.ResponseTimes))
    }
    
    // Normalize response time (lower is better)
    responseTimeScore := 1.0 / (1.0 + avgResponseTime.Seconds())
    
    // Combine success rate and response time
    return (agent.SuccessRate * 0.7) + (responseTimeScore * 0.3)
}

func (pbb *PerformanceBasedBalancer) recordPerformance(agent *PerformanceTrackingAgent, duration time.Duration, success bool) {
    agent.mu.Lock()
    defer agent.mu.Unlock()
    
    // Record response time
    agent.ResponseTimes = append(agent.ResponseTimes, duration)
    if len(agent.ResponseTimes) > 100 {
        agent.ResponseTimes = agent.ResponseTimes[1:]
    }
    
    // Update success rate (exponential moving average)
    alpha := 0.1
    if success {
        agent.SuccessRate = agent.SuccessRate*(1-alpha) + alpha
    } else {
        agent.SuccessRate = agent.SuccessRate * (1 - alpha)
    }
    
    agent.LastUpdate = time.Now()
}
```

## 🔄 Horizontal Scaling Patterns

### **1. Stateless Agent Design**

```go
// Stateless agent that can be scaled horizontally
type StatelessAgent struct {
    name        string
    llmProvider core.ModelProvider
    cache       CacheProvider    // Shared cache
    memory      MemoryProvider   // Shared memory
}

func NewStatelessAgent(name string, provider core.ModelProvider, cache CacheProvider, memory MemoryProvider) *StatelessAgent {
    return &StatelessAgent{
        name:        name,
        llmProvider: provider,
        cache:       cache,
        memory:      memory,
    }
}

func (sa *StatelessAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    // All state is passed in or retrieved from shared resources
    // No local state that would prevent horizontal scaling
    
    // Get context from shared memory
    context := sa.memory.GetContext(state.SessionID)
    
    // Check shared cache
    cacheKey := sa.generateCacheKey(event)
    if cached := sa.cache.Get(cacheKey); cached != nil {
        return cached.(*core.AgentResult), nil
    }
    
    // Process request
    result, err := sa.processRequest(ctx, event, context)
    if err != nil {
        return nil, err
    }
    
    // Store in shared cache
    sa.cache.Set(cacheKey, result, 5*time.Minute)
    
    return result, nil
}

func (sa *StatelessAgent) generateCacheKey(event core.Event) string {
    return fmt.Sprintf("%s:%s:%x", sa.name, event.Type, 
        sha256.Sum256([]byte(fmt.Sprintf("%v", event.Data))))
}
```

### **2. Auto-Scaling Controller**

```go
type AutoScaler struct {
    minInstances    int
    maxInstances    int
    currentInstances int
    targetCPU       float64
    targetMemory    float64
    scaleUpCooldown time.Duration
    scaleDownCooldown time.Duration
    lastScaleAction time.Time
    metrics         MetricsProvider
    orchestrator    Orchestrator
    mu              sync.RWMutex
}

func NewAutoScaler(min, max int, orchestrator Orchestrator) *AutoScaler {
    return &AutoScaler{
        minInstances:      min,
        maxInstances:      max,
        currentInstances:  min,
        targetCPU:         70.0, // 70% CPU utilization
        targetMemory:      80.0, // 80% memory utilization
        scaleUpCooldown:   2 * time.Minute,
        scaleDownCooldown: 5 * time.Minute,
        orchestrator:      orchestrator,
    }
}

func (as *AutoScaler) Start(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            as.evaluateScaling()
        case <-ctx.Done():
            return
        }
    }
}

func (as *AutoScaler) evaluateScaling() {
    as.mu.Lock()
    defer as.mu.Unlock()
    
    // Get current metrics
    cpuUsage := as.metrics.GetCPUUsage()
    memoryUsage := as.metrics.GetMemoryUsage()
    requestRate := as.metrics.GetRequestRate()
    
    // Determine if scaling is needed
    shouldScaleUp := (cpuUsage > as.targetCPU || memoryUsage > as.targetMemory) &&
                     as.currentInstances < as.maxInstances &&
                     time.Since(as.lastScaleAction) > as.scaleUpCooldown
    
    shouldScaleDown := cpuUsage < as.targetCPU*0.5 && memoryUsage < as.targetMemory*0.5 &&
                       as.currentInstances > as.minInstances &&
                       time.Since(as.lastScaleAction) > as.scaleDownCooldown
    
    if shouldScaleUp {
        as.scaleUp()
    } else if shouldScaleDown {
        as.scaleDown()
    }
}

func (as *AutoScaler) scaleUp() {
    newInstances := min(as.currentInstances+1, as.maxInstances)
    if newInstances > as.currentInstances {
        err := as.orchestrator.AddInstance()
        if err == nil {
            as.currentInstances = newInstances
            as.lastScaleAction = time.Now()
            log.Printf("Scaled up to %d instances", as.currentInstances)
        }
    }
}

func (as *AutoScaler) scaleDown() {
    newInstances := max(as.currentInstances-1, as.minInstances)
    if newInstances < as.currentInstances {
        err := as.orchestrator.RemoveInstance()
        if err == nil {
            as.currentInstances = newInstances
            as.lastScaleAction = time.Now()
            log.Printf("Scaled down to %d instances", as.currentInstances)
        }
    }
}
```

### **3. Health-Aware Load Balancing**

```go
type HealthAwareBalancer struct {
    agents      []HealthTrackingAgent
    healthCheck time.Duration
    mu          sync.RWMutex
}

type HealthTrackingAgent struct {
    Agent     core.AgentHandler
    Healthy   bool
    LastCheck time.Time
    Endpoint  string
}

func NewHealthAwareBalancer(agents []core.AgentHandler) *HealthAwareBalancer {
    trackingAgents := make([]HealthTrackingAgent, len(agents))
    for i, agent := range agents {
        trackingAgents[i] = HealthTrackingAgent{
            Agent:   agent,
            Healthy: true,
        }
    }
    
    hab := &HealthAwareBalancer{
        agents:      trackingAgents,
        healthCheck: 30 * time.Second,
    }
    
    // Start health checking
    go hab.startHealthChecking()
    
    return hab
}

func (hab *HealthAwareBalancer) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    healthyAgents := hab.getHealthyAgents()
    if len(healthyAgents) == 0 {
        return nil, errors.New("no healthy agents available")
    }
    
    // Use round-robin among healthy agents
    index := rand.Intn(len(healthyAgents))
    return healthyAgents[index].Execute(ctx, event, state)
}

func (hab *HealthAwareBalancer) getHealthyAgents() []core.AgentHandler {
    hab.mu.RLock()
    defer hab.mu.RUnlock()
    
    var healthy []core.AgentHandler
    for _, agent := range hab.agents {
        if agent.Healthy {
            healthy = append(healthy, agent.Agent)
        }
    }
    
    return healthy
}

func (hab *HealthAwareBalancer) startHealthChecking() {
    ticker := time.NewTicker(hab.healthCheck)
    defer ticker.Stop()
    
    for range ticker.C {
        hab.checkHealth()
    }
}

func (hab *HealthAwareBalancer) checkHealth() {
    hab.mu.Lock()
    defer hab.mu.Unlock()
    
    for i := range hab.agents {
        agent := &hab.agents[i]
        
        // Perform health check
        healthy := hab.performHealthCheck(agent)
        
        if agent.Healthy != healthy {
            agent.Healthy = healthy
            if healthy {
                log.Printf("Agent %d is now healthy", i)
            } else {
                log.Printf("Agent %d is now unhealthy", i)
            }
        }
        
        agent.LastCheck = time.Now()
    }
}

func (hab *HealthAwareBalancer) performHealthCheck(agent *HealthTrackingAgent) bool {
    // Simple health check - try to execute a test request
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    testEvent := core.NewEvent("health_check", "ping")
    testState := &core.State{}
    
    _, err := agent.Agent.Execute(ctx, testEvent, testState)
    return err == nil
}
```

## 🐳 Container Orchestration

### **Docker Compose for Development**

```yaml
# docker-compose.yml
version: '3.8'

services:
  agent-1:
    build: .
    environment:
      - INSTANCE_ID=1
      - REDIS_URL=redis://redis:6379
      - POSTGRES_URL=postgres://user:pass@postgres:5432/agentflow
    depends_on:
      - redis
      - postgres
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M

  agent-2:
    build: .
    environment:
      - INSTANCE_ID=2
      - REDIS_URL=redis://redis:6379
      - POSTGRES_URL=postgres://user:pass@postgres:5432/agentflow
    depends_on:
      - redis
      - postgres
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M

  agent-3:
    build: .
    environment:
      - INSTANCE_ID=3
      - REDIS_URL=redis://redis:6379
      - POSTGRES_URL=postgres://user:pass@postgres:5432/agentflow
    depends_on:
      - redis
      - postgres

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - agent-1
      - agent-2
      - agent-3

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  postgres:
    image: pgvector/pgvector:pg15
    environment:
      POSTGRES_DB: agentflow
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  redis_data:
  postgres_data:
```

### **Kubernetes Deployment**

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agenticgokit-agents
spec:
  replicas: 3
  selector:
    matchLabels:
      app: agenticgokit-agents
  template:
    metadata:
      labels:
        app: agenticgokit-agents
    spec:
      containers:
      - name: agent
        image: agenticgokit:latest
        ports:
        - containerPort: 8080
        env:
        - name: REDIS_URL
          value: "redis://redis-service:6379"
        - name: POSTGRES_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: postgres-url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: agenticgokit-service
spec:
  selector:
    app: agenticgokit-agents
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: agenticgokit-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: agenticgokit-agents
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## 📊 Performance Monitoring

### **Load Balancer Metrics**

```go
type LoadBalancerMetrics struct {
    totalRequests     int64
    requestsPerAgent  map[string]int64
    responseTimesByAgent map[string][]time.Duration
    errorsByAgent     map[string]int64
    mu                sync.RWMutex
}

func NewLoadBalancerMetrics() *LoadBalancerMetrics {
    return &LoadBalancerMetrics{
        requestsPerAgent:     make(map[string]int64),
        responseTimesByAgent: make(map[string][]time.Duration),
        errorsByAgent:        make(map[string]int64),
    }
}

func (lbm *LoadBalancerMetrics) RecordRequest(agentID string, duration time.Duration, success bool) {
    lbm.mu.Lock()
    defer lbm.mu.Unlock()
    
    lbm.totalRequests++
    lbm.requestsPerAgent[agentID]++
    
    // Record response time
    if lbm.responseTimesByAgent[agentID] == nil {
        lbm.responseTimesByAgent[agentID] = make([]time.Duration, 0, 100)
    }
    lbm.responseTimesByAgent[agentID] = append(lbm.responseTimesByAgent[agentID], duration)
    
    // Keep only recent response times
    if len(lbm.responseTimesByAgent[agentID]) > 100 {
        lbm.responseTimesByAgent[agentID] = lbm.responseTimesByAgent[agentID][1:]
    }
    
    // Record errors
    if !success {
        lbm.errorsByAgent[agentID]++
    }
}

func (lbm *LoadBalancerMetrics) GetStats() map[string]interface{} {
    lbm.mu.RLock()
    defer lbm.mu.RUnlock()
    
    stats := map[string]interface{}{
        "total_requests": lbm.totalRequests,
        "agents": make(map[string]interface{}),
    }
    
    for agentID, requests := range lbm.requestsPerAgent {
        agentStats := map[string]interface{}{
            "requests": requests,
            "errors":   lbm.errorsByAgent[agentID],
        }
        
        // Calculate average response time
        if responseTimes := lbm.responseTimesByAgent[agentID]; len(responseTimes) > 0 {
            total := time.Duration(0)
            for _, rt := range responseTimes {
                total += rt
            }
            agentStats["avg_response_time"] = total / time.Duration(len(responseTimes))
        }
        
        stats["agents"].(map[string]interface{})[agentID] = agentStats
    }
    
    return stats
}
```

## 🎯 Best Practices

### **1. Configuration for Scaling**

```toml
# agentflow.toml - Production scaling configuration
[runtime]
max_concurrent_agents = 50
timeout_seconds = 30
enable_metrics = true
metrics_port = 8080

[load_balancer]
strategy = "performance_based"  # round_robin, weighted, least_connections, performance_based
health_check_interval = "30s"
unhealthy_threshold = 3
healthy_threshold = 2

[scaling]
min_instances = 2
max_instances = 20
target_cpu_utilization = 70
target_memory_utilization = 80
scale_up_cooldown = "2m"
scale_down_cooldown = "5m"

[shared_resources]
redis_url = "${REDIS_URL}"
postgres_url = "${POSTGRES_URL}"
vector_db_url = "${VECTOR_DB_URL}"
```

### **2. Health Check Implementation**

```go
func (a *ScalableAgent) HealthCheck() map[string]interface{} {
    health := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    // Check LLM provider
    if err := a.checkLLMProvider(); err != nil {
        health["status"] = "unhealthy"
        health["llm_provider_error"] = err.Error()
    }
    
    // Check shared cache
    if err := a.checkCache(); err != nil {
        health["status"] = "degraded"
        health["cache_error"] = err.Error()
    }
    
    // Check memory system
    if err := a.checkMemory(); err != nil {
        health["status"] = "degraded"
        health["memory_error"] = err.Error()
    }
    
    // Add performance metrics
    health["metrics"] = map[string]interface{}{
        "requests_per_minute": a.getRequestsPerMinute(),
        "avg_response_time":   a.getAverageResponseTime(),
        "error_rate":          a.getErrorRate(),
    }
    
    return health
}
```

### **3. Graceful Shutdown**

```go
func (a *ScalableAgent) GracefulShutdown(ctx context.Context) error {
    log.Println("Starting graceful shutdown...")
    
    // Stop accepting new requests
    a.stopAcceptingRequests()
    
    // Wait for ongoing requests to complete
    done := make(chan struct{})
    go func() {
        a.waitForOngoingRequests()
        close(done)
    }()
    
    select {
    case <-done:
        log.Println("All requests completed")
    case <-ctx.Done():
        log.Println("Shutdown timeout reached, forcing shutdown")
    }
    
    // Clean up resources
    a.cleanup()
    
    log.Println("Graceful shutdown completed")
    return nil
}
```

Load balancing and scaling are essential for building production-ready agent systems that can handle varying loads efficiently and maintain high availability.

## 🚀 Next Steps

- **[Testing Strategies](testing-strategies.md)** - Test your scaled agent systems
- **[Circuit Breaker Patterns](circuit-breaker-patterns.md)** - Add fault tolerance to scaled systems
- **[Production Monitoring](../debugging/performance-monitoring.md)** - Monitor your scaled deployments
- **[Production Deployment](../getting-started/production-deployment.md)** - Deploy your scaled systems
