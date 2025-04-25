package agentflow

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventData holds the payload of an event.
// Using map[string]any is more idiomatic than map[string]interface{}.
type EventData map[string]any

// Event defines the interface for messages passed through the system.
type Event interface {
	GetID() string
	GetTimestamp() time.Time
	GetTargetAgentID() string
	GetSourceAgentID() string
	GetData() EventData             // Returns the payload map
	GetMetadata() map[string]string // Returns the metadata map

	// Mutators (Keep for flexibility, ensure thread-safety)
	SetID(id string)
	SetTargetAgentID(id string)
	SetSourceAgentID(id string)
	SetData(key string, value any) // Changed value type to any
	SetMetadata(key string, value string)
}

// SimpleEvent is a basic implementation of the Event interface.
// Use exported fields for easier JSON handling, protected by mutex via methods.
type SimpleEvent struct {
	// Use exported fields for standard JSON marshalling/unmarshalling
	ID            string            `json:"id"`
	Timestamp     time.Time         `json:"timestamp"`
	TargetAgentID string            `json:"target_agent_id,omitempty"` // Use omitempty if optional
	SourceAgentID string            `json:"source_agent_id,omitempty"` // Use omitempty if optional
	Data          EventData         `json:"data"`                      // Exported map for payload
	Metadata      map[string]string `json:"metadata"`                  // Exported map for metadata

	mu sync.RWMutex `json:"-"` // Exclude mutex from JSON
}

// NewEvent creates a new SimpleEvent with a unique ID and current timestamp.
// TargetAgentID is optional. Data and Metadata are initialized if nil.
func NewEvent(targetAgentID string, data EventData, metadata map[string]string) *SimpleEvent {
	if data == nil {
		data = make(EventData)
	}
	if metadata == nil {
		metadata = make(map[string]string)
	}
	// Ensure ID is generated if not provided (though constructor doesn't take it)
	id := uuid.NewString()

	return &SimpleEvent{
		ID:            id,
		Timestamp:     time.Now(),
		TargetAgentID: targetAgentID,
		Data:          data,
		Metadata:      metadata,
	}
}

// --- Accessors (Implement interface methods) ---

func (e *SimpleEvent) GetID() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.ID // Access exported field
}

func (e *SimpleEvent) GetTimestamp() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Timestamp // Access exported field
}

func (e *SimpleEvent) GetTargetAgentID() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.TargetAgentID // Access exported field
}

func (e *SimpleEvent) GetSourceAgentID() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.SourceAgentID // Access exported field
}

// GetData returns a *copy* of the data map for safety.
func (e *SimpleEvent) GetData() EventData {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// Return a shallow copy
	dataCopy := make(EventData, len(e.Data))
	for k, v := range e.Data {
		dataCopy[k] = v
	}
	return dataCopy // Return copy
}

// GetMetadata returns a *copy* of the metadata map for safety.
func (e *SimpleEvent) GetMetadata() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// Return a shallow copy
	metadataCopy := make(map[string]string, len(e.Metadata))
	for k, v := range e.Metadata {
		metadataCopy[k] = v
	}
	return metadataCopy // Return copy
}

// --- Mutators (Implement interface methods) ---

func (e *SimpleEvent) SetID(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ID = id // Access exported field
}

func (e *SimpleEvent) SetTargetAgentID(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.TargetAgentID = id // Access exported field
}

func (e *SimpleEvent) SetSourceAgentID(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.SourceAgentID = id // Access exported field
}

// SetData sets a specific key-value pair in the Data map.
func (e *SimpleEvent) SetData(key string, value any) { // Changed value type to any
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Data == nil { // Initialize map if nil
		e.Data = make(EventData)
	}
	e.Data[key] = value // Access exported field
}

// SetMetadata sets a specific key-value pair in the Metadata map.
func (e *SimpleEvent) SetMetadata(key string, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Metadata == nil { // Initialize map if nil
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value // Access exported field
}

// Remove custom UnmarshalJSON - rely on standard JSON handling for exported fields.
/*
func (e *SimpleEvent) UnmarshalJSON(data []byte) error {
	type alias SimpleEvent
	aux := &struct {
		Payload interface{} `json:"payload"` // This assumed payload was separate
		*alias
	}{
		alias: (*alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// The logic here put everything under a "payload" key in the data map,
	// which is likely not the desired behavior for a general EventData map.
	// Standard unmarshalling into the exported `Data` field is preferred.
	return nil
}
*/

// Compile-time check to ensure *SimpleEvent implements Event
var _ Event = (*SimpleEvent)(nil)
