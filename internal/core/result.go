package agentflow

import (
	"encoding/json"
	"log"
	"time"
)

// AgentResult encapsulates the outcome of an Agent's execution.
type AgentResult struct {
	OutputState State     `json:"output_state"`
	Error       string    `json:"error,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    time.Duration
}

// MarshalJSON customizes JSON marshalling for AgentResult using a map.
func (r *AgentResult) MarshalJSON() ([]byte, error) {
	durationMs := r.Duration.Milliseconds()

	// Create a map to hold the fields for JSON marshalling
	jsonDataMap := map[string]interface{}{
		"output_state": r.OutputState, // Let json.Marshal handle the State interface
		"error":        r.Error,
		"start_time":   r.StartTime,
		"end_time":     r.EndTime,
		"duration_ms":  durationMs, // Add the duration in milliseconds
	}

	// Remove the error field if it's empty (like omitempty)
	if r.Error == "" {
		delete(jsonDataMap, "error")
	}

	// Marshal the map
	jsonData, err := json.Marshal(jsonDataMap)
	if err != nil {
		log.Printf("MarshalJSON (map): Error during json.Marshal: %v", err)
		return jsonData, err
	}

	// log.Printf("DEBUG MarshalJSON (map) output: %s", string(jsonData)) // Optional debug
	return jsonData, nil
}

// UnmarshalJSON customizes JSON unmarshalling for AgentResult (Manual Approach).
func (r *AgentResult) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Printf("Unmarshal error (raw map): %v", err)
		return err
	}

	if rawOutputState, ok := raw["output_state"]; ok && string(rawOutputState) != "null" {
		var concreteState SimpleState
		if err := json.Unmarshal(rawOutputState, &concreteState); err != nil {
			log.Printf("Unmarshal error (output_state into SimpleState): %v", err)
			return err
		}
		r.OutputState = &concreteState
	} else {
		r.OutputState = nil
	}

	if rawError, ok := raw["error"]; ok {
		if err := json.Unmarshal(rawError, &r.Error); err != nil {
			log.Printf("Unmarshal error (error): %v", err)
			return err
		}
	}

	if rawStartTime, ok := raw["start_time"]; ok {
		if err := json.Unmarshal(rawStartTime, &r.StartTime); err != nil {
			log.Printf("Unmarshal error (start_time): %v", err)
			return err
		}
	}

	if rawEndTime, ok := raw["end_time"]; ok {
		if err := json.Unmarshal(rawEndTime, &r.EndTime); err != nil {
			log.Printf("Unmarshal error (end_time): %v", err)
			return err
		}
	}

	if rawDurationMs, ok := raw["duration_ms"]; ok {
		var durationMs int64
		if err := json.Unmarshal(rawDurationMs, &durationMs); err != nil {
			log.Printf("Unmarshal error (duration_ms): %v --- Raw: %s", err, string(rawDurationMs))
			return err
		}
		r.Duration = time.Duration(durationMs) * time.Millisecond
	} else {
		log.Printf("UnmarshalJSON (manual): duration_ms field not found in JSON")
		r.Duration = 0
	}

	return nil
}
