package scaffold

import (
	"strings"
	"testing"
)

func TestIsValidSemanticVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"v1.0.0", true},
		{"1.0.0", true},
		{"v1.2.3", true},
		{"1.2.3", true},
		{"v1.0.0-alpha", true},
		{"v1.0.0-alpha.1", true},
		{"v1.0.0+build.1", true},
		{"v1.0.0-alpha+build.1", true},
		{"dev", false},
		{"latest", false},
		{"v1.0", false},
		{"1.0", false},
		{"v1", false},
		{"1", false},
		{"", false},
		{"invalid", false},
		{"v1.0.0.0", false}, // Too many parts
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := isValidSemanticVersion(tt.version)
			if result != tt.valid {
				t.Errorf("isValidSemanticVersion(%q) = %v, want %v", tt.version, result, tt.valid)
			}
		})
	}
}

func TestGetAgenticGoKitVersionWithFallback(t *testing.T) {
	version, source := GetAgenticGoKitVersionWithFallback()
	
	// Version should not be empty
	if version == "" {
		t.Error("GetAgenticGoKitVersionWithFallback() returned empty version")
	}
	
	// Version should start with 'v'
	if !strings.HasPrefix(version, "v") {
		t.Errorf("Version %q should start with 'v'", version)
	}
	
	// Version should be valid semantic version
	if !isValidSemanticVersion(version) {
		t.Errorf("Version %q should be valid semantic version", version)
	}
	
	// Source should be one of the expected values
	validSources := []string{"cli-version", "github-api", "fallback"}
	validSource := false
	for _, validSrc := range validSources {
		if source == validSrc {
			validSource = true
			break
		}
	}
	
	if !validSource {
		t.Errorf("Source %q should be one of %v", source, validSources)
	}
	
	t.Logf("Using version %s from source: %s", version, source)
}

func TestAgenticGoKitVersionIsValid(t *testing.T) {
	// Test that the global variable is properly initialized
	if AgenticGoKitVersion == "" {
		t.Error("AgenticGoKitVersion should not be empty")
	}
	
	if !strings.HasPrefix(AgenticGoKitVersion, "v") {
		t.Errorf("AgenticGoKitVersion %q should start with 'v'", AgenticGoKitVersion)
	}
	
	if !isValidSemanticVersion(AgenticGoKitVersion) {
		t.Errorf("AgenticGoKitVersion %q should be valid semantic version", AgenticGoKitVersion)
	}
	
	t.Logf("AgenticGoKitVersion: %s", AgenticGoKitVersion)
}

func TestVersionPrefixHandling(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.0.0", "v1.0.0"},
		{"v1.0.0", "v1.0.0"},
		{"2.1.3", "v2.1.3"},
		{"v2.1.3", "v2.1.3"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Simulate the version prefix handling logic
			version := tt.input
			if !strings.HasPrefix(version, "v") {
				version = "v" + version
			}
			
			if version != tt.expected {
				t.Errorf("Version prefix handling: input %q, got %q, want %q", tt.input, version, tt.expected)
			}
		})
	}
}