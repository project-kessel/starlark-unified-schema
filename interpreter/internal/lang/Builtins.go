package lang

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func registerDefaultBuiltins(l *Loader) {
	l.RegisterBuiltin("reporter", "reporter(name)",
		"Sets the default reporter for the file currently being evaluated. Call once, "+
			"before declaring resources or invoking extension wrappers. The name must be a "+
			"non-empty string. Defaults do not propagate through load() imports. "+
			"Pass reporter= explicitly on a call to override this default.",
		func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var name string
			if err := starlark.UnpackArgs("reporter", args, kwargs, "name", &name); err != nil {
				return nil, err
			}
			if name == "" {
				return nil, fmt.Errorf("reporter: name must not be empty")
			}
			if _, exists := l.reporters[l.currentModule]; exists {
				return nil, fmt.Errorf("reporter: default already declared for %s", l.currentModule)
			}
			l.reporters[l.currentModule] = name
			return starlark.None, nil
		})

	l.RegisterBuiltin("resolve_reporter", "resolve_reporter(value=None)",
		"Resolves an explicit reporter string or, for None, the default set by "+
			"reporter(name) in the file currently being evaluated. Fails if no default "+
			"is declared or the explicit value is not a non-empty string. Used by "+
			"resource() and extension wrappers; custom wrappers can use it too.",
		func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			value := starlark.Value(starlark.None)
			if err := starlark.UnpackArgs("resolve_reporter", args, kwargs, "value?", &value); err != nil {
				return nil, err
			}
			if value == starlark.None {
				if name, ok := l.reporters[l.currentModule]; ok {
					return starlark.String(name), nil
				}
				return nil, fmt.Errorf("resolve_reporter: no default reporter for %s; call reporter(name) first or pass reporter explicitly", l.currentModule)
			}
			if name, ok := starlark.AsString(value); ok && name != "" {
				return value, nil
			}
			return nil, fmt.Errorf("resolve_reporter: reporter must be a non-empty string")
		})

	l.RegisterBuiltin("struct", "struct(**fields)",
		"Creates an immutable struct from the given keyword arguments. Fields are "+
			"read back with attribute access (`s.name`), and the set of field names "+
			"is fixed at construction. Every DSL constructor in kessel.star returns "+
			"a struct tagged with a `kind` field.",
		starlarkstruct.Make)

	l.RegisterBuiltin("println", "println(*values)",
		"Prints its arguments to stdout, for debugging a schema while it compiles. "+
			"Not part of the schema itself — output goes to the terminal, never to a "+
			"generated artifact.",
		func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			chunks := make([]string, 0, len(args))
			for _, v := range args {
				s, err := convert_to_string(v)
				if err != nil {
					return starlark.None, err
				}

				chunks = append(chunks, s)
			}

			fmt.Println(chunks)
			return starlark.None, nil
		})

	const ksl_extension_backward_compatibility_call_name = "call_ksl_extension"
	l.RegisterBuiltin(ksl_extension_backward_compatibility_call_name,
		"call_ksl_extension(reporter, name, namespace, **params)",
		"Invokes a KSL extension that this interpreter does not generate. Every "+
			"keyword argument becomes an extension parameter, and the reference is "+
			"written to the given reporter's namespace.\n\n"+
			"Do not call this directly from a schema file. Call the thin "+
			"per-extension wrapper in `extensions/<namespace>.star` instead — see "+
			"schema/README.md.",
		func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var reporter, name, namespace string
			// kwargs is passed as nil because UnpackPositionalArgs rejects any
			// keyword argument, and here they are all extension parameters.
			if err := starlark.UnpackPositionalArgs(ksl_extension_backward_compatibility_call_name, args, nil, 3, &reporter, &name, &namespace); err != nil {
				return nil, err
			}

			if reporter == "" {
				return nil, fmt.Errorf("%s: reporter is required", ksl_extension_backward_compatibility_call_name)
			}
			if name == "" {
				return nil, fmt.Errorf("%s: name is required", ksl_extension_backward_compatibility_call_name)
			}
			// The extension lives in a namespace this interpreter never emits, so
			// there is nothing for an empty namespace to fall back to.
			if namespace == "" {
				return nil, fmt.Errorf("%s: namespace is required", ksl_extension_backward_compatibility_call_name)
			}

			params := make(map[string]string, len(kwargs))
			for _, kwarg := range kwargs {
				key, err := convert_to_string(kwarg[0])
				if err != nil {
					return nil, fmt.Errorf("%s: %w", ksl_extension_backward_compatibility_call_name, err)
				}

				value, err := convert_to_string(kwarg[1])
				if err != nil {
					return nil, fmt.Errorf("%s: parameter %s must be a string: %w", ksl_extension_backward_compatibility_call_name, key, err)
				}

				params[key] = value
			}

			l.recordExtensionReference(reporter, name, namespace, params)
			return starlark.None, nil
		})
}
