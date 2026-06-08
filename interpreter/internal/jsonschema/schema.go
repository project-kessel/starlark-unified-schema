package jsonschema

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
