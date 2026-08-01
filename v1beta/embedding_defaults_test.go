package v1beta

import (
	"testing"
)

func TestApplyEmbeddingDefaults_OllamaLLM(t *testing.T) {
	mem := &MemoryConfig{Enabled: true, Provider: "chromem"}
	llm := &LLMConfig{Provider: "ollama", Model: "llama3.2", BaseURL: "http://localhost:11434"}

	applyEmbeddingDefaults(mem, llm)

	if got := mem.Options["embedding_provider"]; got != "ollama" {
		t.Errorf("embedding_provider = %q, want %q", got, "ollama")
	}
	// The chat model must never be reused as the embedding model (issue #137).
	if got := mem.Options["embedding_model"]; got != defaultOllamaEmbeddingModel {
		t.Errorf("embedding_model = %q, want %q", got, defaultOllamaEmbeddingModel)
	}
	if got := mem.Options["embedding_url"]; got != "http://localhost:11434" {
		t.Errorf("embedding_url = %q, want LLM BaseURL", got)
	}
	if got := mem.Options["dimensions"]; got != "768" {
		t.Errorf("dimensions = %q, want %q (nomic-embed-text)", got, "768")
	}
}

func TestApplyEmbeddingDefaults_OpenAILLM(t *testing.T) {
	mem := &MemoryConfig{Enabled: true, Provider: "chromem"}
	llm := &LLMConfig{Provider: "openai", Model: "gpt-4o-mini", APIKey: "sk-test"}

	applyEmbeddingDefaults(mem, llm)

	if got := mem.Options["embedding_provider"]; got != "openai" {
		t.Errorf("embedding_provider = %q, want %q", got, "openai")
	}
	if got := mem.Options["embedding_model"]; got != defaultOpenAIEmbeddingModel {
		t.Errorf("embedding_model = %q, want %q", got, defaultOpenAIEmbeddingModel)
	}
	if got := mem.Options["embedding_api_key"]; got != "sk-test" {
		t.Errorf("embedding_api_key = %q, want LLM APIKey", got)
	}
	if got := mem.Options["dimensions"]; got != "1536" {
		t.Errorf("dimensions = %q, want %q (text-embedding-3-small)", got, "1536")
	}
}

func TestApplyEmbeddingDefaults_NonEmbeddingLLMLeftUnset(t *testing.T) {
	mem := &MemoryConfig{Enabled: true, Provider: "chromem"}
	llm := &LLMConfig{Provider: "anthropic", Model: "claude-sonnet-4-20250514"}

	applyEmbeddingDefaults(mem, llm)

	// Anthropic has no embeddings API; the provider must be left unset so
	// memory initialization can apply its own loud fallback, instead of the
	// old behavior of copying the LLM provider/model verbatim.
	if got, ok := mem.Options["embedding_provider"]; ok {
		t.Errorf("embedding_provider = %q, want unset", got)
	}
	if got, ok := mem.Options["embedding_model"]; ok {
		t.Errorf("embedding_model = %q, want unset", got)
	}
}

func TestApplyEmbeddingDefaults_UserSettingsPreserved(t *testing.T) {
	mem := &MemoryConfig{
		Enabled:  true,
		Provider: "pgvector",
		Options: map[string]string{
			"embedding_provider": "ollama",
			"embedding_model":    "mxbai-embed-large",
		},
	}
	llm := &LLMConfig{Provider: "openai", Model: "gpt-4o", APIKey: "sk-test"}

	applyEmbeddingDefaults(mem, llm)

	if got := mem.Options["embedding_provider"]; got != "ollama" {
		t.Errorf("embedding_provider = %q, want user-provided %q", got, "ollama")
	}
	if got := mem.Options["embedding_model"]; got != "mxbai-embed-large" {
		t.Errorf("embedding_model = %q, want user-provided %q", got, "mxbai-embed-large")
	}
	if got := mem.Options["dimensions"]; got != "1024" {
		t.Errorf("dimensions = %q, want %q (mxbai-embed-large)", got, "1024")
	}
}

func TestDimensionsForEmbeddingModel(t *testing.T) {
	cases := map[string]int{
		"nomic-embed-text":        768,
		"nomic-embed-text:latest": 768, // Ollama tag suffix ignored
		"Text-Embedding-3-Large":  3072,
		"unknown-model":           0,
		"":                        0,
	}
	for model, want := range cases {
		if got := dimensionsForEmbeddingModel(model); got != want {
			t.Errorf("dimensionsForEmbeddingModel(%q) = %d, want %d", model, got, want)
		}
	}
}
