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
