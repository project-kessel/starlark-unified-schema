package jsonschema

import (
	"fmt"
	"path/filepath"

	"github.com/project-kessel/starlark-unified-schema/internal/model"
)

type Visitor struct {
	Outputs []OutputEntry
}

func NewVisitor() *Visitor {
	return &Visitor{}
}

func (v *Visitor) VisitResource(r *model.Resource, common []any, reporters map[string][]any) (any, error) {
	commonSchema := buildObjectSchema(common)
	commonSchema.SchemaURI = schemaURI
	v.Outputs = append(v.Outputs, OutputEntry{
		Path:   filepath.Join(r.Name, "common_representation.json"),
		Schema: commonSchema,
	})

	for reporterName, fields := range reporters {
		reporterSchema := buildObjectSchema(fields)
		reporterSchema.SchemaURI = schemaURI
		v.Outputs = append(v.Outputs, OutputEntry{
			Path:   filepath.Join(r.Name, "reporters", reporterName, fmt.Sprintf("%s.json", r.Name)),
			Schema: reporterSchema,
		})
	}

	return nil, nil
}

func (v *Visitor) VisitField(f *model.Field, dataType any) (any, error) {
	schema := dataType.(*Schema)
	if f.Description != nil {
		schema.Description = *f.Description
	}
	return &namedSchema{name: f.Name, schema: schema, required: f.Required}, nil
}

func (v *Visitor) VisitDataType(dt *model.DataType, children []any) (any, error) {
	switch dt.Kind {
	case "text":
		s := &Schema{Type: "string"}
		s.MinLength = dt.MinLength
		s.MaxLength = dt.MaxLength
		if dt.Regex != nil {
			s.Pattern = *dt.Regex
		}
		return s, nil

	case "uuid":
		return &Schema{Type: "string", Format: "uuid"}, nil

	case "numeric_id":
		s := &Schema{Type: "integer"}
		s.Minimum = intPtrToFloatPtr(dt.Min)
		s.Maximum = intPtrToFloatPtr(dt.Max)
		return s, nil

	case "boolean":
		return &Schema{Type: "boolean"}, nil

	case "date_time":
		return &Schema{Type: "string", Format: "date-time"}, nil

	case "enum":
		return &Schema{Type: "string", Enum: dt.Values}, nil

	case "nullable":
		inner := children[0].(*Schema)
		if inner.OneOf != nil {
			schemas := make([]*Schema, len(inner.OneOf)+1)
			copy(schemas, inner.OneOf)
			schemas[len(schemas)-1] = &Schema{Type: "null"}
			return &Schema{OneOf: schemas}, nil
		}
		return &Schema{OneOf: []*Schema{inner, {Type: "null"}}}, nil

	case "union":
		schemas := make([]*Schema, len(children))
		for i, c := range children {
			schemas[i] = c.(*Schema)
		}
		return &Schema{OneOf: schemas}, nil

	case "array":
		return &Schema{Type: "array", Items: children[0].(*Schema)}, nil

	case "object":
		s := &Schema{
			Type:       "object",
			Properties: map[string]*Schema{},
		}
		var required []string
		for i, c := range children {
			ns := c.(*namedSchema)
			s.Properties[ns.name] = ns.schema
			if dt.Properties[i].Required {
				required = append(required, ns.name)
			}
		}
		if len(dt.Required) > 0 {
			r := append([]string(nil), dt.Required...)
			s.Required = &r
		} else if len(required) > 0 {
			s.Required = &required
		}
		return s, nil

	default:
		return nil, fmt.Errorf("unknown data type kind: %q", dt.Kind)
	}
}

func buildObjectSchema(fields []any) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: map[string]*Schema{},
	}

	var required []string
	for _, f := range fields {
		ns := f.(*namedSchema)
		schema.Properties[ns.name] = ns.schema
		if ns.required {
			required = append(required, ns.name)
		}
	}

	if len(required) > 0 {
		schema.Required = &required
	}

	return schema
}

type namedSchema struct {
	name     string
	schema   *Schema
	required bool
}

func intPtrToFloatPtr(v *int) *float64 {
	if v == nil {
		return nil
	}
	f := float64(*v)
	return &f
}
