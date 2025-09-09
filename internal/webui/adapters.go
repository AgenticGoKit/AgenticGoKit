package webui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// WebUIToAgentAdapter provides adapters for converting WebUI concepts to agent concepts
type WebUIToAgentAdapter struct {
	logger core.CoreLogger
}

// NewWebUIToAgentAdapter creates a new adapter instance
func NewWebUIToAgentAdapter(logger core.CoreLogger) *WebUIToAgentAdapter {
	return &WebUIToAgentAdapter{
		logger: logger,
	}
}

// ConvertChatMessageToEvent converts a chat message to an agent event
func (adapter *WebUIToAgentAdapter) ConvertChatMessageToEvent(
	sessionID, message string,
	metadata map[string]interface{},
	userContext map[string]string,
) (*WebUIEvent, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	if strings.TrimSpace(message) == "" {
		return nil, fmt.Errorf("message cannot be empty")
	}

	event := &WebUIEvent{
		ID:          generateEventID(),
		SessionID:   sessionID,
		Type:        "chat_message",
		Message:     message,
		Metadata:    metadata,
		Timestamp:   time.Now(),
		UserContext: userContext,
	}

	// Add default metadata if not provided
	if event.Metadata == nil {
		event.Metadata = make(map[string]interface{})
	}

	// Add message processing metadata
	event.Metadata["message_length"] = len(message)
	event.Metadata["word_count"] = len(strings.Fields(message))
	event.Metadata["processed_at"] = time.Now().Unix()

	return event, nil
}

// ConvertAgentResultToResponse converts an agent result to a WebUI response
func (adapter *WebUIToAgentAdapter) ConvertAgentResultToResponse(
	sessionID, agentName string,
	result core.AgentResult,
) (*AgentResponse, error) {
	response := &AgentResponse{
		ID:        generateResponseID(),
		SessionID: sessionID,
		AgentName: agentName,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Handle error cases
	if result.Error != "" {
		response.Status = "error"
		response.Error = result.Error
		response.Content = "I encountered an error while processing your request."
		return response, nil
	}

	// Extract content from output state
	content := "I've processed your request successfully."
	if result.OutputState != nil {
		// Try different keys for response content
		contentKeys := []string{"response", "content", "output", "result", "message"}
		for _, key := range contentKeys {
			if value, ok := result.OutputState.Get(key); ok {
				if strValue, ok := value.(string); ok && strValue != "" {
					content = strValue
					break
				}
			}
		}
	}

	response.Content = content
	response.Status = "complete"

	// Add execution metadata
	response.Metadata["execution_duration_ms"] = result.Duration.Milliseconds()
	response.Metadata["start_time"] = result.StartTime.Unix()
	response.Metadata["end_time"] = result.EndTime.Unix()

	// Add content metadata
	response.Metadata["response_length"] = len(content)
	response.Metadata["word_count"] = len(strings.Fields(content))

	return response, nil
}

// ExtractUserIntent attempts to extract user intent from a message
func (adapter *WebUIToAgentAdapter) ExtractUserIntent(message string) map[string]interface{} {
	intent := make(map[string]interface{})

	// Basic intent analysis
	lowercaseMsg := strings.ToLower(strings.TrimSpace(message))

	// Question detection
	if strings.Contains(lowercaseMsg, "?") ||
		strings.HasPrefix(lowercaseMsg, "what") ||
		strings.HasPrefix(lowercaseMsg, "how") ||
		strings.HasPrefix(lowercaseMsg, "why") ||
		strings.HasPrefix(lowercaseMsg, "when") ||
		strings.HasPrefix(lowercaseMsg, "where") ||
		strings.HasPrefix(lowercaseMsg, "who") {
		intent["type"] = "question"
		intent["confidence"] = 0.8
	}

	// Command detection
	if strings.HasPrefix(lowercaseMsg, "please") ||
		strings.HasPrefix(lowercaseMsg, "can you") ||
		strings.HasPrefix(lowercaseMsg, "could you") ||
		strings.HasPrefix(lowercaseMsg, "help me") {
		intent["type"] = "request"
		intent["confidence"] = 0.7
	}

	// Greeting detection
	greetings := []string{"hello", "hi", "hey", "good morning", "good afternoon", "good evening"}
	for _, greeting := range greetings {
		if strings.Contains(lowercaseMsg, greeting) {
			intent["type"] = "greeting"
			intent["confidence"] = 0.9
			break
		}
	}

	// Default to conversation if no specific intent detected
	if intent["type"] == nil {
		intent["type"] = "conversation"
		intent["confidence"] = 0.5
	}

	// Add message characteristics
	intent["message_length"] = len(message)
	intent["word_count"] = len(strings.Fields(message))
	intent["has_question_mark"] = strings.Contains(message, "?")
	intent["has_exclamation"] = strings.Contains(message, "!")

	return intent
}

// EnhanceEventWithContext adds additional context to an event based on session history
func (adapter *WebUIToAgentAdapter) EnhanceEventWithContext(
	ctx context.Context,
	event *WebUIEvent,
	sessionManager *EnhancedSessionManager,
) error {
	// Get session information
	session, err := sessionManager.GetSession(event.SessionID)
	if err != nil {
		adapter.logger.Warn().
			Err(err).
			Str("session_id", event.SessionID).
			Msg("Could not get session for context enhancement")
		return nil // Don't fail the event processing
	}

	// Add conversation context
	if event.Metadata == nil {
		event.Metadata = make(map[string]interface{})
	}

	// Message history context
	messageCount := len(session.Messages)
	event.Metadata["conversation_length"] = messageCount

	if messageCount > 0 {
		lastMessage := session.Messages[messageCount-1]
		event.Metadata["last_message_role"] = lastMessage.Role
		event.Metadata["last_message_timestamp"] = lastMessage.Timestamp.Unix()
		event.Metadata["time_since_last_message"] = time.Since(lastMessage.Timestamp).Seconds()
	}

	// Session context
	event.Metadata["session_duration"] = time.Since(session.CreatedAt).Seconds()
	event.Metadata["session_age"] = time.Since(session.CreatedAt).String()

	// User context enhancement
	if event.UserContext == nil {
		event.UserContext = make(map[string]string)
	}

	event.UserContext["session_created"] = session.CreatedAt.Format(time.RFC3339)
	event.UserContext["message_count"] = fmt.Sprintf("%d", messageCount)

	// Add recent conversation context (last 3 messages)
	if messageCount > 0 {
		recentCount := 3
		if messageCount < recentCount {
			recentCount = messageCount
		}

		recentMessages := make([]map[string]interface{}, 0, recentCount)
		for i := messageCount - recentCount; i < messageCount; i++ {
			msg := session.Messages[i]
			recentMessages = append(recentMessages, map[string]interface{}{
				"role":      msg.Role,
				"content":   msg.Content,
				"timestamp": msg.Timestamp.Unix(),
			})
		}
		event.Metadata["recent_messages"] = recentMessages
	}

	return nil
}

// ValidateEvent validates that an event has all required fields
func (adapter *WebUIToAgentAdapter) ValidateEvent(event *WebUIEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if event.ID == "" {
		return fmt.Errorf("event ID cannot be empty")
	}

	if event.SessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	if event.Type == "" {
		return fmt.Errorf("event type cannot be empty")
	}

	if strings.TrimSpace(event.Message) == "" {
		return fmt.Errorf("message cannot be empty")
	}

	if event.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}

	return nil
}

// CreateErrorResponse creates a standardized error response
func (adapter *WebUIToAgentAdapter) CreateErrorResponse(
	sessionID, agentName, errorCode, errorMessage string,
) *AgentResponse {
	return &AgentResponse{
		ID:        generateResponseID(),
		SessionID: sessionID,
		AgentName: agentName,
		Content:   "I encountered an error while processing your request. Please try again.",
		Status:    "error",
		Error:     fmt.Sprintf("%s: %s", errorCode, errorMessage),
		Metadata: map[string]interface{}{
			"error_code":    errorCode,
			"error_message": errorMessage,
			"timestamp":     time.Now().Unix(),
		},
		Timestamp:   time.Now(),
		IsStreaming: false,
	}
}

// CreateProcessingResponse creates a standardized processing response
func (adapter *WebUIToAgentAdapter) CreateProcessingResponse(
	sessionID, agentName, message string,
) *AgentResponse {
	if message == "" {
		message = "Processing your request..."
	}

	return &AgentResponse{
		ID:        generateResponseID(),
		SessionID: sessionID,
		AgentName: agentName,
		Content:   message,
		Status:    "processing",
		Metadata: map[string]interface{}{
			"processing_started": time.Now().Unix(),
		},
		Timestamp:   time.Now(),
		IsStreaming: false,
	}
}

// CreateStreamingResponse creates a streaming response chunk
func (adapter *WebUIToAgentAdapter) CreateStreamingResponse(
	sessionID, agentName, content string,
	chunkIndex, totalChunks int,
	isComplete bool,
) *AgentResponse {
	status := "partial"
	if isComplete {
		status = "complete"
	}

	return &AgentResponse{
		ID:          generateResponseID(),
		SessionID:   sessionID,
		AgentName:   agentName,
		Content:     content,
		Status:      status,
		Timestamp:   time.Now(),
		IsStreaming: true,
		ChunkIndex:  chunkIndex,
		TotalChunks: totalChunks,
		Metadata: map[string]interface{}{
			"chunk_size":    len(content),
			"is_final":      isComplete,
			"chunk_created": time.Now().Unix(),
		},
	}
}

// SplitContentForStreaming splits content into chunks for streaming responses
func (adapter *WebUIToAgentAdapter) SplitContentForStreaming(content string, chunkSize int) []string {
	if chunkSize <= 0 {
		chunkSize = 100 // Default chunk size
	}

	if len(content) <= chunkSize {
		return []string{content}
	}

	chunks := make([]string, 0)
	words := strings.Fields(content)
	currentChunk := ""

	for _, word := range words {
		testChunk := currentChunk
		if testChunk != "" {
			testChunk += " "
		}
		testChunk += word

		if len(testChunk) > chunkSize && currentChunk != "" {
			chunks = append(chunks, currentChunk)
			currentChunk = word
		} else {
			currentChunk = testChunk
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// AgentCapabilityMatcher helps match user requests to agent capabilities
type AgentCapabilityMatcher struct {
	adapter *WebUIToAgentAdapter
}

// NewAgentCapabilityMatcher creates a new capability matcher
func NewAgentCapabilityMatcher(adapter *WebUIToAgentAdapter) *AgentCapabilityMatcher {
	return &AgentCapabilityMatcher{
		adapter: adapter,
	}
}

// MatchBestAgent finds the best agent for a given request
func (matcher *AgentCapabilityMatcher) MatchBestAgent(
	message string,
	availableAgents []AgentInfo,
) (*AgentInfo, float64, error) {
	if len(availableAgents) == 0 {
		return nil, 0, fmt.Errorf("no agents available")
	}

	// For now, return the first available agent
	// TODO: Implement sophisticated capability matching
	bestAgent := &availableAgents[0]
	confidence := 0.5 // Default confidence

	// Simple keyword matching for demonstration
	lowercaseMsg := strings.ToLower(message)

	for i, agent := range availableAgents {
		agentScore := 0.0

		// Check if agent role matches message intent
		if strings.Contains(lowercaseMsg, strings.ToLower(agent.Role)) {
			agentScore += 0.3
		}

		// Check capability matching
		for _, capability := range agent.Capabilities {
			if strings.Contains(lowercaseMsg, strings.ToLower(capability)) {
				agentScore += 0.2
			}
		}

		// Prefer enabled agents
		if agent.IsEnabled {
			agentScore += 0.1
		}

		if agentScore > confidence {
			confidence = agentScore
			bestAgent = &availableAgents[i]
		}
	}

	return bestAgent, confidence, nil
}
