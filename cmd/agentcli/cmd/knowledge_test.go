package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data
const testConfigContent = `
[agent_flow]
name = "test-project"
version = "0.1.0"
provider = "memory"

[agent_memory]
provider = "memory"
connection = "memory"
enable_knowledge_base = true
enable_rag = true
chunk_size = 1000
chunk_overlap = 200

[agent_memory.embedding]
provider = "dummy"
model = "dummy-model"

[llm]
provider = "dummy"
model = "dummy-model"
`

const testMarkdownContent = `# Test Document

This is a test markdown document for testing the knowledge base upload functionality.

## Section 1

This section contains some test content.

## Section 2

This section contains more test content with **bold** and *italic* text.
`

const testTextContent = `This is a simple text document for testing.

It contains multiple lines and paragraphs to test the text processor.
The content should be chunked properly when uploaded to the knowledge base.
`

// Test helper functions

func createTestConfig(t *testing.T) string {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "agentflow.toml")

	err := os.WriteFile(configPath, []byte(testConfigContent), 0644)
	require.NoError(t, err)

	return configPath
}

func createTestFiles(t *testing.T) (string, []string) {
	tempDir := t.TempDir()

	// Create test markdown file
	mdFile := filepath.Join(tempDir, "test.md")
	err := os.WriteFile(mdFile, []byte(testMarkdownContent), 0644)
	require.NoError(t, err)

	// Create test text file
	txtFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(txtFile, []byte(testTextContent), 0644)
	require.NoError(t, err)

	// Create subdirectory with another file
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	subFile := filepath.Join(subDir, "sub.md")
	err = os.WriteFile(subFile, []byte("# Sub Document\n\nThis is in a subdirectory."), 0644)
	require.NoError(t, err)

	return tempDir, []string{mdFile, txtFile, subFile}
}

// Unit Tests

func TestNewKnowledgeManager(t *testing.T) {
	configPath := createTestConfig(t)

	t.Run("Valid Configuration", func(t *testing.T) {
		km, err := NewKnowledgeManager(configPath)
		require.NoError(t, err)
		assert.NotNil(t, km)
		assert.Equal(t, configPath, km.GetConfigPath())
		assert.NotNil(t, km.GetConfig())
	})

	t.Run("Missing Configuration", func(t *testing.T) {
		_, err := NewKnowledgeManager("nonexistent.toml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "No agentflow.toml found")
	})

	t.Run("Default Configuration Path", func(t *testing.T) {
		// Change to temp directory and create config
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		tempDir := t.TempDir()
		err = os.Chdir(tempDir)
		require.NoError(t, err)

		err = os.WriteFile("agentflow.toml", []byte(testConfigContent), 0644)
		require.NoError(t, err)

		km, err := NewKnowledgeManager("")
		require.NoError(t, err)
		assert.NotNil(t, km)
	})
}

func TestDocumentProcessor(t *testing.T) {
	registry := NewProcessorRegistry()

	t.Run("Text Processor", func(t *testing.T) {
		processor, err := registry.GetProcessor("test.txt")
		require.NoError(t, err)
		assert.NotNil(t, processor)

		// Test if it can process text files
		assert.True(t, processor.CanProcess("test.txt"))
		assert.True(t, processor.CanProcess("test.text"))
		assert.False(t, processor.CanProcess("test.md"))
	})

	t.Run("Markdown Processor", func(t *testing.T) {
		processor, err := registry.GetProcessor("test.md")
		require.NoError(t, err)
		assert.NotNil(t, processor)

		// Test if it can process markdown files
		assert.True(t, processor.CanProcess("test.md"))
		assert.True(t, processor.CanProcess("test.markdown"))
		assert.False(t, processor.CanProcess("test.txt"))
	})

	t.Run("Unsupported File Type", func(t *testing.T) {
		_, err := registry.GetProcessor("test.xyz")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no processor found")
	})
}

func TestDocumentProcessing(t *testing.T) {
	_, testFiles := createTestFiles(t)

	t.Run("Process Markdown File", func(t *testing.T) {
		options := ProcessingOptions{
			ChunkSize:       1000,
			ChunkOverlap:    200,
			IncludeMetadata: true,
			Tags:            []string{"test", "markdown"},
		}

		doc, err := CreateDocumentFromFile(testFiles[0], options)
		require.NoError(t, err)
		assert.NotNil(t, doc)

		assert.Equal(t, "Test Document", doc.Title)
		assert.Equal(t, core.DocumentTypeMarkdown, doc.Type)
		assert.Contains(t, doc.Content, "# Test Document")
		assert.Equal(t, []string{"test", "markdown"}, doc.Tags)
		assert.NotNil(t, doc.Metadata)
		assert.Contains(t, doc.Metadata, "file_name")
	})

	t.Run("Process Text File", func(t *testing.T) {
		options := ProcessingOptions{
			ChunkSize:       500,
			ChunkOverlap:    100,
			IncludeMetadata: false,
		}

		doc, err := CreateDocumentFromFile(testFiles[1], options)
		require.NoError(t, err)
		assert.NotNil(t, doc)

		assert.Equal(t, core.DocumentTypeText, doc.Type)
		assert.Contains(t, doc.Content, "simple text document")
		assert.Nil(t, doc.Metadata)
	})
}

func TestChunkDocument(t *testing.T) {
	// Create a large document for chunking
	largeContent := ""
	for i := 0; i < 100; i++ {
		largeContent += "This is sentence number " + string(rune(i)) + ". "
	}

	doc := &core.Document{
		ID:        "test-doc",
		Title:     "Test Document",
		Content:   largeContent,
		Source:    "test.txt",
		Type:      core.DocumentTypeText,
		CreatedAt: time.Now(),
	}

	t.Run("Chunk Large Document", func(t *testing.T) {
		chunks, err := ChunkDocument(doc, 200, 50)
		require.NoError(t, err)
		assert.Greater(t, len(chunks), 1)

		// Verify chunk properties
		for i, chunk := range chunks {
			assert.Equal(t, i+1, chunk.ChunkIndex)
			assert.Equal(t, len(chunks), chunk.ChunkTotal)
			assert.Contains(t, chunk.ID, "chunk")
			assert.Contains(t, chunk.Title, "Part")
			assert.True(t, len(chunk.Content) <= 250) // Allow some flexibility for word boundaries
		}
	})

	t.Run("No Chunking for Small Document", func(t *testing.T) {
		smallDoc := &core.Document{
			ID:        "small-doc",
			Content:   "Small content",
			CreatedAt: time.Now(),
		}

		chunks, err := ChunkDocument(smallDoc, 1000, 100)
		require.NoError(t, err)
		assert.Len(t, chunks, 1)
		assert.Equal(t, smallDoc, chunks[0])
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("parseTagList", func(t *testing.T) {
		tests := []struct {
			input    string
			expected []string
		}{
			{"", nil},
			{"tag1", []string{"tag1"}},
			{"tag1,tag2", []string{"tag1", "tag2"}},
			{"tag1, tag2, tag3", []string{"tag1", "tag2", "tag3"}},
			{" tag1 , tag2 ", []string{"tag1", "tag2"}},
			{"tag1,,tag2", []string{"tag1", "tag2"}},
		}

		for _, test := range tests {
			result := parseTagList(test.input)
			assert.Equal(t, test.expected, result, "Input: %q", test.input)
		}
	})

	t.Run("matchesFilter", func(t *testing.T) {
		tests := []struct {
			text     string
			filter   string
			expected bool
		}{
			{"test.txt", "", true},
			{"test.txt", "test", true},
			{"test.txt", "TEST", true},
			{"test.txt", "txt", true},
			{"test.txt", "pdf", false},
			{"docs/test.txt", "docs/*", true},
			{"docs/test.txt", "*/test.txt", true},
			{"docs/test.txt", "*docs*", true},
		}

		for _, test := range tests {
			result := matchesFilter(test.text, test.filter)
			assert.Equal(t, test.expected, result, "Text: %q, Filter: %q", test.text, test.filter)
		}
	})

	t.Run("tagsMatch", func(t *testing.T) {
		docTags := []string{"test", "markdown", "docs"}

		tests := []struct {
			filterTags []string
			expected   bool
		}{
			{nil, true},
			{[]string{}, true},
			{[]string{"test"}, true},
			{[]string{"TEST"}, true},
			{[]string{"markdown", "other"}, true},
			{[]string{"nonexistent"}, false},
			{[]string{"pdf", "other"}, false},
		}

		for _, test := range tests {
			result := tagsMatch(docTags, test.filterTags)
			assert.Equal(t, test.expected, result, "Filter tags: %v", test.filterTags)
		}
	})
}

// Integration Tests

func TestKnowledgeManagerIntegration(t *testing.T) {
	configPath := createTestConfig(t)
	tempDir, _ := createTestFiles(t)

	km, err := NewKnowledgeManager(configPath)
	require.NoError(t, err)

	err = km.Connect()
	require.NoError(t, err)
	defer km.Close()

	ctx := context.Background()

	t.Run("Upload Documents", func(t *testing.T) {
		options := UploadOptions{
			Recursive:       true,
			Tags:            []string{"test"},
			IncludeMetadata: true,
			ShowProgress:    false,
		}

		err := km.Upload(ctx, []string{tempDir}, options)
		assert.NoError(t, err)
	})

	t.Run("List Documents", func(t *testing.T) {
		options := ListOptions{
			OutputFormat: "table",
			Limit:        10,
		}

		err := km.List(ctx, options)
		assert.NoError(t, err)
	})

	t.Run("Search Knowledge Base", func(t *testing.T) {
		options := SearchOptions{
			OutputFormat:   "table",
			ScoreThreshold: 0.0,
			Limit:          5,
		}

		err := km.Search(ctx, "test", options)
		assert.NoError(t, err)
	})

	t.Run("Validate Knowledge Base", func(t *testing.T) {
		err := km.Validate(ctx)
		assert.NoError(t, err)
	})

	t.Run("Get Statistics", func(t *testing.T) {
		err := km.Stats(ctx, "table")
		assert.NoError(t, err)
	})

	t.Run("Clear with Dry Run", func(t *testing.T) {
		options := ClearOptions{
			DryRun: true,
		}

		err := km.Clear(ctx, options)
		assert.NoError(t, err)
	})
}

func TestOutputFormatters(t *testing.T) {
	// Create sample data for testing formatters
	sampleDocs := []core.KnowledgeResult{
		{
			Content:    "Test content 1",
			Score:      0.95,
			Source:     "test1.md",
			Title:      "Test Document 1",
			DocumentID: "doc1",
			Tags:       []string{"test", "markdown"},
			CreatedAt:  time.Now(),
		},
		{
			Content:    "Test content 2",
			Score:      0.85,
			Source:     "test2.txt",
			Title:      "Test Document 2",
			DocumentID: "doc2",
			Tags:       []string{"test", "text"},
			CreatedAt:  time.Now(),
		},
	}

	t.Run("Table Formatter", func(t *testing.T) {
		formatter := NewFormatter("table")

		// Test document formatting
		output := formatter.FormatDocuments(sampleDocs)
		assert.Contains(t, output, "Found 2 documents")
		assert.Contains(t, output, "Test Document 1")
		assert.Contains(t, output, "test1.md")

		// Test search results formatting
		output = formatter.FormatSearchResults(sampleDocs)
		assert.Contains(t, output, "Found 2 search results")
		assert.Contains(t, output, "0.950")
		assert.Contains(t, output, "0.850")
	})

	t.Run("JSON Formatter", func(t *testing.T) {
		formatter := NewFormatter("json")

		// Test document formatting
		output := formatter.FormatDocuments(sampleDocs)
		assert.Contains(t, output, "\"documents\"")
		assert.Contains(t, output, "\"count\": 2")
		assert.Contains(t, output, "Test Document 1")

		// Test search results formatting
		output = formatter.FormatSearchResults(sampleDocs)
		assert.Contains(t, output, "\"results\"")
		assert.Contains(t, output, "\"count\": 2")
	})

	t.Run("Empty Results", func(t *testing.T) {
		formatter := NewFormatter("table")

		output := formatter.FormatDocuments([]core.KnowledgeResult{})
		assert.Contains(t, output, "No documents found")

		output = formatter.FormatSearchResults([]core.KnowledgeResult{})
		assert.Contains(t, output, "No search results found")
	})
}

// Benchmark Tests

func BenchmarkDocumentProcessing(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "benchmark.md")

	// Create a larger test file
	content := "# Benchmark Document\n\n"
	for i := 0; i < 1000; i++ {
		content += "This is paragraph number " + string(rune(i)) + " with some test content. "
	}

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(b, err)

	options := ProcessingOptions{
		ChunkSize:       1000,
		ChunkOverlap:    200,
		IncludeMetadata: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CreateDocumentFromFile(testFile, options)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChunking(b *testing.B) {
	// Create a large document
	largeContent := ""
	for i := 0; i < 10000; i++ {
		largeContent += "This is sentence number with some content. "
	}

	doc := &core.Document{
		ID:        "benchmark-doc",
		Content:   largeContent,
		CreatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ChunkDocument(doc, 1000, 200)
		if err != nil {
			b.Fatal(err)
		}
	}
}
