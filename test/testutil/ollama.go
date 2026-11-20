// Package testutil holds helpers shared by the repository's test packages.
package testutil

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

// OllamaBaseURL is the Ollama endpoint used by the real-LLM tests.
const OllamaBaseURL = "http://localhost:11434"

// SkipIfOllamaUnavailable skips the test when -short is set, when no Ollama
// server answers at OllamaBaseURL, or when model has not been pulled.
func SkipIfOllamaUnavailable(t testing.TB, model string) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping Ollama integration test in short mode")
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(OllamaBaseURL + "/api/tags")
	if err != nil {
		t.Skipf("skipping: Ollama not reachable at %s: %v", OllamaBaseURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("skipping: Ollama at %s returned %s", OllamaBaseURL, resp.Status)
	}

	var tags struct {
		Models []struct {
			Name  string `json:"name"`
			Model string `json:"model"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		t.Skipf("skipping: cannot decode Ollama model list: %v", err)
	}

	want := model
	if !strings.Contains(want, ":") {
		want += ":latest"
	}
	for _, m := range tags.Models {
		if m.Name == want || m.Model == want {
			return
		}
	}
	t.Skipf("skipping: Ollama model %q is not pulled (run 'ollama pull %s')", model, model)
}
