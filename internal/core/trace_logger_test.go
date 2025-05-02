package agentflow

import (
	"fmt"
	"sync"
	"testing"
	"time"

	// agentflow "kunalkushwaha/agentflow/internal/core" // Keep commented if types are in the same package

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryTraceLogger_LogAndGet(t *testing.T) { // Renamed test for clarity
	logger := NewInMemoryTraceLogger()
	sessionID1 := "session-log-get-1"
	sessionID2 := "session-log-get-2"
	sessionID3 := "session-log-get-nonexistent"
	eventID1 := uuid.NewString()
	eventID2 := uuid.NewString()

	// Log entries for session 1 (simulating agent run)
	entry1_start := TraceEntry{
		SessionID: sessionID1,
		Timestamp: time.Now(),
		Type:      "agent_start",      // Use valid Type
		Hook:      HookBeforeAgentRun, // Use valid Hook
		EventID:   eventID1,
		AgentID:   "compute",
	}
	err := logger.Log(entry1_start)
	require.NoError(t, err, "Logging entry1_start should succeed")

	time.Sleep(1 * time.Millisecond) // Ensure distinct timestamps

	entry1_end := TraceEntry{
		SessionID: sessionID1,
		Timestamp: time.Now(),
		Type:      "agent_end",       // Use valid Type
		Hook:      HookAfterAgentRun, // Use valid Hook
		EventID:   eventID1,
		AgentID:   "compute",
		AgentResult: &AgentResult{ // Example AgentResult
			// FIX: Add routing key to OutputState map, remove RoutingKey field
			OutputState: NewSimpleState(map[string]any{
				"result": 20,
				"route":  "summarizer", // Assuming 'route' is the key used for routing
			}),
			// RoutingKey:  "summarizer", // Field does not exist
		},
	}
	err = logger.Log(entry1_end)
	require.NoError(t, err, "Logging entry1_end should succeed")

	// Log entry for session 2 (simulating event start)
	entry2_start := TraceEntry{
		SessionID: sessionID2,
		Timestamp: time.Now(),
		Type:      "event_start",           // Use valid Type
		Hook:      HookBeforeEventHandling, // Use valid Hook
		EventID:   eventID2,
	}
	err = logger.Log(entry2_start)
	require.NoError(t, err, "Logging entry2_start should succeed")

	// Get and verify session 1
	trace1, err1 := logger.GetTrace(sessionID1)
	require.NoError(t, err1, "Getting trace for session 1 should not error")
	require.Len(t, trace1, 2, "Expected 2 trace entries for session 1")

	// Assert entry 1
	assert.Equal(t, sessionID1, trace1[0].SessionID)
	assert.Equal(t, eventID1, trace1[0].EventID)
	assert.Equal(t, "agent_start", trace1[0].Type)
	assert.Equal(t, HookBeforeAgentRun, trace1[0].Hook)
	assert.Equal(t, "compute", trace1[0].AgentID)

	// Assert entry 2
	assert.Equal(t, sessionID1, trace1[1].SessionID)
	assert.Equal(t, eventID1, trace1[1].EventID)
	assert.Equal(t, "agent_end", trace1[1].Type)
	assert.Equal(t, HookAfterAgentRun, trace1[1].Hook)
	assert.Equal(t, "compute", trace1[1].AgentID)
	require.NotNil(t, trace1[1].AgentResult)
	// assert.Equal(t, "summarizer", trace1[1].AgentResult.RoutingKey) // Field does not exist
	// FIX: Assert routing key within OutputState
	require.NotNil(t, trace1[1].AgentResult.OutputState, "OutputState in AgentResult should not be nil")
	routeVal, routeOk := trace1[1].AgentResult.OutputState.Get("route")
	require.True(t, routeOk, "OutputState should contain 'route' key")
	assert.Equal(t, "summarizer", routeVal, "OutputState 'route' key should be 'summarizer'")

	assert.True(t, trace1[1].Timestamp.After(trace1[0].Timestamp), "Timestamp order should be chronological")

	// Get and verify session 2
	trace2, err2 := logger.GetTrace(sessionID2)
	require.NoError(t, err2, "Getting trace for session 2 should not error")
	require.Len(t, trace2, 1, "Expected 1 trace entry for session 2")
	assert.Equal(t, sessionID2, trace2[0].SessionID)
	assert.Equal(t, eventID2, trace2[0].EventID)
	assert.Equal(t, "event_start", trace2[0].Type)
	assert.Equal(t, HookBeforeEventHandling, trace2[0].Hook)

	// Get and verify non-existent session 3
	trace3, err3 := logger.GetTrace(sessionID3)
	require.NoError(t, err3, "Getting trace for non-existent session should not error")
	assert.Empty(t, trace3, "Expected empty trace for non-existent session 3")
}

func TestInMemoryTraceLogger_Concurrency(t *testing.T) {
	logger := NewInMemoryTraceLogger()
	numGoroutines := 10
	numLogsPerGoroutine := 5 // e.g., start and end for a few steps
	var wg sync.WaitGroup

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(gIndex int) {
			defer wg.Done()
			sessionID := fmt.Sprintf("session-concurrent-%d", gIndex)
			agentID := fmt.Sprintf("agent-%d", gIndex)
			for j := 0; j < numLogsPerGoroutine; j++ {
				eventID := uuid.NewString()
				// Simulate agent start
				entryStart := TraceEntry{
					SessionID: sessionID,
					Timestamp: time.Now(),
					Type:      "agent_start",
					Hook:      HookBeforeAgentRun,
					EventID:   eventID,
					AgentID:   agentID,
				}
				logger.Log(entryStart) // Ignoring error for brevity in test

				time.Sleep(time.Duration(gIndex+1) * time.Millisecond) // Vary sleep

				// Simulate agent end
				entryEnd := TraceEntry{
					SessionID: sessionID,
					Timestamp: time.Now(),
					Type:      "agent_end",
					Hook:      HookAfterAgentRun,
					EventID:   eventID,
					AgentID:   agentID,
					AgentResult: &AgentResult{ // Example result
						// FIX: Add routing key to OutputState map, remove RoutingKey field
						OutputState: NewSimpleState(map[string]any{
							"step":  j,
							"route": fmt.Sprintf("next-agent-%d", j%2), // Example routing
						}),
						// RoutingKey: fmt.Sprintf("next-agent-%d", j%2), // Field does not exist
					},
				}
				logger.Log(entryEnd) // Ignoring error for brevity in test
			}
		}(i)
	}

	wg.Wait() // Wait for all goroutines to finish logging

	// Verify traces for each session
	totalExpectedLogs := numLogsPerGoroutine * 2 // start and end for each "step"
	for i := 0; i < numGoroutines; i++ {
		sessionID := fmt.Sprintf("session-concurrent-%d", i)
		agentID := fmt.Sprintf("agent-%d", i)

		trace, err := logger.GetTrace(sessionID)
		require.NoError(t, err, "Getting trace for session %s should not error", sessionID)
		// Note: Length check might be fragile if timing causes logs from different "steps" to interleave perfectly.
		// A more robust check might involve filtering by eventID or checking pairs.
		// For simplicity, we check total count here.
		assert.Len(t, trace, totalExpectedLogs, "Expected %d total trace entries for session %s", totalExpectedLogs, sessionID)

		// Basic check for content and order within the retrieved (sorted) trace
		for k := 0; k < totalExpectedLogs; k++ {
			assert.Equal(t, sessionID, trace[k].SessionID)
			assert.Equal(t, agentID, trace[k].AgentID) // All entries in this session should have same agent ID
			if k > 0 {
				// FIX: Allow equal timestamps due to sorting and time resolution. Check !Before instead of After.
				assert.True(t, !trace[k].Timestamp.Before(trace[k-1].Timestamp), "Timestamp order incorrect (should be non-decreasing) in session %s", sessionID)
			}
			// Check Type and Hook alternate
			if k%2 == 0 {
				assert.Equal(t, "agent_start", trace[k].Type)
				assert.Equal(t, HookBeforeAgentRun, trace[k].Hook)
			} else {
				assert.Equal(t, "agent_end", trace[k].Type)
				assert.Equal(t, HookAfterAgentRun, trace[k].Hook)
				require.NotNil(t, trace[k].AgentResult)
				// FIX: Check OutputState for routing info if needed
				require.NotNil(t, trace[k].AgentResult.OutputState, "OutputState in AgentResult should not be nil for session %s, index %d", sessionID, k)
				_, routeOk := trace[k].AgentResult.OutputState.Get("route")
				assert.True(t, routeOk, "OutputState should contain 'route' key for session %s, index %d", sessionID, k)
			}
		}
	}
}
