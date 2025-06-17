package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

// Connection represents a connection to an MCP server
type Connection struct {
	ID       string
	ServerID string
	Address  string
}

// Ping tests if the connection is alive
func (c *Connection) Ping(ctx context.Context) error {
	// In a real implementation, this would ping the actual server
	return nil
}

// Close closes the connection
func (c *Connection) Close() error {
	// In a real implementation, this would close the actual connection
	return nil
}

// ConnectionState represents the state of an MCP server connection
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateError
	StateClosed
)

// String returns the string representation of ConnectionState
func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateReconnecting:
		return "reconnecting"
	case StateError:
		return "error"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// PooledConnection represents a pooled connection to an MCP server
type PooledConnection struct {
	ID           string
	ServerID     string
	Connection   *Connection
	State        atomic.Value // ConnectionState
	LastUsed     atomic.Value // time.Time
	LastError    atomic.Value // error
	UseCount     atomic.Int64
	CreatedAt    time.Time
	pool         *ConnectionPool
	healthTicker *time.Ticker
	cancelFunc   context.CancelFunc
	mutex        sync.RWMutex
}

// NewPooledConnection creates a new pooled connection
func NewPooledConnection(serverID string, pool *ConnectionPool) *PooledConnection {
	pc := &PooledConnection{
		ID:        generateConnectionID(),
		ServerID:  serverID,
		CreatedAt: time.Now(),
		pool:      pool,
	}

	pc.State.Store(StateDisconnected)
	pc.LastUsed.Store(time.Now())

	return pc
}

// GetState returns the current connection state
func (pc *PooledConnection) GetState() ConnectionState {
	return pc.State.Load().(ConnectionState)
}

// setState updates the connection state
func (pc *PooledConnection) setState(state ConnectionState) {
	pc.State.Store(state)
}

// GetLastUsed returns the last used time
func (pc *PooledConnection) GetLastUsed() time.Time {
	return pc.LastUsed.Load().(time.Time)
}

// updateLastUsed updates the last used timestamp
func (pc *PooledConnection) updateLastUsed() {
	pc.LastUsed.Store(time.Now())
}

// GetLastError returns the last error
func (pc *PooledConnection) GetLastError() error {
	if err := pc.LastError.Load(); err != nil {
		return err.(error)
	}
	return nil
}

// setLastError updates the last error
func (pc *PooledConnection) setLastError(err error) {
	pc.LastError.Store(err)
}

// Connect establishes connection to the MCP server
func (pc *PooledConnection) Connect(ctx context.Context) error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	if pc.GetState() == StateConnected {
		return nil
	}

	pc.setState(StateConnecting)

	// Create connection context with cancellation
	connCtx, cancel := context.WithCancel(ctx)
	pc.cancelFunc = cancel

	// Attempt to connect
	conn, err := pc.pool.factory.CreateConnection(connCtx, pc.ServerID)
	if err != nil {
		pc.setState(StateError)
		pc.setLastError(err)
		return fmt.Errorf("failed to connect to server %s: %w", pc.ServerID, err)
	}

	pc.Connection = conn
	pc.setState(StateConnected)
	pc.setLastError(nil)
	pc.updateLastUsed()

	// Start health monitoring
	pc.startHealthMonitoring(connCtx)

	return nil
}

// Disconnect closes the connection
func (pc *PooledConnection) Disconnect() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	pc.setState(StateClosed)

	// Cancel context
	if pc.cancelFunc != nil {
		pc.cancelFunc()
	}

	// Stop health monitoring
	if pc.healthTicker != nil {
		pc.healthTicker.Stop()
	}

	// Close connection
	if pc.Connection != nil {
		err := pc.Connection.Close()
		pc.Connection = nil
		return err
	}

	return nil
}

// IsHealthy checks if the connection is healthy
func (pc *PooledConnection) IsHealthy(ctx context.Context) bool {
	if pc.GetState() != StateConnected || pc.Connection == nil {
		return false
	}

	// Perform health check
	err := pc.Connection.Ping(ctx)
	if err != nil {
		pc.setLastError(err)
		return false
	}

	return true
}

// startHealthMonitoring starts periodic health checks
func (pc *PooledConnection) startHealthMonitoring(ctx context.Context) {
	pc.healthTicker = time.NewTicker(pc.pool.config.HealthCheckInterval)

	go func() {
		defer pc.healthTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-pc.healthTicker.C:
				if !pc.IsHealthy(ctx) && pc.GetState() == StateConnected {
					pc.setState(StateError)
					// Trigger reconnection attempt
					go pc.attemptReconnection(ctx)
				}
			}
		}
	}()
}

// attemptReconnection attempts to reconnect to the server
func (pc *PooledConnection) attemptReconnection(ctx context.Context) {
	pc.setState(StateReconnecting)

	backoff := pc.pool.config.ReconnectBackoff
	maxRetries := pc.pool.config.MaxReconnectAttempts

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
			if err := pc.Connect(ctx); err == nil {
				pc.pool.logger.Info().
					Str("server_id", pc.ServerID).
					Str("connection_id", pc.ID).
					Int("attempt", attempt+1).
					Msg("Successfully reconnected to server")
				return
			}

			// Exponential backoff with jitter
			backoff = time.Duration(float64(backoff) * 1.5)
			if backoff > pc.pool.config.MaxReconnectBackoff {
				backoff = pc.pool.config.MaxReconnectBackoff
			}
		}
	}
	pc.setState(StateError)
	pc.pool.logger.Error().
		Str("server_id", pc.ServerID).
		Str("connection_id", pc.ID).
		Int("max_attempts", maxRetries).
		Msg("Failed to reconnect after max attempts")
}

// Use marks the connection as used and returns it for execution
func (pc *PooledConnection) Use() *Connection {
	pc.updateLastUsed()
	pc.UseCount.Add(1)
	return pc.Connection
}

// ConnectionPoolConfig contains configuration for the connection pool
type ConnectionPoolConfig struct {
	// Pool sizing
	MinConnections int           `toml:"min_connections"`
	MaxConnections int           `toml:"max_connections"`
	MaxIdleTime    time.Duration `toml:"max_idle_time"`

	// Health checking
	HealthCheckInterval time.Duration `toml:"health_check_interval"`
	HealthCheckTimeout  time.Duration `toml:"health_check_timeout"`

	// Reconnection
	ReconnectBackoff     time.Duration `toml:"reconnect_backoff"`
	MaxReconnectBackoff  time.Duration `toml:"max_reconnect_backoff"`
	MaxReconnectAttempts int           `toml:"max_reconnect_attempts"`

	// Connection lifecycle
	ConnectionTimeout time.Duration `toml:"connection_timeout"`
	MaxConnectionAge  time.Duration `toml:"max_connection_age"`
}

// DefaultConnectionPoolConfig returns default configuration
func DefaultConnectionPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MinConnections:       2,
		MaxConnections:       10,
		MaxIdleTime:          30 * time.Minute,
		HealthCheckInterval:  30 * time.Second,
		HealthCheckTimeout:   5 * time.Second,
		ReconnectBackoff:     1 * time.Second,
		MaxReconnectBackoff:  30 * time.Second,
		MaxReconnectAttempts: 5,
		ConnectionTimeout:    10 * time.Second,
		MaxConnectionAge:     1 * time.Hour,
	}
}

// ConnectionFactory creates connections to MCP servers
type ConnectionFactory interface {
	CreateConnection(ctx context.Context, serverID string) (*Connection, error)
}

// ConnectionPool manages a pool of connections to MCP servers
type ConnectionPool struct {
	config  *ConnectionPoolConfig
	factory ConnectionFactory
	logger  *zerolog.Logger

	// Server pools - map of serverID to connection pool
	serverPools map[string]*serverPool
	poolMutex   sync.RWMutex

	// Lifecycle
	ctx        context.Context
	cancelFunc context.CancelFunc
	cleanupWG  sync.WaitGroup
}

// serverPool manages connections for a specific server
type serverPool struct {
	serverID    string
	connections []*PooledConnection
	available   chan *PooledConnection
	mutex       sync.RWMutex
	pool        *ConnectionPool
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config *ConnectionPoolConfig, factory ConnectionFactory, logger *zerolog.Logger) *ConnectionPool {
	if config == nil {
		config = DefaultConnectionPoolConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnectionPool{
		config:      config,
		factory:     factory,
		logger:      logger,
		serverPools: make(map[string]*serverPool),
		ctx:         ctx,
		cancelFunc:  cancel,
	}

	// Start cleanup routine
	pool.startCleanupRoutine()

	return pool
}

// GetConnection gets a connection from the pool for the specified server
func (cp *ConnectionPool) GetConnection(ctx context.Context, serverID string) (*PooledConnection, error) {
	sp := cp.getOrCreateServerPool(serverID)

	// Try to get an available connection
	select {
	case conn := <-sp.available:
		if conn.IsHealthy(ctx) {
			return conn, nil
		}
		// Connection is unhealthy, try to get another one
		return cp.createNewConnection(ctx, serverID, sp)
	case <-time.After(cp.config.ConnectionTimeout):
		return nil, errors.New("timeout waiting for available connection")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ReturnConnection returns a connection to the pool
func (cp *ConnectionPool) ReturnConnection(conn *PooledConnection) {
	if conn == nil || conn.GetState() != StateConnected {
		return
	}

	sp := cp.getServerPool(conn.ServerID)
	if sp == nil {
		return
	}

	select {
	case sp.available <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close the connection
		conn.Disconnect()
	}
}

// getOrCreateServerPool gets or creates a server pool
func (cp *ConnectionPool) getOrCreateServerPool(serverID string) *serverPool {
	cp.poolMutex.Lock()
	defer cp.poolMutex.Unlock()

	if sp, exists := cp.serverPools[serverID]; exists {
		return sp
	}

	sp := &serverPool{
		serverID:    serverID,
		connections: make([]*PooledConnection, 0, cp.config.MaxConnections),
		available:   make(chan *PooledConnection, cp.config.MaxConnections),
		pool:        cp,
	}

	cp.serverPools[serverID] = sp

	// Create minimum connections
	go cp.ensureMinConnections(serverID, sp)

	return sp
}

// getServerPool gets an existing server pool
func (cp *ConnectionPool) getServerPool(serverID string) *serverPool {
	cp.poolMutex.RLock()
	defer cp.poolMutex.RUnlock()
	return cp.serverPools[serverID]
}

// createNewConnection creates a new connection for the server
func (cp *ConnectionPool) createNewConnection(ctx context.Context, serverID string, sp *serverPool) (*PooledConnection, error) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	// Check if we've reached max connections
	if len(sp.connections) >= cp.config.MaxConnections {
		return nil, errors.New("maximum connections reached for server")
	}

	// Create new connection
	conn := NewPooledConnection(serverID, cp)
	if err := conn.Connect(ctx); err != nil {
		return nil, err
	}

	sp.connections = append(sp.connections, conn)
	cp.logger.Debug().
		Str("server_id", serverID).
		Str("connection_id", conn.ID).
		Int("pool_size", len(sp.connections)).
		Msg("Created new pooled connection")

	return conn, nil
}

// ensureMinConnections ensures minimum connections are maintained
func (cp *ConnectionPool) ensureMinConnections(serverID string, sp *serverPool) {
	for i := 0; i < cp.config.MinConnections; i++ {
		conn, err := cp.createNewConnection(cp.ctx, serverID, sp)
		if err != nil {
			cp.logger.Error().
				Str("server_id", serverID).
				Err(err).
				Msg("Failed to create minimum connection")
			continue
		}

		// Return to available pool
		select {
		case sp.available <- conn:
		default:
			// Pool is full, which shouldn't happen with min connections
			conn.Disconnect()
		}
	}
}

// startCleanupRoutine starts the connection cleanup routine
func (cp *ConnectionPool) startCleanupRoutine() {
	cp.cleanupWG.Add(1)

	go func() {
		defer cp.cleanupWG.Done()

		ticker := time.NewTicker(1 * time.Minute) // Cleanup every minute
		defer ticker.Stop()

		for {
			select {
			case <-cp.ctx.Done():
				return
			case <-ticker.C:
				cp.cleanup()
			}
		}
	}()
}

// cleanup removes idle and expired connections
func (cp *ConnectionPool) cleanup() {
	cp.poolMutex.RLock()
	serverPools := make([]*serverPool, 0, len(cp.serverPools))
	for _, sp := range cp.serverPools {
		serverPools = append(serverPools, sp)
	}
	cp.poolMutex.RUnlock()

	for _, sp := range serverPools {
		cp.cleanupServerPool(sp)
	}
}

// cleanupServerPool cleans up connections in a server pool
func (cp *ConnectionPool) cleanupServerPool(sp *serverPool) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	now := time.Now()
	activeConnections := make([]*PooledConnection, 0, len(sp.connections))

	for _, conn := range sp.connections {
		shouldRemove := false

		// Check if connection is too old
		if now.Sub(conn.CreatedAt) > cp.config.MaxConnectionAge {
			shouldRemove = true
		}

		// Check if connection has been idle too long
		if now.Sub(conn.GetLastUsed()) > cp.config.MaxIdleTime {
			shouldRemove = true
		}

		// Check if connection is in error state
		if conn.GetState() == StateError || conn.GetState() == StateClosed {
			shouldRemove = true
		}
		if shouldRemove && len(activeConnections) >= cp.config.MinConnections {
			conn.Disconnect()
			cp.logger.Debug().
				Str("server_id", sp.serverID).
				Str("connection_id", conn.ID).
				Str("reason", "cleanup").
				Msg("Cleaned up connection")
		} else {
			activeConnections = append(activeConnections, conn)
		}
	}

	sp.connections = activeConnections
}

// GetStats returns connection pool statistics
func (cp *ConnectionPool) GetStats() map[string]interface{} {
	cp.poolMutex.RLock()
	defer cp.poolMutex.RUnlock()

	stats := map[string]interface{}{
		"total_servers": len(cp.serverPools),
		"servers":       make(map[string]interface{}),
	}

	totalConnections := 0
	totalAvailable := 0

	for serverID, sp := range cp.serverPools {
		sp.mutex.RLock()
		connCount := len(sp.connections)
		availableCount := len(sp.available)
		sp.mutex.RUnlock()

		totalConnections += connCount
		totalAvailable += availableCount

		serverStats := map[string]interface{}{
			"total_connections":     connCount,
			"available_connections": availableCount,
			"active_connections":    connCount - availableCount,
		}

		stats["servers"].(map[string]interface{})[serverID] = serverStats
	}

	stats["total_connections"] = totalConnections
	stats["available_connections"] = totalAvailable
	stats["active_connections"] = totalConnections - totalAvailable

	return stats
}

// Close closes the connection pool and all connections
func (cp *ConnectionPool) Close() error {
	cp.cancelFunc()
	cp.cleanupWG.Wait()

	cp.poolMutex.Lock()
	defer cp.poolMutex.Unlock()

	for _, sp := range cp.serverPools {
		sp.mutex.Lock()
		for _, conn := range sp.connections {
			conn.Disconnect()
		}
		sp.mutex.Unlock()
	}

	cp.serverPools = make(map[string]*serverPool)
	return nil
}

// generateConnectionID generates a unique connection ID
func generateConnectionID() string {
	return fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(),
		atomic.AddInt64(&connectionIDCounter, 1))
}

var connectionIDCounter int64
