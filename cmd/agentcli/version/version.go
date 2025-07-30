package version

import (
	"fmt"
	"runtime"
	"time"
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
var (
	Version   = "dev"
	GitCommit = "unknown"
	GitBranch = "unknown"
	BuildDate = "unknown"
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
func GetDetailedVersionString() string {
	info := GetVersionInfo()
	
	var buildTime string
	if info.BuildDate != "unknown" {
		if t, err := time.Parse(time.RFC3339, info.BuildDate); err == nil {
			buildTime = t.Format("2006-01-02 15:04:05 UTC")
		} else {
			buildTime = info.BuildDate
		}
	} else {
		buildTime = "unknown"
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