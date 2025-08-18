package scaffold

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version"
)

// AgenticGoKitVersion represents the version of AgenticGoKit to use in generated projects
// This is dynamically determined from the CLI version or latest GitHub release
var AgenticGoKitVersion = getAgenticGoKitVersion()

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// getAgenticGoKitVersion dynamically determines the AgenticGoKit version to use
func getAgenticGoKitVersion() string {
	// Strategy 1: Use the CLI's own version if it's a proper release version
	cliVersion := version.Version
	if cliVersion != "" && cliVersion != "dev" && isValidSemanticVersion(cliVersion) {
		// Ensure version has 'v' prefix
		if !strings.HasPrefix(cliVersion, "v") {
			cliVersion = "v" + cliVersion
		}
		return cliVersion
	}

	// Strategy 2: Try to fetch the latest version from GitHub API
	if latestVersion := fetchLatestVersionFromGitHub(); latestVersion != "" {
		return latestVersion
	}

	// Strategy 3: Fallback to a known stable version
	// This should be updated periodically, but serves as a last resort
	return "v0.3.4"
}

// fetchLatestVersionFromGitHub fetches the latest release version from GitHub API
func fetchLatestVersionFromGitHub() string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.github.com/repos/kunalkushwaha/agenticgokit/releases/latest", nil)
	if err != nil {
		return ""
	}

	// Set User-Agent to avoid rate limiting
	req.Header.Set("User-Agent", "AgenticGoKit-CLI")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return ""
	}

	if release.TagName != "" && isValidSemanticVersion(release.TagName) {
		return release.TagName
	}

	return ""
}

// isValidSemanticVersion checks if a version string follows semantic versioning
func isValidSemanticVersion(version string) bool {
	// Remove 'v' prefix if present
	v := strings.TrimPrefix(version, "v")

	// Handle pre-release and build metadata
	// Split on '+' first to handle build metadata
	mainVersion := v
	if idx := strings.Index(v, "+"); idx != -1 {
		mainVersion = v[:idx]
	}

	// Split on '-' to handle pre-release
	if idx := strings.Index(mainVersion, "-"); idx != -1 {
		mainVersion = mainVersion[:idx]
	}

	// Basic semantic version pattern: X.Y.Z (exactly 3 parts)
	parts := strings.Split(mainVersion, ".")
	if len(parts) != 3 {
		return false
	}

	// Check that all three parts are valid numbers
	for _, part := range parts {
		// Check if it's a valid number
		if len(part) == 0 {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
	}

	return true
}

// GetAgenticGoKitVersionWithFallback returns the AgenticGoKit version with explicit fallback handling
func GetAgenticGoKitVersionWithFallback() (string, string) {
	cliVersion := version.Version

	// If CLI version is valid, use it
	if cliVersion != "" && cliVersion != "dev" && isValidSemanticVersion(cliVersion) {
		if !strings.HasPrefix(cliVersion, "v") {
			cliVersion = "v" + cliVersion
		}
		return cliVersion, "cli-version"
	}

	// Try GitHub API
	if latestVersion := fetchLatestVersionFromGitHub(); latestVersion != "" {
		return latestVersion, "github-api"
	}

	// Fallback
	return "v0.3.4", "fallback"
}

// AgentInfo represents information about an agent including its name and purpose
type AgentInfo struct {
	Name        string // User-defined name like "analyzer", "processor"
	FileName    string // File name like "analyzer.go"
	DisplayName string // Capitalized name like "Analyzer"
	Purpose     string // Brief description of the agent's purpose
	Role        string // Agent role like "collaborative", "sequential", "loop"
}

// ProjectConfig represents the configuration for creating a new AgenticGoKit project
type ProjectConfig struct {
	Name          string
	NumAgents     int
	Provider      string
	ResponsibleAI bool
	ErrorHandler  bool

	// MCP configuration
	MCPEnabled         bool
	MCPProduction      bool
	MCPTransport       string
	WithCache          bool
	WithMetrics        bool
	MCPTools           []string
	MCPServers         []string
	CacheBackend       string
	MetricsPort        int
	WithLoadBalancer   bool
	ConnectionPoolSize int
	RetryPolicy        string

	// Multi-agent orchestration configuration
	OrchestrationMode    string
	CollaborativeAgents  []string
	SequentialAgents     []string
	LoopAgent            string
	MaxIterations        int
	OrchestrationTimeout int
	FailureThreshold     float64
	MaxConcurrency       int

	// Visualization configuration
	Visualize          bool
	VisualizeOutputDir string

	// Memory/RAG configuration
	MemoryEnabled       bool
	MemoryProvider      string // inmemory, pgvector, weaviate
	EmbeddingProvider   string // openai, dummy
	EmbeddingModel      string // text-embedding-3-small, etc.
	EmbeddingDimensions int    // Auto-calculated based on embedding model
	RAGEnabled          bool
	RAGChunkSize        int
	RAGOverlap          int
	RAGTopK             int
	RAGScoreThreshold   float64
	HybridSearch        bool
	SessionMemory       bool
}

// TemplateData represents the data structure passed to templates
type TemplateData struct {
	Config         ProjectConfig
	Agent          AgentInfo
	Agents         []AgentInfo
	AgentIndex     int
	TotalAgents    int
	NextAgent      string
	PrevAgent      string
	IsFirstAgent   bool
	IsLastAgent    bool
	SystemPrompt   string
	RoutingComment string
}

// ProjectStructureInfo contains information about the project's directory structure
type ProjectStructureInfo struct {
	AgentsPackage   string // Package name for agents (e.g., "agents")
	InternalPackage string // Package name for internal utilities (e.g., "internal")
	MainPackage     string // Main package name (usually the project name)
	DocsDir         string // Documentation directory name (e.g., "docs")
	AgentsDir       string // Agents directory name (e.g., "agents")
	InternalDir     string // Internal directory name (e.g., "internal")
}

// CustomizationPoint represents a specific area where developers can customize the code
type CustomizationPoint struct {
	Location    string // File and line location (e.g., "agents/agent1.go:45")
	Description string // Human-readable description of what can be customized
	Example     string // Code example showing how to customize
	Required    bool   // Whether this customization is required or optional
	Category    string // Category of customization (e.g., "business_logic", "configuration", "integration")
}

// ImportPathInfo contains information about import paths used in the generated code
type ImportPathInfo struct {
	AgentsImport   string // Import path for agents package
	InternalImport string // Import path for internal package
	ModuleName     string // Go module name
	BaseImportPath string // Base import path for the project
}

// EnhancedTemplateData represents the enhanced data structure passed to templates
type EnhancedTemplateData struct {
	Config         ProjectConfig
	Agent          AgentInfo
	Agents         []AgentInfo
	AgentIndex     int
	TotalAgents    int
	NextAgent      string
	PrevAgent      string
	IsFirstAgent   bool
	IsLastAgent    bool
	SystemPrompt   string
	RoutingComment string

	// New enhanced fields
	ProjectStructure    ProjectStructureInfo
	CustomizationPoints []CustomizationPoint
	ImportPaths         ImportPathInfo

	// Additional metadata
	GenerationTimestamp string
	FrameworkVersion    string
	Features            []string // List of enabled features
}

// CreateProjectStructureInfo creates project structure information
func CreateProjectStructureInfo(config ProjectConfig) ProjectStructureInfo {
	return ProjectStructureInfo{
		AgentsPackage:   "agents",
		InternalPackage: "internal",
		MainPackage:     config.Name,
		DocsDir:         "docs",
		AgentsDir:       "agents",
		InternalDir:     "internal",
	}
}

// CreateImportPathInfo creates import path information with validation
func CreateImportPathInfo(config ProjectConfig) ImportPathInfo {
	// Use safe import path resolution to handle any edge cases
	agentsImport := ResolveImportPathSafe(config.Name, "agents")
	internalImport := ResolveImportPathSafe(config.Name, "internal")

	return ImportPathInfo{
		AgentsImport:   agentsImport,
		InternalImport: internalImport,
		ModuleName:     SanitizeModuleName(config.Name),
		BaseImportPath: SanitizeModuleName(config.Name),
	}
}

// CreateCustomizationPoints creates a list of customization points for the project
func CreateCustomizationPoints(config ProjectConfig) []CustomizationPoint {
	points := []CustomizationPoint{
		{
			Location:    "agents/agent.go:Run method",
			Description: "Customize agent business logic and processing workflow",
			Example:     "// Add your custom processing logic here\nresult := processWithCustomLogic(input)",
			Required:    true,
			Category:    "business_logic",
		},
		{
			Location:    "agents/agent.go:constructor",
			Description: "Add custom dependencies to agent constructors",
			Example:     "func NewAgent(llm ModelProvider, db *sql.DB, config *Config) *AgentHandler",
			Required:    false,
			Category:    "configuration",
		},
		{
			Location:    "main.go:initialization",
			Description: "Add custom initialization logic for external services",
			Example:     "db, err := sql.Open(\"postgres\", connectionString)\nagent := agents.NewAgent(llmProvider, db)",
			Required:    false,
			Category:    "integration",
		},
		{
			Location:    "agentflow.toml:configuration",
			Description: "Customize system configuration for your specific needs",
			Example:     "[custom_section]\napi_endpoint = \"https://your-api.com\"\ntimeout = \"30s\"",
			Required:    false,
			Category:    "configuration",
		},
	}

	// Add memory-specific customization points
	if config.MemoryEnabled {
		points = append(points, CustomizationPoint{
			Location:    "agents/agent.go:memory operations",
			Description: "Customize memory storage and retrieval operations",
			Example:     "// Store custom metadata\nerr := a.memory.StoreWithMetadata(ctx, data, metadata)",
			Required:    false,
			Category:    "integration",
		})
	}

	// Add MCP-specific customization points
	if config.MCPEnabled {
		points = append(points, CustomizationPoint{
			Location:    "agents/agent.go:tool integration",
			Description: "Customize tool usage and validation logic",
			Example:     "// Validate tool call before execution\nif err := validateToolCall(toolName, args); err != nil { return err }",
			Required:    false,
			Category:    "integration",
		})
	}

	// Add orchestration-specific customization points
	switch config.OrchestrationMode {
	case "sequential":
		points = append(points, CustomizationPoint{
			Location:    "agents/agent.go:sequential processing",
			Description: "Customize how agents process input from previous agents",
			Example:     "// Transform input from previous agent\ninput = transformSequentialInput(previousOutput)",
			Required:    false,
			Category:    "business_logic",
		})
	case "collaborative":
		points = append(points, CustomizationPoint{
			Location:    "main.go:result aggregation",
			Description: "Customize how collaborative agent results are combined",
			Example:     "// Custom result aggregation logic\nfinalResult := aggregateResults(collaborativeResults)",
			Required:    false,
			Category:    "business_logic",
		})
	case "loop":
		points = append(points, CustomizationPoint{
			Location:    "agents/agent.go:loop termination",
			Description: "Customize loop termination conditions",
			Example:     "// Custom termination logic\nif customTerminationCondition(result) { break }",
			Required:    false,
			Category:    "business_logic",
		})
	}

	return points
}

// GetEnabledFeatures returns a list of enabled features based on configuration
func GetEnabledFeatures(config ProjectConfig) []string {
	var features []string

	if config.MemoryEnabled {
		features = append(features, "Memory System")
		if config.RAGEnabled {
			features = append(features, "RAG (Retrieval-Augmented Generation)")
		}
		if config.SessionMemory {
			features = append(features, "Session Memory")
		}
		if config.HybridSearch {
			features = append(features, "Hybrid Search")
		}
	}

	if config.MCPEnabled {
		features = append(features, "MCP Tool Integration")
		if config.MCPProduction {
			features = append(features, "Production MCP")
		}
		if config.WithCache {
			features = append(features, "MCP Caching")
		}
		if config.WithMetrics {
			features = append(features, "MCP Metrics")
		}
		if config.WithLoadBalancer {
			features = append(features, "MCP Load Balancer")
		}
	}

	if config.ResponsibleAI {
		features = append(features, "Responsible AI")
	}

	if config.ErrorHandler {
		features = append(features, "Enhanced Error Handling")
	}

	if config.Visualize {
		features = append(features, "Workflow Visualization")
	}

	// Add orchestration mode as a feature
	switch config.OrchestrationMode {
	case "sequential":
		features = append(features, "Sequential Orchestration")
	case "collaborative":
		features = append(features, "Collaborative Orchestration")
	case "loop":
		features = append(features, "Loop Orchestration")
	case "mixed":
		features = append(features, "Mixed Orchestration")
	default:
		features = append(features, "Route-based Orchestration")
	}

	return features
}

// CreateEnhancedTemplateData creates enhanced template data with all the new fields
func CreateEnhancedTemplateData(config ProjectConfig, agent AgentInfo, agents []AgentInfo, agentIndex int) EnhancedTemplateData {
	var nextAgent, prevAgent string

	// Determine next and previous agents
	if agentIndex < len(agents)-1 {
		nextAgent = agents[agentIndex+1].Name
	}
	if agentIndex > 0 {
		prevAgent = agents[agentIndex-1].Name
	}

	// Create routing comment based on orchestration mode
	var routingComment string
	if nextAgent != "" {
		routingComment = fmt.Sprintf("Route to the next agent (%s) in the workflow", nextAgent)
	} else if config.ResponsibleAI {
		routingComment = "Route to Responsible AI for final content check"
	} else {
		routingComment = "Workflow completion"
	}

	return EnhancedTemplateData{
		Config:         config,
		Agent:          agent,
		Agents:         agents,
		AgentIndex:     agentIndex,
		TotalAgents:    len(agents),
		NextAgent:      nextAgent,
		PrevAgent:      prevAgent,
		IsFirstAgent:   agentIndex == 0,
		IsLastAgent:    agentIndex == len(agents)-1,
		SystemPrompt:   CreateSystemPrompt(agent, agentIndex, len(agents), config.OrchestrationMode),
		RoutingComment: routingComment,

		// Enhanced fields
		ProjectStructure:    CreateProjectStructureInfo(config),
		CustomizationPoints: CreateCustomizationPoints(config),
		ImportPaths:         CreateImportPathInfo(config),

		// Metadata
		GenerationTimestamp: time.Now().Format("2006-01-02 15:04:05"),
		FrameworkVersion:    AgenticGoKitVersion,
		Features:            GetEnabledFeatures(config),
	}
}

// CreateSystemPrompt creates a system prompt for an agent based on its role and position
func CreateSystemPrompt(agent AgentInfo, index, total int, orchestrationMode string) string {
	basePrompt := fmt.Sprintf("You are %s, a specialized AI agent", agent.DisplayName)

	// Add purpose
	if agent.Purpose != "" {
		basePrompt += fmt.Sprintf(" whose purpose is to %s", agent.Purpose)
	}

	// Add orchestration context
	switch orchestrationMode {
	case "sequential":
		if index == 0 {
			basePrompt += ". You are the first agent in a sequential workflow. Process the user's input and prepare it for the next agent."
		} else if index == total-1 {
			basePrompt += ". You are the final agent in a sequential workflow. Process the input from previous agents and provide the final response."
		} else {
			basePrompt += fmt.Sprintf(". You are agent %d of %d in a sequential workflow. Process the input from the previous agent and pass refined results to the next agent.", index+1, total)
		}
	case "collaborative":
		basePrompt += fmt.Sprintf(". You are one of %d agents working collaboratively. Process the user's input from your unique perspective and provide your specialized analysis.", total)
	case "loop":
		basePrompt += ". You are part of an iterative loop workflow. Process the input and refine it through multiple iterations until the desired outcome is achieved."
	default:
		basePrompt += ". You are part of a route-based workflow. Process requests that are specifically routed to you based on their content or type."
	}

	basePrompt += " Always provide clear, helpful, and accurate responses."

	return basePrompt
}

// Import path resolution and validation functions

// ValidateGoModuleName validates that a module name follows Go module naming conventions
func ValidateGoModuleName(moduleName string) error {
	if moduleName == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	// Check for invalid characters
	for _, char := range moduleName {
		if !isValidModuleChar(char) {
			return fmt.Errorf("module name '%s' contains invalid character '%c'", moduleName, char)
		}
	}

	// Check for reserved names
	if isReservedModuleName(moduleName) {
		return fmt.Errorf("module name '%s' is reserved", moduleName)
	}

	// Check for proper format (should not start/end with special chars)
	if strings.HasPrefix(moduleName, "-") || strings.HasSuffix(moduleName, "-") ||
		strings.HasPrefix(moduleName, "_") || strings.HasSuffix(moduleName, "_") {
		return fmt.Errorf("module name '%s' cannot start or end with '-' or '_'", moduleName)
	}

	return nil
}

// ValidatePackageName validates that a package name follows Go package naming conventions
func ValidatePackageName(packageName string) error {
	if packageName == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	// Package names should be lowercase
	if packageName != strings.ToLower(packageName) {
		return fmt.Errorf("package name '%s' should be lowercase", packageName)
	}

	// Check for invalid characters (only letters, numbers, underscores)
	for _, char := range packageName {
		if !isValidPackageChar(char) {
			return fmt.Errorf("package name '%s' contains invalid character '%c'", packageName, char)
		}
	}

	// Check for reserved package names
	if isReservedPackageName(packageName) {
		return fmt.Errorf("package name '%s' is reserved", packageName)
	}

	// Package names should not start with numbers
	if len(packageName) > 0 && packageName[0] >= '0' && packageName[0] <= '9' {
		return fmt.Errorf("package name '%s' cannot start with a number", packageName)
	}

	return nil
}

// SanitizeModuleName automatically corrects invalid module names
func SanitizeModuleName(moduleName string) string {
	if moduleName == "" {
		return "my-project"
	}

	// Convert to lowercase and replace invalid characters
	sanitized := strings.ToLower(moduleName)

	// Replace invalid characters with hyphens, but avoid consecutive hyphens
	var result strings.Builder
	lastWasHyphen := false

	for _, char := range sanitized {
		if isValidModuleChar(char) && char != '_' {
			result.WriteRune(char)
			lastWasHyphen = false
		} else if char == ' ' || char == '_' {
			if !lastWasHyphen {
				result.WriteRune('-')
				lastWasHyphen = true
			}
		} else {
			// Skip other invalid characters
		}
	}

	sanitized = result.String()

	// Remove leading/trailing hyphens and underscores
	sanitized = strings.Trim(sanitized, "-_")

	// Ensure it's not empty after sanitization
	if sanitized == "" {
		sanitized = "my-project"
	}

	// Ensure it's not a reserved name
	if isReservedModuleName(sanitized) {
		sanitized = "my-" + sanitized
	}

	return sanitized
}

// SanitizePackageName automatically corrects invalid package names
func SanitizePackageName(packageName string) string {
	if packageName == "" {
		return "mypackage"
	}

	// Convert to lowercase
	sanitized := strings.ToLower(packageName)

	// Replace invalid characters with underscores, collapse consecutive underscores,
	// and avoid introducing leading underscores.
	var result strings.Builder
	lastUnderscore := false
	for _, char := range sanitized {
		// Treat hyphens, spaces, and underscores uniformly as underscore separators
		if char == '-' || char == ' ' || char == '_' {
			// Skip writing underscore if it's leading or duplicate
			if result.Len() == 0 || lastUnderscore {
				lastUnderscore = true
				continue
			}
			result.WriteRune('_')
			lastUnderscore = true
			continue
		}

		if isValidPackageChar(char) {
			result.WriteRune(char)
			lastUnderscore = false
		}
		// Skip other invalid characters silently
	}

	sanitized = result.String()

	// Remove any trailing underscore that may remain
	sanitized = strings.Trim(sanitized, "_")

	// Ensure it's not empty after sanitization
	if sanitized == "" {
		sanitized = "mypackage"
	}

	// Ensure it doesn't start with a number
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "pkg_" + sanitized
	}

	// Minimal reserved handling for sanitizer: only special-case "main"
	// Validation still uses the full reserved keyword list.
	if sanitized == "main" {
		sanitized = "my_" + sanitized
	}

	return sanitized
}

// ResolveImportPath generates the correct import path for a given package within the project
func ResolveImportPath(moduleName, packagePath string) (string, error) {
	if err := ValidateGoModuleName(moduleName); err != nil {
		return "", fmt.Errorf("invalid module name: %w", err)
	}

	if packagePath == "" {
		return moduleName, nil
	}

	// Clean the package path
	packagePath = strings.Trim(packagePath, "/")

	// Validate each component of the package path
	components := strings.Split(packagePath, "/")
	for _, component := range components {
		if err := ValidatePackageName(component); err != nil {
			return "", fmt.Errorf("invalid package path component '%s': %w", component, err)
		}
	}

	return moduleName + "/" + packagePath, nil
}

// ResolveImportPathSafe generates import paths with automatic sanitization
func ResolveImportPathSafe(moduleName, packagePath string) string {
	sanitizedModule := SanitizeModuleName(moduleName)

	if packagePath == "" {
		return sanitizedModule
	}

	// Clean and sanitize the package path
	packagePath = strings.Trim(packagePath, "/")
	components := strings.Split(packagePath, "/")

	var sanitizedComponents []string
	for _, component := range components {
		sanitized := SanitizePackageName(component)
		if sanitized != "" {
			sanitizedComponents = append(sanitizedComponents, sanitized)
		}
	}

	if len(sanitizedComponents) == 0 {
		return sanitizedModule
	}

	return sanitizedModule + "/" + strings.Join(sanitizedComponents, "/")
}

// ValidateImportPaths validates all import paths for a project configuration
func ValidateImportPaths(config ProjectConfig) error {
	// Validate module name
	if err := ValidateGoModuleName(config.Name); err != nil {
		return fmt.Errorf("project name validation failed: %w", err)
	}

	// Validate agents import path
	agentsImport, err := ResolveImportPath(config.Name, "agents")
	if err != nil {
		return fmt.Errorf("agents import path validation failed: %w", err)
	}

	// Validate internal import path
	internalImport, err := ResolveImportPath(config.Name, "internal")
	if err != nil {
		return fmt.Errorf("internal import path validation failed: %w", err)
	}

	// Log successful validation
	fmt.Printf("Import paths validated:\n")
	fmt.Printf("   Module: %s\n", config.Name)
	fmt.Printf("   Agents: %s\n", agentsImport)
	fmt.Printf("   Internal: %s\n", internalImport)

	return nil
}

// Helper functions for character validation

func isValidModuleChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_' || char == '.' || char == '/'
}

func isValidPackageChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_'
}

func isReservedModuleName(name string) bool {
	reserved := []string{
		"main", "test", "example", "internal", "vendor",
		"go", "golang", "std", "builtin", "unsafe",
	}

	for _, r := range reserved {
		if strings.EqualFold(name, r) {
			return true
		}
	}

	return false
}

func isReservedPackageName(name string) bool {
	reserved := []string{
		"main", "test", "init", "import", "package", "var", "const",
		"func", "type", "struct", "interface", "map", "chan", "select",
		"go", "defer", "return", "break", "continue", "fallthrough",
		"if", "else", "switch", "case", "default", "for", "range",
		"builtin", "unsafe", "reflect", "runtime", "sync", "context",
	}

	for _, r := range reserved {
		if name == r {
			return true
		}
	}

	return false
}

// ValidateAndSanitizeProjectConfig validates and sanitizes project configuration
func ValidateAndSanitizeProjectConfig(config *ProjectConfig) error {
	originalName := config.Name

	// Sanitize the project name if needed
	if err := ValidateGoModuleName(config.Name); err != nil {
		fmt.Printf("WARNING: Project name '%s' is invalid: %v\n", config.Name, err)
		config.Name = SanitizeModuleName(config.Name)
		fmt.Printf("Auto-corrected to: '%s'\n", config.Name)
	}

	// Validate that the corrected name is different from original
	if originalName != config.Name {
		fmt.Printf("Project name changed from '%s' to '%s'\n", originalName, config.Name)
	}

	// Validate other configuration aspects
	if config.NumAgents < 1 {
		config.NumAgents = 1
		fmt.Printf("Auto-corrected NumAgents to 1 (minimum required)\n")
	}

	if config.NumAgents > 10 {
		fmt.Printf("WARNING: Large number of agents (%d) may impact performance\n", config.NumAgents)
	}

	// Validate provider
	validProviders := []string{"openai", "azure", "ollama", "anthropic"}
	isValidProvider := false
	for _, provider := range validProviders {
		if config.Provider == provider {
			isValidProvider = true
			break
		}
	}

	if !isValidProvider {
		return fmt.Errorf("invalid provider '%s'. Valid providers: %v", config.Provider, validProviders)
	}

	// Validate orchestration mode
	validModes := []string{"sequential", "collaborative", "loop", "mixed", "route"}
	isValidMode := false
	for _, mode := range validModes {
		if config.OrchestrationMode == mode {
			isValidMode = true
			break
		}
	}

	if !isValidMode && config.OrchestrationMode != "" {
		fmt.Printf("WARNING: Invalid orchestration mode '%s', defaulting to 'sequential'\n", config.OrchestrationMode)
		config.OrchestrationMode = "sequential"
	}

	if config.OrchestrationMode == "" {
		config.OrchestrationMode = "sequential"
	}

	return nil
}

// ShowVersionInfo displays information about the AgenticGoKit version being used
func ShowVersionInfo() {
	version, source := GetAgenticGoKitVersionWithFallback()

	fmt.Printf("Using AgenticGoKit version: %s", version)

	switch source {
	case "cli-version":
		fmt.Printf(" (from CLI version)\n")
	case "github-api":
		fmt.Printf(" (latest from GitHub)\n")
	case "fallback":
		fmt.Printf(" (fallback - consider updating CLI)\n")
		fmt.Printf("   Run the installer to get the latest version:\n")
		fmt.Printf("      curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.sh | bash\n")
	}
}

// CreateProjectWithValidation creates a project with full validation and sanitization
func CreateProjectWithValidation(config ProjectConfig) error {
	// Show version information
	ShowVersionInfo()

	// Validate and sanitize configuration
	if err := ValidateAndSanitizeProjectConfig(&config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create the project with validated configuration
	return CreateAgentProjectModular(config)
}
