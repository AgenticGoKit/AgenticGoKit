package version

import (
	"fmt"
	"runtime"
	"time"
)

// Date format constants for consistent parsing
const (
	// RFC3339Format is the standard format used by all build systems
	// Example: "2006-01-02T15:04:05Z"
	RFC3339Format = time.RFC3339
	
	// LegacyFormat is a fallback for builds without timezone info
	// Example: "2006-01-02T15:04:05"
	LegacyFormat = "2006-01-02T15:04:05"
	
	// DisplayFormat is the human-readable format for version output
	// Example: "2006-01-02 15:04:05 UTC"
	DisplayFormat = "2006-01-02 15:04:05 UTC"
)

// VersionInfo contains all version-related information
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
	Compiler  string `json:"compiler"`
}

// Build-time variables set via ldflags
// BuildDate should be in RFC3339 format (e.g., "2006-01-02T15:04:05Z")
var (
	Version   = "dev"
	GitCommit = "unknown"
	GitBranch = "unknown"
	BuildDate = "1970-01-01T00:00:00Z" // RFC3339 format default instead of "unknown"
)

// GetVersionInfo returns comprehensive version information
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Compiler:  runtime.Compiler,
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	info := GetVersionInfo()
	if info.Version == "dev" {
		commit := info.GitCommit
		if len(commit) > 8 {
			commit = commit[:8]
		}
		return fmt.Sprintf("agentcli %s (commit: %s, built: %s)",
			info.Version, commit, info.BuildDate)
	}
	return fmt.Sprintf("agentcli %s", info.Version)
}

// GetDetailedVersionString returns a detailed multi-line version string
// BuildDate is expected to be in RFC3339 format (e.g., "2006-01-02T15:04:05Z")
func GetDetailedVersionString() string {
	info := GetVersionInfo()
	
	var buildTime string
	// Try to parse RFC3339 format first (standard format from build systems)
	if t, err := time.Parse(RFC3339Format, info.BuildDate); err == nil {
		buildTime = t.Format(DisplayFormat)
	} else {
		// Fallback: try parsing without timezone (for legacy builds)
		if t, err := time.Parse(LegacyFormat, info.BuildDate); err == nil {
			buildTime = t.Format(DisplayFormat)
		} else {
			// If all parsing fails, use the raw value
			if info.BuildDate == "1970-01-01T00:00:00Z" {
				buildTime = "unknown"
			} else {
				buildTime = info.BuildDate
			}
		}
	}
	
	return fmt.Sprintf(`agentcli version %s
Git commit: %s
Git branch: %s
Build date: %s
Go version: %s
Platform: %s
Compiler: %s`,
		info.Version,
		info.GitCommit,
		info.GitBranch,
		buildTime,
		info.GoVersion,
		info.Platform,
		info.Compiler,
	)
}