package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// SessionStorage defines the interface for session persistence
type SessionStorage interface {
	Save(session *ChatSession) error
	Load(sessionID string) (*ChatSession, error)
	Delete(sessionID string) error
	List() ([]string, error)
	Clear() error
}

// InMemoryStorage implements SessionStorage in memory
type InMemoryStorage struct {
	sessions map[string]*ChatSession
	mu       sync.RWMutex
}

// NewInMemoryStorage creates a new in-memory storage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		sessions: make(map[string]*ChatSession),
	}
}

func (s *InMemoryStorage) Save(session *ChatSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a deep copy to avoid race conditions
	sessionCopy := *session
	sessionCopy.Messages = make([]ChatMessage, len(session.Messages))
	copy(sessionCopy.Messages, session.Messages)

	s.sessions[session.ID] = &sessionCopy
	return nil
}

func (s *InMemoryStorage) Load(sessionID string) (*ChatSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Return a copy to avoid external modification
	sessionCopy := *session
	sessionCopy.Messages = make([]ChatMessage, len(session.Messages))
	copy(sessionCopy.Messages, session.Messages)

	return &sessionCopy, nil
}

func (s *InMemoryStorage) Delete(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

func (s *InMemoryStorage) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.sessions))
	for id := range s.sessions {
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *InMemoryStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions = make(map[string]*ChatSession)
	return nil
}

// FileStorage implements SessionStorage with file persistence
type FileStorage struct {
	baseDir string
	mu      sync.RWMutex
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(baseDir string) (*FileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FileStorage{
		baseDir: baseDir,
	}, nil
}

func (s *FileStorage) Save(session *ChatSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.baseDir, session.ID+".json")

	// Create a copy without context fields for JSON serialization
	persistentSession := struct {
		ID        string        `json:"id"`
		Messages  []ChatMessage `json:"messages"`
		CreatedAt time.Time     `json:"created_at"`
		LastUsed  time.Time     `json:"last_used"`
		UserAgent string        `json:"user_agent,omitempty"`
		IPAddress string        `json:"ip_address,omitempty"`
		Active    bool          `json:"active"`
		SavedAt   time.Time     `json:"saved_at"`
	}{
		ID:        session.ID,
		Messages:  session.Messages,
		CreatedAt: session.CreatedAt,
		LastUsed:  session.LastUsed,
		UserAgent: session.UserAgent,
		IPAddress: session.IPAddress,
		Active:    session.Active,
		SavedAt:   time.Now(),
	}

	data, err := json.MarshalIndent(persistentSession, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

func (s *FileStorage) Load(sessionID string) (*ChatSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := filepath.Join(s.baseDir, sessionID+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session %s not found", sessionID)
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var persistentSession struct {
		ID        string        `json:"id"`
		Messages  []ChatMessage `json:"messages"`
		CreatedAt time.Time     `json:"created_at"`
		LastUsed  time.Time     `json:"last_used"`
		UserAgent string        `json:"user_agent,omitempty"`
		IPAddress string        `json:"ip_address,omitempty"`
		Active    bool          `json:"active"`
		SavedAt   time.Time     `json:"saved_at"`
	}

	if err := json.Unmarshal(data, &persistentSession); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Reconstruct the session
	session := &ChatSession{
		ID:        persistentSession.ID,
		Messages:  persistentSession.Messages,
		CreatedAt: persistentSession.CreatedAt,
		LastUsed:  persistentSession.LastUsed,
		UserAgent: persistentSession.UserAgent,
		IPAddress: persistentSession.IPAddress,
		Active:    persistentSession.Active,
	}

	return session, nil
}

func (s *FileStorage) Delete(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.baseDir, sessionID+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}

func (s *FileStorage) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var sessionIDs []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			sessionID := file.Name()[:len(file.Name())-5] // Remove .json extension
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs, nil
}

func (s *FileStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read storage directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(s.baseDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}

// SessionMetrics contains session statistics
type SessionMetrics struct {
	TotalSessions    int           `json:"total_sessions"`
	ActiveSessions   int           `json:"active_sessions"`
	InactiveSessions int           `json:"inactive_sessions"`
	TotalMessages    int           `json:"total_messages"`
	AverageMessages  float64       `json:"average_messages"`
	OldestSession    time.Time     `json:"oldest_session"`
	NewestSession    time.Time     `json:"newest_session"`
	StorageType      string        `json:"storage_type"`
	LastCleanup      time.Time     `json:"last_cleanup"`
	CleanupInterval  time.Duration `json:"cleanup_interval"`
}

// SessionConfig contains configuration for session management
type SessionConfig struct {
	// Storage configuration
	StorageType string `toml:"storage_type"` // "memory" or "file"
	StorageDir  string `toml:"storage_dir"`  // Directory for file storage

	// Session lifecycle
	SessionTimeout  time.Duration `toml:"session_timeout"`  // Time before session expires
	CleanupInterval time.Duration `toml:"cleanup_interval"` // How often to run cleanup
	MaxSessions     int           `toml:"max_sessions"`     // Maximum number of sessions

	// Message limits
	MaxMessages      int           `toml:"max_messages"`      // Maximum messages per session
	MessageRetention time.Duration `toml:"message_retention"` // How long to keep messages

	// Auto-save configuration
	AutoSave     bool          `toml:"auto_save"`     // Whether to auto-save sessions
	SaveInterval time.Duration `toml:"save_interval"` // How often to save sessions
}

// DefaultSessionConfig returns default session configuration
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		StorageType:      "memory",
		StorageDir:       "./sessions",
		SessionTimeout:   24 * time.Hour,
		CleanupInterval:  1 * time.Hour,
		MaxSessions:      1000,
		MaxMessages:      1000,
		MessageRetention: 7 * 24 * time.Hour, // 7 days
		AutoSave:         true,
		SaveInterval:     5 * time.Minute,
	}
}

// Enhanced SessionManager with advanced features
type EnhancedSessionManager struct {
	storage      SessionStorage
	config       *SessionConfig
	coreConfig   *core.Config
	mu           sync.RWMutex
	cleanupTimer *time.Timer
	saveTimer    *time.Timer
	ctx          context.Context
	cancel       context.CancelFunc
	metrics      SessionMetrics

	// Event callbacks
	onSessionCreate func(*ChatSession)
	onSessionUpdate func(*ChatSession)
	onSessionDelete func(string)
	onCleanup       func(int) // Number of sessions cleaned
}

// NewEnhancedSessionManager creates a new enhanced session manager
func NewEnhancedSessionManager(coreConfig *core.Config, sessionConfig *SessionConfig) (*EnhancedSessionManager, error) {
	if sessionConfig == nil {
		sessionConfig = DefaultSessionConfig()
	}

	// Create storage based on configuration
	var storage SessionStorage
	var err error

	switch sessionConfig.StorageType {
	case "memory":
		storage = NewInMemoryStorage()
	case "file":
		storage, err = NewFileStorage(sessionConfig.StorageDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create file storage: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", sessionConfig.StorageType)
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &EnhancedSessionManager{
		storage:    storage,
		config:     sessionConfig,
		coreConfig: coreConfig,
		ctx:        ctx,
		cancel:     cancel,
		metrics: SessionMetrics{
			StorageType:     sessionConfig.StorageType,
			CleanupInterval: sessionConfig.CleanupInterval,
		},
	}

	// Start background routines
	manager.startCleanupRoutine()
	if sessionConfig.AutoSave {
		manager.startSaveRoutine()
	}

	return manager, nil
}

// CreateSession creates a new chat session with enhanced features
func (sm *EnhancedSessionManager) CreateSession(userAgent, ipAddress string) (*ChatSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check session limits
	if sm.config.MaxSessions > 0 {
		sessionIDs, err := sm.storage.List()
		if err != nil {
			return nil, fmt.Errorf("failed to check session count: %w", err)
		}

		if len(sessionIDs) >= sm.config.MaxSessions {
			return nil, fmt.Errorf("maximum number of sessions (%d) reached", sm.config.MaxSessions)
		}
	}

	session := &ChatSession{
		ID:        generateSessionID(),
		Messages:  make([]ChatMessage, 0),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		UserAgent: userAgent,
		IPAddress: ipAddress,
		Active:    true,
		AgentCtx:  sm.ctx,
		Config:    sm.coreConfig,
	}

	// Save to storage
	if err := sm.storage.Save(session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Update metrics
	sm.updateMetricsAfterCreate(session)

	// Trigger callback
	if sm.onSessionCreate != nil {
		sm.onSessionCreate(session)
	}

	return session, nil
}

// GetSession retrieves a session by ID with enhanced error handling
func (sm *EnhancedSessionManager) GetSession(sessionID string) (*ChatSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, err := sm.storage.Load(sessionID)
	if err != nil {
		return nil, err
	}

	// Check if session is expired
	if sm.isSessionExpired(session) {
		// Auto-cleanup expired session
		go func() {
			sm.mu.Lock()
			defer sm.mu.Unlock()
			sm.storage.Delete(sessionID)
			if sm.onSessionDelete != nil {
				sm.onSessionDelete(sessionID)
			}
		}()
		return nil, fmt.Errorf("session %s has expired", sessionID)
	}

	// Update last used timestamp
	session.LastUsed = time.Now()
	session.AgentCtx = sm.ctx
	session.Config = sm.coreConfig

	return session, nil
}

// UpdateSession updates a session in storage
func (sm *EnhancedSessionManager) UpdateSession(session *ChatSession) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session.LastUsed = time.Now()

	if err := sm.storage.Save(session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Trigger callback
	if sm.onSessionUpdate != nil {
		sm.onSessionUpdate(session)
	}

	return nil
}

// AddMessage adds a message to a session with limits checking
func (sm *EnhancedSessionManager) AddMessage(sessionID string, message ChatMessage) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check message limits
	if sm.config.MaxMessages > 0 && len(session.Messages) >= sm.config.MaxMessages {
		// Remove oldest messages to make room
		numToRemove := len(session.Messages) - sm.config.MaxMessages + 1
		session.Messages = session.Messages[numToRemove:]
	}

	// Add message
	session.AddMessage(message)

	// Save updated session
	if err := sm.storage.Save(session); err != nil {
		return fmt.Errorf("failed to save session after adding message: %w", err)
	}

	// Update metrics
	sm.metrics.TotalMessages++

	return nil
}

// ListSessions returns all active sessions with pagination
func (sm *EnhancedSessionManager) ListSessions(offset, limit int) ([]*ChatSession, int, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessionIDs, err := sm.storage.List()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessions []*ChatSession
	for _, id := range sessionIDs {
		session, err := sm.storage.Load(id)
		if err != nil {
			continue // Skip corrupted sessions
		}

		if session.Active && !sm.isSessionExpired(session) {
			sessions = append(sessions, session)
		}
	}

	total := len(sessions)

	// Apply pagination
	if offset >= total {
		return []*ChatSession{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	if limit > 0 {
		sessions = sessions[offset:end]
	}

	return sessions, total, nil
}

// DeleteSession deletes a session
func (sm *EnhancedSessionManager) DeleteSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if err := sm.storage.Delete(sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Trigger callback
	if sm.onSessionDelete != nil {
		sm.onSessionDelete(sessionID)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (sm *EnhancedSessionManager) CleanupExpiredSessions() (int, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionIDs, err := sm.storage.List()
	if err != nil {
		return 0, fmt.Errorf("failed to list sessions for cleanup: %w", err)
	}

	cleaned := 0
	for _, id := range sessionIDs {
		session, err := sm.storage.Load(id)
		if err != nil {
			continue
		}

		if sm.isSessionExpired(session) {
			if err := sm.storage.Delete(id); err == nil {
				cleaned++
				if sm.onSessionDelete != nil {
					sm.onSessionDelete(id)
				}
			}
		}
	}

	// Update metrics
	sm.metrics.LastCleanup = time.Now()

	// Trigger callback
	if sm.onCleanup != nil {
		sm.onCleanup(cleaned)
	}

	return cleaned, nil
}

// GetMetrics returns session metrics
func (sm *EnhancedSessionManager) GetMetrics() (SessionMetrics, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessionIDs, err := sm.storage.List()
	if err != nil {
		return sm.metrics, fmt.Errorf("failed to get session list for metrics: %w", err)
	}

	metrics := sm.metrics
	metrics.TotalSessions = len(sessionIDs)
	metrics.ActiveSessions = 0
	metrics.InactiveSessions = 0
	totalMessages := 0

	var oldestTime, newestTime time.Time

	for _, id := range sessionIDs {
		session, err := sm.storage.Load(id)
		if err != nil {
			continue
		}

		totalMessages += len(session.Messages)

		if session.Active && !sm.isSessionExpired(session) {
			metrics.ActiveSessions++
		} else {
			metrics.InactiveSessions++
		}

		if oldestTime.IsZero() || session.CreatedAt.Before(oldestTime) {
			oldestTime = session.CreatedAt
		}

		if newestTime.IsZero() || session.CreatedAt.After(newestTime) {
			newestTime = session.CreatedAt
		}
	}

	metrics.TotalMessages = totalMessages
	if metrics.TotalSessions > 0 {
		metrics.AverageMessages = float64(totalMessages) / float64(metrics.TotalSessions)
	}
	metrics.OldestSession = oldestTime
	metrics.NewestSession = newestTime

	return metrics, nil
}

// SetCallbacks sets event callbacks
func (sm *EnhancedSessionManager) SetCallbacks(
	onCreate func(*ChatSession),
	onUpdate func(*ChatSession),
	onDelete func(string),
	onCleanup func(int),
) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.onSessionCreate = onCreate
	sm.onSessionUpdate = onUpdate
	sm.onSessionDelete = onDelete
	sm.onCleanup = onCleanup
}

// Stop gracefully stops the session manager
func (sm *EnhancedSessionManager) Stop() error {
	sm.cancel()

	if sm.cleanupTimer != nil {
		sm.cleanupTimer.Stop()
	}

	if sm.saveTimer != nil {
		sm.saveTimer.Stop()
	}

	// Final save of all sessions if auto-save is enabled
	if sm.config.AutoSave {
		return sm.saveAllSessions()
	}

	return nil
}

// Private helper methods

func (sm *EnhancedSessionManager) isSessionExpired(session *ChatSession) bool {
	return time.Since(session.LastUsed) > sm.config.SessionTimeout
}

func (sm *EnhancedSessionManager) updateMetricsAfterCreate(session *ChatSession) {
	sm.metrics.TotalSessions++
	if session.Active {
		sm.metrics.ActiveSessions++
	}
}

func (sm *EnhancedSessionManager) startCleanupRoutine() {
	sm.cleanupTimer = time.AfterFunc(sm.config.CleanupInterval, func() {
		select {
		case <-sm.ctx.Done():
			return
		default:
		}

		sm.CleanupExpiredSessions()
		sm.startCleanupRoutine() // Reschedule
	})
}

func (sm *EnhancedSessionManager) startSaveRoutine() {
	sm.saveTimer = time.AfterFunc(sm.config.SaveInterval, func() {
		select {
		case <-sm.ctx.Done():
			return
		default:
		}

		sm.saveAllSessions()
		sm.startSaveRoutine() // Reschedule
	})
}

func (sm *EnhancedSessionManager) saveAllSessions() error {
	sessionIDs, err := sm.storage.List()
	if err != nil {
		return err
	}

	for _, id := range sessionIDs {
		session, err := sm.storage.Load(id)
		if err != nil {
			continue
		}

		sm.storage.Save(session)
	}

	return nil
}
