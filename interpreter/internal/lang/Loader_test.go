package lang

import (
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/util"
	"github.com/stretchr/testify/assert"

	"go.starlark.net/starlark"
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
	reader := NewInMemorySourceFileReader("schema")
	loader := NewLoaderForReader("schema", reader)
	thread := &starlark.Thread{
		Name: "test",
		Load: loader.Load,
	}

	return loader, reader, thread
}

func addSpyCallback(loader *Loader, f func(v string)) {
	loader.RegisterBuiltin("spy", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		v := ""
		err := starlark.UnpackPositionalArgs("spy", args, kwargs, 1, &v)
		if err != nil {
			return starlark.None, err
		}

		f(v)

		return starlark.None, nil
	})
}
