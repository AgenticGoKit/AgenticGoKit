// Package core provides the public Event interface and related types for AgentFlow.
package core

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const SessionIDKey = "session_id"

// EventData holds the payload of an event.
type EventData map[string]any

// Event defines the interface for messages passed through the system.
type Event interface {
	GetID() string
	GetTimestamp() time.Time
	GetTargetAgentID() string
	GetSourceAgentID() string
	GetData() EventData
	GetMetadata() map[string]string
	GetSessionID() string
	GetMetadataValue(key string) (string, bool)
	SetID(id string)
	SetTargetAgentID(id string)
	SetSourceAgentID(id string)
	SetData(key string, value any)
	SetMetadata(key string, value string)
}

// SimpleEvent is a basic implementation of the Event interface.
type SimpleEvent struct {
	ID            string            `json:"id"`
	Timestamp     time.Time         `json:"timestamp"`
	TargetAgentID string            `json:"target_agent_id,omitempty"`
	SourceAgentID string            `json:"source_agent_id,omitempty"`
	Data          EventData         `json:"data"`
	Metadata      map[string]string `json:"metadata"`
	mu            sync.RWMutex      `json:"-"`
}

// NewEvent creates a new SimpleEvent instance.
func NewEvent(targetAgentID string, data EventData, metadata map[string]string) *SimpleEvent {
	if data == nil {
		data = make(EventData)
	}
	if metadata == nil {
		metadata = make(map[string]string)
	} // Generate unique ID for the event
	id := uuid.NewString()

	return &SimpleEvent{
		ID:            id,
		Timestamp:     time.Now(),
		TargetAgentID: targetAgentID,
		Data:          data,
		Metadata:      metadata,
	}
}

func (e *SimpleEvent) GetID() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.ID
}
func (e *SimpleEvent) GetTimestamp() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Timestamp
}
func (e *SimpleEvent) GetTargetAgentID() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.TargetAgentID
}
func (e *SimpleEvent) GetSourceAgentID() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.SourceAgentID
}
func (e *SimpleEvent) GetData() EventData {
	e.mu.RLock()
	defer e.mu.RUnlock()
	dataCopy := make(EventData, len(e.Data))
	for k, v := range e.Data {
		dataCopy[k] = v
	}
	return dataCopy
}
func (e *SimpleEvent) GetMetadata() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	metadataCopy := make(map[string]string, len(e.Metadata))
	for k, v := range e.Metadata {
		metadataCopy[k] = v
	}
	return metadataCopy
}
func (e *SimpleEvent) GetSessionID() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.Metadata == nil {
		return ""
	}
	return e.Metadata[SessionIDKey]
}
func (e *SimpleEvent) GetMetadataValue(key string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.Metadata == nil {
		return "", false
	}
	val, ok := e.Metadata[key]
	return val, ok
}
func (e *SimpleEvent) SetID(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ID = id
}
func (e *SimpleEvent) SetTargetAgentID(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.TargetAgentID = id
}
func (e *SimpleEvent) SetSourceAgentID(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.SourceAgentID = id
}
func (e *SimpleEvent) SetData(key string, value any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Data == nil {
		e.Data = make(EventData)
	}
	e.Data[key] = value
}
func (e *SimpleEvent) SetMetadata(key string, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
}

var _ Event = (*SimpleEvent)(nil)
