package lang

import (
	"testing"

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

	reader.AddFile("README.md", []byte{})
	reader.AddFile("hello.star", []byte{})

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

	reader.AddFile("hello.star", []byte(`spy("hello")`))

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

	reader.AddFile("values.star", []byte(`message = "hello"`))
	reader.AddFile("hello.star", []byte(`
load("values.star", "message")
spy(message)`))

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

func createDefaultLoaderReaderAndThread() (*Loader, *inmemorySourceFileReader, *starlark.Thread) {
	reader := newInMemorySourceFileReader("schema")
	loader := newLoaderForReader("schema", reader)
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
