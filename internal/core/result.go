package agentflow

import (
	"encoding/json"
	"log" // Keep log for actual error logging
	"time"
)

// AgentResult encapsulates the outcome of an Agent's execution.
type AgentResult struct {
	OutputState *State        `json:"output_state"` // Use pointer to State
	Error       string        `json:"error,omitempty"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"-"` // Exclude raw duration
}

// MarshalJSON customizes JSON marshalling for AgentResult.
func (r *AgentResult) MarshalJSON() ([]byte, error) {
	durationMs := r.Duration.Milliseconds()
	aux := struct {
		OutputState *State    `json:"output_state"`
		Error       string    `json:"error,omitempty"`
		StartTime   time.Time `json:"start_time"`
		EndTime     time.Time `json:"end_time"`
		DurationMs  int64     `json:"duration_ms"`
	}{
		OutputState: r.OutputState,
		Error:       r.Error,
		StartTime:   r.StartTime,
		EndTime:     r.EndTime,
		DurationMs:  durationMs,
	}
	jsonData, err := json.Marshal(aux)
	if err != nil {
		// Keep this log for actual errors
		log.Printf("MarshalJSON: Error during json.Marshal(aux): %v", err)
		return jsonData, err
	}
	return jsonData, nil
}

// UnmarshalJSON customizes JSON unmarshalling for AgentResult (Manual Approach).
func (r *AgentResult) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Printf("Unmarshal error (raw map): %v", err) // Keep error log
		return err
	}

	if rawOutputState, ok := raw["output_state"]; ok && string(rawOutputState) != "null" {
		var state State
		if err := json.Unmarshal(rawOutputState, &state); err != nil {
			log.Printf("Unmarshal error (output_state): %v", err) // Keep error log
			return err
		}
		r.OutputState = &state
	} else {
		r.OutputState = nil
	}

	if rawError, ok := raw["error"]; ok {
		if err := json.Unmarshal(rawError, &r.Error); err != nil {
			log.Printf("Unmarshal error (error): %v", err) // Keep error log
			return err
		}
	}

	if rawStartTime, ok := raw["start_time"]; ok {
		if err := json.Unmarshal(rawStartTime, &r.StartTime); err != nil {
			log.Printf("Unmarshal error (start_time): %v", err) // Keep error log
			return err
		}
	}

	if rawEndTime, ok := raw["end_time"]; ok {
		if err := json.Unmarshal(rawEndTime, &r.EndTime); err != nil {
			log.Printf("Unmarshal error (end_time): %v", err) // Keep error log
			return err
		}
	}

	if rawDurationMs, ok := raw["duration_ms"]; ok {
		var durationMs int64
		if err := json.Unmarshal(rawDurationMs, &durationMs); err != nil {
			// Keep error log, including raw value
			log.Printf("Unmarshal error (duration_ms): %v --- Raw: %s", err, string(rawDurationMs))
			return err
		}
		r.Duration = time.Duration(durationMs) * time.Millisecond
	} else {
		// Keep this log, it indicates potentially malformed input
		log.Printf("UnmarshalJSON (manual): duration_ms field not found in JSON")
		r.Duration = 0
	}

	return nil
}
