package util

import (
	"encoding/json"
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/model"
	"github.com/stretchr/testify/assert"
)

type node map[string]any

type SpyVisitor struct {
	root node
}

func NewSpyVisitor() *SpyVisitor {
	return &SpyVisitor{
		root: make(node),
	}
}

func (v *SpyVisitor) VisitResource(resource model.Resource) error {
	commonFields := make([]any, 0, len(resource.Common))
	for _, f := range resource.Common {
		commonFields = append(commonFields, v.visitField(f))
	}

	reporterGroups := map[string][]any{}
	for name, fields := range resource.Reporters {
		group := make([]any, 0, len(fields))
		for _, f := range fields {
			group = append(group, v.visitField(f))
		}
		reporterGroups[name] = group
	}

	v.root[resource.Name] = node{"common": commonFields, "reporters": reporterGroups}
	return nil
}

func (v *SpyVisitor) visitField(f model.Field) node {
	result := node{"name": f.Name, "required": f.Required, "type": v.visitDataType(f.Type)}
	if f.Description != nil {
		result["description"] = *f.Description
	}
	return result
}

func (v *SpyVisitor) visitDataType(dt model.DataType) node {
	switch dt.Kind {
	case "text":
		return node{"kind": "text", "minLength": dt.MinLength, "maxLength": dt.MaxLength, "regex": dt.Regex}
	case "uuid":
		return node{"kind": "uuid"}
	case "numeric_id":
		return node{"kind": "numeric_id", "min": dt.Min, "max": dt.Max}
	case "boolean":
		return node{"kind": "boolean"}
	case "date_time":
		return node{"kind": "date_time"}
	case "enum":
		return node{"kind": "enum", "values": dt.Values}
	case "nullable":
		return node{"kind": "nullable", "inner": v.visitDataType(*dt.Inner)}
	case "union":
		types := make([]any, 0, len(dt.Members))
		for _, m := range dt.Members {
			types = append(types, v.visitDataType(m))
		}
		return node{"kind": "composite", "types": types}
	case "array":
		return node{"kind": "array", "items": v.visitDataType(*dt.Items)}
	case "object":
		props := make([]any, 0, len(dt.Properties))
		for _, p := range dt.Properties {
			props = append(props, v.visitField(p))
		}
		return node{"kind": "object", "properties": props, "required": dt.Required}
	default:
		return node{"kind": dt.Kind}
	}
}

func (v *SpyVisitor) AssertJSON(t *testing.T, expected string) bool {
	t.Helper()
	actual, err := json.Marshal(v.root)
	if !assert.NoError(t, err) {
		return false
	}
	return assert.JSONEq(t, expected, string(actual))
}
