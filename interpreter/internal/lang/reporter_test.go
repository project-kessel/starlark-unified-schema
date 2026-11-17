package lang

import (
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/util"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func TestReporterDefaultsAreModuleScoped(t *testing.T) {
	for _, first := range []string{"parent.star", "child.star"} {
		t.Run(first, func(t *testing.T) {
			loader, reader, thread := createDefaultLoaderReaderAndThread()
			require.NoError(t, reader.AddRealSchemaFile("kessel.star"))
			util.AddFile(t, reader, "parent.star", `
load("kessel.star", "resource", "uuid")
reporter("rbac")
workspace = resource(id_type=uuid())
`)
			util.AddFile(t, reader, "child.star", `
load("kessel.star", "resource", "uuid")
reporter("features")
load("parent.star", parent="workspace")
workspace = resource(extends=parent)
override = resource(reporter="other", id_type=uuid())
after_override = resource(id_type=uuid())
`)
			_, err := loader.Load(thread, first)
			require.NoError(t, err)
			parent, err := loader.Load(thread, "parent.star")
			require.NoError(t, err)
			child, err := loader.Load(thread, "child.star")
			require.NoError(t, err)
			for _, tc := range []struct {
				globals        starlark.StringDict
				name, reporter string
			}{
				{parent, "workspace", "rbac"},
				{child, "workspace", "features"},
				{child, "override", "other"},
				{child, "after_override", "features"},
			} {
				value, err := tc.globals[tc.name].(*starlarkstruct.Struct).Attr("reporter")
				require.NoError(t, err)
				require.Equal(t, starlark.String(tc.reporter), value)
			}
		})
	}
}

func TestReporterErrors(t *testing.T) {
	for _, tc := range []struct{ name, source, want string }{
		{"missing", `value = resolve_reporter()`, "no default reporter"},
		{"before declaration", `value = resolve_reporter(); reporter("test")`, "no default reporter"},
		{"empty declaration", `reporter("")`, "name must not be empty"},
		{"invalid declaration", `reporter(1)`, "got int, want string"},
		{"duplicate", `reporter("test"); reporter("test")`, "default already declared"},
		{"different duplicate", `reporter("test"); reporter("other")`, "default already declared"},
		{"empty override", `reporter("test"); value = resolve_reporter("")`, "non-empty string"},
		{"invalid override", `reporter("test"); value = resolve_reporter(1)`, "non-empty string"},
		{"no imported default", `load("dependency.star", "value"); result = resolve_reporter()`, "no default reporter"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loader, reader, thread := createDefaultLoaderReaderAndThread()
			util.AddFile(t, reader, "dependency.star", `reporter("dependency"); value = 1`)
			util.AddFile(t, reader, "test.star", tc.source)
			_, err := loader.Load(thread, "test.star")
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestReporterRetryAfterFailedLoad(t *testing.T) {
	loader, reader, thread := createDefaultLoaderReaderAndThread()
	util.AddFile(t, reader, "test.star", `reporter("test"); fail("later failure")`)
	for range 2 {
		_, err := loader.Load(thread, "test.star")
		require.ErrorContains(t, err, "later failure")
	}
}

func TestReporterWrapperErrors(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{`add_v1only_permission("view")`, "no default reporter"},
		{`reporter("test"); add_v1only_permission("view", reporter="")`, "non-empty string"},
		{`reporter("test"); add_v1only_permission("view", reporter=1)`, "non-empty string"},
		{`add_v1only_permission("test", "view", "extra")`, "invalid legacy reporter arguments"},
		{`add_v1only_permission("test", "view", reporter="other")`, "invalid legacy reporter arguments"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			loader, reader, thread := createDefaultLoaderReaderAndThread()
			require.NoError(t, reader.AddRealSchemaFile("extensions/rbac.star"))
			util.AddFile(t, reader, "test.star", `load("extensions/rbac.star", "add_v1only_permission")`+"\n"+tc.source)
			_, err := loader.Load(thread, "test.star")
			require.ErrorContains(t, err, tc.want)
		})
	}
}
