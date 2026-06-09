package lang

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type Loader struct {
	path         string
	modules      map[string]starlark.StringDict
	opts         *syntax.FileOptions
	predeclared  starlark.StringDict
	reader       sourceFileReader
	moduleNames []string
}

func NewLoader(path string) *Loader {
	return newLoaderForReader(path, &filesystemSourceFileReader{})
}

func newLoaderForReader(path string, reader sourceFileReader) *Loader {
	l := &Loader{
		path:         path,
		modules:      map[string]starlark.StringDict{},
		opts:         &syntax.FileOptions{},
		predeclared:  starlark.StringDict{},
		moduleNames: nil,
		reader:       reader,
	}

	registerDefaultBuiltins(l)

	return l
}

func (l *Loader) IsLoaded(name string) bool {
	_, ok := l.modules[name]
	return ok
}

func (l *Loader) Load(thread *starlark.Thread, name string) (starlark.StringDict, error) {
	if m, ok := l.modules[name]; ok {
		return m, nil
	}

	location := path.Join(l.path, name)
	contents, err := l.reader.ReadFile(location)
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

func (l *Loader) GetAllModuleNames() ([]string, error) {
	if l.moduleNames != nil {
		return l.moduleNames, nil
	}

	filepaths, err := l.reader.ListFiles(l.path)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(filepaths))

	for _, path := range filepaths {
		if filepath.Ext(path) != ".star" {
			continue
		}

		names = append(names, path)
	}

	l.moduleNames = names

	return names, nil
}

func (l *Loader) RegisterBuiltin(name string, callback func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)) {
	v := starlark.NewBuiltin(name, callback)
	l.predeclared[name] = v
}

type sourceFileReader interface {
	ReadFile(path string) ([]byte, error)
	ListFiles(path string) ([]string, error)
}

type filesystemSourceFileReader struct{}

func (fs *filesystemSourceFileReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (fs *filesystemSourceFileReader) ListFiles(root string) ([]string, error) {
	var names []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		names = append(names, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return names, nil
}

type inmemorySourceFileReader struct {
	path  string
	files map[string][]byte
}

func newInMemorySourceFileReader(path string) *inmemorySourceFileReader {
	return &inmemorySourceFileReader{
		path:  path,
		files: map[string][]byte{},
	}
}

func (im *inmemorySourceFileReader) AddFile(path string, contents []byte) error {
	pathWithStem := filepath.Join(im.path, path)

	if _, exists := im.files[pathWithStem]; exists {
		return os.ErrExist
	}

	im.files[pathWithStem] = contents

	return nil
}

func (im *inmemorySourceFileReader) ReadFile(path string) ([]byte, error) {
	if contents, found := im.files[path]; found {
		return contents, nil
	} else {
		return nil, os.ErrNotExist
	}
}

func (im *inmemorySourceFileReader) ListFiles(path string) ([]string, error) {
	names := make([]string, 0, len(im.files))

	for name, _ := range im.files {
		relative, err := filepath.Rel(path, name)
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(relative, "..") || relative == ".." {
			continue
		}

		names = append(names, relative)
	}

	return names, nil
}
