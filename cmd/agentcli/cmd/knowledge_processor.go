package cmd

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// DocumentProcessor interface defines how different document types are processed
type DocumentProcessor interface {
	CanProcess(filename string) bool
	Process(ctx context.Context, filepath string, options ProcessingOptions) (*core.Document, error)
	ExtractMetadata(filepath string) (map[string]any, error)
	GetSupportedExtensions() []string
}

// ProcessingOptions contains options for document processing
type ProcessingOptions struct {
	ChunkSize         int
	ChunkOverlap      int
	PreserveStructure bool
	ExtractHeaders    bool
	IncludeMetadata   bool
	Tags              []string
	Source            string // Custom source identifier
}

// ProcessorRegistry manages document processors
type ProcessorRegistry struct {
	processors []DocumentProcessor
}

// NewProcessorRegistry creates a new processor registry with default processors
func NewProcessorRegistry() *ProcessorRegistry {
	registry := &ProcessorRegistry{}

	// Register default processors
	registry.Register(&TextProcessor{})
	registry.Register(&MarkdownProcessor{})
	// PDF support (MVP)
	registry.Register(&PDFProcessor{})

	return registry
}

// Register adds a new processor to the registry
func (pr *ProcessorRegistry) Register(processor DocumentProcessor) {
	pr.processors = append(pr.processors, processor)
}

// GetProcessor returns the appropriate processor for a file
func (pr *ProcessorRegistry) GetProcessor(filename string) (DocumentProcessor, error) {
	for _, processor := range pr.processors {
		if processor.CanProcess(filename) {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("no processor found for file: %s", filename)
}

// GetSupportedExtensions returns all supported file extensions
func (pr *ProcessorRegistry) GetSupportedExtensions() []string {
	var extensions []string
	for _, processor := range pr.processors {
		extensions = append(extensions, processor.GetSupportedExtensions()...)
	}
	return extensions
}

// =============================================================================
// TEXT PROCESSOR
// =============================================================================

// TextProcessor handles plain text files
type TextProcessor struct{}

func (tp *TextProcessor) CanProcess(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".txt" || ext == ".text"
}

func (tp *TextProcessor) GetSupportedExtensions() []string {
	return []string{".txt", ".text"}
}

func (tp *TextProcessor) Process(ctx context.Context, filePath string, options ProcessingOptions) (*core.Document, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// Extract metadata if requested
	var metadata map[string]any
	if options.IncludeMetadata {
		metadata, err = tp.ExtractMetadata(filePath)
		if err != nil {
			// Don't fail the whole operation for metadata extraction errors
			metadata = map[string]any{"metadata_error": err.Error()}
		}
	}

	// Create document
	doc := &core.Document{
		ID:        generateDocumentID(filePath, content),
		Title:     extractTitleFromPath(filePath),
		Content:   string(content),
		Source:    options.Source,
		Type:      core.DocumentTypeText,
		Metadata:  metadata,
		Tags:      options.Tags,
		CreatedAt: time.Now(),
	}

	// Set source to file path if not provided
	if doc.Source == "" {
		doc.Source = filePath
	}

	return doc, nil
}

func (tp *TextProcessor) ExtractMetadata(filePath string) (map[string]any, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	metadata := map[string]any{
		"file_name":      filepath.Base(filePath),
		"file_path":      filePath,
		"file_size":      stat.Size(),
		"modified_time":  stat.ModTime(),
		"file_extension": filepath.Ext(filePath),
		"processor":      "text",
	}

	return metadata, nil
}

// =============================================================================
// MARKDOWN PROCESSOR
// =============================================================================

// MarkdownProcessor handles Markdown files with structure preservation
type MarkdownProcessor struct{}

func (mp *MarkdownProcessor) CanProcess(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".md" || ext == ".markdown"
}

func (mp *MarkdownProcessor) GetSupportedExtensions() []string {
	return []string{".md", ".markdown"}
}

func (mp *MarkdownProcessor) Process(ctx context.Context, filePath string, options ProcessingOptions) (*core.Document, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	contentStr := string(content)

	// Extract title from markdown
	title := mp.extractMarkdownTitle(contentStr)
	if title == "" {
		title = extractTitleFromPath(filePath)
	}

	// Extract metadata if requested
	var metadata map[string]any
	if options.IncludeMetadata {
		metadata, err = mp.ExtractMetadata(filePath)
		if err != nil {
			metadata = map[string]any{"metadata_error": err.Error()}
		}

		// Add markdown-specific metadata
		if options.ExtractHeaders {
			headers := mp.extractHeaders(contentStr)
			metadata["headers"] = headers
		}

		metadata["word_count"] = len(strings.Fields(contentStr))
		metadata["line_count"] = len(strings.Split(contentStr, "\n"))
	}

	// Create document
	doc := &core.Document{
		ID:        generateDocumentID(filePath, content),
		Title:     title,
		Content:   contentStr,
		Source:    options.Source,
		Type:      core.DocumentTypeMarkdown,
		Metadata:  metadata,
		Tags:      options.Tags,
		CreatedAt: time.Now(),
	}

	// Set source to file path if not provided
	if doc.Source == "" {
		doc.Source = filePath
	}

	return doc, nil
}

func (mp *MarkdownProcessor) ExtractMetadata(filePath string) (map[string]any, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	metadata := map[string]any{
		"file_name":      filepath.Base(filePath),
		"file_path":      filePath,
		"file_size":      stat.Size(),
		"modified_time":  stat.ModTime(),
		"file_extension": filepath.Ext(filePath),
		"processor":      "markdown",
	}

	return metadata, nil
}

// extractMarkdownTitle attempts to extract the title from markdown content
func (mp *MarkdownProcessor) extractMarkdownTitle(content string) string {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for H1 headers
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
		// Stop looking after a few lines if no title found
		if line != "" && !strings.HasPrefix(line, "#") {
			break
		}
	}

	return ""
}

// extractHeaders extracts all headers from markdown content
func (mp *MarkdownProcessor) extractHeaders(content string) []map[string]interface{} {
	var headers []map[string]interface{}
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Count the number of # to determine header level
			level := 0
			for _, char := range line {
				if char == '#' {
					level++
				} else {
					break
				}
			}

			if level > 0 && level <= 6 {
				headerText := strings.TrimSpace(line[level:])
				headers = append(headers, map[string]interface{}{
					"level": level,
					"text":  headerText,
					"line":  i + 1,
				})
			}
		}
	}

	return headers
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// generateDocumentID creates a unique ID for a document based on file path and content
func generateDocumentID(filePath string, content []byte) string {
	hasher := md5.New()
	hasher.Write([]byte(filePath))
	hasher.Write(content)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// extractTitleFromPath creates a title from the file path
func extractTitleFromPath(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	title := strings.TrimSuffix(base, ext)

	// Replace underscores and hyphens with spaces
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")

	// Capitalize first letter
	if len(title) > 0 {
		title = strings.ToUpper(title[:1]) + title[1:]
	}

	return title
}

// CreateDocumentFromFile is a convenience function to process a file
func CreateDocumentFromFile(filePath string, options ProcessingOptions) (*core.Document, error) {
	registry := NewProcessorRegistry()
	processor, err := registry.GetProcessor(filePath)
	if err != nil {
		return nil, err
	}

	return processor.Process(context.Background(), filePath, options)
}

// ChunkDocument splits a document into smaller chunks for better embedding
func ChunkDocument(doc *core.Document, chunkSize, chunkOverlap int) ([]*core.Document, error) {
	fmt.Printf("[DEBUG CHUNK] Starting chunking: doc length=%d, chunkSize=%d, chunkOverlap=%d\n", len(doc.Content), chunkSize, chunkOverlap)

	if chunkSize <= 0 {
		return []*core.Document{doc}, nil
	}

	content := doc.Content
	if len(content) <= chunkSize {
		return []*core.Document{doc}, nil
	}

	var chunks []*core.Document
	start := 0
	chunkIndex := 1

	for start < len(content) {
		fmt.Printf("[DEBUG CHUNK] Loop %d: start=%d, content remaining=%d\n", chunkIndex, start, len(content)-start)

		end := start + chunkSize
		if end > len(content) {
			end = len(content)
		}

		// Try to break at word boundaries
		if end < len(content) {
			// Look backwards for a space
			for i := end; i > start && i > end-100; i-- {
				if content[i] == ' ' || content[i] == '\n' {
					end = i
					break
				}
			}
		}

		chunkContent := content[start:end]
		fmt.Printf("[DEBUG CHUNK] Chunk %d: start=%d, end=%d, length=%d, content='%s'\n", chunkIndex, start, end, len(chunkContent), chunkContent[:min(len(chunkContent), 50)])

		// Create chunk document
		chunk := &core.Document{
			ID:         fmt.Sprintf("%s_chunk_%d", doc.ID, chunkIndex),
			Title:      fmt.Sprintf("%s (Part %d)", doc.Title, chunkIndex),
			Content:    chunkContent,
			Source:     doc.Source,
			Type:       doc.Type,
			Metadata:   copyMetadata(doc.Metadata),
			Tags:       doc.Tags,
			CreatedAt:  doc.CreatedAt,
			ChunkIndex: chunkIndex,
		}

		// Add chunk metadata
		if chunk.Metadata == nil {
			chunk.Metadata = make(map[string]any)
		}
		chunk.Metadata["is_chunk"] = true
		chunk.Metadata["chunk_index"] = chunkIndex
		chunk.Metadata["original_document_id"] = doc.ID

		chunks = append(chunks, chunk)

		// Move start position with overlap
		newStart := end - chunkOverlap
		fmt.Printf("[DEBUG CHUNK] Calculating newStart: end=%d - chunkOverlap=%d = %d\n", end, chunkOverlap, newStart)

		if newStart <= start {
			// Ensure we always advance to prevent infinite loops
			// But advance by a reasonable amount, not just 1 character
			newStart = start + (chunkSize / 2) // Advance by half chunk size
			fmt.Printf("[DEBUG CHUNK] newStart <= start, advancing by chunkSize/2: newStart=%d\n", newStart)
			if newStart >= len(content) {
				fmt.Printf("[DEBUG CHUNK] newStart >= content length, breaking\n")
				break // We've reached the end
			}
		}
		start = newStart
		chunkIndex++

		if chunkIndex > 110 {
			fmt.Printf("[DEBUG CHUNK] Safety break at chunk %d to prevent runaway\n", chunkIndex)
			break
		}
	}

	// Update chunk total in all chunks
	for _, chunk := range chunks {
		chunk.ChunkTotal = len(chunks)
		chunk.Metadata["chunk_total"] = len(chunks)
	}

	fmt.Printf("[DEBUG CHUNK] Completed chunking: created %d chunks\n", len(chunks))
	return chunks, nil
}

// copyMetadata creates a copy of metadata map
func copyMetadata(original map[string]any) map[string]any {
	if original == nil {
		return nil
	}

	copy := make(map[string]any)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
