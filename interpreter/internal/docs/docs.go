// Package docs generates an API reference for the Starlark DSL from the
// docstrings in the schema sources.
//
// Extraction is pure runtime introspection: a module is executed through
// lang.Loader, and the resulting globals are inspected for public functions.
// go.starlark.net records a def's docstring on the compiled function, so
// *starlark.Function supplies the name, parameters, defaults and doc text with
// no parsing of its own.
package docs

import (
	"fmt"
	"sort"

	"github.com/project-kessel/starlark-unified-schema/internal/lang"
	"go.starlark.net/starlark"
)

// moduleDocGlobal names the module-level blurb. A """docstring""" at the top of
// a file is not reachable at runtime — the toplevel function is compiled with
// its doc, but *starlark.Program exposes no accessor for it and ExecFileOptions
// returns only the globals — so modules declare prose through this global
// instead. The leading underscore keeps it out of load().
const moduleDocGlobal = "_module_doc"

// Reference is the documented surface of the DSL.
type Reference struct {
	Builtins []lang.BuiltinDoc
	Modules  []Module
}

// Module is one .star file that defines symbols schema authors load().
type Module struct {
	Path    string
	Doc     string
	Symbols []Symbol
}

// Symbol is a public function defined by a module.
type Symbol struct {
	Name      string
	Signature string
	Doc       string
	Line      int32
}

// Generate loads every module under the loader's source directory and collects
// the public functions each one defines.
//
// A module is included only if it defines at least one public function, which
// omits the schema files: they export resource structs and dicts, not an API.
func Generate(loader *lang.Loader) (*Reference, error) {
	builtins := loader.BuiltinDocs()
	for _, b := range builtins {
		// Builtins cannot be introspected, so a missing doc is a registration
		// bug rather than something to paper over with a blank entry.
		if b.Signature == "" || b.Doc == "" {
			return nil, fmt.Errorf("builtin %s is registered without a signature or doc", b.Name)
		}
	}

	names, err := loader.GetAllModuleNames()
	if err != nil {
		return nil, fmt.Errorf("listing modules: %w", err)
	}
	// ListFiles ranges a map for the in-memory reader, and the output is
	// committed, so the order has to be pinned here.
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)

	// A plain thread, deliberately not a Processor: leaving Loader.metadata nil
	// makes recordMetadata a no-op, which in turn drops the load-order
	// constraint that cross-resource relations impose on the compile path.
	thread := &starlark.Thread{Name: "docs", Load: loader.Load}

	ref := &Reference{Builtins: builtins}
	for _, name := range sorted {
		globals, err := loader.Load(thread, name)
		if err != nil {
			// Every module already has to load for `make build-schema` to pass,
			// so a failure here is real. Stopping beats emitting a partial
			// reference that silently drops a module's symbols.
			return nil, fmt.Errorf("documenting %s: %w", name, err)
		}

		module := documentModule(name, globals)
		if len(module.Symbols) == 0 {
			continue
		}
		ref.Modules = append(ref.Modules, module)
	}

	return ref, nil
}

func documentModule(path string, globals starlark.StringDict) Module {
	module := Module{Path: path}

	if doc, ok := globals[moduleDocGlobal].(starlark.String); ok {
		module.Doc = FormatDoc(Dedent(string(doc)))
	}

	// StringDict is a map; sort so ties in source position resolve the same way
	// on every run.
	names := make([]string, 0, len(globals))
	for name := range globals {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		// load()ed names bind file-locally and so never appear here; these are
		// only the symbols this module defines. Underscore names cannot be
		// load()ed, which makes the prefix the public/private boundary.
		if isPrivate(name) {
			continue
		}
		fn, ok := globals[name].(*starlark.Function)
		if !ok {
			continue
		}
		// A global can hold a function defined elsewhere — schema files bind
		// the accessor closures that extensions/*.star wrappers return, for
		// instance. Those are values, not API this module defines, and their
		// name and docstring belong to the other file.
		if fn.Position().Filename() != path {
			continue
		}
		module.Symbols = append(module.Symbols, Symbol{
			Name:      name,
			Signature: Signature(name, fn),
			Doc:       FormatDoc(Dedent(fn.Doc())),
			Line:      fn.Position().Line,
		})
	}

	// Source order groups related constructors together far better than
	// alphabetical order does.
	sort.SliceStable(module.Symbols, func(i, j int) bool {
		return module.Symbols[i].Line < module.Symbols[j].Line
	})

	return module
}

func isPrivate(name string) bool {
	return len(name) > 0 && name[0] == '_'
}
