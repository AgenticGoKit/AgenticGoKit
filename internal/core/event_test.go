package agentflow

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestSimpleEventJSON verifies JSON marshalling/unmarshalling with exported fields.
func TestSimpleEventJSON(t *testing.T) {
	original := &SimpleEvent{
		ID:            uuid.NewString(),
		Timestamp:     time.Now().Truncate(time.Millisecond), // Truncate for comparison
		TargetAgentID: "agent-B",
		SourceAgentID: "agent-A",
		Data:          EventData{"key": "value", "number": 123.45},
		Metadata:      map[string]string{"env": "test", "session": "s123"},
	}

	// Marshal the original event
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal into a new SimpleEvent
	var parsed SimpleEvent
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// --- Compare using Getters ---
	// FIX: Use GetID()
	if original.GetID() != parsed.GetID() {
		t.Errorf("ID mismatch: want %q, got %q", original.GetID(), parsed.GetID())
	}
	// FIX: Use GetTimestamp() - Compare truncated time
	if !original.GetTimestamp().Equal(parsed.GetTimestamp()) {
		t.Errorf("Timestamp mismatch: want %v, got %v", original.GetTimestamp(), parsed.GetTimestamp())
	}
	// FIX: Use GetTargetAgentID()
	if original.GetTargetAgentID() != parsed.GetTargetAgentID() {
		t.Errorf("TargetAgentID mismatch: want %q, got %q", original.GetTargetAgentID(), parsed.GetTargetAgentID())
	}
	// FIX: Use GetSourceAgentID()
	if original.GetSourceAgentID() != parsed.GetSourceAgentID() {
		t.Errorf("SourceAgentID mismatch: want %q, got %q", original.GetSourceAgentID(), parsed.GetSourceAgentID())
	}
	// FIX: Use GetData() - Use reflect.DeepEqual for maps/slices
	if !reflect.DeepEqual(original.GetData(), parsed.GetData()) {
		// Note: JSON unmarshalling converts numbers to float64 by default
		// DeepEqual should handle this comparison correctly.
		t.Errorf("Data mismatch: want %#v, got %#v", original.GetData(), parsed.GetData())
	}
	// FIX: Use GetMetadata() - Use reflect.DeepEqual for maps
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
		// FIX: Use Data field instead of Payload
		{"nil‑metadata", &SimpleEvent{ID: "1", Data: EventData{"p": "p"}, Metadata: nil}},
		// FIX: Remove duplicate slice-payload case from here, it's tested separately below
		// {"slice‑payload", &SimpleEvent{ID: "2", Data: EventData{"nums": []int{1, 2, 3}}, Metadata: map[string]string{"a": "b"}}},
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
			// This comparison is only valid if no type changes occur during marshal/unmarshal
			// For cases like numbers in interface{}, a custom comparison like the one below is needed.
			if !reflect.DeepEqual(c.event, &out) {
				t.Errorf("round‑trip mismatch\ngot %#v\nwant %#v", out, c.event)
			}
		})
	}

	// Keep the standalone test which correctly handles the type change
	t.Run("slice-payload", func(t *testing.T) {
		original := &SimpleEvent{
			ID:       "2",
			Data:     EventData{"nums": []int{1, 2, 3}}, // Original is []int
			Metadata: map[string]string{"a": "b"},
		}
		jsonData, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var roundTrip SimpleEvent
		err = json.Unmarshal(jsonData, &roundTrip)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Define 'expected' to match JSON unmarshaling behavior (float64 in interface{})
		expectedData := EventData{"nums": []interface{}{float64(1), float64(2), float64(3)}} // Expect []interface{} with float64
		expected := &SimpleEvent{
			ID:        "2",
			Timestamp: time.Time{},  // Zero value after unmarshal
			Data:      expectedData, // Use the corrected expectedData
			Metadata:  map[string]string{"a": "b"},
			// mu is zero value
		}

		// Clear non-deterministic fields for comparison
		roundTrip.Timestamp = time.Time{}
		roundTrip.mu = sync.RWMutex{}

		// Use reflect.DeepEqual for robust comparison
		if !reflect.DeepEqual(&roundTrip, expected) {
			t.Errorf("round-trip mismatch\ngot  %#v\nwant %#v", &roundTrip, expected)
		}
	})
}

func TestNewEvent(t *testing.T) {
	target := "agent-test"
	data := EventData{"key": "value", "count": 1}
	meta := map[string]string{"trace_id": "xyz789", "type": "test"}

	event := NewEvent(target, data, meta)

	if event == nil {
		t.Fatal("NewEvent returned nil")
	}
	// FIX: Use GetID() method
	if event.GetID() == "" {
		t.Error("Event ID should not be empty")
	}
	// Rough check for timestamp, ensure it's recent
	// FIX: Use GetTimestamp() method
	if time.Since(event.GetTimestamp()) > time.Second {
		t.Errorf("Event timestamp %v is too old", event.GetTimestamp())
	}
	// FIX: Use GetTargetAgentID() method
	if event.GetTargetAgentID() != target {
		t.Errorf("Expected target agent ID '%s', got '%s'", target, event.GetTargetAgentID())
	}
	// FIX: Use GetSourceAgentID() method
	if event.GetSourceAgentID() != "" { // Default source should be empty
		t.Errorf("Expected empty source agent ID, got '%s'", event.GetSourceAgentID())
	}

	// Check Data
	// FIX: Use GetData() method
	retrievedData := event.GetData()
	if len(retrievedData) != len(data) {
		t.Errorf("Expected data length %d, got %d", len(data), len(retrievedData))
	}
	if val, ok := retrievedData["key"]; !ok || val != "value" {
		t.Errorf("Expected data['key'] == 'value', got %v (ok: %t)", val, ok)
	}
	// Note: JSON unmarshalling makes numbers float64, but here we set an int.
	// The comparison should work as reflect.DeepEqual handles interface{} comparison.
	if val, ok := retrievedData["count"]; !ok || !reflect.DeepEqual(val, 1) {
		t.Errorf("Expected data['count'] == 1, got %v (type %T) (ok: %t)", val, val, ok)
	}

	// Check Metadata
	// FIX: Use GetMetadata() method
	retrievedMeta := event.GetMetadata()
	if len(retrievedMeta) != len(meta) {
		t.Errorf("Expected metadata length %d, got %d", len(meta), len(retrievedMeta))
	}
	if val, ok := retrievedMeta["trace_id"]; !ok || val != "xyz789" {
		t.Errorf("Expected metadata['trace_id'] == 'xyz789', got %v (ok: %t)", val, ok)
	}
	if val, ok := retrievedMeta["type"]; !ok || val != "test" {
		t.Errorf("Expected metadata['type'] == 'test', got %v (ok: %t)", val, ok)
	}
}

func TestSimpleEvent_Setters(t *testing.T) {
	event := NewEvent("initial_target", nil, nil)

	newID := uuid.NewString()
	newTarget := "new_target"
	newSource := "source_agent"
	dataKey := "status"
	dataValue := "updated"
	metaKey := "priority"
	metaValue := "high"

	event.SetID(newID)
	event.SetTargetAgentID(newTarget)
	event.SetSourceAgentID(newSource)
	event.SetData(dataKey, dataValue)
	event.SetMetadata(metaKey, metaValue)

	// FIX: Use GetID() method
	if event.GetID() != newID {
		t.Errorf("SetID failed: expected %s, got %s", newID, event.GetID())
	}
	// FIX: Use GetTargetAgentID() method
	if event.GetTargetAgentID() != newTarget {
		t.Errorf("SetTargetAgentID failed: expected %s, got %s", newTarget, event.GetTargetAgentID())
	}
	// FIX: Use GetSourceAgentID() method
	if event.GetSourceAgentID() != newSource {
		t.Errorf("SetSourceAgentID failed: expected %s, got %s", newSource, event.GetSourceAgentID())
	}

	// FIX: Use GetData() method
	retrievedData := event.GetData()
	if val, ok := retrievedData[dataKey]; !ok || val != dataValue {
		t.Errorf("SetData failed: expected data['%s'] == '%v', got %v (ok: %t)", dataKey, dataValue, val, ok)
	}

	// FIX: Use GetMetadata() method
	retrievedMeta := event.GetMetadata()
	if val, ok := retrievedMeta[metaKey]; !ok || val != metaValue {
		t.Errorf("SetMetadata failed: expected metadata['%s'] == '%s', got %v (ok: %t)", metaKey, metaValue, val, ok)
	}
}

func TestSimpleEvent_NilMapsOnInit(t *testing.T) {
	// FIX: Use NewEvent constructor with nil maps
	event := NewEvent("target", nil, nil)

	if event.GetData() == nil { // <<< FIX: Use GetData()
		t.Error("Data map should not be nil after NewEvent(nil)")
	}
	if len(event.GetData()) != 0 { // <<< FIX: Use GetData()
		t.Error("Data map should be empty after NewEvent(nil)")
	}
	if event.GetMetadata() == nil { // <<< FIX: Use GetMetadata()
		t.Error("Metadata map should not be nil after NewEvent(nil)")
	}
	if len(event.GetMetadata()) != 0 { // <<< FIX: Use GetMetadata()
		t.Error("Metadata map should be empty after NewEvent(nil)")
	}

	// Test setters on initially nil maps
	event.SetData("key", "value")
	if val, ok := event.GetData()["key"]; !ok || val != "value" { // <<< FIX: Use GetData()
		t.Error("SetData failed on initially nil map")
	}

	event.SetMetadata("meta_key", "meta_value")
	if val, ok := event.GetMetadata()["meta_key"]; !ok || val != "meta_value" { // <<< FIX: Use GetMetadata()
		t.Error("SetMetadata failed on initially nil map")
	}
}

// TestSimpleEventMethods verifies the getter and setter methods work correctly.
func TestSimpleEventMethods(t *testing.T) {
	event := NewEvent("target", EventData{"initial": "data"}, map[string]string{"initial": "meta"})

	// Test Getters
	id := event.GetID()
	if id == "" {
		t.Error("GetID() returned empty string")
	}
	ts := event.GetTimestamp()
	if ts.IsZero() {
		t.Error("GetTimestamp() returned zero time")
	}
	if event.GetTargetAgentID() != "target" {
		t.Errorf("GetTargetAgentID() mismatch: want 'target', got %q", event.GetTargetAgentID())
	}
	if !reflect.DeepEqual(event.GetData(), EventData{"initial": "data"}) {
		t.Errorf("GetData() mismatch: want %#v, got %#v", EventData{"initial": "data"}, event.GetData())
	}
	if !reflect.DeepEqual(event.GetMetadata(), map[string]string{"initial": "meta"}) {
		t.Errorf("GetMetadata() mismatch: want %#v, got %#v", map[string]string{"initial": "meta"}, event.GetMetadata())
	}

	// Test Setters
	newID := uuid.NewString()
	event.SetID(newID)
	if event.GetID() != newID {
		t.Errorf("SetID/GetID mismatch: want %q, got %q", newID, event.GetID())
	}

	event.SetTargetAgentID("new_target")
	if event.GetTargetAgentID() != "new_target" {
		t.Errorf("SetTargetAgentID/GetTargetAgentID mismatch: want 'new_target', got %q", event.GetTargetAgentID())
	}

	event.SetSourceAgentID("source_agent")
	if event.GetSourceAgentID() != "source_agent" {
		t.Errorf("SetSourceAgentID/GetSourceAgentID mismatch: want 'source_agent', got %q", event.GetSourceAgentID())
	}

	event.SetData("newData", 123)
	expectedData := EventData{"initial": "data", "newData": 123}
	if !reflect.DeepEqual(event.GetData(), expectedData) {
		t.Errorf("SetData/GetData mismatch: want %#v, got %#v", expectedData, event.GetData())
	}

	event.SetMetadata("newMeta", "abc")
	expectedMeta := map[string]string{"initial": "meta", "newMeta": "abc"}
	if !reflect.DeepEqual(event.GetMetadata(), expectedMeta) {
		t.Errorf("SetMetadata/GetMetadata mismatch: want %#v, got %#v", expectedMeta, event.GetMetadata())
	}

	// Test GetData/GetMetadata returns copies
	metaCopy := event.GetMetadata()
	metaCopy["modified"] = "should_not_affect_original"
	if event.GetMetadata()["modified"] != "" {
		t.Error("GetMetadata() did not return a copy, original was modified")
	}

	dataCopy := event.GetData()
	dataCopy["modified"] = "should_not_affect_original"
	if _, exists := event.GetData()["modified"]; exists {
		t.Error("GetData() did not return a copy, original was modified")
	}
}
