package output

import (
	"fmt"
	"path/filepath"
)

const schemaURI = "http://json-schema.org/draft-07/schema#"

type Schema struct {
	SchemaURI   string             `json:"$schema,omitempty"`
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    *[]string          `json:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	OneOf       []*Schema          `json:"oneOf,omitempty"`
	Enum        []string           `json:"enum,omitempty"`
	Description string             `json:"description,omitempty"`
	Pattern     string             `json:"pattern,omitempty"`
	MinLength   *int               `json:"minLength,omitempty"`
	MaxLength   *int               `json:"maxLength,omitempty"`
	Minimum     *float64           `json:"minimum,omitempty"`
	Maximum     *float64           `json:"maximum,omitempty"`
}

type OutputEntry struct {
	Path   string
	Schema *Schema
}

type JSONSchemaVisitor struct {
	Outputs []OutputEntry
}

func NewJSONSchemaVisitor() *JSONSchemaVisitor {
	return &JSONSchemaVisitor{}
}

func (v *JSONSchemaVisitor) BeginType(name string) {}

func (v *JSONSchemaVisitor) VisitType(name string, commonFields []any, reporterGroups map[string][]any) any {
	commonSchema := v.buildObjectSchema(commonFields, nil)
	commonSchema.SchemaURI = schemaURI
	v.Outputs = append(v.Outputs, OutputEntry{
		Path:   filepath.Join(name, "common_representation.json"),
		Schema: commonSchema,
	})

	for reporterName, fields := range reporterGroups {
		reporterSchema := v.buildObjectSchema(fields, nil)
		reporterSchema.SchemaURI = schemaURI
		v.Outputs = append(v.Outputs, OutputEntry{
			Path:   filepath.Join(name, "reporters", reporterName, fmt.Sprintf("%s.json", name)),
			Schema: reporterSchema,
		})
	}

	return nil
}

func (v *JSONSchemaVisitor) buildObjectSchema(fields []any, explicitRequired []string) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: map[string]*Schema{},
	}

	var derived []string
	for _, f := range fields {
		ns := f.(*namedSchema)
		schema.Properties[ns.name] = ns.schema
		if ns.required {
			derived = append(derived, ns.name)
		}
	}

	required := explicitRequired
	if required == nil {
		required = derived
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

func (v *JSONSchemaVisitor) VisitDataField(name string, required bool, description *string, dataType any) any {
	schema := dataType.(*Schema)
	if description != nil {
		schema.Description = *description
	}
	return &namedSchema{name: name, schema: schema, required: required}
}

func (v *JSONSchemaVisitor) VisitTextDataType(minLength *int, maxLength *int, regex *string) any {
	s := &Schema{Type: "string"}
	s.MinLength = minLength
	s.MaxLength = maxLength
	if regex != nil {
		s.Pattern = *regex
	}
	return s
}

func (v *JSONSchemaVisitor) VisitUUIDDataType() any {
	return &Schema{Type: "string", Format: "uuid"}
}

func (v *JSONSchemaVisitor) VisitNumericIDDataType(min *int, max *int) any {
	s := &Schema{Type: "integer"}
	s.Minimum = intPtrToFloatPtr(min)
	s.Maximum = intPtrToFloatPtr(max)
	return s
}

func (v *JSONSchemaVisitor) VisitBooleanDataType() any {
	return &Schema{Type: "boolean"}
}

func (v *JSONSchemaVisitor) VisitDateTimeDataType() any {
	return &Schema{Type: "string", Format: "date-time"}
}

func (v *JSONSchemaVisitor) VisitEnumDataType(values []string) any {
	return &Schema{Type: "string", Enum: values}
}

func (v *JSONSchemaVisitor) VisitNullableDataType(inner any) any {
	innerSchema := inner.(*Schema)

	if innerSchema.OneOf != nil {
		schemas := make([]*Schema, len(innerSchema.OneOf)+1)
		copy(schemas, innerSchema.OneOf)
		schemas[len(schemas)-1] = &Schema{Type: "null"}
		return &Schema{OneOf: schemas}
	}

	return &Schema{
		OneOf: []*Schema{innerSchema, {Type: "null"}},
	}
}

func (v *JSONSchemaVisitor) VisitCompositeDataType(dataTypes []any) any {
	schemas := make([]*Schema, len(dataTypes))
	for i, dt := range dataTypes {
		schemas[i] = dt.(*Schema)
	}
	return &Schema{OneOf: schemas}
}

func (v *JSONSchemaVisitor) VisitArrayDataType(items any) any {
	return &Schema{
		Type:  "array",
		Items: items.(*Schema),
	}
}

func (v *JSONSchemaVisitor) VisitObjectDataType(properties []any, required []string) any {
	return v.buildObjectSchema(properties, required)
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
