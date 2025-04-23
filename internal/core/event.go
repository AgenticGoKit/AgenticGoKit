package agentflow

import (
	"encoding/json"
)

// Event defines the standard input contract for all workflows.
type Event interface {
	GetID() string
	GetPayload() interface{}
	GetMetadata() map[string]string
}

// SimpleEvent is a basic, JSON‑serializable implementation of Event.
type SimpleEvent struct {
	ID       string            `json:"id"`
	Payload  interface{}       `json:"payload"`
	Metadata map[string]string `json:"metadata"`
}

func (e *SimpleEvent) GetID() string                  { return e.ID }
func (e *SimpleEvent) GetPayload() interface{}        { return e.Payload }
func (e *SimpleEvent) GetMetadata() map[string]string { return e.Metadata }

// UnmarshalJSON implements a custom JSON unmarshal for SimpleEvent
// so that numeric slices come back as []int when possible.
func (e *SimpleEvent) UnmarshalJSON(data []byte) error {
	type alias SimpleEvent
	aux := &struct {
		Payload interface{} `json:"payload"`
		*alias
	}{
		alias: (*alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// if payload is []interface{} (from JSON), try to convert to []int
	if arr, ok := aux.Payload.([]interface{}); ok {
		ints := make([]int, len(arr))
		for i, v := range arr {
			num, ok := v.(float64)
			if !ok {
				// non‐float64 element – fall back to original
				e.Payload = aux.Payload
				return nil
			}
			ints[i] = int(num)
		}
		e.Payload = ints
	} else {
		e.Payload = aux.Payload
	}
	return nil
}
