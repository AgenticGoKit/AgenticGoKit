package agentflow

import (
	"encoding/json"
	"sync"
)

// State represents the data passed between Agents in a workflow.
type State struct {
	mu       sync.RWMutex           `json:"-"` // Still exclude mutex
	data     map[string]interface{} // No JSON tags needed if using custom methods
	metadata map[string]string      // No JSON tags needed if using custom methods
}

// NewState creates an empty State.
func NewState() State {
	return State{
		data:     make(map[string]interface{}),
		metadata: make(map[string]string),
	}
}

// MarshalJSON customizes JSON marshalling for State.
func (s *State) MarshalJSON() ([]byte, error) {
	s.mu.RLock() // Lock for reading
	defer s.mu.RUnlock()

	// Use an auxiliary struct to hold the data for marshalling
	aux := struct {
		Data     map[string]interface{} `json:"data"`
		Metadata map[string]string      `json:"metadata"`
	}{
		Data:     s.data,
		Metadata: s.metadata,
	}
	return json.Marshal(aux)
}

// UnmarshalJSON customizes JSON unmarshalling for State.
func (s *State) UnmarshalJSON(data []byte) error {
	// Use an auxiliary struct to unmarshal into
	aux := struct {
		Data     map[string]interface{} `json:"data"`
		Metadata map[string]string      `json:"metadata"`
	}{} // Initialize aux

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	s.mu.Lock() // Lock for writing
	defer s.mu.Unlock()

	// Ensure internal maps are initialized before assigning
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	if s.metadata == nil {
		s.metadata = make(map[string]string)
	}

	// Assign unmarshalled data
	s.data = aux.Data
	s.metadata = aux.Metadata

	// Handle nil maps from JSON input gracefully
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	if s.metadata == nil {
		s.metadata = make(map[string]string)
	}

	return nil
}

// GetData returns a copy of the data map.
func (s *State) GetData() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Handle nil map case
	if s.data == nil {
		return make(map[string]interface{})
	}
	copiedData := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		copiedData[k] = v
	}
	return copiedData
}

// GetMetadata returns a copy of the metadata map.
func (s *State) GetMetadata() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Handle nil map case
	if s.metadata == nil {
		// Correct syntax: make(map[string]string)
		return make(map[string]string) // Line ~98
	}
	copiedMeta := make(map[string]string, len(s.metadata))
	for k, v := range s.metadata {
		copiedMeta[k] = v
	}
	return copiedMeta
}

// Get retrieves a specific value from the data map.
func (s *State) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.data == nil { // Handle nil map case
		return nil, false
	}
	val, ok := s.data[key]
	return val, ok
}

// Set stores a value in the data map. This should only be called on a cloned State.
func (s *State) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	s.data[key] = value
}

// SetMeta stores a value in the metadata map. This should only be called on a cloned State.
func (s *State) SetMeta(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.metadata == nil {
		s.metadata = make(map[string]string)
	}
	s.metadata[key] = value
}

// Clone creates a deep copy of the State, allowing safe modification.
func (s *State) Clone() State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newState := State{
		// Initialize maps in the new state
		data:     make(map[string]interface{}, len(s.data)),
		metadata: make(map[string]string, len(s.metadata)),
	}

	// Handle nil maps in the source state
	if s.data != nil {
		for k, v := range s.data {
			newState.data[k] = v // Shallow copy of value
		}
	}
	if s.metadata != nil {
		for k, v := range s.metadata {
			newState.metadata[k] = v
		}
	}
	return newState
}
