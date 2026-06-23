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

func (v *SpyVisitor) VisitResource(typeName string, reporter string, commonMembers *output.Members, reporterMembers *output.Members) error {
	entry, exists := v.root[typeName].(node)
	if !exists {
		entry = createNode(map[string]any{"common": nil, "reporters": node{}})
		v.root[typeName] = entry
	}
	if commonMembers != nil && entry["common"] == nil {
		entry["common"] = createNode(map[string]any{"fields": commonMembers.DataFields, "relations": commonMembers.RelationFields, "permissions": commonMembers.Permissions})
	}
	if reporter != "" {
		reporters := entry["reporters"].(node)
		if _, dup := reporters[reporter]; dup {
			return fmt.Errorf("resource %s: reporter '%s' registered more than once", typeName, reporter)
		}
		reporters[reporter] = createNode(map[string]any{"fields": reporterMembers.DataFields, "relations": reporterMembers.RelationFields, "permissions": reporterMembers.Permissions})
	}

	return nil
}

func (v *SpyVisitor) VisitDataField(name string, required bool, description *string, dataType any) any {
	result := createNode(map[string]any{"name": name, "required": required, "type": dataType})
	if description != nil {
		result["description"] = *description
	}
	return result
}

func (v *SpyVisitor) VisitTextDataType(minLength *int, maxLength *int, regex *string) any {
	return createNode(map[string]any{"kind": "text", "minLength": minLength, "maxLength": maxLength, "regex": regex})
}

func (v *SpyVisitor) VisitUUIDDataType() any {
	return createNode(map[string]any{"kind": "uuid"})
}

func (v *SpyVisitor) VisitNumericIDDataType(min *int, max *int) any {
	return createNode(map[string]any{"kind": "numeric_id", "min": min, "max": max})
}

func (v *SpyVisitor) VisitBooleanDataType() any {
	return createNode(map[string]any{"kind": "boolean"})
}

func (v *SpyVisitor) VisitDateTimeDataType() any {
	return createNode(map[string]any{"kind": "date_time"})
}

func (v *SpyVisitor) VisitEnumDataType(values []string) any {
	return createNode(map[string]any{"kind": "enum", "values": values})
}

func (v *SpyVisitor) VisitNullableDataType(inner any) any {
	return createNode(map[string]any{"kind": "nullable", "inner": inner})
}

func (v *SpyVisitor) VisitCompositeDataType(dataTypes []any) any {
	return createNode(map[string]any{"kind": "composite", "types": dataTypes})
}

func (v *SpyVisitor) VisitArrayDataType(items any) any {
	return createNode(map[string]any{"kind": "array", "items": items})
}

func (v *SpyVisitor) VisitObjectDataType(properties []any, required []string) any { //TODO: the individual properties know if they're required, so why do we capture it separately? Also, what about field names?
	return createNode(map[string]any{"kind": "object", "properties": properties, "required": required})
}

func (v *SpyVisitor) VisitAnd(left any, right any) any {
	return createNode(map[string]any{"kind": "and", "left": left, "right": right})
}

func (v *SpyVisitor) VisitOr(left any, right any) any {
	return createNode(map[string]any{"kind": "or", "left": left, "right": right})
}

func (v *SpyVisitor) VisitUnless(left any, right any) any {
	return createNode(map[string]any{"kind": "unless", "left": left, "right": right})
}

func (v *SpyVisitor) VisitReferenceExpression(name string) any {
	return createNode(map[string]any{"kind": "reference", "name": name})
}

func (v *SpyVisitor) VisitSubReferenceExpression(name string, sub string) any {
	return createNode(map[string]any{"kind": "subreference", "name": name, "sub": sub})
}

func (v *SpyVisitor) VisitRelation(name string, reporter string, typeName string, cardinality string, dataType any) any {
	return createNode(map[string]any{"kind": "relation", "name": name, "reporter": reporter, "typeName": typeName, "cardinality": cardinality, "dataType": dataType})
}

func (v *SpyVisitor) BeginPermission(name string) {

}

func createNode(data map[string]any) node {
	result := node{}
	for key, value := range data {
		// Skip nil, empty slices, and empty maps, and nil string/int pointers
		switch v := value.(type) {
		case nil:
			continue
		case []any:
			if len(v) == 0 {
				continue
			}
		case map[string]any:
			if len(v) == 0 {
				continue
			}
		case *string:
			if v == nil {
				continue
			}
		case *int:
			if v == nil {
				continue
			}
		case *bool:
			if v == nil {
				continue
			}
		}

		result[key] = value
	}
	return result
}

// Construct relation expression
func (v *SpyVisitor) VisitPermission(name string, body any) any {
	return createNode(map[string]any{"kind": "relation", "name": name, "body": body})
}

func (v *SpyVisitor) Results() ([]output.OutputEntry, error) {
	return nil, nil
}

func (v *SpyVisitor) AssertJSON(t *testing.T, expected string) bool {
	actual, err := json.Marshal(v.root)
	if !assert.NoError(t, err) {
		return false
	}
	success := assert.JSONEq(t, expected, string(actual))
	if !success {
		t.Logf("Actual JSON did not match expected: %s", string(actual))
	}
	return success
}
