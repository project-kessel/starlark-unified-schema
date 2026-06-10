package output

import (
	"fmt"
	"path/filepath"
	"sort"
)

const schemaURI = "http://json-schema.org/draft-07/schema#"

type node = map[string]any

type OutputEntry struct {
	Path   string
	Schema any
}

type JSONSchemaVisitor struct {
	root node
}

func NewJSONSchemaVisitor() *JSONSchemaVisitor {
	return &JSONSchemaVisitor{root: make(node)}
}

func (v *JSONSchemaVisitor) BeginType(name string) {}

func (v *JSONSchemaVisitor) VisitResource(typeName string, reporter string, commonFields []any, reporterFields []any) error {
	entry, exists := v.root[typeName].(node)
	if !exists {
		entry = node{"common": nil, "reporters": node{}}
		v.root[typeName] = entry
	}
	if commonFields != nil && entry["common"] == nil {
		entry["common"] = commonFields
	}
	if reporter != "" {
		reporters := entry["reporters"].(node)
		if _, dup := reporters[reporter]; dup {
			return fmt.Errorf("resource %s: reporter '%s' registered more than once", typeName, reporter)
		}
		reporters[reporter] = reporterFields
	}
	return nil
}

func (v *JSONSchemaVisitor) Results() []OutputEntry {
	var entries []OutputEntry

	typeNames := make([]string, 0, len(v.root))
	for name := range v.root {
		typeNames = append(typeNames, name)
	}
	sort.Strings(typeNames)

	for _, typeName := range typeNames {
		entry := v.root[typeName].(node)

		var commonFields []any
		if cf, ok := entry["common"].([]any); ok {
			commonFields = cf
		}
		commonSchema := buildObjectSchema(commonFields, nil)
		commonSchema["$schema"] = schemaURI
		entries = append(entries, OutputEntry{
			Path:   filepath.Join(typeName, "common_representation.json"),
			Schema: commonSchema,
		})

		reporters := entry["reporters"].(node)
		reporterNames := make([]string, 0, len(reporters))
		for name := range reporters {
			reporterNames = append(reporterNames, name)
		}
		sort.Strings(reporterNames)

		for _, reporterName := range reporterNames {
			var reporterFields []any
			if rf, ok := reporters[reporterName].([]any); ok {
				reporterFields = rf
			}
			reporterSchema := buildObjectSchema(reporterFields, nil)
			reporterSchema["$schema"] = schemaURI
			entries = append(entries, OutputEntry{
				Path:   filepath.Join(typeName, "reporters", reporterName, fmt.Sprintf("%s.json", typeName)),
				Schema: reporterSchema,
			})
		}
	}

	return entries
}

func buildObjectSchema(fields []any, explicitRequired []string) node {
	properties := node{}
	var derived []string
	for _, f := range fields {
		fn := f.(node)
		name := fn["name"].(string)
		properties[name] = fn["schema"]
		if fn["required"].(bool) {
			derived = append(derived, name)
		}
	}

	required := explicitRequired
	if required == nil {
		required = derived
	}
	if required == nil {
		required = []string{}
	}

	return node{"type": "object", "properties": properties, "required": required}
}

func (v *JSONSchemaVisitor) VisitDataField(name string, required bool, description *string, dataType any) any {
	n := dataType.(node)
	if description != nil {
		n["description"] = *description
	}
	return node{"name": name, "schema": n, "required": required}
}

func (v *JSONSchemaVisitor) VisitTextDataType(minLength *int, maxLength *int, regex *string) any {
	n := node{"type": "string"}
	if minLength != nil {
		n["minLength"] = *minLength
	}
	if maxLength != nil {
		n["maxLength"] = *maxLength
	}
	if regex != nil {
		n["pattern"] = *regex
	}
	return n
}

func (v *JSONSchemaVisitor) VisitUUIDDataType() any {
	return node{"type": "string", "format": "uuid"}
}

func (v *JSONSchemaVisitor) VisitNumericIDDataType(min *int, max *int) any {
	n := node{"type": "integer"}
	if min != nil {
		n["minimum"] = float64(*min)
	}
	if max != nil {
		n["maximum"] = float64(*max)
	}
	return n
}

func (v *JSONSchemaVisitor) VisitBooleanDataType() any {
	return node{"type": "boolean"}
}

func (v *JSONSchemaVisitor) VisitDateTimeDataType() any {
	return node{"type": "string", "format": "date-time"}
}

func (v *JSONSchemaVisitor) VisitEnumDataType(values []string) any {
	return node{"type": "string", "enum": values}
}

func (v *JSONSchemaVisitor) VisitNullableDataType(inner any) any {
	innerNode := inner.(node)
	if oneOf, ok := innerNode["oneOf"]; ok {
		schemas := append(oneOf.([]any), node{"type": "null"})
		return node{"oneOf": schemas}
	}
	return node{"oneOf": []any{innerNode, node{"type": "null"}}}
}

func (v *JSONSchemaVisitor) VisitCompositeDataType(dataTypes []any) any {
	return node{"oneOf": dataTypes}
}

func (v *JSONSchemaVisitor) VisitArrayDataType(items any) any {
	return node{"type": "array", "items": items}
}

func (v *JSONSchemaVisitor) VisitObjectDataType(properties []any, required []string) any {
	return buildObjectSchema(properties, required)
}
