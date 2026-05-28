package lang

import (
	"os"
	"path"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type Loader struct {
	path         string
	modules      map[string]starlark.StringDict
	opts         *syntax.FileOptions
	predeclared  starlark.StringDict
	module_names []string
}

func NewLoader(path string) *Loader {
	l := &Loader{
		path:         path,
		modules:      map[string]starlark.StringDict{},
		opts:         &syntax.FileOptions{},
		predeclared:  starlark.StringDict{},
		module_names: nil,
	}

	registerDefaultBuiltins(l)

	return l
}

func (l *Loader) Load(thread *starlark.Thread, name string) (starlark.StringDict, error) {
	if m, ok := l.modules[name]; ok {
		return m, nil
	}

	location := path.Join(l.path, name)
	contents, err := os.ReadFile(location)
	if err != nil {
		return nil, err
	}

	globals, err := starlark.ExecFileOptions(l.opts, thread, name, contents, l.predeclared)
	if err != nil {
		return nil, err
	}

	l.modules[name] = globals

	return globals, nil
}

func (l *Loader) RegisterBuiltin(name string, callback func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)) {
	v := starlark.NewBuiltin(name, callback)
	l.predeclared[name] = v
}
