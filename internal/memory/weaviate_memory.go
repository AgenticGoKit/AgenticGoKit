package memory

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	agentflow "github.com/kunalkushwaha/agentflow/internal/core" // Import core types

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/schema"
)

// WeaviateMemory implements the agentflow.VectorMemory interface using Weaviate.
type WeaviateMemory struct {
	client    *weaviate.Client
	className string // The Weaviate class name to store vectors in
	// TODO: Add configuration for consistency level, batching, etc.
}

// WeaviateConfig holds configuration options for the Weaviate client.
type WeaviateConfig struct {
	Host       string // e.g., "localhost:8080" or "my-cluster.weaviate.network"
	Scheme     string // e.g., "http" or "https"
	APIKey     string // Optional API Key for authentication
	ClassName  string // The Weaviate class name to use (e.g., "MemoryChunk")
	Dimensions int    // The expected dimension of the vectors
}

// NewWeaviateMemory creates a new WeaviateMemory instance and ensures the class schema exists.
func NewWeaviateMemory(ctx context.Context, config WeaviateConfig) (*WeaviateMemory, error) {
	if config.ClassName == "" {
		return nil, errors.New("Weaviate class name cannot be empty")
	}
	if config.Dimensions <= 0 {
		return nil, errors.New("vector dimensions must be positive")
	}

	clientConfig := weaviate.Config{
		Host:   config.Host,
		Scheme: config.Scheme,
	}
	if config.APIKey != "" {
		clientConfig.AuthConfig = auth.ApiKey{Value: config.APIKey}
	}

	client, err := weaviate.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Weaviate client: %w", err)
	}

	mem := &WeaviateMemory{
		client:    client,
		className: config.ClassName,
	}

	// Ensure the class exists
	err = mem.ensureClassExists(ctx, config.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure Weaviate class '%s': %w", config.ClassName, err)
	}

	log.Printf("WeaviateMemory initialized for class '%s' at %s://%s", config.ClassName, config.Scheme, config.Host)
	return mem, nil
}

// ensureClassExists checks if the configured class exists in Weaviate and creates it if not.
func (m *WeaviateMemory) ensureClassExists(ctx context.Context, dimensions int) error {
	exists, err := m.client.Schema().ClassGetter().WithClassName(m.className).Do(ctx)
	if err != nil {
		// Weaviate might return an error that indicates non-existence, check specific error if possible
		// For now, assume any error means we might need to create it or it's a real connection issue.
		log.Printf("Warning: Failed to check if class '%s' exists (may attempt creation): %v", m.className, err)
	}
	if exists != nil {
		log.Printf("Weaviate class '%s' already exists.", m.className)
		// TODO: Validate existing schema matches expected dimensions/config?
		return nil
	}

	log.Printf("Weaviate class '%s' not found, attempting to create...", m.className)
	classObj := &models.Class{
		Class:       m.className,
		Description: "Stores vector embeddings and metadata for Agentflow memory",
		// Use a generic vector index config, assuming embeddings are generated externally
		Vectorizer: "none", // Important: We provide vectors directly
		VectorIndexConfig: map[string]interface{}{
			// Using HNSW as a common default, adjust as needed
			"distance": "cosine", // Common distance metric for embeddings
			// Other HNSW parameters (efConstruction, maxConnections, ef) can be tuned
		},
		Properties: []*models.Property{{
			Name:        "item_id", // Store our internal ID separately
			DataType:    []string{string(schema.DataTypeText)},
			Description: "The unique ID provided by Agentflow",
			// Index this field for potential filtering later
			IndexFilterable: &[]bool{true}[0],
			IndexSearchable: &[]bool{false}[0], // Not typically needed for search
		},
		// We will store other metadata dynamically within the object properties
		},
	}

	err = m.client.Schema().ClassCreator().WithClass(classObj).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Weaviate class '%s': %w", m.className, err)
	}
	log.Printf("Successfully created Weaviate class '%s'", m.className)
	return nil
}

// Store saves a vector embedding and its associated metadata to Weaviate.
func (m *WeaviateMemory) Store(ctx context.Context, id string, embedding []float32, metadata map[string]any) error {
	if id == "" {
		return errors.New("item ID cannot be empty")
	}
	if len(embedding) == 0 {
		return errors.New("embedding cannot be empty")
	}

	// Weaviate properties must be map[string]interface{}
	properties := make(map[string]interface{}, len(metadata)+1)
	properties["item_id"] = id // Store our ID in a specific property

	// Copy metadata, ensuring keys are valid Weaviate property names (basic sanitization)
	for k, v := range metadata {
		propName := m.sanitizePropertyName(k)
		if propName == "item_id" || propName == "id" || propName == "_vector" { // Avoid reserved/conflicting names
			log.Printf("Warning: Skipping metadata key '%s' as it conflicts with reserved Weaviate names", k)
			continue
		}
		// TODO: Add more robust type checking/conversion if needed
		properties[propName] = v
	}

	// Weaviate uses UUIDs internally, but we can provide our own vector
	// We will use the provided `id` to potentially check for existence first if needed,
	// but a simple Create/Replace is often sufficient.

	_, err := m.client.Data().Creator().
		WithClassName(m.className).
		WithProperties(properties).
		WithVector(embedding).
		// WithID(generateUUID(id)) // Optionally generate a deterministic UUID from our ID
		Do(ctx)

	if err != nil {
		// TODO: Implement update logic if Create fails due to existence?
		// Weaviate's default is create-or-replace if an ID is provided, but we aren't providing one here.
		// A common pattern is Delete + Create or using WithID and letting Weaviate handle upsert.
		// For simplicity now, we just create. If it needs update, delete first.
		// Let's try delete + create for upsert semantics.
		log.Printf("Initial store failed for ID '%s', attempting delete then create: %v", id, err)

		// Attempt to delete based on our custom item_id property
		delErr := m.deleteByID(ctx, id)
		if delErr != nil {
			log.Printf("Failed to delete existing item with item_id '%s' during upsert attempt: %v", id, delErr)
			// Return the original creation error
			return fmt.Errorf("failed to store item '%s' and failed to delete potential existing item: %w", id, err)
		}

		// Retry creation after deletion
		_, err = m.client.Data().Creator().
			WithClassName(m.className).
			WithProperties(properties).
			WithVector(embedding).
			Do(ctx)

		if err != nil {
			return fmt.Errorf("failed to store item '%s' even after delete attempt: %w", id, err)
		}
		log.Printf("Successfully stored item '%s' via delete and create", id)
	}

	return nil
}

// deleteByID attempts to delete an object based on the custom 'item_id' property.
func (m *WeaviateMemory) deleteByID(ctx context.Context, itemID string) error {
	where := filters.Where().
		WithPath([]string{"item_id"}).
		WithOperator(filters.Equal).
		WithValueText(itemID)

	result, err := m.client.Batch().ObjectsBatchDeleter().
		WithClassName(m.className).
		WithWhere(where).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("batch delete request failed for item_id '%s': %w", itemID, err)
	}

	// Check results for errors during deletion
	if result != nil && result.Results != nil {
		if result.Results.Failed > 0 {
			// Log details if available
			errMsg := fmt.Sprintf("failed to delete %d objects", result.Results.Failed)
			if len(result.Results.Objects) > 0 {
				// Find first error message
				for _, obj := range result.Results.Objects {
					if obj.Errors != nil && len(obj.Errors.Error) > 0 {
						errMsg = fmt.Sprintf("%s - first error: %s", errMsg, obj.Errors.Error[0].Message)
						break
					}
				}
			}
			return errors.New(errMsg)
		}
		log.Printf("Batch delete completed for item_id '%s': matched=%d, successful=%d", itemID, result.Results.Matches, result.Results.Successful)
	}

	return nil
}

// Query performs a vector similarity search in Weaviate.
func (m *WeaviateMemory) Query(ctx context.Context, embedding []float32, topK int) ([]agentflow.QueryResult, error) {
	if len(embedding) == 0 {
		return nil, errors.New("query embedding cannot be empty")
	}
	if topK <= 0 {
		return nil, errors.New("topK must be positive")
	}

	// Define the fields to retrieve (our item_id and all other metadata)
	// We need to know the metadata fields beforehand or use GraphQL introspection,
	// or retrieve the whole object and parse. Let's retrieve known fields + _additional fields.
	fields := []graphql.Field{
		{Name: "item_id"}, // Our custom ID field
		{Name: "_additional", Fields: []graphql.Field{
			{Name: "distance"}, // Similarity score (Weaviate calls it distance)
			{Name: "id"},       // Weaviate's internal UUID
			// {Name: "vector"}, // Optionally retrieve the vector
		}},
		// TODO: How to retrieve all *other* metadata properties dynamically?
		// We might need to list all properties from schema or use a wildcard if available.
		// For now, assume metadata needs to be explicitly listed or parsed from a raw object.
		// Let's try retrieving the full object and extracting later.
	}

	// Perform the nearVector search
	search := m.client.GraphQL().NearVectorArgBuilder().
		WithVector(embedding)

	resp, err := m.client.GraphQL().Get().
		WithClassName(m.className).
		WithFields(fields...). // Request item_id and _additional fields
		WithNearVector(search).
		WithLimit(topK).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("Weaviate query failed: %w", err)
	}

	// Process the response
	results := []agentflow.QueryResult{}
	getRespData, ok := resp.Data["Get"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format from Weaviate: 'Get' field missing or not a map")
	}

	classData, ok := getRespData[m.className].([]interface{})
	if !ok {
		// It's okay if the class data is missing (no results), but not if it's the wrong type
		if getRespData[m.className] != nil {
			return nil, fmt.Errorf("unexpected response format from Weaviate: '%s' field not a slice", m.className)
		}
		// No results found
		return results, nil
	}

	for _, item := range classData {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("Warning: Skipping result item, expected map[string]interface{}, got %T", item)
			continue
		}

		qr := agentflow.QueryResult{
			Metadata: make(map[string]any),
		}

		// Extract item_id
		if idVal, ok := itemMap["item_id"]; ok {
			if idStr, ok := idVal.(string); ok {
				qr.ID = idStr
			}
		}

		// Extract score (distance) from _additional
		if additional, ok := itemMap["_additional"].(map[string]interface{}); ok {
			if distVal, ok := additional["distance"]; ok {
				// Weaviate distance is float64, convert to float32 for our interface
				if distF64, ok := distVal.(float64); ok {
					qr.Score = float32(distF64)
				}
			}
			// weaviateUUID, _ := additional["id"].(string) // Store if needed
		}

		// Extract other properties as metadata
		for key, val := range itemMap {
			if key != "item_id" && key != "_additional" {
				qr.Metadata[m.desanitizePropertyName(key)] = val // Restore original key name if possible
			}
		}

		// Only add if we successfully got our ID
		if qr.ID != "" {
			results = append(results, qr)
		} else {
			log.Printf("Warning: Skipping result item without a valid 'item_id': %+v", itemMap)
		}
	}

	return results, nil
}

// sanitizePropertyName converts a metadata key into a valid Weaviate property name.
// Weaviate property names must start with a lowercase letter and contain only [a-zA-Z0-9_].
// This is a basic example; more robust sanitization might be needed.
func (m *WeaviateMemory) sanitizePropertyName(key string) string {
	// Replace invalid characters with underscore
	sanitized := strings.Builder{}
	for i, r := range key {
		if i == 0 {
			if r >= 'a' && r <= 'z' {
				sanitized.WriteRune(r)
			} else {
				// Prepend if first char is invalid
				sanitized.WriteString("prop_")
				if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
					sanitized.WriteRune(r)
				} else {
					sanitized.WriteRune('_')
				}
			}
		} else {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
				sanitized.WriteRune(r)
			} else {
				sanitized.WriteRune('_')
			}
		}
	}
	// Ensure first char is lowercase if it wasn't handled by prepending "prop_"
	final := sanitized.String()
	if len(final) > 0 && final[0] >= 'A' && final[0] <= 'Z' && !strings.HasPrefix(final, "prop_") {
		final = strings.ToLower(string(final[0])) + final[1:]
	}

	// Handle case where key was initially empty or all invalid chars
	if final == "" || final == "prop_" {
		return "prop_unnamed"
	}

	return final
}

// desanitizePropertyName is a placeholder. A robust implementation would require
// storing the original mapping or using a reversible sanitization scheme.
func (m *WeaviateMemory) desanitizePropertyName(sanitizedKey string) string {
	// Simple placeholder - assumes no sanitization was actually needed or reverses basic cases
	if strings.HasPrefix(sanitizedKey, "prop_") {
		// Cannot reliably reverse this without stored mapping
		return sanitizedKey // Return sanitized for now
	}
	return sanitizedKey
}

// TODO: Implement generateUUID(id string) if deterministic UUIDs are needed for WithID()
