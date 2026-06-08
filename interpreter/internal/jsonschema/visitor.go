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

func (v *Visitor) VisitResource(resource model.Resource) error {
	commonSchema := v.buildFieldsSchema(resource.Common)
	commonSchema.SchemaURI = schemaURI
	v.Outputs = append(v.Outputs, OutputEntry{
		Path:   filepath.Join(resource.Name, "common_representation.json"),
		Schema: commonSchema,
	})

	for reporterName, fields := range resource.Reporters {
		reporterSchema := v.buildFieldsSchema(fields)
		reporterSchema.SchemaURI = schemaURI
		v.Outputs = append(v.Outputs, OutputEntry{
			Path:   filepath.Join(resource.Name, "reporters", reporterName, fmt.Sprintf("%s.json", resource.Name)),
			Schema: reporterSchema,
		})
	}

	return nil
}

func (v *Visitor) buildFieldsSchema(fields []model.Field) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: map[string]*Schema{},
	}

	var required []string
	for _, f := range fields {
		schema.Properties[f.Name] = v.convertField(f)
		if f.Required {
			required = append(required, f.Name)
		}
	}

	if len(required) > 0 {
		r := append([]string(nil), required...)
		schema.Required = &r
	} else {
		r := []string{}
		schema.Required = &r
	}

	return schema
}

func (v *Visitor) convertField(f model.Field) *Schema {
	schema := v.convertDataType(f.Type)
	if f.Description != nil {
		schema.Description = *f.Description
	}
	return schema
}

func (v *Visitor) convertDataType(dt model.DataType) *Schema {
	switch dt.Kind {
	case "text":
		s := &Schema{Type: "string"}
		s.MinLength = dt.MinLength
		s.MaxLength = dt.MaxLength
		if dt.Regex != nil {
			s.Pattern = *dt.Regex
		}
		return s

	case "uuid":
		return &Schema{Type: "string", Format: "uuid"}

	case "numeric_id":
		s := &Schema{Type: "integer"}
		s.Minimum = intPtrToFloatPtr(dt.Min)
		s.Maximum = intPtrToFloatPtr(dt.Max)
		return s

	case "boolean":
		return &Schema{Type: "boolean"}

	case "date_time":
		return &Schema{Type: "string", Format: "date-time"}

	case "enum":
		return &Schema{Type: "string", Enum: dt.Values}

	case "nullable":
		innerSchema := v.convertDataType(*dt.Inner)
		if innerSchema.OneOf != nil {
			schemas := make([]*Schema, len(innerSchema.OneOf)+1)
			copy(schemas, innerSchema.OneOf)
			schemas[len(schemas)-1] = &Schema{Type: "null"}
			return &Schema{OneOf: schemas}
		}
		return &Schema{
			OneOf: []*Schema{innerSchema, {Type: "null"}},
		}

	case "union":
		schemas := make([]*Schema, len(dt.Members))
		for i, m := range dt.Members {
			schemas[i] = v.convertDataType(m)
		}
		return &Schema{OneOf: schemas}

	case "array":
		return &Schema{
			Type:  "array",
			Items: v.convertDataType(*dt.Items),
		}

	case "object":
		s := &Schema{
			Type:       "object",
			Properties: map[string]*Schema{},
		}
		var required []string
		for _, prop := range dt.Properties {
			s.Properties[prop.Name] = v.convertDataType(prop.Type)
			if prop.Required {
				required = append(required, prop.Name)
			}
		}
		if len(dt.Required) > 0 {
			r := append([]string(nil), dt.Required...)
			s.Required = &r
		} else if len(required) > 0 {
			s.Required = &required
		}
		return s

	default:
		return &Schema{}
	}
}

func intPtrToFloatPtr(v *int) *float64 {
	if v == nil {
		return nil
	}
	f := float64(*v)
	return &f
}
