package memory

import (
	"context"
	"errors"
	"fmt"
	"log"

	agentflow "kunalkushwaha/agentflow/internal/core" // Import core types
)

// FallbackVectorMemory implements the agentflow.VectorMemory interface by attempting
// operations on a primary VectorMemory instance and falling back to a secondary
// instance if the primary fails.
type FallbackVectorMemory struct {
	primary   agentflow.VectorMemory
	secondary agentflow.VectorMemory
}

// NewFallbackVectorMemory creates a new FallbackVectorMemory instance.
// Both primary and secondary must be non-nil.
func NewFallbackVectorMemory(primary, secondary agentflow.VectorMemory) (*FallbackVectorMemory, error) {
	if primary == nil {
		return nil, errors.New("primary vector memory cannot be nil")
	}
	if secondary == nil {
		return nil, errors.New("secondary vector memory cannot be nil")
	}
	log.Println("FallbackVectorMemory initialized.")
	return &FallbackVectorMemory{
		primary:   primary,
		secondary: secondary,
	}, nil
}

// Store attempts to save the vector using the primary memory. If that fails,
// it logs the error and attempts to use the secondary memory.
func (m *FallbackVectorMemory) Store(ctx context.Context, id string, embedding []float32, metadata map[string]any) error {
	errPrimary := m.primary.Store(ctx, id, embedding, metadata)
	if errPrimary == nil {
		return nil // Success with primary
	}

	log.Printf("Warning: FallbackVectorMemory primary Store failed for id '%s', falling back to secondary. Primary error: %v", id, errPrimary)

	errSecondary := m.secondary.Store(ctx, id, embedding, metadata)
	if errSecondary != nil {
		// Log the secondary error as well
		log.Printf("Error: FallbackVectorMemory secondary Store also failed for id '%s'. Secondary error: %v", id, errSecondary)
		// Return a combined error or just the secondary one? Let's combine for more info.
		return fmt.Errorf("primary store failed (%w) and secondary store failed (%v)", errPrimary, errSecondary)
	}

	log.Printf("Info: FallbackVectorMemory successfully used secondary Store for id '%s'", id)
	return nil // Success with secondary
}

// Query attempts to perform the search using the primary memory. If that fails,
// it logs the error and attempts to use the secondary memory.
func (m *FallbackVectorMemory) Query(ctx context.Context, embedding []float32, topK int) ([]agentflow.QueryResult, error) {
	resultsPrimary, errPrimary := m.primary.Query(ctx, embedding, topK)
	if errPrimary == nil {
		return resultsPrimary, nil // Success with primary
	}

	log.Printf("Warning: FallbackVectorMemory primary Query failed, falling back to secondary. Primary error: %v", errPrimary)

	resultsSecondary, errSecondary := m.secondary.Query(ctx, embedding, topK)
	if errSecondary != nil {
		// Log the secondary error as well
		log.Printf("Error: FallbackVectorMemory secondary Query also failed. Secondary error: %v", errSecondary)
		// Return a combined error or just the secondary one? Combine.
		return nil, fmt.Errorf("primary query failed (%w) and secondary query failed (%v)", errPrimary, errSecondary)
	}

	log.Printf("Info: FallbackVectorMemory successfully used secondary Query")
	return resultsSecondary, nil // Success with secondary
}

// TODO: Implement Close() method if underlying memory stores need closing?
// This would require checking if primary/secondary implement an optional Closer interface.
// Example:
// func (m *FallbackVectorMemory) Close() error {
//     var errPrimary, errSecondary error
//     if closer, ok := m.primary.(interface{ Close() }); ok {
//         // Call closer.Close() and handle error
//     }
//     if closer, ok := m.secondary.(interface{ Close() }); ok {
//         // Call closer.Close() and handle error
//     }
//     // Combine errors if necessary
//     return combinedError
// }
