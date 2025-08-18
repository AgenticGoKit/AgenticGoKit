package mcp

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// MetricsConfig contains configuration for metrics collection
type MetricsConfig struct {
	Enabled          bool          `toml:"enabled"`
	Port             int           `toml:"port"`
	Path             string        `toml:"path"`
	UpdateInterval   time.Duration `toml:"update_interval"`
	HistogramBuckets []float64     `toml:"histogram_buckets"`
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:        true,
		Port:           9090,
		Path:           "/metrics",
		UpdateInterval: 30 * time.Second,
		HistogramBuckets: []float64{
			0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
		},
	}
}

// MCPMetrics collects and exposes MCP-related metrics
type MCPMetrics struct {
	config   *MetricsConfig
	registry *prometheus.Registry
	logger   *zerolog.Logger

	// Connection metrics
	connectionTotal      *prometheus.CounterVec
	connectionDuration   *prometheus.HistogramVec
	connectionPoolSize   *prometheus.GaugeVec
	connectionPoolActive *prometheus.GaugeVec
	connectionErrors     *prometheus.CounterVec

	// Tool execution metrics
	toolExecutionTotal    *prometheus.CounterVec
	toolExecutionDuration *prometheus.HistogramVec
	toolExecutionErrors   *prometheus.CounterVec
	toolExecutionRetries  *prometheus.CounterVec

	// Cache metrics
	cacheHits        *prometheus.CounterVec
	cacheMisses      *prometheus.CounterVec
	cacheEvictions   *prometheus.CounterVec
	cacheSize        *prometheus.GaugeVec
	cacheMemoryUsage *prometheus.GaugeVec

	// Circuit breaker metrics
	circuitBreakerState       *prometheus.GaugeVec
	circuitBreakerTransitions *prometheus.CounterVec

	// Performance metrics
	responseSize          *prometheus.HistogramVec
	concurrentConnections *prometheus.GaugeVec
	queueSize             *prometheus.GaugeVec

	// HTTP server for metrics
	httpServer *http.Server
	serverMux  sync.Mutex
}

// NewMCPMetrics creates a new metrics collector
func NewMCPMetrics(config *MetricsConfig, logger *zerolog.Logger) *MCPMetrics {
	if config == nil {
		config = DefaultMetricsConfig()
	}

	registry := prometheus.NewRegistry()

	m := &MCPMetrics{
		config:   config,
		registry: registry,
		logger:   logger,
	}

	m.initializeMetrics()

	if config.Enabled {
		m.startMetricsServer()
	}

	return m
}

// initializeMetrics initializes all Prometheus metrics
func (m *MCPMetrics) initializeMetrics() {
	// Connection metrics
	m.connectionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_connections_total",
			Help: "Total number of MCP connections",
		},
		[]string{"server_id", "status"},
	)

	m.connectionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_connection_duration_seconds",
			Help:    "Duration of MCP connections",
			Buckets: m.config.HistogramBuckets,
		},
		[]string{"server_id"},
	)

	m.connectionPoolSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_connection_pool_size",
			Help: "Current size of connection pools",
		},
		[]string{"server_id"},
	)

	m.connectionPoolActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_connection_pool_active",
			Help: "Number of active connections in pool",
		},
		[]string{"server_id"},
	)

	m.connectionErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_connection_errors_total",
			Help: "Total number of connection errors",
		},
		[]string{"server_id", "error_type"},
	)

	// Tool execution metrics
	m.toolExecutionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_tool_executions_total",
			Help: "Total number of tool executions",
		},
		[]string{"server_id", "tool_name", "status"},
	)

	m.toolExecutionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_tool_execution_duration_seconds",
			Help:    "Duration of tool executions",
			Buckets: m.config.HistogramBuckets,
		},
		[]string{"server_id", "tool_name"},
	)

	m.toolExecutionErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_tool_execution_errors_total",
			Help: "Total number of tool execution errors",
		},
		[]string{"server_id", "tool_name", "error_type"},
	)

	m.toolExecutionRetries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_tool_execution_retries_total",
			Help: "Total number of tool execution retries",
		},
		[]string{"server_id", "tool_name"},
	)

	// Cache metrics
	m.cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type", "tool_name"},
	)

	m.cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type", "tool_name"},
	)

	m.cacheEvictions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_cache_evictions_total",
			Help: "Total number of cache evictions",
		},
		[]string{"cache_type", "eviction_reason"},
	)

	m.cacheSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_cache_size_items",
			Help: "Current number of items in cache",
		},
		[]string{"cache_type"},
	)

	m.cacheMemoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_cache_memory_usage_bytes",
			Help: "Current memory usage of cache",
		},
		[]string{"cache_type"},
	)

	// Circuit breaker metrics
	m.circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_circuit_breaker_state",
			Help: "Current state of circuit breaker (0=closed, 1=open, 2=half-open)",
		},
		[]string{"server_id", "tool_name"},
	)

	m.circuitBreakerTransitions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_circuit_breaker_transitions_total",
			Help: "Total number of circuit breaker state transitions",
		},
		[]string{"server_id", "tool_name", "from_state", "to_state"},
	)

	// Performance metrics
	m.responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_response_size_bytes",
			Help:    "Size of tool execution responses",
			Buckets: prometheus.ExponentialBuckets(256, 2, 10), // 256B to 128KB
		},
		[]string{"server_id", "tool_name"},
	)

	m.concurrentConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_concurrent_connections",
			Help: "Number of concurrent connections",
		},
		[]string{"server_id"},
	)

	m.queueSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_queue_size",
			Help: "Current size of execution queue",
		},
		[]string{"server_id"},
	)

	// Register all metrics
	m.registry.MustRegister(
		m.connectionTotal,
		m.connectionDuration,
		m.connectionPoolSize,
		m.connectionPoolActive,
		m.connectionErrors,
		m.toolExecutionTotal,
		m.toolExecutionDuration,
		m.toolExecutionErrors,
		m.toolExecutionRetries,
		m.cacheHits,
		m.cacheMisses,
		m.cacheEvictions,
		m.cacheSize,
		m.cacheMemoryUsage,
		m.circuitBreakerState,
		m.circuitBreakerTransitions,
		m.responseSize,
		m.concurrentConnections,
		m.queueSize,
	)
}

// startMetricsServer starts the HTTP server for metrics
func (m *MCPMetrics) startMetricsServer() {
	m.serverMux.Lock()
	defer m.serverMux.Unlock()

	if m.httpServer != nil {
		return
	}

	mux := http.NewServeMux()
	mux.Handle(m.config.Path, promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))

	m.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.config.Port),
		Handler: mux,
	}
	go func() {
		m.logger.Info().
			Int("port", m.config.Port).
			Str("path", m.config.Path).
			Msg("Starting metrics server")
		if err := m.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.Error().Err(err).Msg("Metrics server error")
		}
	}()
}

// RecordConnection records connection metrics
func (m *MCPMetrics) RecordConnection(serverID string, status string, duration time.Duration) {
	m.connectionTotal.WithLabelValues(serverID, status).Inc()
	if duration > 0 {
		m.connectionDuration.WithLabelValues(serverID).Observe(duration.Seconds())
	}
}

// UpdateConnectionPool updates connection pool metrics
func (m *MCPMetrics) UpdateConnectionPool(serverID string, size, active int) {
	m.connectionPoolSize.WithLabelValues(serverID).Set(float64(size))
	m.connectionPoolActive.WithLabelValues(serverID).Set(float64(active))
}

// RecordConnectionError records connection error metrics
func (m *MCPMetrics) RecordConnectionError(serverID, errorType string) {
	m.connectionErrors.WithLabelValues(serverID, errorType).Inc()
}

// RecordToolExecution records tool execution metrics
func (m *MCPMetrics) RecordToolExecution(serverID, toolName, status string, duration time.Duration, responseSize int) {
	m.toolExecutionTotal.WithLabelValues(serverID, toolName, status).Inc()
	m.toolExecutionDuration.WithLabelValues(serverID, toolName).Observe(duration.Seconds())

	if responseSize > 0 {
		m.responseSize.WithLabelValues(serverID, toolName).Observe(float64(responseSize))
	}
}

// RecordToolExecutionError records tool execution error metrics
func (m *MCPMetrics) RecordToolExecutionError(serverID, toolName, errorType string) {
	m.toolExecutionErrors.WithLabelValues(serverID, toolName, errorType).Inc()
}

// RecordToolExecutionRetry records tool execution retry metrics
func (m *MCPMetrics) RecordToolExecutionRetry(serverID, toolName string) {
	m.toolExecutionRetries.WithLabelValues(serverID, toolName).Inc()
}

// RecordCacheHit records cache hit metrics
func (m *MCPMetrics) RecordCacheHit(cacheType, toolName string) {
	m.cacheHits.WithLabelValues(cacheType, toolName).Inc()
}

// RecordCacheMiss records cache miss metrics
func (m *MCPMetrics) RecordCacheMiss(cacheType, toolName string) {
	m.cacheMisses.WithLabelValues(cacheType, toolName).Inc()
}

// RecordCacheEviction records cache eviction metrics
func (m *MCPMetrics) RecordCacheEviction(cacheType, reason string) {
	m.cacheEvictions.WithLabelValues(cacheType, reason).Inc()
}

// UpdateCacheSize updates cache size metrics
func (m *MCPMetrics) UpdateCacheSize(cacheType string, size int) {
	m.cacheSize.WithLabelValues(cacheType).Set(float64(size))
}

// UpdateCacheMemoryUsage updates cache memory usage metrics
func (m *MCPMetrics) UpdateCacheMemoryUsage(cacheType string, bytes int64) {
	m.cacheMemoryUsage.WithLabelValues(cacheType).Set(float64(bytes))
}

// RecordCircuitBreakerState records circuit breaker state
func (m *MCPMetrics) RecordCircuitBreakerState(serverID, toolName string, state int) {
	m.circuitBreakerState.WithLabelValues(serverID, toolName).Set(float64(state))
}

// RecordCircuitBreakerTransition records circuit breaker state transition
func (m *MCPMetrics) RecordCircuitBreakerTransition(serverID, toolName, fromState, toState string) {
	m.circuitBreakerTransitions.WithLabelValues(serverID, toolName, fromState, toState).Inc()
}

// UpdateConcurrentConnections updates concurrent connections metric
func (m *MCPMetrics) UpdateConcurrentConnections(serverID string, count int) {
	m.concurrentConnections.WithLabelValues(serverID).Set(float64(count))
}

// UpdateQueueSize updates queue size metric
func (m *MCPMetrics) UpdateQueueSize(serverID string, size int) {
	m.queueSize.WithLabelValues(serverID).Set(float64(size))
}

// GetMetricsSnapshot returns a snapshot of current metrics
func (m *MCPMetrics) GetMetricsSnapshot() map[string]interface{} {
	metricFamilies, err := m.registry.Gather()
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to gather metrics")
		return nil
	}

	snapshot := make(map[string]interface{})

	for _, mf := range metricFamilies {
		metricName := mf.GetName()
		metrics := mf.GetMetric()

		metricData := make([]map[string]interface{}, len(metrics))

		for i, metric := range metrics {
			data := map[string]interface{}{
				"labels": make(map[string]string),
			}

			// Extract labels
			for _, label := range metric.GetLabel() {
				data["labels"].(map[string]string)[label.GetName()] = label.GetValue()
			}

			// Extract value based on metric type
			switch mf.GetType() {
			case 0: // COUNTER
				data["value"] = metric.GetCounter().GetValue()
			case 1: // GAUGE
				data["value"] = metric.GetGauge().GetValue()
			case 4: // HISTOGRAM
				hist := metric.GetHistogram()
				data["sample_count"] = hist.GetSampleCount()
				data["sample_sum"] = hist.GetSampleSum()
				buckets := make([]map[string]interface{}, len(hist.GetBucket()))
				for j, bucket := range hist.GetBucket() {
					buckets[j] = map[string]interface{}{
						"upper_bound":      bucket.GetUpperBound(),
						"cumulative_count": bucket.GetCumulativeCount(),
					}
				}
				data["buckets"] = buckets
			}

			metricData[i] = data
		}

		snapshot[metricName] = metricData
	}

	return snapshot
}

// Close stops the metrics server
func (m *MCPMetrics) Close() error {
	m.serverMux.Lock()
	defer m.serverMux.Unlock()

	if m.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.httpServer.Shutdown(ctx)
	}

	return nil
}

// HealthChecker provides health check endpoints for MCP services
type HealthChecker struct {
	connectionPool *ConnectionPool
	cacheManager   core.MCPCacheManager
	metrics        *MCPMetrics
	logger         *zerolog.Logger
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(connectionPool *ConnectionPool, cacheManager core.MCPCacheManager, metrics *MCPMetrics, logger *zerolog.Logger) *HealthChecker {
	return &HealthChecker{
		connectionPool: connectionPool,
		cacheManager:   cacheManager,
		metrics:        metrics,
		logger:         logger,
	}
}

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// CheckHealth performs a comprehensive health check
func (h *HealthChecker) CheckHealth(ctx context.Context) map[string]*HealthStatus {
	results := make(map[string]*HealthStatus)

	// Check connection pool health
	results["connection_pool"] = h.checkConnectionPoolHealth(ctx)

	// Check cache health
	results["cache"] = h.checkCacheHealth(ctx)

	// Check metrics health
	results["metrics"] = h.checkMetricsHealth(ctx)

	// Overall status
	overallStatus := "healthy"
	for _, status := range results {
		if status.Status != "healthy" {
			overallStatus = "unhealthy"
			break
		}
	}

	results["overall"] = &HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"components_checked": len(results) - 1,
		},
	}

	return results
}

// checkConnectionPoolHealth checks connection pool health
func (h *HealthChecker) checkConnectionPoolHealth(ctx context.Context) *HealthStatus {
	start := time.Now()

	if h.connectionPool == nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Error:     "connection pool not initialized",
		}
	}

	stats := h.connectionPool.GetStats()

	status := "healthy"
	details := map[string]interface{}{
		"total_servers":      stats["total_servers"],
		"total_connections":  stats["total_connections"],
		"active_connections": stats["active_connections"],
	}

	// Check if we have any connections
	if stats["total_connections"].(int) == 0 {
		status = "warning"
		details["warning"] = "no connections available"
	}

	return &HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Details:   details,
	}
}

// checkCacheHealth checks cache health
func (h *HealthChecker) checkCacheHealth(ctx context.Context) *HealthStatus {
	start := time.Now()

	if h.cacheManager == nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Error:     "cache manager not initialized",
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats, err := h.cacheManager.GetGlobalStats(ctx)
	if err != nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Error:     fmt.Sprintf("failed to get cache stats: %v", err),
		}
	}
	// Convert stats to map for Details field
	statsMap := map[string]interface{}{
		"hit_rate":        stats.HitRate,
		"hit_count":       stats.HitCount,
		"miss_count":      stats.MissCount,
		"total_size":      stats.TotalSize,
		"total_keys":      stats.TotalKeys,
		"eviction_count":  stats.EvictionCount,
		"average_latency": stats.AverageLatency,
		"last_cleanup":    stats.LastCleanup,
	}

	return &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Details:   statsMap,
	}
}

// checkMetricsHealth checks metrics system health
func (h *HealthChecker) checkMetricsHealth(ctx context.Context) *HealthStatus {
	start := time.Now()

	if h.metrics == nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Error:     "metrics not initialized",
		}
	}

	snapshot := h.metrics.GetMetricsSnapshot()

	return &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Details: map[string]interface{}{
			"metrics_count": len(snapshot),
		},
	}
}
