package agentkit_test

import (
	"encoding/json"
	"testing"

	"github.com/agenticgokit/agentkit"
)

func TestSchemaOfStruct(t *testing.T) {
	type nested struct {
		Code string `json:"code"`
	}
	type args struct {
		Query    string   `json:"query" jsonschema:"description=Search terms"`
		Limit    int      `json:"limit,omitempty"`
		Enabled  bool     `json:"enabled"`
		Ratio    float64  `json:"ratio"`
		Tags     []string `json:"tags,omitempty"`
		Optional *string  `json:"optional,omitempty"`
		Mode     string   `json:"mode" jsonschema:"enum=fast|slow"`
		Inner    nested   `json:"inner"`
		Hidden   string   `json:"-"`
		unexp    string   //nolint:unused // verifies unexported fields are skipped
	}

	s := agentkit.SchemaOf[args]()

	if s.Type != "object" {
		t.Fatalf("Type = %q, want object", s.Type)
	}

	for name, want := range map[string]string{
		"query":   "string",
		"limit":   "integer",
		"enabled": "boolean",
		"ratio":   "number",
		"tags":    "array",
		"inner":   "object",
	} {
		p := s.Properties[name]
		if p == nil {
			t.Errorf("property %q missing", name)
			continue
		}
		if p.Type != want {
			t.Errorf("property %q type = %q, want %q", name, p.Type, want)
		}
	}

	if _, ok := s.Properties["Hidden"]; ok {
		t.Error(`field tagged json:"-" leaked into the schema`)
	}
	if _, ok := s.Properties["unexp"]; ok {
		t.Error("unexported field leaked into the schema")
	}

	if got := s.Properties["query"].Description; got != "Search terms" {
		t.Errorf("description = %q, want %q", got, "Search terms")
	}
	if got := s.Properties["mode"].Enum; len(got) != 2 || got[0] != "fast" || got[1] != "slow" {
		t.Errorf("enum = %v, want [fast slow]", got)
	}
	if s.Properties["tags"].Items == nil || s.Properties["tags"].Items.Type != "string" {
		t.Error("array items type not derived")
	}
	if s.Properties["inner"].Properties["code"] == nil {
		t.Error("nested struct properties not derived")
	}

	required := map[string]bool{}
	for _, r := range s.Required {
		required[r] = true
	}
	if !required["query"] {
		t.Error("query should be required")
	}
	if required["limit"] {
		t.Error("omitempty field should not be required")
	}
	if required["optional"] {
		t.Error("pointer field should not be required")
	}
}

func TestSchemaHandlesRecursiveTypes(t *testing.T) {
	type node struct {
		Name     string  `json:"name"`
		Children []*node `json:"children,omitempty"`
	}
	// Must terminate rather than recurse forever.
	s := agentkit.SchemaOf[node]()
	if s.Properties["children"] == nil {
		t.Fatal("children property missing")
	}
}

func TestSchemaJSONIsValid(t *testing.T) {
	type out struct {
		Score int `json:"score"`
	}
	raw := agentkit.SchemaJSON[out]()
	var back map[string]any
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	if back["type"] != "object" {
		t.Errorf("type = %v, want object", back["type"])
	}
}
