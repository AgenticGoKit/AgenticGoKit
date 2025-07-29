package scaffold

import (
	"testing"
)

func TestValidateGoModuleName(t *testing.T) {
	tests := []struct {
		name        string
		moduleName  string
		expectError bool
	}{
		{"valid simple name", "myproject", false},
		{"valid with hyphens", "my-project", false},
		{"valid with dots", "github.com/user/project", false},
		{"empty name", "", true},
		{"reserved name", "main", true},
		{"starts with hyphen", "-project", true},
		{"ends with hyphen", "project-", true},
		{"starts with underscore", "_project", true},
		{"ends with underscore", "project_", true},
		{"valid complex path", "github.com/user/my-project", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGoModuleName(tt.moduleName)
			if tt.expectError && err == nil {
				t.Errorf("expected error for module name '%s', but got none", tt.moduleName)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for module name '%s': %v", tt.moduleName, err)
			}
		})
	}
}

func TestValidatePackageName(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		expectError bool
	}{
		{"valid simple name", "mypackage", false},
		{"valid with underscores", "my_package", false},
		{"empty name", "", true},
		{"uppercase letters", "MyPackage", true},
		{"starts with number", "1package", true},
		{"contains hyphens", "my-package", true},
		{"reserved name", "main", true},
		{"reserved keyword", "func", true},
		{"valid single letter", "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageName(tt.packageName)
			if tt.expectError && err == nil {
				t.Errorf("expected error for package name '%s', but got none", tt.packageName)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for package name '%s': %v", tt.packageName, err)
			}
		})
	}
}

func TestSanitizeModuleName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid name unchanged", "myproject", "myproject"},
		{"uppercase to lowercase", "MyProject", "myproject"},
		{"spaces to hyphens", "my project", "my-project"},
		{"underscores to hyphens", "my_project", "my-project"},
		{"remove invalid chars", "my@project!", "myproject"},
		{"empty string", "", "my-project"},
		{"reserved name", "main", "my-main"},
		{"leading/trailing hyphens", "-project-", "project"},
		{"multiple spaces", "my   project", "my-project"},
		{"mixed invalid chars", "My Project@123!", "my-project123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeModuleName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeModuleName(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizePackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid name unchanged", "mypackage", "mypackage"},
		{"uppercase to lowercase", "MyPackage", "mypackage"},
		{"hyphens to underscores", "my-package", "my_package"},
		{"spaces to underscores", "my package", "my_package"},
		{"remove invalid chars", "my@package!", "mypackage"},
		{"empty string", "", "mypackage"},
		{"reserved name", "main", "my_main"},
		{"starts with number", "1package", "pkg_1package"},
		{"leading/trailing underscores", "_package_", "package"},
		{"multiple spaces", "my   package", "my_package"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizePackageName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizePackageName(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestResolveImportPath(t *testing.T) {
	tests := []struct {
		name        string
		moduleName  string
		packagePath string
		expected    string
		expectError bool
	}{
		{"simple module", "myproject", "", "myproject", false},
		{"with package path", "myproject", "agents", "myproject/agents", false},
		{"nested package path", "myproject", "internal/config", "myproject/internal/config", false},
		{"invalid module name", "my@project", "agents", "", true},
		{"invalid package name", "myproject", "my-agents", "", true},
		{"empty module name", "", "agents", "", true},
		{"valid complex module", "github.com/user/project", "agents", "github.com/user/project/agents", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveImportPath(tt.moduleName, tt.packagePath)
			if tt.expectError && err == nil {
				t.Errorf("expected error for ResolveImportPath(%s, %s), but got none", tt.moduleName, tt.packagePath)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for ResolveImportPath(%s, %s): %v", tt.moduleName, tt.packagePath, err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("ResolveImportPath(%s, %s) = %s, expected %s", tt.moduleName, tt.packagePath, result, tt.expected)
			}
		})
	}
}

func TestResolveImportPathSafe(t *testing.T) {
	tests := []struct {
		name        string
		moduleName  string
		packagePath string
		expected    string
	}{
		{"simple module", "myproject", "", "myproject"},
		{"with package path", "myproject", "agents", "myproject/agents"},
		{"sanitize module name", "My@Project", "agents", "myproject/agents"},
		{"sanitize package path", "myproject", "my-agents", "myproject/my_agents"},
		{"empty module name", "", "agents", "my-project/agents"},
		{"empty package path", "myproject", "", "myproject"},
		{"complex sanitization", "My Project@123", "My-Agents!", "my-project123/my_agents"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveImportPathSafe(tt.moduleName, tt.packagePath)
			if result != tt.expected {
				t.Errorf("ResolveImportPathSafe(%s, %s) = %s, expected %s", tt.moduleName, tt.packagePath, result, tt.expected)
			}
		})
	}
}

func TestValidateImportPaths(t *testing.T) {
	tests := []struct {
		name        string
		config      ProjectConfig
		expectError bool
	}{
		{
			name: "valid configuration",
			config: ProjectConfig{
				Name:      "myproject",
				NumAgents: 2,
				Provider:  "openai",
			},
			expectError: false,
		},
		{
			name: "invalid project name",
			config: ProjectConfig{
				Name:      "my@project",
				NumAgents: 2,
				Provider:  "openai",
			},
			expectError: true,
		},
		{
			name: "empty project name",
			config: ProjectConfig{
				Name:      "",
				NumAgents: 2,
				Provider:  "openai",
			},
			expectError: true,
		},
		{
			name: "reserved project name",
			config: ProjectConfig{
				Name:      "main",
				NumAgents: 2,
				Provider:  "openai",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImportPaths(tt.config)
			if tt.expectError && err == nil {
				t.Errorf("expected error for ValidateImportPaths with config %+v, but got none", tt.config)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for ValidateImportPaths with config %+v: %v", tt.config, err)
			}
		})
	}
}

func TestValidateAndSanitizeProjectConfig(t *testing.T) {
	tests := []struct {
		name           string
		inputConfig    ProjectConfig
		expectedName   string
		expectedAgents int
		expectError    bool
	}{
		{
			name: "valid configuration unchanged",
			inputConfig: ProjectConfig{
				Name:              "myproject",
				NumAgents:         2,
				Provider:          "openai",
				OrchestrationMode: "sequential",
			},
			expectedName:   "myproject",
			expectedAgents: 2,
			expectError:    false,
		},
		{
			name: "sanitize invalid name",
			inputConfig: ProjectConfig{
				Name:              "My@Project!",
				NumAgents:         2,
				Provider:          "openai",
				OrchestrationMode: "sequential",
			},
			expectedName:   "myproject",
			expectedAgents: 2,
			expectError:    false,
		},
		{
			name: "fix zero agents",
			inputConfig: ProjectConfig{
				Name:              "myproject",
				NumAgents:         0,
				Provider:          "openai",
				OrchestrationMode: "sequential",
			},
			expectedName:   "myproject",
			expectedAgents: 1,
			expectError:    false,
		},
		{
			name: "invalid provider",
			inputConfig: ProjectConfig{
				Name:              "myproject",
				NumAgents:         2,
				Provider:          "invalid",
				OrchestrationMode: "sequential",
			},
			expectedName:   "myproject",
			expectedAgents: 2,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.inputConfig // Make a copy
			err := ValidateAndSanitizeProjectConfig(&config)
			
			if tt.expectError && err == nil {
				t.Errorf("expected error for ValidateAndSanitizeProjectConfig, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for ValidateAndSanitizeProjectConfig: %v", err)
			}
			
			if !tt.expectError {
				if config.Name != tt.expectedName {
					t.Errorf("expected name %s, got %s", tt.expectedName, config.Name)
				}
				if config.NumAgents != tt.expectedAgents {
					t.Errorf("expected %d agents, got %d", tt.expectedAgents, config.NumAgents)
				}
			}
		})
	}
}