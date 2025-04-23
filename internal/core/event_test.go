package agentflow

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestSimpleEventJSON verifies JSON marshalling/unmarshalling.
func TestSimpleEventJSON(t *testing.T) {
	original := &SimpleEvent{
		ID:       "evt-123",
		Payload:  map[string]interface{}{"key": "value"},
		Metadata: map[string]string{"env": "test"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var parsed SimpleEvent
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if original.GetID() != parsed.GetID() {
		t.Errorf("ID mismatch: want %q, got %q", original.GetID(), parsed.GetID())
	}
	if !reflect.DeepEqual(original.GetPayload(), parsed.GetPayload()) {
		t.Errorf("Payload mismatch: want %#v, got %#v", original.GetPayload(), parsed.GetPayload())
	}
	if !reflect.DeepEqual(original.GetMetadata(), parsed.GetMetadata()) {
		t.Errorf("Metadata mismatch: want %#v, got %#v", original.GetMetadata(), parsed.GetMetadata())
	}
}

func TestSimpleEventEdgeCases(t *testing.T) {
	cases := []struct {
		name  string
		event *SimpleEvent
	}{
		{"empty", &SimpleEvent{}},
		{"nil‑metadata", &SimpleEvent{ID: "1", Payload: "p", Metadata: nil}},
		{"slice‑payload", &SimpleEvent{ID: "2", Payload: []int{1, 2, 3}, Metadata: map[string]string{"a": "b"}}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, err := json.Marshal(c.event)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			var out SimpleEvent
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if !reflect.DeepEqual(c.event, &out) {
				t.Errorf("round‑trip mismatch\ngot %#v\nwant %#v", out, c.event)
			}
		})
	}
}
