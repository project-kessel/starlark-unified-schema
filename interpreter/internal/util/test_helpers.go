package util

import (
	"cmp"
	"slices"
	"strings"
	"testing"

	"github.com/project-kessel/ksl-schema-language/pkg/intermediate"
	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

// The following were copied from the output folder and probably need to be shared somehow
func VerifyJSONSchemaResults(t *testing.T, v output.SchemaVisitor, examples map[string]JsonSchemaTestCase) {
	t.Helper()

	results, err := v.Results()
	if !assert.NoError(t, err, "error getting results") {
		return
	}

	resultPaths := make(map[string]bool, len(results))
	for _, entry := range results {
		resultPaths[entry.Path] = true
	}
	for path := range examples {
		assert.True(t, resultPaths[path], "no result entry found for example path: %s", path)
	}

	for _, entry := range results {
		example, ok := examples[entry.Path]
		if !assert.True(t, ok, "no test case found for path: %s", entry.Path) {
			return
		}

		schema := gojsonschema.NewBytesLoader(entry.Contents)

		if example.Valid != nil {
			for _, valid := range example.Valid {
				data := gojsonschema.NewStringLoader(valid)
				result, err := gojsonschema.Validate(schema, data)

				if !assert.NoError(t, err, "error validating valid json") {
					return
				}

				assert.True(t, result.Valid(), "json was unexpectedly invalid. path=%s, contents=%s\n\nErrors: %+v", entry.Path, valid, result.Errors())
			}
		}

		if example.Invalid != nil {
			for _, invalid := range example.Invalid {
				data := gojsonschema.NewStringLoader(invalid)
				result, err := gojsonschema.Validate(schema, data)

				if !assert.NoError(t, err, "error validating invalid json") {
					return
				}

				assert.False(t, result.Valid(), "json was unexpectedly valid. path=%s, contents=%s", entry.Path, invalid)
			}
		}
	}
}

type JsonSchemaTestCase struct {
	Valid   []string
	Invalid []string
}

func VerifyKSILResults(t *testing.T, visitor output.SchemaVisitor, expected map[string]*intermediate.Namespace) {
	t.Helper()

	outputs, err := visitor.Results()
	if !assert.NoError(t, err) {
		return
	}

	used := map[string]bool{}

	for _, output := range outputs {
		expectedEntry, ok := expected[output.Path]
		if !assert.Truef(t, ok, "no expected entry found for path: %s", output.Path) {
			return
		}

		used[output.Path] = true

		_, err := expectedEntry.ToSemantic()
		assert.NoError(t, err, "expected result failed to project to a semantic model for path: %s", output.Path)

		actualNS, err := intermediate.Load(strings.NewReader(string(output.Contents)))
		if !assert.NoErrorf(t, err, "error loading output for entry %s into a namespace. Contents: %s", output.Path, output.Contents) {
			return
		}
		sortKSILNamespace(actualNS)
		sortKSILNamespace(expectedEntry)

		var expectedJS strings.Builder
		assert.NoError(t, intermediate.Store(expectedEntry, &expectedJS))

		var actualJS strings.Builder
		assert.NoError(t, intermediate.Store(actualNS, &actualJS))

		assert.JSONEqf(t, expectedJS.String(), actualJS.String(), "expected and actual json did not match. Expected: %s\n\nActual: %s", expectedJS.String(), string(output.Contents))
	}

	for path := range expected {
		assert.Truef(t, used[path], "Expected path %s was not used", path)
	}
}

func sortKSILNamespace(ns *intermediate.Namespace) {
	slices.SortStableFunc(ns.Types, func(l, r *intermediate.Type) int { return cmp.Compare(l.Name, r.Name) })
	for _, t := range ns.Types {
		slices.SortStableFunc(t.Relations, func(l, r *intermediate.Relation) int { return cmp.Compare(l.Name, r.Name) })
	}
	slices.SortStableFunc(ns.ExtensionDefinitions, func(l, r *intermediate.ExtensionDefinition) int { return cmp.Compare(l.Name, r.Name) })
}
