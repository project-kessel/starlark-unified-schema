package jsonschema

import (
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisitorRejectsUnknownDataTypeKind(t *testing.T) {
	visitor := NewVisitor()
	resource := model.Resource{
		Name: "host",
		Reporters: map[string][]model.Field{
			"hbi": {
				{Name: "bad_field", Type: model.DataType{Kind: "unknown_kind"}},
			},
		},
	}

	err := resource.Accept(visitor)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown data type kind: "unknown_kind"`)
}
