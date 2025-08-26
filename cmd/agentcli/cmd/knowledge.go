package cmd

import (
	"github.com/spf13/cobra"

	// Import embedding factory to enable Ollama and other embedding services
	_ "github.com/kunalkushwaha/agenticgokit/internal/embedding"
)

// Knowledge command flags
var (
	knowledgeConfigPath string
	outputFormat        string
	recursive           bool
	tags                string
	includeMetadata     bool
	limit               int
	scoreThreshold      float32
	filterSource        string
	filterType          string
	filterTags          string
	dryRun              bool
	force               bool
	interactive         bool
	showProgress        bool
)

// knowledgeCmd represents the knowledge management command
var knowledgeCmd = &cobra.Command{
	Use:   "knowledge",
	Short: "Manage knowledge base documents and content",
	Long: `Comprehensive knowledge base management commands for AgenticGoKit.

The knowledge commands provide tools to upload, search, list, validate, and manage
your AgenticGoKit knowledge base. These commands replace the debug-oriented 
'agentcli memory --docs' commands with production-ready management tools.

BASIC USAGE:
  # Upload documents to knowledge base
  agentcli knowledge upload ./docs/

  # Search knowledge base
  agentcli knowledge search "machine learning concepts"

  # List all documents
  agentcli knowledge list

  # Validate knowledge base health
  agentcli knowledge validate

  # Show knowledge base statistics
  agentcli knowledge stats

  # Clear knowledge base (with confirmation)
  agentcli knowledge clear

DOCUMENT MANAGEMENT:
  # Upload single file
  agentcli knowledge upload document.pdf

  # Upload directory recursively with tags
  agentcli knowledge upload ./docs/ --recursive --tags "docs,reference"

  # Upload with custom metadata
  agentcli knowledge upload ./code/ --metadata --recursive

OUTPUT FORMATS:
  # Table format (default)
  agentcli knowledge list

  # JSON format for scripting
  agentcli knowledge list --output json

FILTERING AND SEARCH:
  # Search with score threshold
  agentcli knowledge search "AI agents" --threshold 0.8

  # List with filtering
  agentcli knowledge list --source "docs/" --type "pdf" --tags "reference"

  # Limit results
  agentcli knowledge search "query" --limit 5

MAINTENANCE:
  # Validate configuration and connectivity
  agentcli knowledge validate

  # Show detailed statistics
  agentcli knowledge stats

  # Clear by source
  agentcli knowledge clear --source "old-docs/"

  # Interactive clear with confirmation
  agentcli knowledge clear --interactive

CONFIGURATION:
  # Use specific config file
  agentcli knowledge list --config-path /path/to/agentflow.toml

REQUIREMENTS:
- Must be run from an AgentFlow project directory (containing agentflow.toml)
- Knowledge base must be enabled in agentflow.toml (enable_knowledge_base = true)
- Memory provider must be configured and accessible

For detailed help on any subcommand, use: agentcli knowledge <command> --help`,
}

// knowledgeUploadCmd represents the upload command
var knowledgeUploadCmd = &cobra.Command{
	Use:   "upload [files/directories...]",
	Short: "Upload documents to the knowledge base",
	Long: `Upload documents and directories to the knowledge base.

Supports multiple file formats including:
- Text files (.txt)
- Markdown files (.md, .markdown)
- PDF files (.pdf) - requires PDF processing library
- Code files (.go, .py, .js, .java, etc.)
- HTML files (.html, .htm)

EXAMPLES:
  # Upload single file
  agentcli knowledge upload document.pdf

  # Upload multiple files
  agentcli knowledge upload doc1.md doc2.txt manual.pdf

  # Upload directory
  agentcli knowledge upload ./documentation/

  # Upload recursively with tags
  agentcli knowledge upload ./docs/ --recursive --tags "documentation,reference"

  # Upload with metadata extraction
  agentcli knowledge upload ./code/ --metadata --recursive

FLAGS:
  --recursive       Process directories recursively
  --tags           Add comma-separated tags to all uploaded documents
  --metadata       Extract and include file metadata
  --config-path    Path to agentflow.toml file
  --show-progress  Show detailed progress during upload`,
	Args: cobra.MinimumNArgs(1),
	RunE: runKnowledgeUpload,
}

// knowledgeListCmd represents the list command
var knowledgeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List documents in the knowledge base",
	Long: `List all documents in the knowledge base with optional filtering.

EXAMPLES:
  # List all documents
  agentcli knowledge list

  # List with JSON output
  agentcli knowledge list --output json

  # Filter by source
  agentcli knowledge list --source "docs/"

  # Filter by type and tags
  agentcli knowledge list --type "pdf" --tags "reference,manual"

  # Limit results
  agentcli knowledge list --limit 20

FLAGS:
  --output         Output format: table (default) or json
  --source         Filter by source path/pattern
  --type           Filter by document type (pdf, txt, md, etc.)
  --tags           Filter by tags (comma-separated)
  --limit          Maximum number of results to show
  --config-path    Path to agentflow.toml file`,
	RunE: runKnowledgeList,
}

// knowledgeSearchCmd represents the search command
var knowledgeSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search the knowledge base",
	Long: `Search the knowledge base using semantic similarity.

EXAMPLES:
  # Basic search
  agentcli knowledge search "machine learning concepts"

  # Search with score threshold
  agentcli knowledge search "AI agents" --threshold 0.8

  # Search with result limit
  agentcli knowledge search "documentation" --limit 5

  # Search with JSON output
  agentcli knowledge search "query" --output json

FLAGS:
  --threshold      Minimum similarity score (0.0-1.0)
  --limit          Maximum number of results
  --output         Output format: table (default) or json
  --config-path    Path to agentflow.toml file`,
	Args: cobra.ExactArgs(1),
	RunE: runKnowledgeSearch,
}

// knowledgeValidateCmd represents the validate command
var knowledgeValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate knowledge base configuration and health",
	Long: `Validate the knowledge base configuration, connectivity, and health.

Performs comprehensive checks including:
- Configuration file validation
- Memory provider connectivity
- Embedding service health
- Search functionality test
- Performance benchmarks

EXAMPLES:
  # Basic validation
  agentcli knowledge validate

  # Validation with custom config
  agentcli knowledge validate --config-path /path/to/agentflow.toml

FLAGS:
  --config-path    Path to agentflow.toml file`,
	RunE: runKnowledgeValidate,
}

// knowledgeStatsCmd represents the stats command
var knowledgeStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show knowledge base statistics",
	Long: `Display comprehensive statistics about the knowledge base.

Shows information including:
- Document count by type
- Storage usage statistics
- Search performance metrics
- Configuration summary
- Provider health status

EXAMPLES:
  # Show statistics
  agentcli knowledge stats

  # Statistics with JSON output
  agentcli knowledge stats --output json

FLAGS:
  --output         Output format: table (default) or json
  --config-path    Path to agentflow.toml file`,
	RunE: runKnowledgeStats,
}

// knowledgeClearCmd represents the clear command
var knowledgeClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear documents from the knowledge base",
	Long: `Clear documents from the knowledge base with optional filtering.

EXAMPLES:
  # Clear all documents (with confirmation)
  agentcli knowledge clear --interactive

  # Clear by source
  agentcli knowledge clear --source "old-docs/"

  # Clear by type
  agentcli knowledge clear --type "pdf"

  # Clear by tags
  agentcli knowledge clear --tags "deprecated,old"

  # Dry run to preview what would be deleted
  agentcli knowledge clear --dry-run

  # Force clear without confirmation
  agentcli knowledge clear --force

FLAGS:
  --source         Clear documents from specific source
  --type           Clear documents of specific type
  --tags           Clear documents with specific tags
  --dry-run        Show what would be deleted without actually deleting
  --force          Skip confirmation prompts
  --interactive    Interactive mode with detailed confirmation
  --config-path    Path to agentflow.toml file`,
	RunE: runKnowledgeClear,
}

func init() {
	// Add knowledge command to root
	rootCmd.AddCommand(knowledgeCmd)

	// Add subcommands
	knowledgeCmd.AddCommand(knowledgeUploadCmd)
	knowledgeCmd.AddCommand(knowledgeListCmd)
	knowledgeCmd.AddCommand(knowledgeSearchCmd)
	knowledgeCmd.AddCommand(knowledgeValidateCmd)
	knowledgeCmd.AddCommand(knowledgeStatsCmd)
	knowledgeCmd.AddCommand(knowledgeClearCmd)

	// Global flags
	knowledgeCmd.PersistentFlags().StringVar(&knowledgeConfigPath, "config-path", "", "Path to agentflow.toml file (default: ./agentflow.toml)")

	// Upload command flags
	knowledgeUploadCmd.Flags().BoolVar(&recursive, "recursive", false, "Process directories recursively")
	knowledgeUploadCmd.Flags().StringVar(&tags, "tags", "", "Comma-separated tags to add to documents")
	knowledgeUploadCmd.Flags().BoolVar(&includeMetadata, "metadata", false, "Extract and include file metadata")
	knowledgeUploadCmd.Flags().BoolVar(&showProgress, "show-progress", true, "Show upload progress")

	// List command flags
	knowledgeListCmd.Flags().StringVar(&outputFormat, "output", "table", "Output format: table or json")
	knowledgeListCmd.Flags().StringVar(&filterSource, "source", "", "Filter by source path/pattern")
	knowledgeListCmd.Flags().StringVar(&filterType, "type", "", "Filter by document type")
	knowledgeListCmd.Flags().StringVar(&filterTags, "tags", "", "Filter by tags (comma-separated)")
	knowledgeListCmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = no limit)")

	// Search command flags
	knowledgeSearchCmd.Flags().Float32Var(&scoreThreshold, "threshold", 0.0, "Minimum similarity score (0.0-1.0)")
	knowledgeSearchCmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results")
	knowledgeSearchCmd.Flags().StringVar(&outputFormat, "output", "table", "Output format: table or json")

	// Stats command flags
	knowledgeStatsCmd.Flags().StringVar(&outputFormat, "output", "table", "Output format: table or json")

	// Clear command flags
	knowledgeClearCmd.Flags().StringVar(&filterSource, "source", "", "Clear documents from specific source")
	knowledgeClearCmd.Flags().StringVar(&filterType, "type", "", "Clear documents of specific type")
	knowledgeClearCmd.Flags().StringVar(&filterTags, "tags", "", "Clear documents with specific tags")
	knowledgeClearCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without deleting")
	knowledgeClearCmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompts")
	knowledgeClearCmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive mode with detailed confirmation")
}

// Command handler functions - implementations will be added in subsequent tasks

func runKnowledgeUpload(cmd *cobra.Command, args []string) error {
	// Create and connect knowledge manager
	km, err := NewKnowledgeManager(knowledgeConfigPath)
	if err != nil {
		return err
	}

	if err := km.Connect(); err != nil {
		return err
	}
	defer km.Close()

	// Parse tags
	uploadTags := parseTagList(tags)

	// Create upload options
	options := UploadOptions{
		Recursive:       recursive,
		Tags:            uploadTags,
		IncludeMetadata: includeMetadata,
		ShowProgress:    showProgress,
		BatchSize:       10, // Default batch size
	}

	// Upload files
	return km.Upload(cmd.Context(), args, options)
}

func runKnowledgeList(cmd *cobra.Command, args []string) error {
	// Create and connect knowledge manager
	km, err := NewKnowledgeManager(knowledgeConfigPath)
	if err != nil {
		return err
	}

	if err := km.Connect(); err != nil {
		return err
	}
	defer km.Close()

	// Parse filter tags
	filterTagsList := parseTagList(filterTags)

	// Create list options
	options := ListOptions{
		OutputFormat: outputFormat,
		FilterSource: filterSource,
		FilterType:   filterType,
		FilterTags:   filterTagsList,
		Limit:        limit,
	}

	// List documents
	return km.List(cmd.Context(), options)
}

func runKnowledgeSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	// Create and connect knowledge manager
	km, err := NewKnowledgeManager(knowledgeConfigPath)
	if err != nil {
		return err
	}

	if err := km.Connect(); err != nil {
		return err
	}
	defer km.Close()

	// Create search options
	options := SearchOptions{
		OutputFormat:   outputFormat,
		ScoreThreshold: scoreThreshold,
		Limit:          limit,
	}

	// Search knowledge base
	return km.Search(cmd.Context(), query, options)
}

func runKnowledgeValidate(cmd *cobra.Command, args []string) error {
	// Create and connect knowledge manager
	km, err := NewKnowledgeManager(knowledgeConfigPath)
	if err != nil {
		return err
	}

	if err := km.Connect(); err != nil {
		return err
	}
	defer km.Close()

	// Validate knowledge base
	return km.Validate(cmd.Context())
}

func runKnowledgeStats(cmd *cobra.Command, args []string) error {
	// Create and connect knowledge manager
	km, err := NewKnowledgeManager(knowledgeConfigPath)
	if err != nil {
		return err
	}

	if err := km.Connect(); err != nil {
		return err
	}
	defer km.Close()

	// Get knowledge base statistics
	return km.Stats(cmd.Context(), outputFormat)
}

func runKnowledgeClear(cmd *cobra.Command, args []string) error {
	// Create and connect knowledge manager
	km, err := NewKnowledgeManager(knowledgeConfigPath)
	if err != nil {
		return err
	}

	if err := km.Connect(); err != nil {
		return err
	}
	defer km.Close()

	// Parse filter tags
	filterTagsList := parseTagList(filterTags)

	// Create clear options
	options := ClearOptions{
		FilterSource: filterSource,
		FilterType:   filterType,
		FilterTags:   filterTagsList,
		DryRun:       dryRun,
		Force:        force,
		Interactive:  interactive,
	}

	// Clear documents
	return km.Clear(cmd.Context(), options)
}
