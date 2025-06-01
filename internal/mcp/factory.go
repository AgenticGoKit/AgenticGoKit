package mcp

import (
	"fmt"
	"sync"

	"github.com/kunalkushwaha/agentflow/internal/mcp/client/mark3labs"
)

// ClientFactory implements MCPClientFactory interface
type ClientFactory struct {
	creators map[string]ClientCreator
	mu       sync.RWMutex
}

// ClientCreator is a function that creates an MCP client instance
type ClientCreator func(config map[string]interface{}) (MCPClient, error)

// NewClientFactory creates a new client factory
func NewClientFactory() *ClientFactory {
	factory := &ClientFactory{
		creators: make(map[string]ClientCreator),
	}

	// Register default implementations
	factory.registerDefaultClients()

	return factory
}

// RegisterClient registers a new client implementation
func (f *ClientFactory) RegisterClient(clientType string, creator ClientCreator) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators[clientType] = creator
}

// CreateClient creates a client of the specified type
func (f *ClientFactory) CreateClient(clientType string, config map[string]interface{}) (MCPClient, error) {
	f.mu.RLock()
	creator, exists := f.creators[clientType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported MCP client type: %s", clientType)
	}

	return creator(config)
}

// SupportedClients returns a list of supported client types
func (f *ClientFactory) SupportedClients() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	clients := make([]string, 0, len(f.creators))
	for clientType := range f.creators {
		clients = append(clients, clientType)
	}
	return clients
}

// DefaultClient returns the default client type
func (f *ClientFactory) DefaultClient() string {
	return "mark3labs" // Current default implementation
}

// registerDefaultClients registers the built-in client implementations
func (f *ClientFactory) registerDefaultClients() {
	// Register mark3labs implementation
	f.RegisterClient("mark3labs", func(config map[string]interface{}) (MCPClient, error) {
		client, err := mark3labs.NewMark3LabsClient(config)
		if err != nil {
			return nil, err
		}
		// Since mark3labs returns interface{} to avoid import cycle,
		// we need to assert it implements MCPClient when the actual implementation is ready
		if mcpClient, ok := client.(MCPClient); ok {
			return mcpClient, nil
		}
		return nil, fmt.Errorf("mark3labs client does not implement MCPClient interface")
	})

	// Register custom implementation (placeholder for future)
	f.RegisterClient("custom", func(config map[string]interface{}) (MCPClient, error) {
		return NewCustomClient(config)
	})

	// Register mock implementation for testing
	f.RegisterClient("mock", func(config map[string]interface{}) (MCPClient, error) {
		return NewMockClient(config)
	})
}

// Global factory instance
var globalFactory = NewClientFactory()

// RegisterGlobalClient registers a client implementation globally
func RegisterGlobalClient(clientType string, creator ClientCreator) {
	globalFactory.RegisterClient(clientType, creator)
}

// CreateGlobalClient creates a client using the global factory
func CreateGlobalClient(clientType string, config map[string]interface{}) (MCPClient, error) {
	return globalFactory.CreateClient(clientType, config)
}

// GetSupportedClients returns supported clients from global factory
func GetSupportedClients() []string {
	return globalFactory.SupportedClients()
}

// GetDefaultClient returns the default client type
func GetDefaultClient() string {
	return globalFactory.DefaultClient()
}
