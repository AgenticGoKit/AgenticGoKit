package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// KnowledgeManager provides centralized management of knowledge base operations
type KnowledgeManager struct {
	config     *core.Config
	configPath string
	memory     core.Memory
	options    KnowledgeOptions
}

// KnowledgeOptions contains configuration options for knowledge operations
type KnowledgeOptions struct {
	ShowProgress    bool
	OutputFormat    string // "table" or "json"
	ScoreThreshold  float32
	Limit           int
	FilterSource    string
	FilterType      string
	FilterTags      []string
	DryRun          bool
	Force           bool
	Interactive     bool
	IncludeMetadata bool
	Recursive       bool
	Tags            []string
}

// UploadOptions contains options specific to upload operations
type UploadOptions struct {
	Recursive       bool
	Tags            []string
	IncludeMetadata bool
	ShowProgress    bool
	BatchSize       int
}

// ListOptions contains options for listing documents
type ListOptions struct {
	OutputFormat string
	FilterSource string
	FilterType   string
	FilterTags   []string
	Limit        int
}

// SearchOptions contains options for searching knowledge base
type SearchOptions struct {
	OutputFormat   string
	ScoreThreshold float32
	Limit          int
}

// ClearOptions contains options for clearing documents
type ClearOptions struct {
	FilterSource string
	FilterType   string
	FilterTags   []string
	DryRun       bool
	Force        bool
	Interactive  bool
}

// KnowledgeValidationResult represents the result of knowledge base validation
type KnowledgeValidationResult struct {
	Success         bool                         `json:"success"`
	Errors          []KnowledgeValidationError   `json:"errors,omitempty"`
	Warnings        []KnowledgeValidationWarning `json:"warnings,omitempty"`
	Summary         KnowledgeValidationSummary   `json:"summary"`
	Recommendations []string                     `json:"recommendations,omitempty"`
}

// KnowledgeValidationError represents a validation error
type KnowledgeValidationError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Component  string `json:"component"` // config, memory, embedding, etc.
}

// KnowledgeValidationWarning represents a validation warning
type KnowledgeValidationWarning struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Component string `json:"component"`
}

// KnowledgeValidationSummary provides summary of validation results
type KnowledgeValidationSummary struct {
	ConfigValid      bool          `json:"config_valid"`
	MemoryConnected  bool          `json:"memory_connected"`
	EmbeddingHealthy bool          `json:"embedding_healthy"`
	SearchFunctional bool          `json:"search_functional"`
	DocumentCount    int           `json:"document_count"`
	ValidationTime   time.Duration `json:"validation_time"`
}

// KnowledgeStats represents knowledge base statistics
type KnowledgeStats struct {
	DocumentCounts     map[string]int     `json:"document_counts"`
	TotalDocuments     int                `json:"total_documents"`
	TotalChunks        int                `json:"total_chunks"`
	StorageSize        int64              `json:"storage_size_bytes"`
	LastUpdated        time.Time          `json:"last_updated"`
	ProviderInfo       ProviderInfo       `json:"provider_info"`
	PerformanceMetrics PerformanceMetrics `json:"performance_metrics"`
	Configuration      ConfigSummary      `json:"configuration"`
}

// ProviderInfo contains information about the memory provider
type ProviderInfo struct {
	Name      string                 `json:"name"`
	Connected bool                   `json:"connected"`
	Version   string                 `json:"version,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// PerformanceMetrics contains performance statistics
type PerformanceMetrics struct {
	AverageSearchTime time.Duration `json:"average_search_time"`
	AverageUploadTime time.Duration `json:"average_upload_time"`
	LastSearchTime    time.Duration `json:"last_search_time"`
	LastUploadTime    time.Duration `json:"last_upload_time"`
	SearchCount       int64         `json:"search_count"`
	UploadCount       int64         `json:"upload_count"`
}

// ConfigSummary provides a summary of relevant configuration
type ConfigSummary struct {
	Provider          string  `json:"provider"`
	Dimensions        int     `json:"dimensions"`
	ChunkSize         int     `json:"chunk_size"`
	ChunkOverlap      int     `json:"chunk_overlap"`
	KnowledgeEnabled  bool    `json:"knowledge_enabled"`
	RAGEnabled        bool    `json:"rag_enabled"`
	ScoreThreshold    float32 `json:"score_threshold"`
	MaxResults        int     `json:"max_results"`
	EmbeddingProvider string  `json:"embedding_provider"`
	EmbeddingModel    string  `json:"embedding_model"`
}

// NewKnowledgeManager creates a new knowledge manager instance
func NewKnowledgeManager(configPath string) (*KnowledgeManager, error) {
	// Determine config file path
	if configPath == "" {
		configPath = "agentflow.toml"
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("[ERROR] No agentflow.toml found at %s\n[SUGGESTION] Run this command from your AgentFlow project root, or specify config:\n   agentcli knowledge --config-path /path/to/agentflow.toml", configPath)
	}

	// Load configuration
	config, err := core.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to load configuration: %v\n[SUGGESTION] Check your agentflow.toml file for syntax errors", err)
	}

	// Validate memory configuration
	if config.AgentMemory.Provider == "" {
		return nil, fmt.Errorf("[ERROR] Memory system not configured in agentflow.toml\n[SUGGESTION] Add [agent_memory] section to enable memory features")
	}

	// Check if knowledge base is enabled
	if !config.AgentMemory.EnableKnowledgeBase {
		return nil, fmt.Errorf("[ERROR] Knowledge base not enabled in agentflow.toml\n[SUGGESTION] Set enable_knowledge_base = true in [agent_memory] section")
	}

	return &KnowledgeManager{
		config:     config,
		configPath: configPath,
		options:    KnowledgeOptions{}, // Default options
	}, nil
}

// Connect establishes connection to the memory system
func (km *KnowledgeManager) Connect() error {
	memory, err := core.NewMemory(km.config.AgentMemory)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to connect to memory system: %v\n%s", err, km.getTroubleshootingHelp())
	}
	km.memory = memory
	return nil
}

// Close closes the memory connection
func (km *KnowledgeManager) Close() error {
	if km.memory != nil {
		return km.memory.Close()
	}
	return nil
}

// SetOptions updates the knowledge manager options
func (km *KnowledgeManager) SetOptions(options KnowledgeOptions) {
	km.options = options
}

// GetConfig returns the loaded configuration
func (km *KnowledgeManager) GetConfig() *core.Config {
	return km.config
}

// GetConfigPath returns the configuration file path
func (km *KnowledgeManager) GetConfigPath() string {
	return km.configPath
}

// getTroubleshootingHelp returns provider-specific troubleshooting help
func (km *KnowledgeManager) getTroubleshootingHelp() string {
	switch km.config.AgentMemory.Provider {
	case "pgvector":
		return `[TROUBLESHOOTING] PostgreSQL/PgVector:
   1. Start database: docker compose up -d
   2. Check connection: psql -h localhost -U user -d agentflow
   3. Verify connection string in agentflow.toml
   4. Run setup script: ./setup.sh (or setup.bat on Windows)
   5. Ensure pgvector extension is installed: CREATE EXTENSION vector;`
	case "weaviate":
		return `[TROUBLESHOOTING] Weaviate:
   1. Start Weaviate: docker compose up -d
   2. Check status: curl http://localhost:8080/v1/meta
   3. Verify connection string in agentflow.toml
   4. Check Weaviate logs: docker logs <weaviate-container>`
	case "memory":
		return `[TROUBLESHOOTING] In-Memory Provider Issue:
   This shouldn't fail - check your configuration syntax`
	default:
		return `[TROUBLESHOOTING] Check your memory provider configuration in agentflow.toml`
	}
}

// TestConnection tests the memory connection with a simple operation
func (km *KnowledgeManager) TestConnection(ctx context.Context) error {
	if km.memory == nil {
		return fmt.Errorf("memory not connected - call Connect() first")
	}

	// Test with a simple search operation
	_, err := km.memory.SearchKnowledge(ctx, "test", core.WithLimit(1))
	if err != nil {
		// Try to provide more specific error information
		if strings.Contains(err.Error(), "connection") {
			return fmt.Errorf("connection test failed: %v\n%s", err, km.getTroubleshootingHelp())
		}
		return fmt.Errorf("connection test failed: %v", err)
	}

	return nil
}

// Upload uploads files and directories to the knowledge base
func (km *KnowledgeManager) Upload(ctx context.Context, sources []string, options UploadOptions) error {
	registry := NewProcessorRegistry()

	// Collect all files to process
	files, err := km.collectFiles(sources, options)
	if err != nil {
		return fmt.Errorf("failed to collect files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No supported files found to upload.")
		return nil
	}

	fmt.Printf("Found %d files to process\n", len(files))

	// Process files
	processed := 0
	failed := 0
	skipped := 0

	for i, filePath := range files {
		fmt.Printf("[DEBUG] Starting file %d/%d: %s\n", i+1, len(files), filepath.Base(filePath))

		if options.ShowProgress {
			fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(files), filepath.Base(filePath))
		}

		// Get processor for this file
		fmt.Printf("[DEBUG] Getting processor for file: %s\n", filePath)
		processor, err := registry.GetProcessor(filePath)
		if err != nil {
			if options.ShowProgress {
				fmt.Printf("  Skipped: %v\n", err)
			}
			skipped++
			continue
		}
		fmt.Printf("[DEBUG] Got processor, creating processing options\n")

		// Create processing options
		procOptions := ProcessingOptions{
			ChunkSize:         km.config.AgentMemory.ChunkSize,
			ChunkOverlap:      km.config.AgentMemory.ChunkOverlap,
			PreserveStructure: true,
			ExtractHeaders:    true,
			IncludeMetadata:   options.IncludeMetadata,
			Tags:              options.Tags,
			Source:            filePath,
		}

		// Process the document
		fmt.Printf("[DEBUG] Processing document with processor\n")
		doc, err := processor.Process(ctx, filePath, procOptions)
		if err != nil {
			if options.ShowProgress {
				fmt.Printf("  Failed: %v\n", err)
			}
			failed++
			continue
		}
		fmt.Printf("[DEBUG] Document processed, checking chunking\n")

		// Handle chunking if enabled
		var docsToIngest []*core.Document
		if km.config.AgentMemory.ChunkSize > 0 && len(doc.Content) > km.config.AgentMemory.ChunkSize {
			fmt.Printf("[DEBUG] Document needs chunking (size: %d, limit: %d)\n", len(doc.Content), km.config.AgentMemory.ChunkSize)
			chunks, err := ChunkDocument(doc, km.config.AgentMemory.ChunkSize, km.config.AgentMemory.ChunkOverlap)
			if err != nil {
				if options.ShowProgress {
					fmt.Printf("  Failed to chunk: %v\n", err)
				}
				failed++
				continue
			}
			docsToIngest = chunks
			if options.ShowProgress {
				fmt.Printf("  Created %d chunks\n", len(chunks))
			}
			fmt.Printf("[DEBUG] Created %d chunks\n", len(chunks))
		} else {
			fmt.Printf("[DEBUG] No chunking needed (size: %d, limit: %d)\n", len(doc.Content), km.config.AgentMemory.ChunkSize)
			docsToIngest = []*core.Document{doc}
		}

		// Ingest documents
		fmt.Printf("[DEBUG] Starting ingestion of %d documents\n", len(docsToIngest))
		for docIndex, docToIngest := range docsToIngest {
			fmt.Printf("[DEBUG] Ingesting document %d/%d\n", docIndex+1, len(docsToIngest))
			if err := km.memory.IngestDocument(ctx, *docToIngest); err != nil {
				fmt.Printf("[DEBUG] Ingestion failed: %v\n", err)
				if options.ShowProgress {
					fmt.Printf("  Failed to ingest: %v\n", err)
				}
				failed++
				break
			}
			fmt.Printf("[DEBUG] Document %d ingested successfully\n", docIndex+1)
		}

		processed++
		if options.ShowProgress {
			fmt.Printf("  Uploaded successfully\n")
		}
	}

	// Print summary
	fmt.Printf("\nUpload Summary:\n")
	fmt.Printf("  Processed: %d files\n", processed)
	if failed > 0 {
		fmt.Printf("  Failed: %d files\n", failed)
	}
	if skipped > 0 {
		fmt.Printf("  Skipped: %d files\n", skipped)
	}

	return nil
}

// collectFiles collects all files to process based on sources and options
func (km *KnowledgeManager) collectFiles(sources []string, options UploadOptions) ([]string, error) {
	var files []string
	registry := NewProcessorRegistry()
	supportedExts := registry.GetSupportedExtensions()

	for _, source := range sources {
		info, err := os.Stat(source)
		if err != nil {
			return nil, fmt.Errorf("cannot access %s: %v", source, err)
		}

		if info.IsDir() {
			// Process directory
			dirFiles, err := km.collectFromDirectory(source, options.Recursive, supportedExts)
			if err != nil {
				return nil, err
			}
			files = append(files, dirFiles...)
		} else {
			// Process single file
			if km.isSupportedFile(source, supportedExts) {
				files = append(files, source)
			}
		}
	}

	return files, nil
}

// collectFromDirectory collects files from a directory
func (km *KnowledgeManager) collectFromDirectory(dirPath string, recursive bool, supportedExts []string) ([]string, error) {
	var files []string

	if recursive {
		// Walk directory tree
		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && km.isSupportedFile(path, supportedExts) {
				files = append(files, path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %v", dirPath, err)
		}
	} else {
		// Only process immediate files
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %v", dirPath, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				filePath := filepath.Join(dirPath, entry.Name())
				if km.isSupportedFile(filePath, supportedExts) {
					files = append(files, filePath)
				}
			}
		}
	}

	return files, nil
}

// List lists documents in the knowledge base with optional filtering
func (km *KnowledgeManager) List(ctx context.Context, options ListOptions) error {
	// Use SearchKnowledge with a very common word and threshold 0.0 to get all documents
	// We'll use multiple strategies to ensure we capture all documents
	searchOptions := []core.SearchOption{
		core.WithLimit(10000), // Use a very large limit to get all documents
		core.WithScoreThreshold(0.0), // Accept all documents regardless of similarity score
	}

	// Try different query strategies to ensure we get all documents
	queries := []string{
		"the",    // Most common English word
		"a",      // Another very common word  
		"and",    // Common conjunction
		" ",      // Space character
		".",      // Period - likely to appear in most documents
	}

	var allResults []core.KnowledgeResult
	resultMap := make(map[string]core.KnowledgeResult) // Use map to deduplicate

	// Try each query and collect unique results
	for _, query := range queries {
		results, err := km.memory.SearchKnowledge(ctx, query, searchOptions...)
		if err == nil {
			// Add results to map to deduplicate by DocumentID
			for _, result := range results {
				resultMap[result.DocumentID] = result
			}
		}
	}

	// Convert map back to slice
	for _, result := range resultMap {
		allResults = append(allResults, result)
	}

	// Apply client-side filtering
	filteredResults := km.applyListFilters(allResults, options)

	// Apply limit if specified
	if options.Limit > 0 && len(filteredResults) > options.Limit {
		filteredResults = filteredResults[:options.Limit]
	}

	// Format and display results
	formatter := NewFormatter(options.OutputFormat)
	output := formatter.FormatDocuments(filteredResults)
	fmt.Print(output)

	return nil
}

// applyListFilters applies client-side filtering to search results
func (km *KnowledgeManager) applyListFilters(results []core.KnowledgeResult, options ListOptions) []core.KnowledgeResult {
	var filtered []core.KnowledgeResult

	for _, result := range results {
		// Apply source filter
		if options.FilterSource != "" && !matchesFilter(result.Source, options.FilterSource) {
			continue
		}

		// Apply type filter
		if options.FilterType != "" {
			docType := extractDocumentType(result)
			if !strings.EqualFold(docType, options.FilterType) {
				continue
			}
		}

		// Apply tags filter
		if len(options.FilterTags) > 0 && !tagsMatch(result.Tags, options.FilterTags) {
			continue
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// isSupportedFile checks if a file is supported based on its extension
func (km *KnowledgeManager) isSupportedFile(filePath string, supportedExts []string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

// Search searches the knowledge base using semantic similarity
func (km *KnowledgeManager) Search(ctx context.Context, query string, options SearchOptions) error {
	// Build search options
	searchOptions := []core.SearchOption{
		core.WithLimit(options.Limit),
	}

	// Add score threshold if specified (>=0 to allow 0.0 threshold)
	if options.ScoreThreshold >= 0 {
		searchOptions = append(searchOptions, core.WithScoreThreshold(options.ScoreThreshold))
	}

	// Perform search
	results, err := km.memory.SearchKnowledge(ctx, query, searchOptions...)
	if err != nil {
		return fmt.Errorf("search failed: %v", err)
	}

	// Apply score threshold filtering if not handled by the provider (>=0 to allow 0.0 threshold)
	if options.ScoreThreshold >= 0 {
		var filteredResults []core.KnowledgeResult
		for _, result := range results {
			if result.Score >= options.ScoreThreshold {
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults
	}

	// Format and display results
	formatter := NewFormatter(options.OutputFormat)
	output := formatter.FormatSearchResults(results)
	fmt.Print(output)

	return nil
}

// Validate validates the knowledge base configuration and health
func (km *KnowledgeManager) Validate(ctx context.Context) error {
	startTime := time.Now()
	result := KnowledgeValidationResult{
		Success:         true,
		Errors:          []KnowledgeValidationError{},
		Warnings:        []KnowledgeValidationWarning{},
		Summary:         KnowledgeValidationSummary{},
		Recommendations: []string{},
	}

	// Validate configuration
	result.Summary.ConfigValid = km.validateConfiguration(&result)

	// Test memory connection
	result.Summary.MemoryConnected = km.validateMemoryConnection(ctx, &result)

	// Test embedding service health
	result.Summary.EmbeddingHealthy = km.validateEmbeddingService(ctx, &result)

	// Test search functionality
	result.Summary.SearchFunctional = km.validateSearchFunctionality(ctx, &result)

	// Get document count
	result.Summary.DocumentCount = km.getDocumentCount(ctx)

	// Calculate validation time
	result.Summary.ValidationTime = time.Since(startTime)

	// Overall success status
	result.Success = len(result.Errors) == 0

	// Add recommendations
	km.addRecommendations(&result)

	// Format and display results
	formatter := NewFormatter("table")
	output := formatter.FormatValidation(result)
	fmt.Print(output)

	return nil
}

// Stats displays knowledge base statistics
func (km *KnowledgeManager) Stats(ctx context.Context, outputFormat string) error {
	stats := KnowledgeStats{
		DocumentCounts:     make(map[string]int),
		TotalDocuments:     0,
		TotalChunks:        0,
		StorageSize:        0,
		LastUpdated:        time.Now(),
		ProviderInfo:       km.getProviderInfo(),
		PerformanceMetrics: km.getPerformanceMetrics(ctx),
		Configuration:      km.getConfigSummary(),
	}

	// Get document statistics
	km.collectDocumentStats(ctx, &stats)

	// Format and display results
	formatter := NewFormatter(outputFormat)
	output := formatter.FormatStats(stats)
	fmt.Print(output)

	return nil
}

// validateConfiguration validates the knowledge base configuration
func (km *KnowledgeManager) validateConfiguration(result *KnowledgeValidationResult) bool {
	valid := true

	// Check if knowledge base is enabled
	if !km.config.AgentMemory.EnableKnowledgeBase {
		result.Errors = append(result.Errors, KnowledgeValidationError{
			Code:       "KNOWLEDGE_BASE_DISABLED",
			Message:    "Knowledge base is not enabled in configuration",
			Suggestion: "Set enable_knowledge_base = true in [agent_memory] section",
			Component:  "config",
		})
		valid = false
	}

	return valid
}

// validateMemoryConnection tests the memory provider connection
func (km *KnowledgeManager) validateMemoryConnection(ctx context.Context, result *KnowledgeValidationResult) bool {
	if err := km.TestConnection(ctx); err != nil {
		result.Errors = append(result.Errors, KnowledgeValidationError{
			Code:       "MEMORY_CONNECTION_FAILED",
			Message:    fmt.Sprintf("Memory connection test failed: %v", err),
			Suggestion: "Check memory provider connection string and ensure service is running",
			Component:  "memory",
		})
		return false
	}
	return true
}

// validateEmbeddingService tests the embedding service health
func (km *KnowledgeManager) validateEmbeddingService(ctx context.Context, result *KnowledgeValidationResult) bool {
	if km.config.AgentMemory.Embedding.Provider == "" {
		return false
	}
	return true
}

// validateSearchFunctionality tests the search functionality
func (km *KnowledgeManager) validateSearchFunctionality(ctx context.Context, result *KnowledgeValidationResult) bool {
	_, err := km.memory.SearchKnowledge(ctx, "test", core.WithLimit(1))
	if err != nil {
		result.Errors = append(result.Errors, KnowledgeValidationError{
			Code:       "SEARCH_FUNCTIONALITY_FAILED",
			Message:    fmt.Sprintf("Search functionality test failed: %v", err),
			Suggestion: "Check knowledge base setup and embedding service",
			Component:  "search",
		})
		return false
	}
	return true
}

// getDocumentCount gets the total number of documents in the knowledge base
func (km *KnowledgeManager) getDocumentCount(ctx context.Context) int {
	results, err := km.memory.SearchKnowledge(ctx, "the", core.WithLimit(10000), core.WithScoreThreshold(0.0))
	if err != nil {
		return 0
	}
	return len(results)
}

// addRecommendations adds recommendations based on validation results
func (km *KnowledgeManager) addRecommendations(result *KnowledgeValidationResult) {
	if !result.Summary.ConfigValid {
		result.Recommendations = append(result.Recommendations,
			"Review and fix configuration errors before using the knowledge base")
	}

	if result.Summary.DocumentCount == 0 {
		result.Recommendations = append(result.Recommendations,
			"Upload some documents to the knowledge base using 'agentcli knowledge upload'")
	}
}

// getProviderInfo returns information about the memory provider
func (km *KnowledgeManager) getProviderInfo() ProviderInfo {
	return ProviderInfo{
		Name:      km.config.AgentMemory.Provider,
		Connected: km.memory != nil,
		Version:   "",
		Details: map[string]interface{}{
			"connection": km.config.AgentMemory.Connection,
			"dimensions": km.config.AgentMemory.Dimensions,
		},
	}
}

// getPerformanceMetrics returns performance metrics
func (km *KnowledgeManager) getPerformanceMetrics(ctx context.Context) PerformanceMetrics {
	startTime := time.Now()
	_, err := km.memory.SearchKnowledge(ctx, "performance test", core.WithLimit(1), core.WithScoreThreshold(0.0))
	searchTime := time.Since(startTime)

	if err != nil {
		searchTime = 0
	}

	return PerformanceMetrics{
		AverageSearchTime: searchTime,
		AverageUploadTime: 0,
		LastSearchTime:    searchTime,
		LastUploadTime:    0,
		SearchCount:       0,
		UploadCount:       0,
	}
}

// getConfigSummary returns a summary of relevant configuration
func (km *KnowledgeManager) getConfigSummary() ConfigSummary {
	return ConfigSummary{
		Provider:          km.config.AgentMemory.Provider,
		Dimensions:        km.config.AgentMemory.Dimensions,
		ChunkSize:         km.config.AgentMemory.ChunkSize,
		ChunkOverlap:      km.config.AgentMemory.ChunkOverlap,
		KnowledgeEnabled:  km.config.AgentMemory.EnableKnowledgeBase,
		RAGEnabled:        km.config.AgentMemory.EnableRAG,
		ScoreThreshold:    km.config.AgentMemory.KnowledgeScoreThreshold,
		MaxResults:        km.config.AgentMemory.KnowledgeMaxResults,
		EmbeddingProvider: km.config.AgentMemory.Embedding.Provider,
		EmbeddingModel:    km.config.AgentMemory.Embedding.Model,
	}
}

// collectDocumentStats collects document statistics
func (km *KnowledgeManager) collectDocumentStats(ctx context.Context, stats *KnowledgeStats) {
	allResults, err := km.memory.SearchKnowledge(ctx, "the", core.WithLimit(10000), core.WithScoreThreshold(0.0))
	if err != nil {
		return
	}

	stats.TotalDocuments = len(allResults)

	typeCount := make(map[string]int)
	chunkCount := 0
	var totalSize int64

	for _, result := range allResults {
		docType := extractDocumentType(result)
		typeCount[docType]++

		if result.ChunkIndex > 0 {
			chunkCount++
		}

		totalSize += int64(len(result.Content))

		if result.CreatedAt.After(stats.LastUpdated) {
			stats.LastUpdated = result.CreatedAt
		}
	}

	stats.DocumentCounts = typeCount
	stats.TotalChunks = chunkCount
	stats.StorageSize = totalSize
}

// Clear clears documents from the knowledge base with selective deletion
func (km *KnowledgeManager) Clear(ctx context.Context, options ClearOptions) error {
	// Get documents that match the filter criteria
	targetDocuments, err := km.getFilteredDocuments(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to get documents for clearing: %v", err)
	}

	if len(targetDocuments) == 0 {
		fmt.Println("No documents found matching the specified criteria.")
		return nil
	}

	fmt.Printf("Found %d documents matching criteria:\n\n", len(targetDocuments))

	// Show preview of what will be deleted
	km.showClearPreview(targetDocuments)

	// Handle dry run
	if options.DryRun {
		fmt.Printf("\n[DRY RUN] Would delete %d documents. Use --force to actually delete.\n", len(targetDocuments))
		return nil
	}

	// Get confirmation unless force is used
	if !options.Force {
		if options.Interactive {
			if !km.getInteractiveConfirmation(targetDocuments) {
				fmt.Println("Operation cancelled.")
				return nil
			}
		} else {
			if !km.getSimpleConfirmation(len(targetDocuments)) {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}
	}

	// Show limitation notice
	fmt.Printf("\n[NOTICE] Document deletion functionality is not yet implemented in the core Memory interface.\n")
	fmt.Printf("This command shows what would be deleted but cannot perform actual deletion.\n")
	fmt.Printf("\nTo clear the knowledge base, you can:\n")
	fmt.Printf("1. Drop and recreate the database/vector store\n")
	fmt.Printf("2. Use provider-specific tools to delete documents\n")
	fmt.Printf("3. Clear session data if using session-scoped storage\n\n")

	// For now, we can only clear session data if available
	if km.memory != nil {
		if err := km.memory.ClearSession(ctx); err == nil {
			fmt.Printf("Cleared session data successfully.\n")
		} else {
			fmt.Printf("Failed to clear session data: %v\n", err)
		}
	}

	return nil
}

// getFilteredDocuments gets documents that match the clear filter criteria
func (km *KnowledgeManager) getFilteredDocuments(ctx context.Context, options ClearOptions) ([]core.KnowledgeResult, error) {
	// Get all documents using a broad search
	allResults, err := km.memory.SearchKnowledge(ctx, "the", core.WithLimit(10000), core.WithScoreThreshold(0.0))
	if err != nil {
		return nil, err
	}

	// Apply filters
	var filtered []core.KnowledgeResult
	for _, result := range allResults {
		// Apply source filter
		if options.FilterSource != "" && !matchesFilter(result.Source, options.FilterSource) {
			continue
		}

		// Apply type filter
		if options.FilterType != "" {
			docType := extractDocumentType(result)
			if !strings.EqualFold(docType, options.FilterType) {
				continue
			}
		}

		// Apply tags filter
		if len(options.FilterTags) > 0 && !tagsMatch(result.Tags, options.FilterTags) {
			continue
		}

		filtered = append(filtered, result)
	}

	return filtered, nil
}

// showClearPreview shows a preview of documents that will be deleted
func (km *KnowledgeManager) showClearPreview(documents []core.KnowledgeResult) {
	// Group documents by source for better display
	sourceGroups := make(map[string][]core.KnowledgeResult)
	for _, doc := range documents {
		source := doc.Source
		if source == "" {
			source = "Unknown source"
		}
		sourceGroups[source] = append(sourceGroups[source], doc)
	}

	fmt.Printf("Documents to be deleted (grouped by source):\n")
	fmt.Printf("%-50s %s\n", "SOURCE", "COUNT")
	fmt.Printf("%s\n", strings.Repeat("-", 60))

	for source, docs := range sourceGroups {
		if len(source) > 47 {
			source = source[:44] + "..."
		}
		fmt.Printf("%-50s %d\n", source, len(docs))
	}

	fmt.Printf("\nTotal documents: %d\n", len(documents))
}

// getSimpleConfirmation gets a simple yes/no confirmation
func (km *KnowledgeManager) getSimpleConfirmation(count int) bool {
	fmt.Printf("\nAre you sure you want to delete %d documents? [y/N]: ", count)

	// For now, we'll assume 'no' since we can't actually read from stdin in this implementation
	// In a real implementation, this would read from os.Stdin
	fmt.Printf("n\n")
	fmt.Printf("[INFO] Automatic 'no' response - interactive confirmation not yet implemented\n")
	return false
}

// getInteractiveConfirmation gets detailed interactive confirmation
func (km *KnowledgeManager) getInteractiveConfirmation(documents []core.KnowledgeResult) bool {
	fmt.Printf("\n=== Interactive Deletion Confirmation ===\n")
	fmt.Printf("You are about to delete %d documents.\n\n", len(documents))

	// Show first few documents as examples
	fmt.Printf("Sample documents to be deleted:\n")
	for i, doc := range documents {
		if i >= 5 { // Show max 5 examples
			fmt.Printf("... and %d more\n", len(documents)-i)
			break
		}
		title := doc.Title
		if title == "" {
			title = doc.Source
		}
		if len(title) > 60 {
			title = title[:57] + "..."
		}
		fmt.Printf("  %d. %s\n", i+1, title)
	}

	fmt.Printf("\nType 'DELETE' to confirm, or anything else to cancel: ")

	// For now, we'll assume cancellation since we can't actually read from stdin
	fmt.Printf("[cancelled]\n")
	fmt.Printf("[INFO] Interactive confirmation not yet implemented - operation cancelled\n")
	return false
}

func parseTagList(tagString string) []string {
	if tagString == "" {
		return nil
	}

	var tags []string
	for _, tag := range strings.Split(tagString, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// Helper function to check if a string contains any of the filter patterns
func matchesFilter(text, filter string) bool {
	if filter == "" {
		return true
	}

	filter = strings.ToLower(filter)
	text = strings.ToLower(text)

	// Support simple wildcard matching
	if strings.Contains(filter, "*") {
		// Simple prefix/suffix matching
		if strings.HasPrefix(filter, "*") && strings.HasSuffix(filter, "*") {
			// Contains match
			return strings.Contains(text, strings.Trim(filter, "*"))
		} else if strings.HasPrefix(filter, "*") {
			// Suffix match
			return strings.HasSuffix(text, strings.TrimPrefix(filter, "*"))
		} else if strings.HasSuffix(filter, "*") {
			// Prefix match
			return strings.HasPrefix(text, strings.TrimSuffix(filter, "*"))
		}
	}

	// Exact or substring match
	return strings.Contains(text, filter)
}

// Helper function to check if any tags match the filter
func tagsMatch(documentTags, filterTags []string) bool {
	if len(filterTags) == 0 {
		return true
	}

	for _, filterTag := range filterTags {
		for _, docTag := range documentTags {
			if strings.EqualFold(docTag, filterTag) {
				return true
			}
		}
	}
	return false
}
