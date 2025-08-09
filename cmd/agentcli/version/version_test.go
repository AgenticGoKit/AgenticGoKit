package version

import (
	"strings"
	"testing"
	"time"
)

func TestGetDetailedVersionString(t *testing.T) {
	tests := []struct {
		name        string
		buildDate   string
		expectTime  string
		expectError bool
	}{
		{
			name:       "RFC3339 format",
			buildDate:  "2024-01-15T10:30:45Z",
			expectTime: "2024-01-15 10:30:45 UTC",
		},
		{
			name:       "Legacy format without timezone",
			buildDate:  "2024-01-15T10:30:45",
			expectTime: "2024-01-15 10:30:45 UTC",
		},
		{
			name:       "Default unknown date",
			buildDate:  "1970-01-01T00:00:00Z",
			expectTime: "unknown",
		},
		{
			name:       "Invalid date format",
			buildDate:  "invalid-date",
			expectTime: "invalid-date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			originalBuildDate := BuildDate
			defer func() { BuildDate = originalBuildDate }()

			// Set test build date
			BuildDate = tt.buildDate

			// Get detailed version string
			result := GetDetailedVersionString()

			// Check if expected time appears in result
			if !strings.Contains(result, tt.expectTime) {
				t.Errorf("Expected build time '%s' not found in result:\n%s", tt.expectTime, result)
			}

			// Verify the result contains all expected fields
			expectedFields := []string{
				"agentcli version",
				"Git commit:",
				"Git branch:",
				"Build date:",
				"Go version:",
				"Platform:",
				"Compiler:",
			}

			for _, field := range expectedFields {
				if !strings.Contains(result, field) {
					t.Errorf("Expected field '%s' not found in result:\n%s", field, result)
				}
			}
		})
	}
}

func TestVersionInfoConsistency(t *testing.T) {
	// Test that GetVersionInfo returns consistent data
	info1 := GetVersionInfo()
	info2 := GetVersionInfo()

	if info1.Version != info2.Version {
		t.Errorf("Version inconsistent: %s != %s", info1.Version, info2.Version)
	}

	if info1.GitCommit != info2.GitCommit {
		t.Errorf("GitCommit inconsistent: %s != %s", info1.GitCommit, info2.GitCommit)
	}

	if info1.BuildDate != info2.BuildDate {
		t.Errorf("BuildDate inconsistent: %s != %s", info1.BuildDate, info2.BuildDate)
	}
}

func TestBuildDateFormats(t *testing.T) {
	// Test that our build systems produce parseable dates
	testDates := []string{
		"2024-01-15T10:30:45Z",     // Makefile/build.sh format
		"2024-01-15T10:30:45Z",     // build.ps1 format
		"1970-01-01T00:00:00Z",     // Default fallback
	}

	for _, dateStr := range testDates {
		t.Run("Parse_"+dateStr, func(t *testing.T) {
			// Should parse as RFC3339
			if _, err := time.Parse(time.RFC3339, dateStr); err != nil {
				t.Errorf("Date '%s' should parse as RFC3339: %v", dateStr, err)
			}
		})
	}
}

func TestGetVersionString(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalGitCommit := GitCommit
	originalBuildDate := BuildDate
	defer func() {
		Version = originalVersion
		GitCommit = originalGitCommit
		BuildDate = originalBuildDate
	}()

	// Test dev version
	Version = "dev"
	GitCommit = "abc123def456"
	BuildDate = "2024-01-15T10:30:45Z"

	result := GetVersionString()
	expected := "agentcli dev (commit: abc123de, built: 2024-01-15T10:30:45Z)"
	if result != expected {
		t.Errorf("Expected: %s, Got: %s", expected, result)
	}

	// Test release version
	Version = "v1.0.0"
	result = GetVersionString()
	expected = "agentcli v1.0.0"
	if result != expected {
		t.Errorf("Expected: %s, Got: %s", expected, result)
	}
}