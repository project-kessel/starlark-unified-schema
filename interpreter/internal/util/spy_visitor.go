package util

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"github.com/stretchr/testify/assert"
)

type SpyVisitor struct {
	root node
}

var _ output.SchemaVisitor = (*SpyVisitor)(nil)

func NewSpyVisitor() *SpyVisitor {
	return &SpyVisitor{
		root: make(node),
	}
}

/*
The idea here is to implement this visitor alongside the others where each function captures the available data as a node (an alias for map[string]any) and returns it.
Then at the top, some top-level container (ex: VisitReporter) becomes a container on the root.
The content of a visitor can then be asserted as equivalent to a given golden json.
*/
type node map[string]any

func (v *SpyVisitor) BeginType(name string) {}

func (v *SpyVisitor) VisitResource(typeName string, reporter string, commonFields []any, reporterFields []any) error {
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

func (v *SpyVisitor) VisitDataField(name string, required bool, description *string, dataType any) any {
	result := node{"name": name, "required": required, "type": dataType}
	if description != nil {
		result["description"] = *description
	}
	return result
}

func (v *SpyVisitor) VisitTextDataType(minLength *int, maxLength *int, regex *string) any {
	return node{"kind": "text", "minLength": minLength, "maxLength": maxLength, "regex": regex}
}

func (v *SpyVisitor) VisitUUIDDataType() any {
	return node{"kind": "uuid"}
}

func (v *SpyVisitor) VisitNumericIDDataType(min *int, max *int) any {
	return node{"kind": "numeric_id", "min": min, "max": max}
}

func (v *SpyVisitor) VisitBooleanDataType() any {
	return node{"kind": "boolean"}
}

func (v *SpyVisitor) VisitDateTimeDataType() any {
	return node{"kind": "date_time"}
}

func (v *SpyVisitor) VisitEnumDataType(values []string) any {
	return node{"kind": "enum", "values": values}
}

func (v *SpyVisitor) VisitNullableDataType(inner any) any {
	return node{"kind": "nullable", "inner": inner}
}

func (v *SpyVisitor) VisitCompositeDataType(dataTypes []any) any {
	return node{"kind": "composite", "types": dataTypes}
}

func (v *SpyVisitor) VisitArrayDataType(items any) any {
	return node{"kind": "array", "items": items}
}

func (v *SpyVisitor) VisitObjectDataType(properties []any, required []string) any {
	return node{"kind": "object", "properties": properties, "required": required}
}

func (v *SpyVisitor) Results() ([]output.OutputEntry, error) {
	return nil, nil
}

func (v *SpyVisitor) AssertJSON(t *testing.T, expected string) bool {
	actual, err := json.Marshal(v.root)
	if !assert.NoError(t, err) {
		return false
	}
	return assert.JSONEq(t, expected, string(actual))
}
