package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// OutputFormatter interface for formatting knowledge base output
type OutputFormatter interface {
	FormatDocuments(docs []core.KnowledgeResult) string
	FormatSearchResults(results []core.KnowledgeResult) string
	FormatStats(stats KnowledgeStats) string
	FormatValidation(result KnowledgeValidationResult) string
}

// TableFormatter formats output as readable tables
type TableFormatter struct{}

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

// NewFormatter creates the appropriate formatter based on format type
func NewFormatter(format string) OutputFormatter {
	switch strings.ToLower(format) {
	case "json":
		return &JSONFormatter{}
	default:
		return &TableFormatter{}
	}
}

// TableFormatter implementations

func (tf *TableFormatter) FormatDocuments(docs []core.KnowledgeResult) string {
	if len(docs) == 0 {
		return "No documents found."
	}

	var output strings.Builder

	// Header
	output.WriteString(fmt.Sprintf("Found %d documents:\n\n", len(docs)))
	output.WriteString(fmt.Sprintf("%-40s %-15s %-12s %-20s %s\n", "TITLE/SOURCE", "TYPE", "CHUNKS", "UPDATED", "TAGS"))
	output.WriteString(strings.Repeat("-", 100) + "\n")

	// Documents
	for _, doc := range docs {
		title := doc.Title
		if title == "" {
			title = doc.Source
		}
		if len(title) > 37 {
			title = title[:34] + "..."
		}

		// Extract document type from source or metadata
		docType := extractDocumentType(doc)

		// Format chunks info
		chunksInfo := ""
		if doc.ChunkIndex > 0 {
			chunksInfo = fmt.Sprintf("%d/%d", doc.ChunkIndex, doc.Metadata["chunk_total"])
		} else {
			chunksInfo = "1/1"
		}

		// Format update time
		updated := doc.CreatedAt.Format("Jan 02, 15:04")

		// Format tags
		tagsStr := strings.Join(doc.Tags, ",")
		if len(tagsStr) > 20 {
			tagsStr = tagsStr[:17] + "..."
		}

		output.WriteString(fmt.Sprintf("%-40s %-15s %-12s %-20s %s\n",
			title, docType, chunksInfo, updated, tagsStr))
	}

	return output.String()
}

func (tf *TableFormatter) FormatSearchResults(results []core.KnowledgeResult) string {
	if len(results) == 0 {
		return "No search results found."
	}

	var output strings.Builder

	// Header
	output.WriteString(fmt.Sprintf("Found %d search results:\n\n", len(results)))
	output.WriteString(fmt.Sprintf("%-6s %-40s %-50s %s\n", "SCORE", "SOURCE", "CONTENT", "TAGS"))
	output.WriteString(strings.Repeat("-", 120) + "\n")

	// Results
	for _, result := range results {
		score := fmt.Sprintf("%.3f", result.Score)

		source := result.Source
		if len(source) > 37 {
			source = source[:34] + "..."
		}

		content := strings.ReplaceAll(result.Content, "\n", " ")
		if len(content) > 47 {
			content = content[:44] + "..."
		}

		tagsStr := strings.Join(result.Tags, ",")
		if len(tagsStr) > 15 {
			tagsStr = tagsStr[:12] + "..."
		}

		output.WriteString(fmt.Sprintf("%-6s %-40s %-50s %s\n",
			score, source, content, tagsStr))
	}

	return output.String()
}

func (tf *TableFormatter) FormatStats(stats KnowledgeStats) string {
	var output strings.Builder

	output.WriteString("Knowledge Base Statistics\n")
	output.WriteString("========================\n\n")

	// Overview
	output.WriteString("Overview:\n")
	output.WriteString(fmt.Sprintf("  Total Documents: %d\n", stats.TotalDocuments))
	output.WriteString(fmt.Sprintf("  Total Chunks: %d\n", stats.TotalChunks))
	output.WriteString(fmt.Sprintf("  Storage Size: %s\n", formatBytes(stats.StorageSize)))
	output.WriteString(fmt.Sprintf("  Last Updated: %s\n", stats.LastUpdated.Format("2006-01-02 15:04:05")))
	output.WriteString("\n")

	// Document types
	if len(stats.DocumentCounts) > 0 {
		output.WriteString("Document Types:\n")
		for docType, count := range stats.DocumentCounts {
			output.WriteString(fmt.Sprintf("  %s: %d\n", docType, count))
		}
		output.WriteString("\n")
	}

	// Provider info
	output.WriteString("Provider Information:\n")
	output.WriteString(fmt.Sprintf("  Name: %s\n", stats.ProviderInfo.Name))
	output.WriteString(fmt.Sprintf("  Connected: %t\n", stats.ProviderInfo.Connected))
	if stats.ProviderInfo.Version != "" {
		output.WriteString(fmt.Sprintf("  Version: %s\n", stats.ProviderInfo.Version))
	}
	output.WriteString("\n")

	// Configuration
	output.WriteString("Configuration:\n")
	output.WriteString(fmt.Sprintf("  Provider: %s\n", stats.Configuration.Provider))
	output.WriteString(fmt.Sprintf("  Dimensions: %d\n", stats.Configuration.Dimensions))
	output.WriteString(fmt.Sprintf("  Chunk Size: %d\n", stats.Configuration.ChunkSize))
	output.WriteString(fmt.Sprintf("  Chunk Overlap: %d\n", stats.Configuration.ChunkOverlap))
	output.WriteString(fmt.Sprintf("  Knowledge Enabled: %t\n", stats.Configuration.KnowledgeEnabled))
	output.WriteString(fmt.Sprintf("  RAG Enabled: %t\n", stats.Configuration.RAGEnabled))
	output.WriteString(fmt.Sprintf("  Score Threshold: %.2f\n", stats.Configuration.ScoreThreshold))
	output.WriteString(fmt.Sprintf("  Embedding: %s/%s\n", stats.Configuration.EmbeddingProvider, stats.Configuration.EmbeddingModel))
	output.WriteString("\n")

	// Performance metrics
	output.WriteString("Performance Metrics:\n")
	output.WriteString(fmt.Sprintf("  Average Search Time: %v\n", stats.PerformanceMetrics.AverageSearchTime))
	output.WriteString(fmt.Sprintf("  Average Upload Time: %v\n", stats.PerformanceMetrics.AverageUploadTime))
	output.WriteString(fmt.Sprintf("  Total Searches: %d\n", stats.PerformanceMetrics.SearchCount))
	output.WriteString(fmt.Sprintf("  Total Uploads: %d\n", stats.PerformanceMetrics.UploadCount))

	return output.String()
}

func (tf *TableFormatter) FormatValidation(result KnowledgeValidationResult) string {
	var output strings.Builder

	output.WriteString("Knowledge Base Validation Results\n")
	output.WriteString("=================================\n\n")

	// Overall status
	if result.Success {
		output.WriteString("✓ Validation PASSED\n\n")
	} else {
		output.WriteString("✗ Validation FAILED\n\n")
	}

	// Summary
	output.WriteString("Validation Summary:\n")
	output.WriteString(fmt.Sprintf("  Configuration Valid: %s\n", checkMark(result.Summary.ConfigValid)))
	output.WriteString(fmt.Sprintf("  Memory Connected: %s\n", checkMark(result.Summary.MemoryConnected)))
	output.WriteString(fmt.Sprintf("  Embedding Healthy: %s\n", checkMark(result.Summary.EmbeddingHealthy)))
	output.WriteString(fmt.Sprintf("  Search Functional: %s\n", checkMark(result.Summary.SearchFunctional)))
	output.WriteString(fmt.Sprintf("  Document Count: %d\n", result.Summary.DocumentCount))
	output.WriteString(fmt.Sprintf("  Validation Time: %v\n", result.Summary.ValidationTime))
	output.WriteString("\n")

	// Errors
	if len(result.Errors) > 0 {
		output.WriteString("Errors:\n")
		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("  ✗ [%s] %s: %s\n", err.Component, err.Code, err.Message))
			if err.Suggestion != "" {
				output.WriteString(fmt.Sprintf("    Suggestion: %s\n", err.Suggestion))
			}
		}
		output.WriteString("\n")
	}

	// Warnings
	if len(result.Warnings) > 0 {
		output.WriteString("Warnings:\n")
		for _, warn := range result.Warnings {
			output.WriteString(fmt.Sprintf("  ⚠ [%s] %s: %s\n", warn.Component, warn.Code, warn.Message))
		}
		output.WriteString("\n")
	}

	// Recommendations
	if len(result.Recommendations) > 0 {
		output.WriteString("Recommendations:\n")
		for _, rec := range result.Recommendations {
			output.WriteString(fmt.Sprintf("  • %s\n", rec))
		}
	}

	return output.String()
}

// JSONFormatter implementations

func (jf *JSONFormatter) FormatDocuments(docs []core.KnowledgeResult) string {
	data := map[string]interface{}{
		"documents": docs,
		"count":     len(docs),
		"timestamp": time.Now(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "Failed to marshal JSON: %v"}`, err)
	}

	return string(jsonData)
}

func (jf *JSONFormatter) FormatSearchResults(results []core.KnowledgeResult) string {
	data := map[string]interface{}{
		"results":   results,
		"count":     len(results),
		"timestamp": time.Now(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "Failed to marshal JSON: %v"}`, err)
	}

	return string(jsonData)
}

func (jf *JSONFormatter) FormatStats(stats KnowledgeStats) string {
	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "Failed to marshal JSON: %v"}`, err)
	}

	return string(jsonData)
}

func (jf *JSONFormatter) FormatValidation(result KnowledgeValidationResult) string {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "Failed to marshal JSON: %v"}`, err)
	}

	return string(jsonData)
}

// Helper functions

func extractDocumentType(doc core.KnowledgeResult) string {
	// Try to get type from metadata first
	if docType, ok := doc.Metadata["type"].(string); ok {
		return strings.ToUpper(docType)
	}

	// Extract from source file extension
	source := doc.Source
	if lastDot := strings.LastIndex(source, "."); lastDot != -1 {
		ext := strings.ToLower(source[lastDot+1:])
		switch ext {
		case "pdf":
			return "PDF"
		case "md", "markdown":
			return "MD"
		case "txt":
			return "TXT"
		case "html", "htm":
			return "HTML"
		case "go":
			return "GO"
		case "py":
			return "PYTHON"
		case "js":
			return "JS"
		case "java":
			return "JAVA"
		default:
			return strings.ToUpper(ext)
		}
	}

	return "UNKNOWN"
}

func checkMark(success bool) string {
	if success {
		return "✓"
	}
	return "✗"
}
