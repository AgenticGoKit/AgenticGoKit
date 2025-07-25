# Deployment & Operations Guides

Guides for deploying and operating AgenticGoKit applications in production.

## Available Guides

### Docker Deployment
Containerize your AgenticGoKit applications with Docker, including multi-stage builds and optimization techniques.

**When to use:** Deploying agents in containerized environments or cloud platforms.

### Monitoring
Set up comprehensive monitoring for agent performance, including metrics, logging, and alerting.

**When to use:** Running agents in production and need visibility into system health and performance.

### Scaling
Scale AgenticGoKit applications horizontally, including load balancing and distributed deployment patterns.

**When to use:** Handling increased load or building high-availability agent systems.

## Deployment Patterns

Common deployment patterns:

### Single Instance
- Simple applications with low traffic
- Development and testing environments
- Proof-of-concept deployments

### Load Balanced
- Multiple instances behind a load balancer
- Horizontal scaling for increased throughput
- High availability with failover

### Microservices
- Agents deployed as separate services
- Independent scaling and deployment
- Service mesh integration

## Production Checklist

Before deploying to production:

- [ ] **Configuration management** - Externalized configuration
- [ ] **Security** - Proper authentication and authorization
- [ ] **Monitoring** - Metrics, logging, and alerting set up
- [ ] **Backup and recovery** - Data backup and disaster recovery plans
- [ ] **Performance testing** - Load testing and capacity planning
- [ ] **Documentation** - Operational runbooks and procedures

## Infrastructure Requirements

Typical production requirements:
- **Compute** - CPU and memory based on agent complexity
- **Storage** - Persistent storage for state and memory systems
- **Network** - Reliable network connectivity for LLM APIs
- **Monitoring** - Observability infrastructure

## Next Steps

For production deployment:
1. Start with Docker containerization
2. Add monitoring for observability
3. Scale with appropriate patterns as needed
4. Reference [Best Practices](../development/best-practices.md) for operational excellence