---
title: Document Ingestion
description: Learn how to ingest and process documents using AgenticGoKit's current Document API, including chunking strategies, metadata extraction, and batch processing.
---

# Document Ingestion and Knowledge Base Management

## Overview

Document ingestion is a critical component of building comprehensive knowledge bases in AgenticGoKit. This tutorial covers the complete pipeline from raw documents to searchable knowledge, including document processing, chunking strategies, metadata extraction, and optimization techniques using the current Document API.

Effective document ingestion enables agents to access and reason over large collections of structured and unstructured data.

## Prerequisites

- Understanding of [Memory Systems Overview](README.md)
- Familiarity with [Vector Databases](vector-databases.md)
- Knowledge of document formats (PDF, Markdown, HTML, etc.)
- Basic understanding of text processing and NLP concepts

> Current PDF reader (MVP) — limitations and notes
>
> The project includes a minimal PDF processor intended as an MVP to enable basic PDF-to-text extraction for ingestion. Please note the following limitations and recommendations when working with PDF documents today:
>
> - No OCR: The built-in PDF reader extracts embedded text only. Scanned PDFs (image-only pages) will not be OCR'd and will return little or no text. If you need scanned PDF support, enable an OCR fallback (for example, Tesseract via `gosseract`) or pre-convert PDFs to text.
> - Layout and ordering: Complex layouts (multi-column text, tables, sidebars) may produce poorly ordered plain text. Expect imperfect sentence order for academic papers or magazines with multi-column formatting.
> - Metadata depth: The MVP extracts basic metadata (file size, modified time, page count). Rich metadata such as embedded title/author or structured attachments/images may not be available.
> - Performance and large files: Very large PDFs (hundreds of pages) are read into memory for extraction in the MVP. For production workloads, use streaming/external tooling or per-page processing to reduce memory usage.
> - Encrypted PDFs: Password-protected PDFs are not currently handled; they will fail with a readable error explaining the encryption.
>
> Recommended fallbacks and best practices:
>
> - For scanned/image PDFs use an OCR step (Tesseract) before ingestion or enable an OCR fallback in the processor.
> - For high-fidelity extraction (layout/tables/figures) prefer an external tool like `pdftotext` (Poppler) or a commercial parser (UniDoc) and feed the cleaned text into ingestion.
> - For very large collections, pre-process PDFs into plain text files split by page or section and ingest those files to avoid memory spikes.
>
> The documentation and CLI will be updated as we expand the PDF processor (page-aware chunking, OCR fallback, image extraction). If you rely heavily on PDF ingestion today, consider adding a small preprocessing step in your pipeline that runs `pdftotext` or OCR before using `agentcli knowledge upload`.

## Document Ingestion Pipeline

### Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Raw           │───▶│   Document       │───▶│   Text          │
│   Documents     │    │   Parser         │    │   Extraction    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Vector        │◀───│   Embedding      │◀───│   Text          │
│   Storage       │    │   Generation     │    │   Chunking      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
                                                ┌─────────────────┐
                                                │   Metadata      │
                                                │   Extraction    │
                                                └─────────────────┘
```

## Document Types and Structure

### 1. Current Document Structure

```go
// Document structure for ingestion using current API
type Document struct {
    ID         string         `json:"id"`
    Title      string         `json:"title,omitempty"`
    Content    string         `json:"content"`
    Source     string         `json:"source,omitempty"` // URL, file path, etc.
    Type       DocumentType   `json:"type,omitempty"`   // PDF, TXT, WEB, etc.
    Metadata   map[string]any `json:"metadata,omitempty"`
    Tags       []string       `json:"tags,omitempty"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at,omitempty"`
    ChunkIndex int            `json:"chunk_index,omitempty"` // For chunked documents
    ChunkTotal int            `json:"chunk_total,omitempty"`
}

// Supported document types
const (
    DocumentTypePDF      DocumentType = "pdf"
    DocumentTypeText     DocumentType = "txt"
    DocumentTypeMarkdown DocumentType = "md"
    DocumentTypeWeb      DocumentType = "web"
    DocumentTypeCode     DocumentType = "code"
    DocumentTypeJSON     DocumentType = "json"
)
```

### 2. Basic Document Ingestion

```go
package main

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "log"
    "os"
    "strings"
    "sync"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create memory with current configuration
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "memory", // or "pgvector", "weaviate"
        Connection: "memory",
        MaxResults: 10,
        Dimensions: 1536,
        AutoEmbed:  true,
        
        // RAG-enhanced settings
        EnableRAG:               true,
        EnableKnowledgeBase:     true,
        KnowledgeMaxResults:     20,
        KnowledgeScoreThreshold: 0.7,
        ChunkSize:               1000,
        ChunkOverlap:            200,
        
        Embedding: core.EmbeddingConfig{
            Provider: "dummy", // Use "openai" for production
            Model:    "dummy-model",
        },
        
        Documents: core.DocumentConfig{
            AutoChunk:                true,
            SupportedTypes:           []string{"pdf", "txt", "md", "web", "code"},
            MaxFileSize:              "10MB",
            EnableMetadataExtraction: true,
            EnableURLScraping:        true,
        },
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    defer memory.Close()
    
    // Demonstrate document ingestion
    err = ingestBasicDocument(memory)
    if err != nil {
        log.Fatalf("Failed to ingest document: %v", err)
    }
    
    err = ingestMultipleDocuments(memory)
    if err != nil {
        log.Fatalf("Failed to ingest multiple documents: %v", err)
    }
}

func ingestBasicDocument(memory core.Memory) error {
    ctx := context.Background()
    
    // Create a document using current Document structure
    doc := core.Document{
        ID:      "ml-intro-001",
        Title:   "Introduction to Machine Learning",
        Content: `Machine learning is a subset of artificial intelligence that enables computers to learn and make decisions from data without being explicitly programmed for every task. It involves algorithms that can identify patterns, make predictions, and improve their performance over time.

Key concepts in machine learning include:
- Supervised learning: Learning from labeled data
- Unsupervised learning: Finding patterns in unlabeled data  
- Reinforcement learning: Learning through interaction and feedback
- Neural networks: Computing systems inspired by biological neural networks
- Deep learning: Machine learning using deep neural networks

Applications of machine learning are widespread, including image recognition, natural language processing, recommendation systems, and autonomous vehicles.`,
        Source:  "textbook-chapter-1.pdf",
        Type:    core.DocumentTypePDF,
        Metadata: map[string]any{
            "author":     "Dr. Jane Smith",
            "chapter":    1,
            "subject":    "machine-learning",
            "difficulty": "beginner",
            "language":   "english",
            "page_count": 15,
            "isbn":       "978-0123456789",
        },
        Tags:      []string{"ml", "ai", "introduction", "supervised-learning", "neural-networks"},
        CreatedAt: time.Now(),
    }
    
    // Ingest the document using current API
    err := memory.IngestDocument(ctx, doc)
    if err != nil {
        return fmt.Errorf("failed to ingest document: %w", err)
    }
    
    fmt.Printf("Successfully ingested document: %s\n", doc.Title)
    
    // Test search to verify ingestion
    results, err := memory.SearchKnowledge(ctx, "machine learning concepts",
        core.WithLimit(5),
        core.WithTags([]string{"ml"}),
    )
    if err != nil {
        return fmt.Errorf("failed to search ingested document: %w", err)
    }
    
    fmt.Printf("Found %d results when searching for ingested content\n", len(results))
    return nil
}
```

### 3. Batch Document Ingestion

```go
func ingestMultipleDocuments(memory core.Memory) error {
    ctx := context.Background()
    
    // Prepare multiple documents using current Document structure
    documents := []core.Document{
        {
            ID:      "nn-fundamentals-002",
            Title:   "Neural Networks Fundamentals",
            Content: `Neural networks are computing systems inspired by biological neural networks that constitute animal brains. They are based on a collection of connected units or nodes called artificial neurons, which loosely model the neurons in a biological brain.

Key components of neural networks:
- Neurons (nodes): Basic processing units
- Weights: Connection strengths between neurons
- Activation functions: Functions that determine neuron output
- Layers: Groups of neurons (input, hidden, output layers)
- Backpropagation: Learning algorithm for training networks

Types of neural networks:
- Feedforward networks: Information flows in one direction
- Recurrent networks: Networks with feedback connections
- Convolutional networks: Specialized for processing grid-like data
- Transformer networks: Attention-based architectures`,
            Source:  "textbook-chapter-2.pdf",
            Type:    core.DocumentTypePDF,
            Metadata: map[string]any{
                "author":     "Dr. Jane Smith",
                "chapter":    2,
                "subject":    "neural-networks",
                "difficulty": "intermediate",
                "language":   "english",
                "page_count": 22,
            },
            Tags:      []string{"neural-networks", "deep-learning", "backpropagation", "cnn", "rnn"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "data-preprocessing-003",
            Title:   "Data Preprocessing Techniques",
            Content: `Data preprocessing is a crucial step in machine learning pipelines that involves cleaning, transforming, and preparing raw data for analysis and model training.

Common preprocessing techniques:
- Data cleaning: Handling missing values, outliers, and inconsistencies
- Data transformation: Scaling, normalization, and encoding
- Feature selection: Choosing relevant features for modeling
- Feature engineering: Creating new features from existing data
- Data splitting: Dividing data into training, validation, and test sets

Data quality issues:
- Missing values: Can be handled by imputation or removal
- Outliers: May need to be identified and treated appropriately
- Inconsistent formats: Require standardization
- Duplicate records: Should be identified and removed
- Noise: Random errors that can affect model performance`,
            Source:  "textbook-chapter-3.pdf",
            Type:    core.DocumentTypePDF,
            Metadata: map[string]any{
                "author":     "Dr. Jane Smith",
                "chapter":    3,
                "subject":    "data-preprocessing",
                "difficulty": "beginner",
                "language":   "english",
                "page_count": 18,
            },
            Tags:      []string{"data-science", "preprocessing", "feature-engineering", "data-cleaning"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "python-ml-code-004",
            Title:   "Python Machine Learning Implementation",
            Content: `# Machine Learning with Python

import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_squared_error

# Load and prepare data
def load_data(filename):
    """Load dataset from CSV file"""
    data = pd.read_csv(filename)
    return data

def preprocess_data(data):
    """Clean and preprocess the dataset"""
    # Handle missing values
    data = data.dropna()
    
    # Feature scaling
    from sklearn.preprocessing import StandardScaler
    scaler = StandardScaler()
    
    return data, scaler

# Train machine learning model
def train_model(X_train, y_train):
    """Train a linear regression model"""
    model = LinearRegression()
    model.fit(X_train, y_train)
    return model

# Evaluate model performance
def evaluate_model(model, X_test, y_test):
    """Evaluate model and return metrics"""
    predictions = model.predict(X_test)
    mse = mean_squared_error(y_test, predictions)
    return mse, predictions

# Main execution
if __name__ == "__main__":
    # Load and preprocess data
    data = load_data("dataset.csv")
    processed_data, scaler = preprocess_data(data)
    
    # Split data
    X = processed_data.drop('target', axis=1)
    y = processed_data['target']
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)
    
    # Train and evaluate model
    model = train_model(X_train, y_train)
    mse, predictions = evaluate_model(model, X_test, y_test)
    
    print(f"Model MSE: {mse}")`,
            Source:  "ml_example.py",
            Type:    core.DocumentTypeCode,
            Metadata: map[string]any{
                "programming_language": "python",
                "framework":           "scikit-learn",
                "topic":              "machine-learning",
                "difficulty":         "intermediate",
                "lines_of_code":      65,
                "functions":          []string{"load_data", "preprocess_data", "train_model", "evaluate_model"},
            },
            Tags:      []string{"python", "scikit-learn", "linear-regression", "code-example"},
            CreatedAt: time.Now(),
        },
    }
    
    // Batch ingest documents using current API
    err := memory.IngestDocuments(ctx, documents)
    if err != nil {
        return fmt.Errorf("failed to ingest documents: %w", err)
    }
    
    fmt.Printf("Successfully ingested %d documents\n", len(documents))
    
    // Test search across all ingested documents
    results, err := memory.SearchKnowledge(ctx, "neural networks and data preprocessing",
        core.WithLimit(10),
        core.WithScoreThreshold(0.5),
    )
    if err != nil {
        return fmt.Errorf("failed to search ingested documents: %w", err)
    }
    
    fmt.Printf("Found %d results across all ingested documents:\n", len(results))
    for _, result := range results {
        fmt.Printf("- %s (Score: %.3f, Source: %s)\n", 
            result.Title, result.Score, result.Source)
    }
    
    return nil
}
```

## Document Validation and Error Handling

### 1. Document Validation

```go
func validateDocument(doc core.Document) error {
    // Validate required fields
    if doc.ID == "" {
        return fmt.Errorf("document ID is required")
    }
    
    if doc.Content == "" {
        return fmt.Errorf("document content is required")
    }
    
    // Validate document type
    validTypes := map[core.DocumentType]bool{
        core.DocumentTypePDF:      true,
        core.DocumentTypeText:     true,
        core.DocumentTypeMarkdown: true,
        core.DocumentTypeWeb:      true,
        core.DocumentTypeCode:     true,
        core.DocumentTypeJSON:     true,
    }
    
    if doc.Type != "" && !validTypes[doc.Type] {
        return fmt.Errorf("invalid document type: %s", doc.Type)
    }
    
    // Validate content size
    maxContentSize := 10 * 1024 * 1024 // 10MB
    if len(doc.Content) > maxContentSize {
        return fmt.Errorf("document content too large: %d bytes (max: %d)", 
            len(doc.Content), maxContentSize)
    }
    
    // Validate metadata
    if doc.Metadata != nil {
        for key, value := range doc.Metadata {
            if key == "" {
                return fmt.Errorf("metadata key cannot be empty")
            }
            
            // Check for reasonable metadata value types
            switch value.(type) {
            case string, int, int64, float64, bool, []string, []int, map[string]any:
                // Valid types
            default:
                return fmt.Errorf("invalid metadata value type for key %s: %T", key, value)
            }
        }
    }
    
    // Validate tags
    for _, tag := range doc.Tags {
        if tag == "" {
            return fmt.Errorf("empty tag not allowed")
        }
        if len(tag) > 100 {
            return fmt.Errorf("tag too long: %s (max: 100 chars)", tag)
        }
    }
    
    return nil
}

func validateAndIngestDocument(memory core.Memory, doc core.Document) error {
    ctx := context.Background()
    
    // Validate document
    if err := validateDocument(doc); err != nil {
        return fmt.Errorf("document validation failed: %w", err)
    }
    
    // Set timestamps if not provided
    if doc.CreatedAt.IsZero() {
        doc.CreatedAt = time.Now()
    }
    
    // Generate ID if not provided
    if doc.ID == "" {
        doc.ID = generateDocumentID(doc)
    }
    
    // Ingest with error handling
    err := memory.IngestDocument(ctx, doc)
    if err != nil {
        return fmt.Errorf("failed to ingest document %s: %w", doc.ID, err)
    }
    
    fmt.Printf("Successfully ingested document: %s\n", doc.ID)
    return nil
}

func generateDocumentID(doc core.Document) string {
    // Generate ID based on content hash and timestamp
    hasher := sha256.New()
    hasher.Write([]byte(doc.Title + doc.Content + doc.Source))
    hash := hex.EncodeToString(hasher.Sum(nil))
    return fmt.Sprintf("doc-%s-%d", hash[:8], time.Now().Unix())
}
```

### 2. Batch Processing with Error Recovery

```go
func ingestDocumentsWithErrorRecovery(memory core.Memory, documents []core.Document) error {
    ctx := context.Background()
    
    var successCount, errorCount int
    var errors []error
    
    for i, doc := range documents {
        err := validateAndIngestDocument(memory, doc)
        if err != nil {
            errorCount++
            errors = append(errors, fmt.Errorf("document %d (%s): %w", i, doc.ID, err))
            log.Printf("Failed to ingest document %d: %v", i, err)
            continue
        }
        successCount++
    }
    
    fmt.Printf("Ingestion completed: %d successful, %d failed\n", successCount, errorCount)
    
    if errorCount > 0 {
        fmt.Printf("Errors encountered:\n")
        for _, err := range errors {
            fmt.Printf("- %v\n", err)
        }
        
        // Return error if more than 50% failed
        if errorCount > len(documents)/2 {
            return fmt.Errorf("batch ingestion failed: %d/%d documents failed", 
                errorCount, len(documents))
        }
    }
    
    return nil
}
```

### 3. Metadata Extraction and Enhancement

```go
func enhanceDocumentMetadata(doc *core.Document) error {
    // Extract basic statistics
    wordCount := len(strings.Fields(doc.Content))
    charCount := len(doc.Content)
    lineCount := len(strings.Split(doc.Content, "\n"))
    
    // Initialize metadata if nil
    if doc.Metadata == nil {
        doc.Metadata = make(map[string]any)
    }
    
    // Add basic statistics
    doc.Metadata["word_count"] = wordCount
    doc.Metadata["char_count"] = charCount
    doc.Metadata["line_count"] = lineCount
    
    // Extract language (simple heuristic)
    language := detectLanguage(doc.Content)
    doc.Metadata["detected_language"] = language
    
    // Extract key phrases (simplified)
    keyPhrases := extractKeyPhrases(doc.Content)
    doc.Metadata["key_phrases"] = keyPhrases
    
    // Estimate reading time
    readingTime := estimateReadingTime(wordCount)
    doc.Metadata["estimated_reading_time"] = readingTime
    
    // Extract document structure for markdown
    if doc.Type == core.DocumentTypeMarkdown {
        structure := extractMarkdownStructure(doc.Content)
        doc.Metadata["structure"] = structure
    }
    
    // Extract code information for code documents
    if doc.Type == core.DocumentTypeCode {
        codeInfo := extractCodeInformation(doc.Content, doc.Source)
        for key, value := range codeInfo {
            doc.Metadata[key] = value
        }
    }
    
    return nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

## Advanced Document Processing

### 1. Document Processing with Chunking

```go
func demonstrateDocumentChunking(memory core.Memory) error {
    ctx := context.Background()
    
    // Create a large document that will be automatically chunked
    largeContent := `
# Comprehensive Guide to Machine Learning

## Chapter 1: Introduction to Machine Learning

Machine learning is a method of data analysis that automates analytical model building. It is a branch of artificial intelligence based on the idea that systems can learn from data, identify patterns and make decisions with minimal human intervention.

### Historical Background

The concept of machine learning has been around since the 1950s. Arthur Samuel, an American pioneer in the field of computer gaming and artificial intelligence, coined the term "machine learning" in 1959. He defined it as a "field of study that gives computers the ability to learn without being explicitly programmed."

### Types of Machine Learning

There are three main types of machine learning:

1. **Supervised Learning**: This is where the algorithm learns from labeled training data, helping it to predict outcomes for unforeseen data. Examples include classification and regression problems.

2. **Unsupervised Learning**: Here, the algorithm tries to learn the underlying structure of data without any labeled examples. Common techniques include clustering and dimensionality reduction.

3. **Reinforcement Learning**: This type involves an agent learning to make decisions by taking actions in an environment to maximize some notion of cumulative reward.

## Chapter 2: Supervised Learning Algorithms

Supervised learning algorithms build a mathematical model of training data that contains both inputs and desired outputs. The training data consists of a set of training examples.

### Linear Regression

Linear regression is a linear approach to modeling the relationship between a scalar response and one or more explanatory variables. The case of one explanatory variable is called simple linear regression; for more than one, the process is called multiple linear regression.

### Decision Trees

A decision tree is a decision support tool that uses a tree-like model of decisions and their possible consequences, including chance event outcomes, resource costs, and utility. Decision trees are commonly used in operations research, specifically in decision analysis, to help identify a strategy most likely to reach a goal.

### Support Vector Machines

Support Vector Machines (SVMs) are supervised learning models with associated learning algorithms that analyze data for classification and regression analysis. Given a set of training examples, each marked as belonging to one of two categories, an SVM training algorithm builds a model that assigns new examples to one category or the other.

## Chapter 3: Unsupervised Learning

Unsupervised learning is a type of machine learning that looks for previously undetected patterns in a data set with no pre-existing labels and with a minimum of human supervision.

### Clustering

Cluster analysis or clustering is the task of grouping a set of objects in such a way that objects in the same group (called a cluster) are more similar (in some sense) to each other than to those in other groups (clusters).

### Principal Component Analysis

Principal component analysis (PCA) is a statistical procedure that uses an orthogonal transformation to convert a set of observations of possibly correlated variables into a set of values of linearly uncorrelated variables called principal components.

## Chapter 4: Deep Learning

Deep learning is part of a broader family of machine learning methods based on artificial neural networks with representation learning. Learning can be supervised, semi-supervised or unsupervised.

### Neural Network Architecture

Deep learning architectures such as deep neural networks, deep belief networks, recurrent neural networks and convolutional neural networks have been applied to fields including computer vision, speech recognition, natural language processing, machine translation, bioinformatics and drug design.

### Convolutional Neural Networks

A convolutional neural network (CNN, or ConvNet) is a class of deep neural networks, most commonly applied to analyzing visual imagery. CNNs are regularized versions of multilayer perceptrons.

### Recurrent Neural Networks

A recurrent neural network (RNN) is a class of artificial neural networks where connections between nodes form a directed graph along a temporal sequence. This allows it to exhibit temporal dynamic behavior.
`
    
    // Create document that will be automatically chunked due to size
    doc := core.Document{
        ID:      "ml-comprehensive-guide",
        Title:   "Comprehensive Guide to Machine Learning",
        Content: largeContent,
        Source:  "ml-comprehensive-guide.md",
        Type:    core.DocumentTypeMarkdown,
        Metadata: map[string]any{
            "author":       "AI Learning Team",
            "topic":        "machine-learning",
            "difficulty":   "comprehensive",
            "word_count":   len(strings.Fields(largeContent)),
            "char_count":   len(largeContent),
            "chapters":     4,
            "content_type": "educational",
        },
        Tags:      []string{"ml", "comprehensive", "guide", "supervised", "unsupervised", "deep-learning"},
        CreatedAt: time.Now(),
    }
    
    // Ingest document - it will be automatically chunked based on configuration
    err := memory.IngestDocument(ctx, doc)
    if err != nil {
        return fmt.Errorf("failed to ingest large document: %w", err)
    }
    
    fmt.Printf("Successfully ingested large document: %s\n", doc.Title)
    
    // Search for content that should be in different chunks
    searchQueries := []string{
        "introduction to machine learning",
        "supervised learning algorithms",
        "clustering and PCA",
        "convolutional neural networks",
    }
    
    for _, query := range searchQueries {
        results, err := memory.SearchKnowledge(ctx, query,
            core.WithLimit(3),
            core.WithScoreThreshold(0.6),
            core.WithSources([]string{"ml-comprehensive-guide.md"}),
        )
        if err != nil {
            fmt.Printf("Search failed for '%s': %v\n", query, err)
            continue
        }
        
        fmt.Printf("\nSearch results for '%s':\n", query)
        for _, result := range results {
            fmt.Printf("- Score: %.3f\n", result.Score)
            fmt.Printf("  Content: %s...\n", truncateString(result.Content, 100))
            if result.ChunkIndex > 0 {
                fmt.Printf("  Chunk: %d\n", result.ChunkIndex)
            }
        }
    }
    
    return nil
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}
```

### 2. Multi-Modal Document Processing

```go
func demonstrateMultiModalIngestion(memory core.Memory) error {
    ctx := context.Background()
    
    // Different types of documents with rich metadata
    documents := []core.Document{
        // Research paper
        {
            ID:      "research-paper-001",
            Title:   "Attention Is All You Need",
            Content: `Abstract: The dominant sequence transduction models are based on complex recurrent or convolutional neural networks that include an encoder and a decoder. The best performing models also connect the encoder and decoder through an attention mechanism. We propose a new simple network architecture, the Transformer, based solely on attention mechanisms, dispensing with recurrence and convolutions entirely.

Introduction: Recurrent neural networks, long short-term memory and gated recurrent neural networks in particular, have been firmly established as state of the art approaches in sequence modeling and transduction problems such as language modeling and machine translation.

The Transformer Model: The Transformer follows this overall architecture using stacked self-attention and point-wise, fully connected layers for both the encoder and decoder, shown in the left and right halves of Figure 1, respectively.`,
            Source:  "attention_is_all_you_need.pdf",
            Type:    core.DocumentTypePDF,
            Metadata: map[string]any{
                "authors":        []string{"Ashish Vaswani", "Noam Shazeer", "Niki Parmar"},
                "publication":    "NIPS 2017",
                "citations":      50000,
                "research_area":  "natural-language-processing",
                "model_type":     "transformer",
                "contribution":   "attention-mechanism",
                "impact_factor":  9.8,
            },
            Tags: []string{"transformer", "attention", "nlp", "research", "neural-networks"},
            CreatedAt: time.Now(),
        },
        
        // Code repository
        {
            ID:      "transformer-implementation",
            Title:   "Transformer Implementation in PyTorch",
            Content: `"""
PyTorch implementation of the Transformer model from "Attention Is All You Need"
"""

import torch
import torch.nn as nn
import torch.nn.functional as F
import math

class MultiHeadAttention(nn.Module):
    def __init__(self, d_model, n_heads):
        super(MultiHeadAttention, self).__init__()
        self.d_model = d_model
        self.n_heads = n_heads
        self.d_k = d_model // n_heads
        
        self.W_q = nn.Linear(d_model, d_model)
        self.W_k = nn.Linear(d_model, d_model)
        self.W_v = nn.Linear(d_model, d_model)
        self.W_o = nn.Linear(d_model, d_model)
        
    def scaled_dot_product_attention(self, Q, K, V, mask=None):
        scores = torch.matmul(Q, K.transpose(-2, -1)) / math.sqrt(self.d_k)
        
        if mask is not None:
            scores = scores.masked_fill(mask == 0, -1e9)
            
        attention_weights = F.softmax(scores, dim=-1)
        output = torch.matmul(attention_weights, V)
        
        return output, attention_weights
    
    def forward(self, query, key, value, mask=None):
        batch_size = query.size(0)
        
        # Linear transformations and reshape
        Q = self.W_q(query).view(batch_size, -1, self.n_heads, self.d_k).transpose(1, 2)
        K = self.W_k(key).view(batch_size, -1, self.n_heads, self.d_k).transpose(1, 2)
        V = self.W_v(value).view(batch_size, -1, self.n_heads, self.d_k).transpose(1, 2)
        
        # Apply attention
        attention_output, attention_weights = self.scaled_dot_product_attention(Q, K, V, mask)
        
        # Concatenate heads and apply output projection
        attention_output = attention_output.transpose(1, 2).contiguous().view(
            batch_size, -1, self.d_model)
        output = self.W_o(attention_output)
        
        return output

class TransformerBlock(nn.Module):
    def __init__(self, d_model, n_heads, d_ff, dropout=0.1):
        super(TransformerBlock, self).__init__()
        self.attention = MultiHeadAttention(d_model, n_heads)
        self.norm1 = nn.LayerNorm(d_model)
        self.norm2 = nn.LayerNorm(d_model)
        
        self.feed_forward = nn.Sequential(
            nn.Linear(d_model, d_ff),
            nn.ReLU(),
            nn.Linear(d_ff, d_model)
        )
        
        self.dropout = nn.Dropout(dropout)
        
    def forward(self, x, mask=None):
        # Self-attention with residual connection
        attn_output = self.attention(x, x, x, mask)
        x = self.norm1(x + self.dropout(attn_output))
        
        # Feed-forward with residual connection
        ff_output = self.feed_forward(x)
        x = self.norm2(x + self.dropout(ff_output))
        
        return x`,
            Source:  "transformer.py",
            Type:    core.DocumentTypeCode,
            Metadata: map[string]any{
                "programming_language": "python",
                "framework":           "pytorch",
                "model_architecture":  "transformer",
                "lines_of_code":       85,
                "complexity":          "advanced",
                "classes":             []string{"MultiHeadAttention", "TransformerBlock"},
                "functions":           []string{"scaled_dot_product_attention", "forward"},
                "dependencies":        []string{"torch", "torch.nn", "torch.nn.functional", "math"},
            },
            Tags: []string{"pytorch", "transformer", "attention", "implementation", "python"},
            CreatedAt: time.Now(),
        },
        
        // Tutorial/Blog post
        {
            ID:      "transformer-tutorial",
            Title:   "Understanding Transformers: A Visual Guide",
            Content: `# Understanding Transformers: A Visual Guide

Transformers have revolutionized natural language processing and are now being applied to computer vision, audio processing, and many other domains. This guide will help you understand how transformers work through visual explanations and intuitive examples.

## What Makes Transformers Special?

Before transformers, most sequence-to-sequence models relied on recurrent neural networks (RNNs) or convolutional neural networks (CNNs). These architectures had limitations:

- **RNNs**: Process sequences step by step, making them slow and prone to vanishing gradients
- **CNNs**: Good at capturing local patterns but struggle with long-range dependencies

Transformers solve these problems by using **attention mechanisms** that can directly connect any two positions in a sequence, regardless of their distance.

## The Attention Mechanism

The core innovation of transformers is the attention mechanism. Think of attention as a way for the model to decide which parts of the input to focus on when processing each element.

### Self-Attention in Action

Imagine you're reading the sentence: "The cat sat on the mat because it was comfortable."

When processing the word "it", the model needs to figure out what "it" refers to. Self-attention allows the model to look at all previous words and determine that "it" most likely refers to "the cat" or "the mat" based on context.

### Multi-Head Attention

Instead of having just one attention mechanism, transformers use multiple "attention heads" that can focus on different types of relationships:

- **Head 1**: Might focus on syntactic relationships (subject-verb-object)
- **Head 2**: Might focus on semantic relationships (synonyms, antonyms)
- **Head 3**: Might focus on positional relationships (nearby words)

## The Complete Transformer Architecture

A transformer consists of two main components:

1. **Encoder**: Processes the input sequence and creates rich representations
2. **Decoder**: Generates the output sequence using the encoder's representations

### Key Components:

- **Positional Encoding**: Since transformers don't process sequences in order, they need a way to understand position
- **Layer Normalization**: Helps with training stability
- **Residual Connections**: Allow gradients to flow better during training
- **Feed-Forward Networks**: Add non-linearity and processing power

## Applications Beyond NLP

While transformers started in NLP, they're now used in:

- **Computer Vision**: Vision Transformer (ViT) for image classification
- **Audio Processing**: Speech recognition and music generation
- **Protein Folding**: AlphaFold uses transformer-like architectures
- **Code Generation**: GitHub Copilot and similar tools

## Conclusion

Transformers represent a paradigm shift in how we think about sequence modeling. By replacing recurrence with attention, they enable parallel processing and better capture of long-range dependencies.`,
            Source:  "transformer-guide.md",
            Type:    core.DocumentTypeMarkdown,
            Metadata: map[string]any{
                "content_type":    "tutorial",
                "difficulty":      "intermediate",
                "reading_time":    "15 minutes",
                "topic":          "transformers",
                "target_audience": "ml-practitioners",
                "visual_elements": true,
                "interactive":     false,
                "last_updated":    time.Now().Format("2006-01-02"),
            },
            Tags: []string{"transformers", "attention", "tutorial", "visual-guide", "nlp"},
            CreatedAt: time.Now(),
        }currence with attention, they've enabled more efficient training and better performance across a wide range of tasks. Understanding transformers is crucial for anyone working in modern AI and machine learning. better performance across many tasks.

The key insight is that attention allows models to directly access any part of the input when making decisions, leading to better understanding of context and relationships in data.`,
            Source:  "transformer-visual-guide.md",
            Type:    core.DocumentTypeMarkdown,
            Metadata: map[string]any{
                "content_type":    "tutorial",
                "target_audience": "intermediate",
                "reading_time":    "15 minutes",
                "topics":          []string{"transformers", "attention", "deep-learning"},
                "difficulty":      "intermediate",
                "format":          "visual-guide",
                "word_count":      len(strings.Fields("Understanding Transformers: A Visual Guide...")),
            },
            Tags: []string{"tutorial", "transformers", "attention", "visual-guide", "deep-learning"},
            CreatedAt: time.Now(),
        },
    }
    
    // Ingest all documents
    err := memory.IngestDocuments(ctx, documents)
    if err != nil {
        return fmt.Errorf("failed to ingest multi-modal documents: %w", err)
    }
    
    fmt.Printf("Successfully ingested %d multi-modal documents\n", len(documents))
    
    // Demonstrate different search strategies
    searchScenarios := []struct {
        query   string
        options []core.SearchOption
        description string
    }{
        {
            query: "transformer attention mechanism",
            options: []core.SearchOption{
                core.WithLimit(5),
                core.WithScoreThreshold(0.7),
            },
            description: "General search across all document types",
        },
        {
            query: "pytorch implementation",
            options: []core.SearchOption{
                core.WithLimit(3),
                core.WithDocumentTypes([]core.DocumentType{core.DocumentTypeCode}),
            },
            description: "Search only in code documents",
        },
        {
            query: "visual explanation of transformers",
            options: []core.SearchOption{
                core.WithLimit(3),
                core.WithTags([]string{"tutorial", "visual-guide"}),
            },
            description: "Search for tutorial content",
        },
    }
    
    for _, scenario := range searchScenarios {
        fmt.Printf("\n%s:\n", scenario.description)
        results, err := memory.SearchKnowledge(ctx, scenario.query, scenario.options...)
        if err != nil {
            fmt.Printf("Search failed: %v\n", err)
            continue
        }
        
        for _, result := range results {
            fmt.Printf("- %s (Score: %.3f, Type: %s)\n", 
                result.Title, result.Score, getDocumentTypeFromSource(result.Source))
            fmt.Printf("  Content: %s...\n", truncateString(result.Content, 80))
        }
    }
    
    return nil
}

func getDocumentTypeFromSource(source string) string {
    if strings.HasSuffix(source, ".pdf") {
        return "PDF"
    } else if strings.HasSuffix(source, ".py") {
        return "Code"
    } else if strings.HasSuffix(source, ".md") {
        return "Markdown"
    }
    return "Unknown"
}
```

## Knowledge Base Search and Retrieval

### 1. Advanced Search Patterns

```go
func demonstrateAdvancedSearch(memory core.Memory) error {
    ctx := context.Background()
    
    // Complex search with multiple filters
    results, err := memory.SearchKnowledge(ctx, "machine learning neural networks",
        core.WithLimit(10),
        core.WithScoreThreshold(0.6),
        core.WithTags([]string{"neural-networks", "ml"}),
        core.WithDocumentTypes([]core.DocumentType{
            core.DocumentTypePDF,
            core.DocumentTypeMarkdown,
        }),
    )
    if err != nil {
        return fmt.Errorf("advanced search failed: %w", err)
    }
    
    fmt.Printf("Advanced search found %d results:\n", len(results))
    for _, result := range results {
        fmt.Printf("- %s (Score: %.3f)\n", result.Title, result.Score)
        fmt.Printf("  Source: %s, Document ID: %s\n", result.Source, result.DocumentID)
        if len(result.Tags) > 0 {
            fmt.Printf("  Tags: %v\n", result.Tags)
        }
        if result.ChunkIndex > 0 {
            fmt.Printf("  Chunk: %d/%d\n", result.ChunkIndex+1, result.ChunkTotal)
        }
        fmt.Println()
    }
    
    // Hybrid search combining personal memory and knowledge base
    hybridResult, err := memory.SearchAll(ctx, "python machine learning implementation",
        core.WithLimit(15),
        core.WithIncludePersonal(true),
        core.WithIncludeKnowledge(true),
        core.WithScoreThreshold(0.5),
    )
    if err != nil {
        return fmt.Errorf("hybrid search failed: %w", err)
    }
    
    fmt.Printf("Hybrid search results:\n")
    fmt.Printf("Personal Memory: %d results\n", len(hybridResult.PersonalMemory))
    fmt.Printf("Knowledge Base: %d results\n", len(hybridResult.Knowledge))
    fmt.Printf("Total Results: %d\n", hybridResult.TotalResults)
    fmt.Printf("Search Time: %v\n", hybridResult.SearchTime)
    
    return nil
}
```

### 2. RAG Context Building

```go
func demonstrateRAGContextBuilding(memory core.Memory) error {
    ctx := context.Background()
    
    // Build comprehensive RAG context for a complex query
    query := "How do I implement a transformer model in PyTorch with attention mechanisms?"
    
    ragContext, err := memory.BuildContext(ctx, query,
        core.WithMaxTokens(4000),
        core.WithPersonalWeight(0.2),   // Less weight on personal memory
        core.WithKnowledgeWeight(0.8),  // More weight on knowledge base
        core.WithHistoryLimit(3),       // Include recent conversation
        core.WithIncludeSources(true),  // Include source attribution
    )
    if err != nil {
        return fmt.Errorf("failed to build RAG context: %w", err)
    }
    
    fmt.Printf("RAG Context for: %s\n", ragContext.Query)
    fmt.Printf("Token Count: %d\n", ragContext.TokenCount)
    fmt.Printf("Sources: %v\n", ragContext.Sources)
    fmt.Printf("Knowledge Results: %d\n", len(ragContext.Knowledge))
    fmt.Printf("Personal Memory Results: %d\n", len(ragContext.PersonalMemory))
    fmt.Printf("Chat History: %d messages\n", len(ragContext.ChatHistory))
    
    // The context text is formatted for LLM consumption
    fmt.Printf("\nFormatted Context (first 500 chars):\n%s...\n", 
        truncateString(ragContext.ContextText, 500))
    
    return nil
}
```

## Production Optimization

### 1. Batch Processing Pipeline

```go
type DocumentBatchProcessor struct {
    memory      core.Memory
    concurrency int
    batchSize   int
    metrics     *ProcessingMetrics
}

type ProcessingMetrics struct {
    DocumentsProcessed int64         `json:"documents_processed"`
    ProcessingTime     time.Duration `json:"processing_time"`
    ErrorCount         int64         `json:"error_count"`
    AverageDocSize     float64       `json:"average_doc_size"`
    mu                 sync.RWMutex
}

func NewDocumentBatchProcessor(memory core.Memory, concurrency, batchSize int) *DocumentBatchProcessor {
    return &DocumentBatchProcessor{
        memory:      memory,
        concurrency: concurrency,
        batchSize:   batchSize,
        metrics:     &ProcessingMetrics{},
    }
}

func (dbp *DocumentBatchProcessor) ProcessDocuments(ctx context.Context, documents []core.Document) error {
    start := time.Now()
    
    // Process documents in batches
    for i := 0; i < len(documents); i += dbp.batchSize {
        end := i + dbp.batchSize
        if end > len(documents) {
            end = len(documents)
        }
        
        batch := documents[i:end]
        err := dbp.processBatch(ctx, batch)
        if err != nil {
            dbp.metrics.mu.Lock()
            dbp.metrics.ErrorCount++
            dbp.metrics.mu.Unlock()
            return fmt.Errorf("failed to process batch %d-%d: %w", i, end-1, err)
        }
        
        fmt.Printf("Processed batch %d-%d (%d documents)\n", i, end-1, len(batch))
    }
    
    // Update metrics
    dbp.metrics.mu.Lock()
    dbp.metrics.DocumentsProcessed += int64(len(documents))
    dbp.metrics.ProcessingTime = time.Since(start)
    
    totalSize := 0
    for _, doc := range documents {
        totalSize += len(doc.Content)
    }
    dbp.metrics.AverageDocSize = float64(totalSize) / float64(len(documents))
    dbp.metrics.mu.Unlock()
    
    return nil
}

func (dbp *DocumentBatchProcessor) processBatch(ctx context.Context, documents []core.Document) error {
    // Use worker pool for concurrent processing
    jobs := make(chan core.Document, len(documents))
    results := make(chan error, len(documents))
    
    // Start workers
    for w := 0; w < dbp.concurrency; w++ {
        go dbp.worker(ctx, jobs, results)
    }
    
    // Send jobs
    for _, doc := range documents {
        jobs <- doc
    }
    close(jobs)
    
    // Collect results
    var errors []error
    for i := 0; i < len(documents); i++ {
        if err := <-results; err != nil {
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("batch processing failed with %d errors: %v", len(errors), errors[0])
    }
    
    return nil
}

func (dbp *DocumentBatchProcessor) worker(ctx context.Context, jobs <-chan core.Document, results chan<- error) {
    for doc := range jobs {
        err := dbp.memory.IngestDocument(ctx, doc)
        results <- err
    }
}

func (dbp *DocumentBatchProcessor) GetMetrics() ProcessingMetrics {
    dbp.metrics.mu.RLock()
    defer dbp.metrics.mu.RUnlock()
    return *dbp.metrics
}
```

### 2. Document Validation and Quality Control

```go
type DocumentValidator struct {
    minContentLength int
    maxContentLength int
    requiredFields   []string
    allowedTypes     []core.DocumentType
}

func NewDocumentValidator() *DocumentValidator {
    return &DocumentValidator{
        minContentLength: 50,    // Minimum 50 characters
        maxContentLength: 100000, // Maximum 100KB
        requiredFields:   []string{"ID", "Title", "Content", "Source"},
        allowedTypes: []core.DocumentType{
            core.DocumentTypePDF,
            core.DocumentTypeText,
            core.DocumentTypeMarkdown,
            core.DocumentTypeWeb,
            core.DocumentTypeCode,
            core.DocumentTypeJSON,
        },
    }
}

func (dv *DocumentValidator) ValidateDocument(doc core.Document) error {
    // Check required fields
    if doc.ID == "" {
        return fmt.Errorf("document ID is required")
    }
    if doc.Title == "" {
        return fmt.Errorf("document title is required")
    }
    if doc.Content == "" {
        return fmt.Errorf("document content is required")
    }
    if doc.Source == "" {
        return fmt.Errorf("document source is required")
    }
    
    // Check content length
    if len(doc.Content) < dv.minContentLength {
        return fmt.Errorf("document content too short: %d chars (minimum: %d)", 
            len(doc.Content), dv.minContentLength)
    }
    if len(doc.Content) > dv.maxContentLength {
        return fmt.Errorf("document content too long: %d chars (maximum: %d)", 
            len(doc.Content), dv.maxContentLength)
    }
    
    // Check document type
    if doc.Type != "" {
        validType := false
        for _, allowedType := range dv.allowedTypes {
            if doc.Type == allowedType {
                validType = true
                break
            }
        }
        if !validType {
            return fmt.Errorf("invalid document type: %s", doc.Type)
        }
    }
    
    // Check for duplicate IDs (would need external tracking)
    // This is a simplified check - in production, you'd check against existing documents
    
    return nil
}

func (dv *DocumentValidator) ValidateDocuments(documents []core.Document) ([]core.Document, []error) {
    var validDocs []core.Document
    var errors []error
    
    for i, doc := range documents {
        if err := dv.ValidateDocument(doc); err != nil {
            errors = append(errors, fmt.Errorf("document %d (%s): %w", i, doc.ID, err))
        } else {
            validDocs = append(validDocs, doc)
        }
    }
    
    return validDocs, errors
}
```

## Best Practices and Recommendations

### 1. Document Structure Guidelines

```go
// Best practices for document creation
func createOptimalDocument(id, title, content, source string, docType core.DocumentType) core.Document {
    return core.Document{
        ID:      id, // Use meaningful, unique IDs
        Title:   title, // Clear, descriptive titles
        Content: cleanAndOptimizeContent(content),
        Source:  source, // Always include source for attribution
        Type:    docType,
        Metadata: map[string]any{
            "created_by":    "document-processor",
            "version":       "1.0",
            "language":      detectLanguage(content),
            "word_count":    len(strings.Fields(content)),
            "char_count":    len(content),
            "content_hash":  generateContentHash(content), // For deduplication
        },
        Tags:      extractRelevantTags(content, title),
        CreatedAt: time.Now(),
    }
}

func cleanAndOptimizeContent(content string) string {
    // Remove excessive whitespace
    content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
    
    // Remove control characters
    content = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(content, "")
    
    // Normalize line endings
    content = strings.ReplaceAll(content, "\r\n", "\n")
    content = strings.ReplaceAll(content, "\r", "\n")
    
    // Trim whitespace
    content = strings.TrimSpace(content)
    
    return content
}

func extractRelevantTags(content, title string) []string {
    // Simple tag extraction based on common keywords
    var tags []string
    
    text := strings.ToLower(content + " " + title)
    
    // Technology tags
    techKeywords := map[string]string{
        "machine learning": "ml",
        "neural network":   "neural-networks",
        "deep learning":    "deep-learning",
        "python":          "python",
        "pytorch":         "pytorch",
        "tensorflow":      "tensorflow",
        "transformer":     "transformer",
        "attention":       "attention",
    }
    
    for keyword, tag := range techKeywords {
        if strings.Contains(text, keyword) {
            tags = append(tags, tag)
        }
    }
    
    // Remove duplicates
    return removeDuplicateTags(tags)
}

func removeDuplicateTags(tags []string) []string {
    seen := make(map[string]bool)
    var result []string
    
    for _, tag := range tags {
        if !seen[tag] {
            seen[tag] = true
            result = append(result, tag)
        }
    }
    
    return result
}

func generateContentHash(content string) string {
    // Simple hash for content deduplication
    hash := sha256.Sum256([]byte(content))
    return hex.EncodeToString(hash[:])
}

func detectLanguage(content string) string {
    // Simple language detection
    englishWords := []string{"the", "and", "is", "in", "to", "of", "a", "that"}
    words := strings.Fields(strings.ToLower(content))
    
    if len(words) == 0 {
        return "unknown"
    }
    
    englishCount := 0
    for _, word := range words {
        for _, englishWord := range englishWords {
            if word == englishWord {
                englishCount++
                break
            }
        }
    }
    
    if float64(englishCount)/float64(len(words)) > 0.05 {
        return "english"
    }
    
    return "unknown"
}
```

### 2. Document Batch Processor

```go
type DocumentBatchProcessor struct {
    memory              core.Memory
    batchSize          int
    maxConcurrency     int
    documentsProcessed int64
    errorCount         int64
    processingTime     time.Duration
    totalDocSize       int64
    mu                 sync.RWMutex
}

type ProcessingMetrics struct {
    DocumentsProcessed int64
    ErrorCount         int64
    ProcessingTime     time.Duration
    AverageDocSize     float64
}

func NewDocumentBatchProcessor(memory core.Memory, batchSize, maxConcurrency int) *DocumentBatchProcessor {
    return &DocumentBatchProcessor{
        memory:         memory,
        batchSize:      batchSize,
        maxConcurrency: maxConcurrency,
    }
}

func (p *DocumentBatchProcessor) ProcessDocuments(ctx context.Context, documents []core.Document) error {
    start := time.Now()
    
    // Process documents in batches
    for i := 0; i < len(documents); i += p.batchSize {
        end := i + p.batchSize
        if end > len(documents) {
            end = len(documents)
        }
        
        batch := documents[i:end]
        err := p.memory.IngestDocuments(ctx, batch)
        if err != nil {
            p.mu.Lock()
            p.errorCount++
            p.mu.Unlock()
            return fmt.Errorf("failed to process batch %d-%d: %w", i, end, err)
        }
        
        // Update metrics
        p.mu.Lock()
        p.documentsProcessed += int64(len(batch))
        for _, doc := range batch {
            p.totalDocSize += int64(len(doc.Content))
        }
        p.mu.Unlock()
    }
    
    p.mu.Lock()
    p.processingTime += time.Since(start)
    p.mu.Unlock()
    
    return nil
}

func (p *DocumentBatchProcessor) GetMetrics() ProcessingMetrics {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    avgSize := 0.0
    if p.documentsProcessed > 0 {
        avgSize = float64(p.totalDocSize) / float64(p.documentsProcessed)
    }
    
    return ProcessingMetrics{
        DocumentsProcessed: p.documentsProcessed,
        ErrorCount:         p.errorCount,
        ProcessingTime:     p.processingTime,
        AverageDocSize:     avgSize,
    }
}
```

### 3. Performance Monitoring

```go
func monitorIngestionPerformance(processor *DocumentBatchProcessor) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        metrics := processor.GetMetrics()
        
        fmt.Printf("=== Document Ingestion Metrics ===\n")
        fmt.Printf("Documents Processed: %d\n", metrics.DocumentsProcessed)
        fmt.Printf("Processing Time: %v\n", metrics.ProcessingTime)
        fmt.Printf("Error Count: %d\n", metrics.ErrorCount)
        fmt.Printf("Average Document Size: %.2f chars\n", metrics.AverageDocSize)
        
        if metrics.DocumentsProcessed > 0 && metrics.ProcessingTime > 0 {
            rate := float64(metrics.DocumentsProcessed) / metrics.ProcessingTime.Seconds()
            fmt.Printf("Processing Rate: %.2f docs/second\n", rate)
        }
        
        fmt.Println()
    }
}
```

## Conclusion

Document ingestion is a critical component of building effective knowledge bases in AgenticGoKit. Key takeaways:

- Use the current Document API with proper structure and metadata
- Implement proper validation and error handling
- Consider chunking strategies for large documents
- Use batch processing for better performance
- Monitor ingestion metrics for optimization
- Leverage advanced search capabilities for retrieval

Effective document ingestion enables agents to access and reason over comprehensive knowledge bases, significantly enhancing their capabilities.

## Next Steps

With your documents ingested, unlock their full potential:

### 🧠 **Intelligent Retrieval**
- **[RAG Implementation](rag-implementation.md)** - Build retrieval-augmented generation systems
- Use your ingested documents to provide intelligent, context-aware responses

### 🏗️ **Advanced Search**
- **[Knowledge Bases](knowledge-bases.md)** - Create comprehensive knowledge systems
- Advanced search patterns, filtering, and multi-modal content handling

### ⚡ **Performance & Scale**
- **[Memory Optimization](memory-optimization.md)** - Advanced performance tuning
- Optimize ingestion pipelines, caching, and scaling strategies

::: tip Content Strategy
📚 **Quality over Quantity**: Focus on high-quality, well-structured documents  
🏷️ **Rich Metadata**: Use comprehensive tagging and metadata for better retrieval  
🔄 **Incremental Updates**: Implement strategies for updating existing content
:::

## Foundation Topics

- **[Vector Databases](vector-databases.md)** - Ensure optimal storage backend
- **[Basic Memory Operations](basic-memory.md)** - Review memory fundamentals

## Complete Integration Example

Here's a comprehensive example showing memory integration with agents, orchestration, and real-world scenarios:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

// CustomerSupportAgent demonstrates memory integration in a real-world scenario
type CustomerSupportAgent struct {
    name   string
    memory core.Memory
    llm    core.LLMProvider
}

func NewCustomerSupportAgent(name string, memory core.Memory, llm core.LLMProvider) *CustomerSupportAgent {
    return &CustomerSupportAgent{
        name:   name,
        memory: memory,
        llm:    llm,
    }
}

func (csa *CustomerSupportAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Extract customer query and ID
    query, _ := state.Get("customer_query")
    customerID, _ := state.Get("customer_id")
    
    queryStr := query.(string)
    custID := customerID.(string)
    
    // Set customer-specific session
    sessionID := fmt.Sprintf("customer-%s", custID)
    ctx = csa.memory.SetSession(ctx, sessionID)
    
    // Store customer interaction
    err := csa.memory.AddMessage(ctx, "user", queryStr)
    if err != nil {
        log.Printf("Failed to store customer message: %v", err)
    }
    
    // Get customer context and history
    customerContext, err := csa.buildCustomerContext(ctx, custID, queryStr)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to build customer context: %w", err)
    }
    
    // Search knowledge base for relevant information
    knowledgeResults, err := csa.memory.SearchKnowledge(ctx, queryStr,
        core.WithLimit(5),
        core.WithScoreThreshold(0.7),
        core.WithTags([]string{"support", "faq", "troubleshooting"}),
    )
    if err != nil {
        log.Printf("Knowledge search failed: %v", err)
    }
    
    // Build comprehensive RAG context
    ragContext, err := csa.memory.BuildContext(ctx, queryStr,
        core.WithMaxTokens(3000),
        core.WithPersonalWeight(0.4), // Higher weight for customer history
        core.WithKnowledgeWeight(0.6),
        core.WithHistoryLimit(5),
        core.WithIncludeSources(true),
    )
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to build RAG context: %w", err)
    }
    
    // Generate personalized response
    response, err := csa.generateSupportResponse(ctx, queryStr, customerContext, ragContext)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("response generation failed: %w", err)
    }
    
    // Store response and update customer profile
    err = csa.memory.AddMessage(ctx, "assistant", response)
    if err != nil {
        log.Printf("Failed to store response: %v", err)
    }
    
    // Learn from interaction
    csa.updateCustomerProfile(ctx, custID, queryStr, response)
    
    // Return result
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("customer_id", custID)
    outputState.Set("knowledge_sources", len(knowledgeResults))
    outputState.Set("context_tokens", ragContext.TokenCount)
    
    return core.AgentResult{OutputState: outputState}, nil
}

func (csa *CustomerSupportAgent) buildCustomerContext(ctx context.Context, customerID, query string) (map[string]any, error) {
    context := make(map[string]any)
    
    // Get customer preferences
    if prefs, err := csa.memory.Recall(ctx, fmt.Sprintf("customer_%s_preferences", customerID)); err == nil {
        context["preferences"] = prefs
    }
    
    // Get customer tier/status
    if tier, err := csa.memory.Recall(ctx, fmt.Sprintf("customer_%s_tier", customerID)); err == nil {
        context["tier"] = tier
    } else {
        context["tier"] = "standard" // Default
    }
    
    // Get previous issues
    previousIssues, err := csa.memory.Query(ctx, fmt.Sprintf("customer %s issues", customerID), 3)
    if err == nil {
        context["previous_issues"] = previousIssues
    }
    
    return context, nil
}

func (csa *CustomerSupportAgent) generateSupportResponse(ctx context.Context, query string, customerContext map[string]any, ragContext *core.RAGContext) (string, error) {
    // Build comprehensive prompt
    prompt := fmt.Sprintf(`You are a helpful customer support agent. Provide personalized assistance based on the customer's context and available knowledge.

Customer Context:
- Tier: %v
- Previous Issues: %d
- Preferences: %v

Knowledge Base Context:
%s

Current Query: %s

Please provide a helpful, personalized response that:
1. Addresses the customer's specific question
2. Takes into account their tier and history
3. Uses information from the knowledge base when relevant
4. Maintains a professional and empathetic tone

Response:`, 
        customerContext["tier"],
        len(customerContext["previous_issues"].([]core.Result)),
        customerContext["preferences"],
        ragContext.ContextText,
        query)
    
    response, err := csa.llm.Generate(ctx, prompt)
    if err != nil {
        return "", err
    }
    
    // Add source attribution if available
    if len(ragContext.Sources) > 0 {
        response += fmt.Sprintf("\n\nReference: %s", ragContext.Sources[0])
    }
    
    return response, nil
}

func (csa *CustomerSupportAgent) updateCustomerProfile(ctx context.Context, customerID, query, response string) {
    // Store interaction pattern
    interaction := fmt.Sprintf("Query: %s | Response: %s", query, response)
    csa.memory.Store(ctx, interaction, "customer-interaction", customerID)
    
    // Update customer preferences based on interaction
    if contains(query, "urgent") || contains(query, "emergency") {
        csa.memory.Remember(ctx, fmt.Sprintf("customer_%s_prefers_urgent_handling", customerID), true)
    }
    
    // Track issue categories
    category := categorizeQuery(query)
    csa.memory.Store(ctx, fmt.Sprintf("Customer %s had %s issue", customerID, category), 
        "issue-tracking", customerID, category)
}

func contains(text, substr string) bool {
    return len(text) > 0 && len(substr) > 0 && 
           strings.Contains(strings.ToLower(text), strings.ToLower(substr))
}

func categorizeQuery(query string) string {
    query = strings.ToLower(query)
    if contains(query, "billing") || contains(query, "payment") {
        return "billing"
    } else if contains(query, "technical") || contains(query, "bug") {
        return "technical"
    } else if contains(query, "account") || contains(query, "login") {
        return "account"
    }
    return "general"
}

// Multi-Agent Memory Coordination Example
type MemoryCoordinator struct {
    sharedMemory core.Memory
    agents       map[string]*CustomerSupportAgent
}

func NewMemoryCoordinator(sharedMemory core.Memory) *MemoryCoordinator {
    return &MemoryCoordinator{
        sharedMemory: sharedMemory,
        agents:       make(map[string]*CustomerSupportAgent),
    }
}

func (mc *MemoryCoordinator) HandleCustomerEscalation(ctx context.Context, customerID, issue string) error {
    // Store escalation in shared memory
    escalation := fmt.Sprintf("Customer %s escalated issue: %s", customerID, issue)
    err := mc.sharedMemory.Store(ctx, escalation, "escalation", "high-priority", customerID)
    if err != nil {
        return fmt.Errorf("failed to store escalation: %w", err)
    }
    
    // Search for similar escalations
    similarIssues, err := mc.sharedMemory.SearchKnowledge(ctx, issue,
        core.WithLimit(3),
        core.WithTags([]string{"escalation", "high-priority"}),
        core.WithScoreThreshold(0.8),
    )
    if err != nil {
        log.Printf("Failed to search similar escalations: %v", err)
    }
    
    // Notify relevant agents
    for agentName, agent := range mc.agents {
        log.Printf("Notifying agent %s of escalation for customer %s", agentName, customerID)
        // In a real system, you'd send events to agents
    }
    
    log.Printf("Escalation handled: %d similar issues found", len(similarIssues))
    return nil
}

// Production deployment example
func main() {
    // Create production-ready memory system
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "memory", // Use "pgvector" for production
        Connection: "memory",
        MaxResults: 15,
        Dimensions: 1536,
        AutoEmbed:  true,
        
        // Optimized for customer support
        EnableRAG:               true,
        EnableKnowledgeBase:     true,
        KnowledgeMaxResults:     10,
        KnowledgeScoreThreshold: 0.7,
        ChunkSize:               1000,
        ChunkOverlap:            200,
        
        RAGMaxContextTokens: 3000,
        RAGPersonalWeight:   0.4, // Higher weight for customer history
        RAGKnowledgeWeight:  0.6,
        RAGIncludeSources:   true,
        
        Embedding: core.EmbeddingConfig{
            Provider: "dummy", // Use "openai" for production
            Model:    "dummy-model",
        },
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    defer memory.Close()
    
    // Populate knowledge base with support documentation
    err = populateSupportKnowledgeBase(memory)
    if err != nil {
        log.Fatalf("Failed to populate knowledge base: %v", err)
    }
    
    // Create customer support agent
    llm := &MockLLM{}
    supportAgent := NewCustomerSupportAgent("support-agent-1", memory, llm)
    
    // Create memory coordinator for multi-agent scenarios
    coordinator := NewMemoryCoordinator(memory)
    coordinator.agents["support-agent-1"] = supportAgent
    
    // Simulate customer interactions
    customers := []struct {
        id    string
        query string
    }{
        {"cust-001", "I'm having trouble with my billing statement"},
        {"cust-002", "My account login isn't working"},
        {"cust-001", "This is urgent - I need help with the same billing issue"},
    }
    
    ctx := context.Background()
    
    for _, customer := range customers {
        fmt.Printf("\n=== Customer %s Query ===\n", customer.id)
        fmt.Printf("Query: %s\n", customer.query)
        
        // Create event and state
        event := core.NewEvent(
            "support-agent-1",
            core.EventData{"customer_query": customer.query, "customer_id": customer.id},
            map[string]string{"session_id": fmt.Sprintf("customer-%s", customer.id)},
        )
        
        state := core.NewState()
        state.Set("customer_query", customer.query)
        state.Set("customer_id", customer.id)
        
        // Process customer query
        result, err := supportAgent.Run(ctx, event, state)
        if err != nil {
            log.Printf("Agent error: %v", err)
            continue
        }
        
        response, _ := result.OutputState.Get("response")
        sources, _ := result.OutputState.Get("knowledge_sources")
        tokens, _ := result.OutputState.Get("context_tokens")
        
        fmt.Printf("Response: %s\n", response)
        fmt.Printf("Knowledge Sources Used: %v\n", sources)
        fmt.Printf("Context Tokens: %v\n", tokens)
        
        // Simulate escalation for urgent issues
        if contains(customer.query, "urgent") {
            coordinator.HandleCustomerEscalation(ctx, customer.id, customer.query)
        }
    }
    
    fmt.Println("\n🎉 Memory integration example completed successfully!")
}

func populateSupportKnowledgeBase(memory core.Memory) error {
    ctx := context.Background()
    
    supportDocs := []core.Document{
        {
            ID:      "billing-faq",
            Title:   "Billing FAQ",
            Content: "Common billing questions and answers. If you're having trouble with your billing statement, check that all charges are correct and contact support if you find discrepancies.",
            Source:  "support-docs/billing-faq.md",
            Type:    core.DocumentTypeMarkdown,
            Tags:    []string{"support", "faq", "billing"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "login-troubleshooting",
            Title:   "Login Troubleshooting Guide",
            Content: "Steps to resolve login issues: 1) Check your username and password, 2) Clear browser cache, 3) Try password reset, 4) Contact support if issues persist.",
            Source:  "support-docs/login-help.md",
            Type:    core.DocumentTypeMarkdown,
            Tags:    []string{"support", "troubleshooting", "account", "login"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "escalation-procedures",
            Title:   "Escalation Procedures",
            Content: "For urgent issues: 1) Acknowledge customer urgency, 2) Gather all relevant information, 3) Escalate to senior support, 4) Follow up within 2 hours.",
            Source:  "support-docs/escalation.md",
            Type:    core.DocumentTypeMarkdown,
            Tags:    []string{"support", "escalation", "procedures", "urgent"},
            CreatedAt: time.Now(),
        },
    }
    
    return memory.IngestDocuments(ctx, supportDocs)
}

// Mock LLM for demonstration
type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, prompt string) (string, error) {
    if contains(prompt, "billing") {
        return "I understand you're having billing concerns. Let me help you resolve this issue. Based on our records and support documentation, I recommend checking your billing statement for any discrepancies and I can help you understand the charges.", nil
    } else if contains(prompt, "login") {
        return "I can help you with your login issue. Let's start by trying these troubleshooting steps: first, please verify your username and password are correct, then try clearing your browser cache. If that doesn't work, we can proceed with a password reset.", nil
    } else if contains(prompt, "urgent") {
        return "I understand this is urgent and I'm prioritizing your request. I've escalated this to our senior support team and you can expect a follow-up within 2 hours. In the meantime, let me provide immediate assistance based on your previous interactions.", nil
    }
    return "Thank you for contacting support. I'm here to help you with your inquiry. Let me search our knowledge base for the most relevant information to assist you.", nil
}
```

This comprehensive example demonstrates:

1. **Customer-Specific Memory**: Using sessions and customer IDs for personalized experiences
2. **Knowledge Base Integration**: Searching support documentation with relevant tags
3. **RAG Implementation**: Building context from customer history and knowledge base
4. **Learning and Adaptation**: Updating customer profiles based on interactions
5. **Multi-Agent Coordination**: Sharing memory across multiple agents
6. **Production Patterns**: Error handling, logging, and escalation procedures
7. **Real-World Scenarios**: Customer support use case with practical applications

## Further Reading

- [Document Processing Best Practices](../../reference/best-practices/document-processing.md)
- [API Reference: Document Types](../../reference/api/memory.md#document-types)
- [Examples: Document Ingestion](../../examples/document-ingestion/)