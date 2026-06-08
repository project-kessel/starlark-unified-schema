package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SpyVisitor struct {
	root node
}

func NewSpyVisitor() *SpyVisitor {
	return &SpyVisitor{
		root: make(node),
	}
}

/***
The idea here is to implement this visitor alongside the others where each function captures the available data as a node (an alias for map[string]any) and returns it.
Then at the top, some top-level container (ex: VisitReporter) becomes a container on the root.
The content of a visitor can then be asserted as equivalent to a given golden json.

***/

func (v *SpyVisitor) TempAdd(name string, value any) {
	//This function (and associated test) can be removed when there are actual functions
	v.root[name] = value
}

type node map[string]any

func (v *SpyVisitor) AssertJSON(t *testing.T, expected string) bool {
	actual, err := json.Marshal(v.root)
	if !assert.NoError(t, err) {
		return false
	}

	return assert.JSONEq(t, expected, string(actual))
}

func (V *SpyVisitor) VisitAnd(left any, right any) any {
	return node{
		"kind":  "and",
		"left":  left,
		"right": right,
	}
}

func (V *SpyVisitor) VisitOr(left any, right any) any {
	return node{
		"kind":  "or",
		"left":  left,
		"right": right,
	}
}

func (V *SpyVisitor) VisitUnless(left any, right any) any {
	return node{
		"kind":  "unless",
		"left":  left,
		"right": right,
	}
}

func (V *SpyVisitor) VisitRelationExpression(name string) any {
	return node{
		"kind": "ref",
		"name": name,
	}
}

func (V *SpyVisitor) VisitSubRelationExpression(name string, sub string) any {
	return node{
		"kind": "subref",
		"name": name,
		"sub":  sub,
	}
}

func (V *SpyVisitor) VisitAssignableExpression(typeNamespace string, typeName string, cardinality string) any {
	return node{
		"kind":          "assignable",
		"typeNamespace": typeNamespace,
		"typeName":      typeName,
		"cardinality":   cardinality,
	}
}

func (V *SpyVisitor) BeginRelation(name string) {

}

// Construct relation expression
func (V *SpyVisitor) VisitRelation(name string, body any) any {
	return node{
		"kind": "relation",
		"name": name,
		"body": body,
	}
}

func (V *SpyVisitor) BeginType(namespace string, name string) {

}

// Construct type expression
func (V *SpyVisitor) VisitType(namespace string, name string, relations []any) any {
	return node{
		"kind":      "type",
		"namespace": namespace,
		"name":      name,
		"relations": relations,
	}
}
