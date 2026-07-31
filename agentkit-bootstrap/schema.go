package agentkit

import (
	"encoding/json"
	"reflect"
	"strings"
)

// Schema is the subset of JSON Schema used to describe tool inputs and
// structured outputs. It is generated from Go types so a schema can never
// drift from the function signature it describes.
type Schema struct {
	Type                 string             `json:"type,omitempty"`
	Description          string             `json:"description,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	Enum                 []string           `json:"enum,omitempty"`
	AdditionalProperties *bool              `json:"additionalProperties,omitempty"`
}

// SchemaOf derives a JSON Schema from a Go type.
//
// Field names come from json tags. A field is required unless it is a pointer
// or carries omitempty. The jsonschema tag adds metadata:
//
//	Field string `json:"field" jsonschema:"description=What it is,enum=a|b"`
func SchemaOf[T any]() *Schema {
	var zero T
	return schemaOfType(reflect.TypeOf(&zero).Elem(), map[reflect.Type]bool{})
}

// SchemaJSON returns SchemaOf marshaled, for passing to providers.
func SchemaJSON[T any]() []byte {
	b, err := json.Marshal(SchemaOf[T]())
	if err != nil {
		// Schema values are plain data with no unmarshalable fields; a failure
		// here means the generator itself is broken.
		panic("agentkit: schema marshal failed: " + err.Error())
	}
	return b
}

func schemaOfType(t reflect.Type, seen map[reflect.Type]bool) *Schema {
	if t == nil {
		return &Schema{}
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Slice, reflect.Array:
		// []byte marshals as a base64 string, so describe it as one.
		if t.Elem().Kind() == reflect.Uint8 {
			return &Schema{Type: "string"}
		}
		return &Schema{Type: "array", Items: schemaOfType(t.Elem(), seen)}
	case reflect.Map:
		yes := true
		return &Schema{Type: "object", AdditionalProperties: &yes}
	case reflect.Interface:
		return &Schema{}
	case reflect.Struct:
		return structSchema(t, seen)
	default:
		return &Schema{}
	}
}

func structSchema(t reflect.Type, seen map[reflect.Type]bool) *Schema {
	// Recursive types would otherwise loop forever; emit an open object.
	if seen[t] {
		yes := true
		return &Schema{Type: "object", AdditionalProperties: &yes}
	}
	seen[t] = true
	defer delete(seen, t)

	s := &Schema{Type: "object", Properties: map[string]*Schema{}}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name, omitempty, skip := jsonFieldName(f)
		if skip {
			continue
		}

		// Embedded structs without a json name are flattened, matching
		// encoding/json.
		if f.Anonymous && name == "" {
			if inner := schemaOfType(f.Type, seen); inner.Properties != nil {
				for k, v := range inner.Properties {
					s.Properties[k] = v
				}
				s.Required = append(s.Required, inner.Required...)
			}
			continue
		}
		if name == "" {
			name = f.Name
		}

		fs := schemaOfType(f.Type, seen)
		required := !omitempty && f.Type.Kind() != reflect.Pointer
		applyJSONSchemaTag(f.Tag.Get("jsonschema"), fs, &required)

		s.Properties[name] = fs
		if required {
			s.Required = append(s.Required, name)
		}
	}
	return s
}

func jsonFieldName(f reflect.StructField) (name string, omitempty, skip bool) {
	tag := f.Tag.Get("json")
	if tag == "-" {
		return "", false, true
	}
	parts := strings.Split(tag, ",")
	if len(parts) > 0 {
		name = parts[0]
	}
	for _, o := range parts[1:] {
		if o == "omitempty" {
			omitempty = true
		}
	}
	if name == "" && !f.Anonymous {
		name = f.Name
	}
	return name, omitempty, false
}

// applyJSONSchemaTag reads `jsonschema:"description=...,enum=a|b,required"`.
func applyJSONSchemaTag(tag string, s *Schema, required *bool) {
	if tag == "" {
		return
	}
	for _, opt := range strings.Split(tag, ",") {
		switch {
		case opt == "required":
			*required = true
		case opt == "optional":
			*required = false
		case strings.HasPrefix(opt, "description="):
			s.Description = strings.TrimPrefix(opt, "description=")
		case strings.HasPrefix(opt, "enum="):
			s.Enum = strings.Split(strings.TrimPrefix(opt, "enum="), "|")
		}
	}
}
