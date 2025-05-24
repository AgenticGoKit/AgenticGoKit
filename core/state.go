// Package core provides the public State interface and related types for AgentFlow.
package core

import (
	"encoding/json"
	"sync"
)

// State represents the data and metadata passed between agents or stored in sessions.
// Implementations must be thread-safe.
type State interface {
	// Get retrieves a value from the data map.
	Get(key string) (any, bool)
	// Set adds or updates a value in the data map.
	Set(key string, value any)
	// GetMeta retrieves a value from the metadata map.
	GetMeta(key string) (string, bool)
	// SetMeta adds or updates a value in the metadata map.
	SetMeta(key string, value string)
	// Keys returns a slice of all keys present in the data map.
	Keys() []string
	// MetaKeys returns a slice of all keys present in the metadata map.
	MetaKeys() []string
	// Clone creates a deep copy of the state.
	Clone() State
	// Merge copies data and metadata from another state into this one.
	// Existing keys in the destination state will be overwritten by keys from the source state.
	Merge(source State)
}

// SimpleState is a basic thread-safe implementation of State using maps.
type SimpleState struct {
	mu   sync.RWMutex      `json:"-"`
	data map[string]any    `json:"data"`
	meta map[string]string `json:"meta"`
}

// Compile-time check to ensure SimpleState implements State
var _ State = (*SimpleState)(nil)

// NewState creates an empty SimpleState.
func NewState() *SimpleState {
	return &SimpleState{
		data: make(map[string]any),
		meta: make(map[string]string),
	}
}

// stateJSON is a helper struct for JSON marshaling/unmarshaling SimpleState.
type stateJSON struct {
	Data map[string]any    `json:"data"`
	Meta map[string]string `json:"meta"`
}

// MarshalJSON implements the json.Marshaler interface for SimpleState.
func (s *SimpleState) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	temp := stateJSON{
		Data: make(map[string]any, len(s.data)),
		Meta: make(map[string]string, len(s.meta)),
	}
	for k, v := range s.data {
		temp.Data[k] = v
	}
	for k, v := range s.meta {
		temp.Meta[k] = v
	}

	return json.Marshal(temp)
}

// UnmarshalJSON implements the json.Unmarshaler interface for SimpleState.
func (s *SimpleState) UnmarshalJSON(b []byte) error {
	var temp stateJSON
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		s.data = make(map[string]any)
	}
	if s.meta == nil {
		s.meta = make(map[string]string)
	}

	for k, v := range temp.Data {
		s.data[k] = v
	}
	for k, v := range temp.Meta {
		s.meta[k] = v
	}

	return nil
}

// Get retrieves a value from the data map.
func (s *SimpleState) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// Set adds or updates a value in the data map.
func (s *SimpleState) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		s.data = make(map[string]any)
	}
	s.data[key] = value
}

// GetMeta retrieves a value from the metadata map.
func (s *SimpleState) GetMeta(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.meta[key]
	return val, ok
}

// SetMeta adds or updates a value in the metadata map.
func (s *SimpleState) SetMeta(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.meta == nil {
		s.meta = make(map[string]string)
	}
	s.meta[key] = value
}

// Keys returns a slice of all keys present in the data map.
func (s *SimpleState) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// MetaKeys returns a slice of all keys present in the metadata map.
func (s *SimpleState) MetaKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.meta))
	for k := range s.meta {
		keys = append(keys, k)
	}
	return keys
}

// Clone creates a deep copy of the SimpleState.
func (s *SimpleState) Clone() State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newState := NewState()
	if s.data != nil {
		for k, v := range s.data {
			newState.data[k] = v
		}
	}
	if s.meta != nil {
		for k, v := range s.meta {
			newState.meta[k] = v
		}
	}
	return newState
}

// Merge copies data and metadata from the source state into this state.
func (s *SimpleState) Merge(source State) {
	if source == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range source.Keys() {
		if value, ok := source.Get(key); ok {
			s.data[key] = value
		}
	}

	for _, key := range source.MetaKeys() {
		if value, ok := source.GetMeta(key); ok {
			s.meta[key] = value
		}
	}
}

// NewStateWithData creates a new SimpleState initialized with the provided data map.
func NewStateWithData(data map[string]any) State {
	s := NewState()
	if data != nil {
		for k, v := range data {
			s.data[k] = v
		}
	}
	return s
}

// NewSimpleState creates a new state instance with optional initial data.
func NewSimpleState(initialData map[string]any) *SimpleState {
	s := NewState()
	if initialData != nil {
		for k, v := range initialData {
			s.data[k] = v
		}
	}
	return s
}
