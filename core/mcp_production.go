package core

import (
	"errors"
	"time"
)

// ProductionConfig contains all production-level configuration
type ProductionConfig struct {
	// Connection pooling configuration
	ConnectionPool ConnectionPoolConfig `toml:"connection_pool"`

	// Retry policy configuration
	RetryPolicy RetryPolicyConfig `toml:"retry_policy"`

	// Load balancing configuration
	LoadBalancer LoadBalancerConfig `toml:"load_balancer"`

	// Metrics configuration
	Metrics MetricsConfig `toml:"metrics"`

	// Health check configuration
	HealthCheck HealthCheckConfig `toml:"health_check"`

	// Cache configuration
	Cache CacheConfig `toml:"cache"`
	// Circuit breaker configuration
	CircuitBreaker ProductionCircuitBreakerConfig `toml:"circuit_breaker"`
}

// ConnectionPoolConfig contains connection pooling settings
type ConnectionPoolConfig struct {
	MinConnections       int           `toml:"min_connections"`
	MaxConnections       int           `toml:"max_connections"`
	MaxIdleTime          time.Duration `toml:"max_idle_time"`
	HealthCheckInterval  time.Duration `toml:"health_check_interval"`
	HealthCheckTimeout   time.Duration `toml:"health_check_timeout"`
	ReconnectBackoff     time.Duration `toml:"reconnect_backoff"`
	MaxReconnectBackoff  time.Duration `toml:"max_reconnect_backoff"`
	MaxReconnectAttempts int           `toml:"max_reconnect_attempts"`
	ConnectionTimeout    time.Duration `toml:"connection_timeout"`
	MaxConnectionAge     time.Duration `toml:"max_connection_age"`
}

// RetryPolicyConfig contains retry policy settings
type RetryPolicyConfig struct {
	Strategy             string                     `toml:"strategy"` // exponential, linear, adaptive
	BaseDelay            time.Duration              `toml:"base_delay"`
	MaxDelay             time.Duration              `toml:"max_delay"`
	MaxAttempts          int                        `toml:"max_attempts"`
	Multiplier           float64                    `toml:"multiplier"`
	Jitter               float64                    `toml:"jitter"`
	RetryableErrors      []string                   `toml:"retryable_errors"`
	NonRetryableErrors   []string                   `toml:"non_retryable_errors"`
	ToolSpecificPolicies map[string]ToolRetryConfig `toml:"tool_specific_policies"`
}

// ToolRetryConfig contains tool-specific retry configuration
type ToolRetryConfig struct {
	Strategy    string        `toml:"strategy"`
	BaseDelay   time.Duration `toml:"base_delay"`
	MaxDelay    time.Duration `toml:"max_delay"`
	MaxAttempts int           `toml:"max_attempts"`
}

// LoadBalancerConfig contains load balancer settings
type LoadBalancerConfig struct {
	Strategy              string        `toml:"strategy"` // round_robin, least_connections, etc.
	HealthCheckInterval   time.Duration `toml:"health_check_interval"`
	HealthCheckTimeout    time.Duration `toml:"health_check_timeout"`
	UnhealthyThreshold    int           `toml:"unhealthy_threshold"`
	HealthyThreshold      int           `toml:"healthy_threshold"`
	FailoverEnabled       bool          `toml:"failover_enabled"`
	CircuitBreakerEnabled bool          `toml:"circuit_breaker_enabled"`
}

// MetricsConfig contains metrics settings
type MetricsConfig struct {
	Enabled           bool          `toml:"enabled"`
	Port              int           `toml:"port"`
	Path              string        `toml:"path"`
	UpdateInterval    time.Duration `toml:"update_interval"`
	HistogramBuckets  []float64     `toml:"histogram_buckets"`
	PrometheusEnabled bool          `toml:"prometheus_enabled"`
}

// HealthCheckConfig contains health check settings
type HealthCheckConfig struct {
	Enabled        bool          `toml:"enabled"`
	Port           int           `toml:"port"`
	Path           string        `toml:"path"`
	Interval       time.Duration `toml:"interval"`
	Timeout        time.Duration `toml:"timeout"`
	ChecksRequired int           `toml:"checks_required"`
}

// CacheConfig contains cache settings (extending existing)
type CacheConfig struct {
	// Existing cache config
	Type    string        `toml:"type"`
	TTL     time.Duration `toml:"ttl"`
	MaxSize int           `toml:"max_size"`

	// Production-specific settings
	BackgroundCleanup  bool          `toml:"background_cleanup"`
	CleanupInterval    time.Duration `toml:"cleanup_interval"`
	MemoryLimit        int64         `toml:"memory_limit"`
	CompressionEnabled bool          `toml:"compression_enabled"`
	PersistenceEnabled bool          `toml:"persistence_enabled"`
	PersistencePath    string        `toml:"persistence_path"`

	// Distributed cache settings
	Redis RedisConfig `toml:"redis"`
}

// RedisConfig contains Redis cache settings
type RedisConfig struct {
	Enabled    bool          `toml:"enabled"`
	Address    string        `toml:"address"`
	Password   string        `toml:"password"`
	Database   int           `toml:"database"`
	PoolSize   int           `toml:"pool_size"`
	Timeout    time.Duration `toml:"timeout"`
	MaxRetries int           `toml:"max_retries"`
}

// ProductionCircuitBreakerConfig contains circuit breaker settings (extending existing)
type ProductionCircuitBreakerConfig struct {
	// Existing circuit breaker config
	FailureThreshold int           `toml:"failure_threshold"`
	SuccessThreshold int           `toml:"success_threshold"`
	Timeout          time.Duration `toml:"timeout"`

	// Production-specific settings
	HalfOpenMaxCalls    int           `toml:"half_open_max_calls"`
	OpenStateTimeout    time.Duration `toml:"open_state_timeout"`
	MetricsEnabled      bool          `toml:"metrics_enabled"`
	NotificationEnabled bool          `toml:"notification_enabled"`
}

// DefaultProductionConfig returns production-ready default configuration
func DefaultProductionConfig() *ProductionConfig {
	return &ProductionConfig{
		ConnectionPool: ConnectionPoolConfig{
			MinConnections:       5,
			MaxConnections:       50,
			MaxIdleTime:          10 * time.Minute,
			HealthCheckInterval:  30 * time.Second,
			HealthCheckTimeout:   5 * time.Second,
			ReconnectBackoff:     1 * time.Second,
			MaxReconnectBackoff:  30 * time.Second,
			MaxReconnectAttempts: 5,
			ConnectionTimeout:    10 * time.Second,
			MaxConnectionAge:     1 * time.Hour,
		},
		RetryPolicy: RetryPolicyConfig{
			Strategy:    "exponential",
			BaseDelay:   1 * time.Second,
			MaxDelay:    30 * time.Second,
			MaxAttempts: 5,
			Multiplier:  2.0,
			Jitter:      0.1,
			RetryableErrors: []string{
				"network_error",
				"timeout_error",
				"throttled_error",
			},
			NonRetryableErrors: []string{
				"authentication_error",
				"authorization_error",
				"validation_error",
			},
			ToolSpecificPolicies: make(map[string]ToolRetryConfig),
		},
		LoadBalancer: LoadBalancerConfig{
			Strategy:              "round_robin",
			HealthCheckInterval:   30 * time.Second,
			HealthCheckTimeout:    5 * time.Second,
			UnhealthyThreshold:    3,
			HealthyThreshold:      2,
			FailoverEnabled:       true,
			CircuitBreakerEnabled: true,
		},
		Metrics: MetricsConfig{
			Enabled:           true,
			Port:              9090,
			Path:              "/metrics",
			UpdateInterval:    30 * time.Second,
			PrometheusEnabled: true,
			HistogramBuckets: []float64{
				0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
			},
		},
		HealthCheck: HealthCheckConfig{
			Enabled:        true,
			Port:           8080,
			Path:           "/health",
			Interval:       15 * time.Second,
			Timeout:        5 * time.Second,
			ChecksRequired: 2,
		},
		Cache: CacheConfig{
			Type:               "memory",
			TTL:                1 * time.Hour,
			MaxSize:            10000,
			BackgroundCleanup:  true,
			CleanupInterval:    5 * time.Minute,
			MemoryLimit:        1024 * 1024 * 1024, // 1GB
			CompressionEnabled: true,
			PersistenceEnabled: false,
			Redis: RedisConfig{
				Enabled:    false,
				Address:    "localhost:6379",
				Database:   0,
				PoolSize:   10,
				Timeout:    5 * time.Second,
				MaxRetries: 3,
			},
		}, CircuitBreaker: ProductionCircuitBreakerConfig{
			FailureThreshold:    5,
			SuccessThreshold:    3,
			Timeout:             60 * time.Second,
			HalfOpenMaxCalls:    3,
			OpenStateTimeout:    30 * time.Second,
			MetricsEnabled:      true,
			NotificationEnabled: true,
		},
	}
}

// DevelopmentConfig returns development-friendly configuration
func DevelopmentConfig() *ProductionConfig {
	config := DefaultProductionConfig()

	// Reduce timeouts and thresholds for faster development
	config.ConnectionPool.MinConnections = 1
	config.ConnectionPool.MaxConnections = 5
	config.ConnectionPool.HealthCheckInterval = 10 * time.Second

	config.RetryPolicy.BaseDelay = 500 * time.Millisecond
	config.RetryPolicy.MaxDelay = 5 * time.Second
	config.RetryPolicy.MaxAttempts = 3

	config.LoadBalancer.HealthCheckInterval = 10 * time.Second
	config.LoadBalancer.UnhealthyThreshold = 2

	config.Cache.MaxSize = 1000
	config.Cache.TTL = 10 * time.Minute
	config.Cache.MemoryLimit = 100 * 1024 * 1024 // 100MB

	config.CircuitBreaker.FailureThreshold = 3
	config.CircuitBreaker.Timeout = 30 * time.Second

	return config
}

// TestConfig returns configuration suitable for testing
func TestConfig() *ProductionConfig {
	config := DevelopmentConfig()

	// Further reduce for tests
	config.ConnectionPool.MinConnections = 1
	config.ConnectionPool.MaxConnections = 2
	config.ConnectionPool.HealthCheckInterval = 1 * time.Second
	config.ConnectionPool.ReconnectBackoff = 100 * time.Millisecond

	config.RetryPolicy.BaseDelay = 10 * time.Millisecond
	config.RetryPolicy.MaxDelay = 100 * time.Millisecond
	config.RetryPolicy.MaxAttempts = 2

	config.Cache.MaxSize = 100
	config.Cache.TTL = 1 * time.Minute
	config.Cache.BackgroundCleanup = false
	config.Cache.MemoryLimit = 10 * 1024 * 1024 // 10MB

	config.Metrics.Enabled = false
	config.HealthCheck.Enabled = false

	return config
}

// ValidateConfig validates the production configuration
func (c *ProductionConfig) ValidateConfig() error {
	// Validate connection pool
	if c.ConnectionPool.MinConnections > c.ConnectionPool.MaxConnections {
		return errors.New("connection pool min_connections cannot be greater than max_connections")
	}

	if c.ConnectionPool.MinConnections < 0 || c.ConnectionPool.MaxConnections < 1 {
		return errors.New("connection pool connections must be positive")
	}

	// Validate retry policy
	if c.RetryPolicy.MaxAttempts < 1 {
		return errors.New("retry policy max_attempts must be at least 1")
	}

	if c.RetryPolicy.BaseDelay <= 0 || c.RetryPolicy.MaxDelay <= 0 {
		return errors.New("retry policy delays must be positive")
	}

	if c.RetryPolicy.BaseDelay > c.RetryPolicy.MaxDelay {
		return errors.New("retry policy base_delay cannot be greater than max_delay")
	}

	// Validate cache
	if c.Cache.MaxSize < 1 {
		return errors.New("cache max_size must be at least 1")
	}

	if c.Cache.TTL <= 0 {
		return errors.New("cache TTL must be positive")
	}

	// Validate circuit breaker
	if c.CircuitBreaker.FailureThreshold < 1 || c.CircuitBreaker.SuccessThreshold < 1 {
		return errors.New("circuit breaker thresholds must be at least 1")
	}

	return nil
}
