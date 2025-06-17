package mcp

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

// LoadBalancingStrategy defines the load balancing strategy
type LoadBalancingStrategy int

const (
	RoundRobin LoadBalancingStrategy = iota
	LeastConnections
	WeightedRoundRobin
	Random
	HealthBased
	ResponseTimeBased
)

// String returns the string representation of LoadBalancingStrategy
func (s LoadBalancingStrategy) String() string {
	switch s {
	case RoundRobin:
		return "round_robin"
	case LeastConnections:
		return "least_connections"
	case WeightedRoundRobin:
		return "weighted_round_robin"
	case Random:
		return "random"
	case HealthBased:
		return "health_based"
	case ResponseTimeBased:
		return "response_time_based"
	default:
		return "unknown"
	}
}

// ServerEndpoint represents a server endpoint with metadata
type ServerEndpoint struct {
	ID              string
	Address         string
	Weight          int
	MaxConnections  int
	HealthScore     atomic.Value // float64
	ResponseTime    atomic.Value // time.Duration
	ConnectionCount atomic.Int64
	FailureCount    atomic.Int64
	LastHealthCheck atomic.Value // time.Time
	Available       atomic.Bool
	Tools           []string
	Metadata        map[string]string
}

// NewServerEndpoint creates a new server endpoint
func NewServerEndpoint(id, address string, weight int, tools []string) *ServerEndpoint {
	ep := &ServerEndpoint{
		ID:             id,
		Address:        address,
		Weight:         weight,
		MaxConnections: 100, // Default max connections
		Tools:          tools,
		Metadata:       make(map[string]string),
	}

	ep.HealthScore.Store(1.0)
	ep.ResponseTime.Store(time.Duration(0))
	ep.LastHealthCheck.Store(time.Now())
	ep.Available.Store(true)

	return ep
}

// GetHealthScore returns the current health score
func (ep *ServerEndpoint) GetHealthScore() float64 {
	return ep.HealthScore.Load().(float64)
}

// SetHealthScore updates the health score
func (ep *ServerEndpoint) SetHealthScore(score float64) {
	ep.HealthScore.Store(score)
}

// GetResponseTime returns the current response time
func (ep *ServerEndpoint) GetResponseTime() time.Duration {
	return ep.ResponseTime.Load().(time.Duration)
}

// SetResponseTime updates the response time
func (ep *ServerEndpoint) SetResponseTime(duration time.Duration) {
	ep.ResponseTime.Store(duration)
}

// GetLastHealthCheck returns the last health check time
func (ep *ServerEndpoint) GetLastHealthCheck() time.Time {
	return ep.LastHealthCheck.Load().(time.Time)
}

// SetLastHealthCheck updates the last health check time
func (ep *ServerEndpoint) SetLastHealthCheck(t time.Time) {
	ep.LastHealthCheck.Store(t)
}

// IsAvailable returns whether the endpoint is available
func (ep *ServerEndpoint) IsAvailable() bool {
	return ep.Available.Load()
}

// SetAvailable updates the availability status
func (ep *ServerEndpoint) SetAvailable(available bool) {
	ep.Available.Store(available)
}

// IncConnections increments the connection count
func (ep *ServerEndpoint) IncConnections() {
	ep.ConnectionCount.Add(1)
}

// DecConnections decrements the connection count
func (ep *ServerEndpoint) DecConnections() {
	ep.ConnectionCount.Add(-1)
}

// GetConnections returns the current connection count
func (ep *ServerEndpoint) GetConnections() int64 {
	return ep.ConnectionCount.Load()
}

// IncFailures increments the failure count
func (ep *ServerEndpoint) IncFailures() {
	ep.FailureCount.Add(1)
}

// GetFailures returns the current failure count
func (ep *ServerEndpoint) GetFailures() int64 {
	return ep.FailureCount.Load()
}

// HasTool checks if the endpoint supports a specific tool
func (ep *ServerEndpoint) HasTool(toolName string) bool {
	for _, tool := range ep.Tools {
		if tool == toolName {
			return true
		}
	}
	return false
}

// LoadBalancer manages multiple server endpoints and provides load balancing
type LoadBalancer struct {
	strategy   LoadBalancingStrategy
	endpoints  map[string]*ServerEndpoint
	toolMap    map[string][]*ServerEndpoint // Maps tools to available endpoints
	mutex      sync.RWMutex
	roundRobin map[string]*int64 // Round robin counters per tool (pointers for atomic ops)
	logger     *zerolog.Logger

	// Health checking
	healthChecker *EndpointHealthChecker

	// Metrics
	metrics *MCPMetrics
}

// LoadBalancerConfig contains load balancer configuration
type LoadBalancerConfig struct {
	Strategy              LoadBalancingStrategy `toml:"strategy"`
	HealthCheckInterval   time.Duration         `toml:"health_check_interval"`
	HealthCheckTimeout    time.Duration         `toml:"health_check_timeout"`
	UnhealthyThreshold    int                   `toml:"unhealthy_threshold"`
	HealthyThreshold      int                   `toml:"healthy_threshold"`
	FailoverEnabled       bool                  `toml:"failover_enabled"`
	CircuitBreakerEnabled bool                  `toml:"circuit_breaker_enabled"`
}

// DefaultLoadBalancerConfig returns default load balancer configuration
func DefaultLoadBalancerConfig() *LoadBalancerConfig {
	return &LoadBalancerConfig{
		Strategy:              RoundRobin,
		HealthCheckInterval:   30 * time.Second,
		HealthCheckTimeout:    5 * time.Second,
		UnhealthyThreshold:    3,
		HealthyThreshold:      2,
		FailoverEnabled:       true,
		CircuitBreakerEnabled: true,
	}
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(config *LoadBalancerConfig, logger *zerolog.Logger, metrics *MCPMetrics) *LoadBalancer {
	if config == nil {
		config = DefaultLoadBalancerConfig()
	}
	lb := &LoadBalancer{
		strategy:   config.Strategy,
		endpoints:  make(map[string]*ServerEndpoint),
		toolMap:    make(map[string][]*ServerEndpoint),
		roundRobin: make(map[string]*int64),
		logger:     logger,
		metrics:    metrics,
	}

	// Initialize health checker
	lb.healthChecker = NewEndpointHealthChecker(config, lb, logger)

	return lb
}

// AddEndpoint adds a server endpoint to the load balancer
func (lb *LoadBalancer) AddEndpoint(endpoint *ServerEndpoint) error {
	if endpoint == nil {
		return errors.New("endpoint cannot be nil")
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// Add to endpoints map
	lb.endpoints[endpoint.ID] = endpoint

	// Update tool mapping
	for _, tool := range endpoint.Tools {
		if _, exists := lb.toolMap[tool]; !exists {
			lb.toolMap[tool] = make([]*ServerEndpoint, 0)
		}
		lb.toolMap[tool] = append(lb.toolMap[tool], endpoint)
	}

	// Start health checking for this endpoint
	lb.healthChecker.AddEndpoint(endpoint)
	lb.logger.Info().
		Str("endpoint_id", endpoint.ID).
		Str("address", endpoint.Address).
		Interface("tools", endpoint.Tools).
		Msg("Added endpoint to load balancer")

	return nil
}

// RemoveEndpoint removes a server endpoint from the load balancer
func (lb *LoadBalancer) RemoveEndpoint(endpointID string) error {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	endpoint, exists := lb.endpoints[endpointID]
	if !exists {
		return fmt.Errorf("endpoint %s not found", endpointID)
	}

	// Remove from endpoints map
	delete(lb.endpoints, endpointID)

	// Update tool mapping
	for _, tool := range endpoint.Tools {
		if endpoints, exists := lb.toolMap[tool]; exists {
			filtered := make([]*ServerEndpoint, 0, len(endpoints))
			for _, ep := range endpoints {
				if ep.ID != endpointID {
					filtered = append(filtered, ep)
				}
			}
			lb.toolMap[tool] = filtered
		}
	}

	// Stop health checking for this endpoint
	lb.healthChecker.RemoveEndpoint(endpointID)
	lb.logger.Info().
		Str("endpoint_id", endpointID).
		Msg("Removed endpoint from load balancer")

	return nil
}

// SelectEndpoint selects an endpoint for executing a tool
func (lb *LoadBalancer) SelectEndpoint(toolName string) (*ServerEndpoint, error) {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	// Get available endpoints for the tool
	candidates, exists := lb.toolMap[toolName]
	if !exists || len(candidates) == 0 {
		return nil, fmt.Errorf("no endpoints available for tool %s", toolName)
	}

	// Filter available endpoints
	available := make([]*ServerEndpoint, 0, len(candidates))
	for _, ep := range candidates {
		if ep.IsAvailable() && ep.GetHealthScore() > 0.1 {
			available = append(available, ep)
		}
	}

	if len(available) == 0 {
		return nil, fmt.Errorf("no healthy endpoints available for tool %s", toolName)
	}

	// Select endpoint based on strategy
	var selected *ServerEndpoint
	var err error

	switch lb.strategy {
	case RoundRobin:
		selected = lb.selectRoundRobin(toolName, available)
	case LeastConnections:
		selected = lb.selectLeastConnections(available)
	case WeightedRoundRobin:
		selected = lb.selectWeightedRoundRobin(toolName, available)
	case Random:
		selected = lb.selectRandom(available)
	case HealthBased:
		selected = lb.selectHealthBased(available)
	case ResponseTimeBased:
		selected = lb.selectResponseTimeBased(available)
	default:
		selected = lb.selectRoundRobin(toolName, available)
	}

	if selected == nil {
		return nil, fmt.Errorf("failed to select endpoint for tool %s", toolName)
	}

	// Increment connection count
	selected.IncConnections()

	return selected, err
}

// ReleaseEndpoint releases an endpoint after use
func (lb *LoadBalancer) ReleaseEndpoint(endpoint *ServerEndpoint) {
	if endpoint != nil {
		endpoint.DecConnections()
	}
}

// selectRoundRobin selects endpoint using round robin strategy
func (lb *LoadBalancer) selectRoundRobin(toolName string, endpoints []*ServerEndpoint) *ServerEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	// Get or create counter for this tool
	if _, exists := lb.roundRobin[toolName]; !exists {
		lb.roundRobin[toolName] = new(int64)
	}

	counter := atomic.AddInt64(lb.roundRobin[toolName], 1)
	index := int(counter-1) % len(endpoints)
	return endpoints[index]
}

// selectLeastConnections selects endpoint with least connections
func (lb *LoadBalancer) selectLeastConnections(endpoints []*ServerEndpoint) *ServerEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	var selected *ServerEndpoint
	minConnections := int64(-1)

	for _, ep := range endpoints {
		connections := ep.GetConnections()
		if minConnections == -1 || connections < minConnections {
			minConnections = connections
			selected = ep
		}
	}

	return selected
}

// selectWeightedRoundRobin selects endpoint using weighted round robin
func (lb *LoadBalancer) selectWeightedRoundRobin(toolName string, endpoints []*ServerEndpoint) *ServerEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	// Calculate total weight
	totalWeight := 0
	for _, ep := range endpoints {
		totalWeight += ep.Weight
	}

	if totalWeight == 0 {
		return lb.selectRoundRobin(toolName, endpoints)
	}

	// Get or create counter for this tool
	if _, exists := lb.roundRobin[toolName]; !exists {
		lb.roundRobin[toolName] = new(int64)
	}

	// Select based on weight
	counter := atomic.AddInt64(lb.roundRobin[toolName], 1)
	target := int(counter) % totalWeight

	currentWeight := 0
	for _, ep := range endpoints {
		currentWeight += ep.Weight
		if currentWeight > target {
			return ep
		}
	}

	return endpoints[0]
}

// selectRandom selects a random endpoint
func (lb *LoadBalancer) selectRandom(endpoints []*ServerEndpoint) *ServerEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	index := rand.Intn(len(endpoints))
	return endpoints[index]
}

// selectHealthBased selects endpoint based on health score
func (lb *LoadBalancer) selectHealthBased(endpoints []*ServerEndpoint) *ServerEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	// Sort by health score (descending)
	sorted := make([]*ServerEndpoint, len(endpoints))
	copy(sorted, endpoints)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetHealthScore() > sorted[j].GetHealthScore()
	})

	return sorted[0]
}

// selectResponseTimeBased selects endpoint based on response time
func (lb *LoadBalancer) selectResponseTimeBased(endpoints []*ServerEndpoint) *ServerEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	// Sort by response time (ascending)
	sorted := make([]*ServerEndpoint, len(endpoints))
	copy(sorted, endpoints)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetResponseTime() < sorted[j].GetResponseTime()
	})

	return sorted[0]
}

// GetEndpoints returns all endpoints
func (lb *LoadBalancer) GetEndpoints() map[string]*ServerEndpoint {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	result := make(map[string]*ServerEndpoint)
	for id, ep := range lb.endpoints {
		result[id] = ep
	}

	return result
}

// GetEndpointsForTool returns endpoints that support a specific tool
func (lb *LoadBalancer) GetEndpointsForTool(toolName string) []*ServerEndpoint {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	if endpoints, exists := lb.toolMap[toolName]; exists {
		result := make([]*ServerEndpoint, len(endpoints))
		copy(result, endpoints)
		return result
	}

	return nil
}

// GetStats returns load balancer statistics
func (lb *LoadBalancer) GetStats() map[string]interface{} {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	stats := map[string]interface{}{
		"strategy":        lb.strategy.String(),
		"total_endpoints": len(lb.endpoints),
		"tools_supported": len(lb.toolMap),
		"endpoints":       make(map[string]interface{}),
	}

	healthyCount := 0
	totalConnections := int64(0)

	for id, ep := range lb.endpoints {
		if ep.IsAvailable() {
			healthyCount++
		}

		connections := ep.GetConnections()
		totalConnections += connections

		endpointStats := map[string]interface{}{
			"available":     ep.IsAvailable(),
			"health_score":  ep.GetHealthScore(),
			"connections":   connections,
			"failures":      ep.GetFailures(),
			"response_time": ep.GetResponseTime().String(),
			"tools":         ep.Tools,
		}

		stats["endpoints"].(map[string]interface{})[id] = endpointStats
	}

	stats["healthy_endpoints"] = healthyCount
	stats["total_connections"] = totalConnections

	return stats
}

// Start starts the load balancer
func (lb *LoadBalancer) Start(ctx context.Context) error {
	return lb.healthChecker.Start(ctx)
}

// Stop stops the load balancer
func (lb *LoadBalancer) Stop() error {
	return lb.healthChecker.Stop()
}

// EndpointHealthChecker performs health checks on endpoints
type EndpointHealthChecker struct {
	config       *LoadBalancerConfig
	loadBalancer *LoadBalancer
	logger       *zerolog.Logger
	endpoints    map[string]*ServerEndpoint
	mutex        sync.RWMutex
	ctx          context.Context
	cancelFunc   context.CancelFunc
	wg           sync.WaitGroup
}

// NewEndpointHealthChecker creates a new endpoint health checker
func NewEndpointHealthChecker(config *LoadBalancerConfig, lb *LoadBalancer, logger *zerolog.Logger) *EndpointHealthChecker {
	return &EndpointHealthChecker{
		config:       config,
		loadBalancer: lb,
		logger:       logger,
		endpoints:    make(map[string]*ServerEndpoint),
	}
}

// AddEndpoint adds an endpoint for health checking
func (hc *EndpointHealthChecker) AddEndpoint(endpoint *ServerEndpoint) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.endpoints[endpoint.ID] = endpoint
}

// RemoveEndpoint removes an endpoint from health checking
func (hc *EndpointHealthChecker) RemoveEndpoint(endpointID string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	delete(hc.endpoints, endpointID)
}

// Start starts the health checker
func (hc *EndpointHealthChecker) Start(ctx context.Context) error {
	hc.ctx, hc.cancelFunc = context.WithCancel(ctx)

	hc.wg.Add(1)
	go hc.healthCheckLoop()
	hc.logger.Info().
		Dur("interval", hc.config.HealthCheckInterval).
		Msg("Started endpoint health checker")

	return nil
}

// Stop stops the health checker
func (hc *EndpointHealthChecker) Stop() error {
	if hc.cancelFunc != nil {
		hc.cancelFunc()
	}

	hc.wg.Wait()

	hc.logger.Info().Msg("Stopped endpoint health checker")
	return nil
}

// healthCheckLoop runs the health check loop
func (hc *EndpointHealthChecker) healthCheckLoop() {
	defer hc.wg.Done()

	ticker := time.NewTicker(hc.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks on all endpoints
func (hc *EndpointHealthChecker) performHealthChecks() {
	hc.mutex.RLock()
	endpoints := make([]*ServerEndpoint, 0, len(hc.endpoints))
	for _, ep := range hc.endpoints {
		endpoints = append(endpoints, ep)
	}
	hc.mutex.RUnlock()

	for _, endpoint := range endpoints {
		go hc.checkEndpointHealth(endpoint)
	}
}

// checkEndpointHealth performs a health check on a single endpoint
func (hc *EndpointHealthChecker) checkEndpointHealth(endpoint *ServerEndpoint) {
	ctx, cancel := context.WithTimeout(hc.ctx, hc.config.HealthCheckTimeout)
	defer cancel()

	start := time.Now()
	healthy := hc.performHealthCheck(ctx, endpoint)
	duration := time.Since(start)

	endpoint.SetResponseTime(duration)
	endpoint.SetLastHealthCheck(time.Now())

	if healthy {
		// Update health score positively
		currentScore := endpoint.GetHealthScore()
		newScore := currentScore + (1.0-currentScore)*0.1 // Gradual improvement
		endpoint.SetHealthScore(newScore)
		if !endpoint.IsAvailable() {
			endpoint.SetAvailable(true)
			hc.logger.Info().
				Str("endpoint_id", endpoint.ID).
				Float64("health_score", newScore).
				Msg("Endpoint marked as healthy")
		}
	} else {
		// Update health score negatively
		endpoint.IncFailures()
		currentScore := endpoint.GetHealthScore()
		newScore := currentScore * 0.8 // Gradual degradation
		endpoint.SetHealthScore(newScore)

		if newScore < 0.1 && endpoint.IsAvailable() {
			endpoint.SetAvailable(false)
			hc.logger.Warn().
				Str("endpoint_id", endpoint.ID).
				Float64("health_score", newScore).
				Int64("failures", endpoint.GetFailures()).
				Msg("Endpoint marked as unhealthy")
		}
	}

	// Update metrics
	if hc.loadBalancer.metrics != nil {
		status := "success"
		if !healthy {
			status = "failure"
		}
		hc.loadBalancer.metrics.RecordConnection(endpoint.ID, status, duration)
	}
}

// performHealthCheck performs the actual health check
func (hc *EndpointHealthChecker) performHealthCheck(ctx context.Context, endpoint *ServerEndpoint) bool {
	// This is a placeholder implementation
	// In a real implementation, you would:
	// 1. Make a connection to the endpoint
	// 2. Send a health check request (e.g., ping)
	// 3. Verify the response

	// For now, we'll simulate a health check
	select {
	case <-ctx.Done():
		return false
	case <-time.After(time.Duration(rand.Intn(100)) * time.Millisecond):
		// Simulate 95% success rate
		return rand.Float32() < 0.95
	}
}
