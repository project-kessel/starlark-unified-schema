package main

import (
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"github.com/project-kessel/starlark-unified-schema/internal/output/ksil"
	"github.com/project-kessel/starlark-unified-schema/internal/util"
	"github.com/stretchr/testify/require"
)

// Compare both artifact families through the real loader, DSL, processor, and
// visitors. In particular, wrapper functions belong to imported modules but
// must use the reporter of the schema that invokes them.
func TestReporterDefaultOutputCompatibility(t *testing.T) {
	const imports = `
load("kessel.star", "resource", "uuid")
load("extensions/rbac.star", "add_v1_based_permission", "add_v1only_permission", "add_contingent_permission", "add_unified_permission")
load("extensions/hbi.star", "expose_host_permission")
`
	cases := []struct{ name, source string }{
		{"legacy positional", `
item = resource("test", id_type=uuid())
add_v1_based_permission("test", "app", "items", "read", "view")
add_v1only_permission("test", "app_items_write")
add_contingent_permission("test", "view", "edit", "both")
add_unified_permission("test", "app", "items", "read")
expose_host_permission("test", "view", "host_view")
add_v1only_permission("other", "other_items_read")
`},
		{"default", `
reporter("test")
item = resource(id_type=uuid())
add_v1_based_permission("app", "items", "read", "view")
add_v1only_permission("app_items_write")
add_contingent_permission("view", "edit", "both")
add_unified_permission("app", "items", "read")
expose_host_permission("view", "host_view")
add_v1only_permission("other_items_read", reporter="other")
`},
		{"keyword override", `
reporter("unused")
item = resource(reporter="test", id_type=uuid())
add_v1_based_permission(reporter="test", app="app", resource="items", verb="read", v2_perm="view")
add_v1only_permission(reporter="test", perm="app_items_write")
add_contingent_permission(reporter="test", first="view", second="edit", contingent="both")
add_unified_permission(reporter="test", app="app", resource="items", verb="read")
expose_host_permission(reporter="test", v2_perm="view", host_perm="host_view")
add_v1only_permission(reporter="other", perm="other_items_read")
`},
	}
	var baseline map[string]string
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			processor, reader := setupForTest(t)
			for _, file := range []string{"extensions/rbac.star", "extensions/hbi.star"} {
				require.NoError(t, reader.AddRealSchemaFile(file))
			}
			util.AddFile(t, reader, "test.star", imports+tc.source)
			artifacts := map[string]string{}
			for format, visitor := range map[string]output.SchemaVisitor{
				"jsonschema/": output.NewJSONSchemaVisitor(),
				"ksil/":       ksil.NewKSILVisitor(),
			} {
				require.NoError(t, processor.Process(visitor, "test.star"))
				entries, err := visitor.Results()
				require.NoError(t, err)
				for _, entry := range entries {
					artifacts[format+entry.Path] = string(entry.Contents)
				}
			}
			require.Contains(t, artifacts, "ksil/test.json")
			require.Contains(t, artifacts, "ksil/other.json")
			require.Contains(t, artifacts, "jsonschema/item/reporters/test/item.json")
			if baseline == nil {
				baseline = artifacts
				return
			}
			require.Len(t, artifacts, len(baseline))
			for path, expected := range baseline {
				require.JSONEq(t, expected, artifacts[path], path)
			}
		})
	}
}
