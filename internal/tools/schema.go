package tools

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

// InferSchema reflects Go struct T's fields and `json`/`jsonschema` tags into
// a JSON Schema object, for use as core.FunctionDefinition.Parameters.
//
// Mirrors eino's utils.InferTool (github.com/eino-contrib/jsonschema-backed
// goStruct2ParamsOneOf): the parameter schema is derived from a typed args
// struct instead of hand-authored per tool, so the schema can't drift from
// the struct Call() actually decodes.
func InferSchema[T any]() (map[string]interface{}, error) {
	r := &jsonschema.Reflector{
		DoNotReference:            true,
		ExpandedStruct:            true,
		AllowAdditionalProperties: false,
	}
	var zero T
	js := r.Reflect(zero)
	// $schema/$id (the latter defaulting to a repo URL) describe the schema
	// document itself, not the tool's parameters — a tool-call parameter
	// schema doesn't need them, so clear before marshaling rather than
	// stripping keys back out of the encoded map.
	js.Version = ""
	js.ID = ""

	b, err := json.Marshal(js)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}
