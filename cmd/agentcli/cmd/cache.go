package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/spf13/cobra"
)

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage and monitor MCP tool result caches",
	Long: `The cache command provides comprehensive management and monitoring capabilities 
for MCP tool result caches. This includes viewing cache statistics, managing cache 
entries, and optimizing cache performance.

Key capabilities:
  * View cache statistics and performance metrics
  * List cached tool results with filtering options
  * Clear specific cache entries or entire caches
  * Invalidate cache entries by pattern matching
  * Monitor cache hit rates and memory usage
  * Warm up caches with frequently used tools
  * Export and import cache configurations

Examples:
  # View overall cache statistics
  agentcli cache stats

  # List all cached entries
  agentcli cache list

  # List cached entries for a specific tool
  agentcli cache list --tool web_search

  # Clear all caches
  agentcli cache clear --all

  # Clear caches for a specific server
  agentcli cache clear --server web-service

  # Invalidate caches matching a pattern
  agentcli cache invalidate "web-*"

  # Show detailed cache information
  agentcli cache info --tool web_search --server web-service`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no subcommand is provided
		cmd.Help()
	},
}

// Cache command flags
var (
	cacheServer  string
	cacheTool    string
	cacheAll     bool
	cachePattern string
	cacheFormat  string
	cacheVerbose bool
	cacheLimit   int
	cacheSortBy  string
)

// cacheStatsCmd shows cache statistics
var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display cache performance statistics",
	Long: `Display comprehensive cache performance statistics including hit rates,
memory usage, cache sizes, and performance metrics for all MCP tool caches.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize cache manager
		cacheManager, err := initializeCacheManager()
		if err != nil {
			return fmt.Errorf("failed to initialize cache manager: %w", err)
		}

		// Get global cache statistics
		stats, err := cacheManager.GetGlobalStats(ctx)
		if err != nil {
			return fmt.Errorf("failed to get cache statistics: %w", err)
		}

		// Display statistics based on format
		switch cacheFormat {
		case "json":
			return displayStatsJSON(stats)
		case "table":
			return displayStatsTable(stats)
		default:
			return displayStatsDefault(stats)
		}
	},
}

// cacheListCmd lists cache entries
var cacheListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cached tool results",
	Long: `List cached tool results with optional filtering by server, tool, or pattern.
Displays cache keys, TTL information, access counts, and result summaries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize cache manager
		cacheManager, err := initializeCacheManager()
		if err != nil {
			return fmt.Errorf("failed to initialize cache manager: %w", err)
		}

		// Get cache entries (this would require extending the interface)
		entries, err := getCacheEntries(ctx, cacheManager)
		if err != nil {
			return fmt.Errorf("failed to get cache entries: %w", err)
		}

		// Apply filters
		filteredEntries := filterCacheEntries(entries)

		// Sort entries
		sortCacheEntries(filteredEntries)

		// Display entries
		return displayCacheEntries(filteredEntries)
	},
}

// cacheClearCmd clears cache entries
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear cache entries",
	Long: `Clear cache entries with options to target specific servers, tools, or clear all caches.
Use with caution as this will remove cached results and may impact performance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize cache manager
		cacheManager, err := initializeCacheManager()
		if err != nil {
			return fmt.Errorf("failed to initialize cache manager: %w", err)
		}

		// Confirm destructive operation
		if !confirmClearOperation() {
			fmt.Println("Cache clear operation cancelled.")
			return nil
		}

		// Clear caches based on flags
		if cacheAll {
			return clearAllCaches(ctx, cacheManager)
		} else if cacheServer != "" {
			return clearServerCaches(ctx, cacheManager, cacheServer)
		} else if cacheTool != "" {
			return clearToolCaches(ctx, cacheManager, cacheTool)
		}

		return fmt.Errorf("please specify --all, --server, or --tool")
	},
}

// cacheInvalidateCmd invalidates cache entries by pattern
var cacheInvalidateCmd = &cobra.Command{
	Use:   "invalidate [pattern]",
	Short: "Invalidate cache entries matching a pattern",
	Long: `Invalidate cache entries that match the specified pattern. Patterns can include
wildcards (*) and can target servers, tools, or specific cache keys.

Examples:
  agentcli cache invalidate "web-*"          # All web-service caches
  agentcli cache invalidate "*:web_search"   # All web_search tools
  agentcli cache invalidate "nlp-service:*"  # All tools from nlp-service`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		pattern := args[0]

		// Initialize cache manager
		cacheManager, err := initializeCacheManager()
		if err != nil {
			return fmt.Errorf("failed to initialize cache manager: %w", err)
		}

		// Invalidate by pattern
		err = cacheManager.InvalidateByPattern(ctx, pattern)
		if err != nil {
			return fmt.Errorf("failed to invalidate cache entries: %w", err)
		}

		fmt.Printf("✅ Successfully invalidated cache entries matching pattern: %s\n", pattern)
		return nil
	},
}

// cacheInfoCmd shows detailed cache information
var cacheInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show detailed cache information",
	Long: `Show detailed information about specific cache instances, including configuration,
statistics, memory usage, and cached entries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize cache manager
		cacheManager, err := initializeCacheManager()
		if err != nil {
			return fmt.Errorf("failed to initialize cache manager: %w", err)
		}

		// Show cache information
		return showCacheInfo(ctx, cacheManager)
	},
}

// cacheWarmCmd warms up caches
var cacheWarmCmd = &cobra.Command{
	Use:   "warm",
	Short: "Warm up caches with frequently used tools",
	Long: `Pre-populate caches by executing frequently used tools with common parameters.
This can improve initial performance by ensuring commonly accessed results are cached.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize cache manager
		cacheManager, err := initializeCacheManager()
		if err != nil {
			return fmt.Errorf("failed to initialize cache manager: %w", err)
		}

		// Warm up caches
		return warmUpCaches(ctx, cacheManager)
	},
}

// Initialize cache manager (placeholder - would integrate with actual factory)
func initializeCacheManager() (core.MCPCacheManager, error) {
	// This would integrate with the actual cache manager factory
	// For now, return an error indicating the feature needs configuration
	return nil, fmt.Errorf("cache manager not configured - please ensure MCP agent with caching is running")
}

// Display functions
func displayStatsJSON(stats core.MCPCacheStats) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func displayStatsTable(stats core.MCPCacheStats) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "METRIC\tVALUE\tUNIT")
	fmt.Fprintln(w, "------\t-----\t----")
	fmt.Fprintf(w, "Total Keys\t%d\tkeys\n", stats.TotalKeys)
	fmt.Fprintf(w, "Hit Count\t%d\thits\n", stats.HitCount)
	fmt.Fprintf(w, "Miss Count\t%d\tmisses\n", stats.MissCount)
	fmt.Fprintf(w, "Hit Rate\t%.2f%%\tpercentage\n", stats.HitRate)
	fmt.Fprintf(w, "Evictions\t%d\tevictions\n", stats.EvictionCount)
	fmt.Fprintf(w, "Total Size\t%s\tbytes\n", formatBytes(stats.TotalSize))
	fmt.Fprintf(w, "Avg Latency\t%v\tduration\n", stats.AverageLatency)
	fmt.Fprintf(w, "Last Cleanup\t%v\ttime\n", stats.LastCleanup.Format(time.RFC3339))

	return nil
}

func displayStatsDefault(stats core.MCPCacheStats) error {
	fmt.Println("🗂️  MCP Cache Statistics")
	fmt.Println("========================")
	fmt.Printf("📊 Cache Performance:\n")
	fmt.Printf("   • Total Keys: %d\n", stats.TotalKeys)
	fmt.Printf("   • Hit Rate: %.2f%% (%d hits, %d misses)\n", stats.HitRate, stats.HitCount, stats.MissCount)
	fmt.Printf("   • Average Latency: %v\n", stats.AverageLatency)
	fmt.Printf("\n💾 Memory Usage:\n")
	fmt.Printf("   • Total Size: %s\n", formatBytes(stats.TotalSize))
	fmt.Printf("   • Evictions: %d\n", stats.EvictionCount)
	fmt.Printf("\n🧹 Maintenance:\n")
	fmt.Printf("   • Last Cleanup: %v\n", stats.LastCleanup.Format(time.RFC3339))

	return nil
}

// Cache entry management (these would require extending the cache interface)
type CacheEntry struct {
	Key         core.MCPCacheKey
	Result      core.MCPToolResult
	Timestamp   time.Time
	TTL         time.Duration
	AccessCount int
	Size        int64
}

func getCacheEntries(ctx context.Context, cacheManager core.MCPCacheManager) ([]CacheEntry, error) {
	// This would require extending the MCPCacheManager interface to list entries
	// For now, return placeholder data
	return []CacheEntry{}, fmt.Errorf("cache entry listing not yet implemented")
}

func filterCacheEntries(entries []CacheEntry) []CacheEntry {
	var filtered []CacheEntry

	for _, entry := range entries {
		// Apply server filter
		if cacheServer != "" && entry.Key.ServerName != cacheServer {
			continue
		}

		// Apply tool filter
		if cacheTool != "" && entry.Key.ToolName != cacheTool {
			continue
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

func sortCacheEntries(entries []CacheEntry) {
	switch cacheSortBy {
	case "time":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.After(entries[j].Timestamp)
		})
	case "access":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].AccessCount > entries[j].AccessCount
		})
	case "size":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Size > entries[j].Size
		})
	default: // name
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Key.ToolName < entries[j].Key.ToolName
		})
	}
}

func displayCacheEntries(entries []CacheEntry) error {
	if len(entries) == 0 {
		fmt.Println("No cache entries found matching the specified criteria.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "SERVER\tTOOL\tAGE\tTTL\tACCESS\tSIZE\tHASH")
	fmt.Fprintln(w, "------\t----\t---\t---\t------\t----\t----")

	for i, entry := range entries {
		if cacheLimit > 0 && i >= cacheLimit {
			fmt.Printf("\n... and %d more entries (use --limit to see more)\n", len(entries)-i)
			break
		}

		age := time.Since(entry.Timestamp)
		ttlRemaining := entry.TTL - age

		fmt.Fprintf(w, "%s\t%s\t%v\t%v\t%d\t%s\t%s\n",
			entry.Key.ServerName,
			entry.Key.ToolName,
			formatDuration(age),
			formatDuration(ttlRemaining),
			entry.AccessCount,
			formatBytes(entry.Size),
			entry.Key.Hash[:8])
	}

	return nil
}

// Clear operations
func confirmClearOperation() bool {
	if cacheAll {
		fmt.Print("⚠️  This will clear ALL cache entries. Are you sure? (y/N): ")
	} else {
		fmt.Print("⚠️  This will clear cache entries. Are you sure? (y/N): ")
	}

	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func clearAllCaches(ctx context.Context, cacheManager core.MCPCacheManager) error {
	// This would require extending the interface to support clearing all caches
	return fmt.Errorf("clear all caches not yet implemented")
}

func clearServerCaches(ctx context.Context, cacheManager core.MCPCacheManager, server string) error {
	// Clear all caches for a specific server
	err := cacheManager.InvalidateByPattern(ctx, server)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Cleared all cache entries for server: %s\n", server)
	return nil
}

func clearToolCaches(ctx context.Context, cacheManager core.MCPCacheManager, tool string) error {
	// Clear all caches for a specific tool
	pattern := "*:" + tool
	err := cacheManager.InvalidateByPattern(ctx, pattern)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Cleared all cache entries for tool: %s\n", tool)
	return nil
}

// Cache information display
func showCacheInfo(ctx context.Context, cacheManager core.MCPCacheManager) error {
	// Get statistics
	stats, err := cacheManager.GetGlobalStats(ctx)
	if err != nil {
		return err
	}

	fmt.Println("🔍 Detailed Cache Information")
	fmt.Println("=============================")

	// Show basic stats
	displayStatsDefault(stats)

	// Show per-tool/server breakdown if verbose
	if cacheVerbose {
		fmt.Println("\n📋 Cache Breakdown:")
		// This would show per-tool and per-server statistics
		fmt.Println("   (Detailed breakdown not yet implemented)")
	}

	return nil
}

// Cache warming
func warmUpCaches(ctx context.Context, cacheManager core.MCPCacheManager) error {
	fmt.Println("🔥 Warming up caches...")
	fmt.Println("This feature is not yet implemented.")
	fmt.Println("Future implementation will:")
	fmt.Println("   • Execute common tool combinations")
	fmt.Println("   • Pre-populate frequently accessed results")
	fmt.Println("   • Use machine learning to predict cache needs")

	return nil
}

// Utility functions
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}

	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.0fh", d.Hours())
	} else {
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	}
}

func init() {
	// Add cache command to root
	rootCmd.AddCommand(cacheCmd)

	// Add subcommands
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheInvalidateCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
	cacheCmd.AddCommand(cacheWarmCmd)

	// Global cache flags
	cacheCmd.PersistentFlags().StringVar(&cacheFormat, "format", "default", "Output format (default, table, json)")
	cacheCmd.PersistentFlags().BoolVar(&cacheVerbose, "verbose", false, "Show verbose output")

	// Stats command flags
	cacheStatsCmd.Flags().StringVar(&cacheFormat, "format", "default", "Output format (default, table, json)")

	// List command flags
	cacheListCmd.Flags().StringVar(&cacheServer, "server", "", "Filter by server name")
	cacheListCmd.Flags().StringVar(&cacheTool, "tool", "", "Filter by tool name")
	cacheListCmd.Flags().IntVar(&cacheLimit, "limit", 50, "Limit number of entries shown (0 for no limit)")
	cacheListCmd.Flags().StringVar(&cacheSortBy, "sort", "name", "Sort by (name, time, access, size)")

	// Clear command flags
	cacheClearCmd.Flags().BoolVar(&cacheAll, "all", false, "Clear all cache entries")
	cacheClearCmd.Flags().StringVar(&cacheServer, "server", "", "Clear entries for specific server")
	cacheClearCmd.Flags().StringVar(&cacheTool, "tool", "", "Clear entries for specific tool")

	// Info command flags
	cacheInfoCmd.Flags().StringVar(&cacheServer, "server", "", "Show info for specific server")
	cacheInfoCmd.Flags().StringVar(&cacheTool, "tool", "", "Show info for specific tool")
}
