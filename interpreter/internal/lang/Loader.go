package lang

import (
	"fmt"
	"os"
	"path"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type Loader struct {
	path        string
	modules     map[string]starlark.StringDict
	resources   []ResourceDefinition
	opts        *syntax.FileOptions
	predeclared starlark.StringDict
}

func NewLoader(basePath string) *Loader {
	l := &Loader{
		path:        basePath,
		modules:     map[string]starlark.StringDict{},
		opts:        &syntax.FileOptions{},
		predeclared: starlark.StringDict{},
	}

	l.registerDefaultBuiltins()

	return l
}

func (l *Loader) Load(thread *starlark.Thread, name string) (starlark.StringDict, error) {
	if m, ok := l.modules[name]; ok {
		return m, nil
	}

	location := path.Join(l.path, name)
	contents, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("error reading module %s: %w", name, err)
	}

	globals, err := starlark.ExecFileOptions(l.opts, thread, name, contents, l.predeclared)
	if err != nil {
		return nil, err
	}

	l.modules[name] = globals
	return globals, nil
}

func (l *Loader) GetModuleNames() ([]string, error) {
	entries, err := os.ReadDir(l.path)
	if err != nil {
		return nil, fmt.Errorf("error reading schema directory %s: %w", l.path, err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".star") {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

func (l *Loader) Resources() []ResourceDefinition {
	return l.resources
}
