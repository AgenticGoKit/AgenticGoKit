package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/ledongthuc/pdf"
)

// PDFProcessor handles PDF files using a lightweight extraction library.
type PDFProcessor struct{}

func (pp *PDFProcessor) CanProcess(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".pdf"
}

func (pp *PDFProcessor) GetSupportedExtensions() []string {
	return []string{".pdf"}
}

func (pp *PDFProcessor) Process(ctx context.Context, filePath string, options ProcessingOptions) (*core.Document, error) {
	// Ensure file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %v", filePath, err)
	}

	// Open PDF
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pdf %s: %v", filePath, err)
	}
	defer f.Close()

	// Try to extract plain text
	reader, err := r.GetPlainText()
	if err != nil {
		return nil, fmt.Errorf("failed to extract text from pdf %s: %v", filePath, err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, fmt.Errorf("failed to read extracted text for %s: %v", filePath, err)
	}

	content := buf.String()

	// Build metadata
	metadata := map[string]any{
		"file_name":  filepath.Base(filePath),
		"file_path":  filePath,
		"processor":  "pdf",
		"page_count": r.NumPage(),
	}

	// If IncludeMetadata is requested, add file stats
	if options.IncludeMetadata {
		if stat, err := os.Stat(filePath); err == nil {
			metadata["file_size"] = stat.Size()
			metadata["modified_time"] = stat.ModTime()
		}
	}

	title := extractTitleFromPath(filePath)

	doc := &core.Document{
		ID:        generateDocumentID(filePath, []byte(content)),
		Title:     title,
		Content:   content,
		Source:    options.Source,
		Type:      core.DocumentTypePDF,
		Metadata:  metadata,
		Tags:      options.Tags,
		CreatedAt: time.Now(),
	}

	if doc.Source == "" {
		doc.Source = filePath
	}

	return doc, nil
}

func (pp *PDFProcessor) ExtractMetadata(filePath string) (map[string]any, error) {
	// Basic metadata (file stat + page count if available)
	md := map[string]any{}
	if stat, err := os.Stat(filePath); err == nil {
		md["file_name"] = filepath.Base(filePath)
		md["file_path"] = filePath
		md["file_size"] = stat.Size()
		md["modified_time"] = stat.ModTime()
		md["processor"] = "pdf"
	} else {
		return nil, err
	}

	// Try to open PDF for page count
	f, r, err := pdf.Open(filePath)
	if err == nil {
		md["page_count"] = r.NumPage()
		f.Close()
	}

	return md, nil
}

