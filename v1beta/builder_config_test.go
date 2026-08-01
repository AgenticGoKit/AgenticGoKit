package v1beta

import (
	"testing"
	"time"
)

// TestWithConfigPreservesBuilderName verifies that WithConfig does not
// discard the name passed to NewBuilder (issue #137, additional finding A).
func TestWithConfigPreservesBuilderName(t *testing.T) {
	b := NewBuilder("sales").WithConfig(&Config{
		LLM: LLMConfig{Provider: "mock", Model: "mock-model"},
	})

	sb, ok := b.(*streamlinedBuilder)
	if !ok {
		t.Fatalf("unexpected builder type %T", b)
	}
	if sb.config.Name != "sales" {
		t.Errorf("config.Name = %q, want %q from NewBuilder", sb.config.Name, "sales")
	}
	if sb.config.Timeout <= 0 {
		t.Errorf("config.Timeout = %v, want builder default preserved", sb.config.Timeout)
	}
}

// TestMemoryEnabledFalseIsHonored verifies that an explicitly disabled
// MemoryConfig actually disables memory (issue #137, additional finding B).
func TestMemoryEnabledFalseIsHonored(t *testing.T) {
	agent, err := NewBuilder("no-memory").WithConfig(&Config{
		LLM: LLMConfig{Provider: "mock", Model: "mock-model"},
		Memory: &MemoryConfig{
			Enabled:  false,
			Provider: "chromem",
		},
	}).Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if mem := agent.Memory(); mem != nil {
		t.Errorf("agent.Memory() = %T, want nil when Memory.Enabled=false", mem)
	}
}

// TestMemoryEnabledTrueProvidesMemory verifies the enabled path still works.
func TestMemoryEnabledTrueProvidesMemory(t *testing.T) {
	agent, err := NewBuilder("with-memory").WithConfig(&Config{
		LLM: LLMConfig{Provider: "mock", Model: "mock-model"},
		Memory: &MemoryConfig{
			Enabled:  true,
			Provider: "chromem",
			Options:  map[string]string{"embedding_provider": "dummy"},
		},
	}).Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if agent.Memory() == nil {
		t.Error("agent.Memory() = nil, want memory provider when Enabled=true")
	}
}

// TestWithConfigExplicitValuesWin verifies user-provided values are not
// overwritten by builder defaults.
func TestWithConfigExplicitValuesWin(t *testing.T) {
	b := NewBuilder("ignored").WithConfig(&Config{
		Name:    "explicit",
		Timeout: 42 * time.Second,
		LLM:     LLMConfig{Provider: "mock", Model: "mock-model"},
	})

	sb := b.(*streamlinedBuilder)
	if sb.config.Name != "explicit" {
		t.Errorf("config.Name = %q, want %q", sb.config.Name, "explicit")
	}
	if sb.config.Timeout != 42*time.Second {
		t.Errorf("config.Timeout = %v, want 42s", sb.config.Timeout)
	}
}
