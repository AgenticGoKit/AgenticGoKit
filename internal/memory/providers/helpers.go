// Package providers contains internal memory provider implementations.
package providers

import (
	"fmt"
	"strings"
	"time"
)

// Helper functions shared across all memory providers

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func calculateScore(content, query string) float32 {
	content = strings.ToLower(content)
	query = strings.ToLower(query)
	
	if content == query {
		return 1.0
	}
	
	if strings.Contains(content, query) {
		return 0.8
	}
	
	// Simple word matching
	queryWords := strings.Fields(query)
	contentWords := strings.Fields(content)
	matches := 0
	
	for _, qw := range queryWords {
		for _, cw := range contentWords {
			if strings.Contains(cw, qw) || strings.Contains(qw, cw) {
				matches++
				break
			}
		}
	}
	
	if len(queryWords) > 0 {
		return float32(matches) / float32(len(queryWords)) * 0.6
	}
	
	return 0.0
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsAnyTag(tags []string, searchTags []string) bool {
	for _, tag := range tags {
		for _, searchTag := range searchTags {
			if tag == searchTag {
				return true
			}
		}
	}
	return false
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

func estimateTokenCount(text string) int {
	// Rough estimation: 1 token per 4 characters
	return len(text) / 4
}