package lang

import (
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func TestLoaderGetAllModulesWithEmptyDirectory(t *testing.T) {
	loader, _, _ := createDefaultLoaderReaderAndThread()

	modules, err := loader.GetAllModuleNames()

	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, modules, 0)
}

func TestLoaderIgnoresNonStarFiles(t *testing.T) {
	loader, reader, _ := createDefaultLoaderReaderAndThread()

	util.AddFile(t, reader, "README.md", "")
	util.AddFile(t, reader, "hello.star", "")

	names, err := loader.GetAllModuleNames()
	if !assert.NoError(t, err) {
		return
	}
	if !assert.ElementsMatch(t, []string{"hello.star"}, names) {
		return
	}
}

func TestLoaderWithSingleFile(t *testing.T) {
	values := []string{}
	loader, reader, thread := createDefaultLoaderReaderAndThread()
	addSpyCallback(loader, func(v string) { values = append(values, v) })

	util.AddFile(t, reader, "hello.star", `spy("hello")`)

	names, err := loader.GetAllModuleNames()
	if !assert.NoError(t, err) {
		return
	}
	if !assert.ElementsMatch(t, []string{"hello.star"}, names) {
		return
	}

	loader.Load(thread, "hello.star")

	assert.ElementsMatch(t, []string{"hello"}, values)

}

func TestLoaderWithDependency(t *testing.T) {
	values := []string{}
	loader, reader, thread := createDefaultLoaderReaderAndThread()
	addSpyCallback(loader, func(v string) { values = append(values, v) })

	util.AddFile(t, reader, "values.star", `message = "hello"`)
	util.AddFile(t, reader, "hello.star", `
load("values.star", "message")
spy(message)`)

	names, err := loader.GetAllModuleNames()
	if !assert.NoError(t, err) {
		return
	}
	if !assert.ElementsMatch(t, []string{"hello.star", "values.star"}, names) {
		return
	}

	_, err = loader.Load(thread, "hello.star")
	assert.NoError(t, err)

	assert.ElementsMatch(t, []string{"hello"}, values)
}

func TestLoaderWithNonResourceStruct(t *testing.T) {
	loader, reader, thread := createDefaultLoaderReaderAndThread()
	util.AddFile(t, reader, "hello.star", `
banana = struct(
	message = "hello"
)
`)

	metadata := map[resourceType]meta{}
	loader.SetMetadata(metadata)

	_, err := loader.Load(thread, "hello.star")
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, metadata, 0)
}

func createDefaultLoaderReaderAndThread() (*Loader, *InmemorySourceFileReader, *starlark.Thread) {
	reader := NewInMemorySourceFileReader("schema", "../../../schema")
	loader := NewLoaderForReader("schema", reader)
	thread := &starlark.Thread{
		Name: "test",
		Load: loader.Load,
	}

	return loader, reader, thread
}

func addSpyCallback(loader *Loader, f func(v string)) {
	loader.RegisterBuiltin("spy", "spy(value)", "Records a value for assertion in tests.",
		func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			v := ""
			err := starlark.UnpackPositionalArgs("spy", args, kwargs, 1, &v)
			if err != nil {
				return starlark.None, err
			}

			f(v)

			return starlark.None, nil
		})
}

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
