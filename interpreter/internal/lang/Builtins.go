package lang

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func registerDefaultBuiltins(l *Loader) {
	l.RegisterBuiltin("struct", starlarkstruct.Make)

	l.RegisterBuiltin("println", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

	// Invokes a KSL extension that this interpreter does not generate. Schema authors are expected to call thin
	// per-extension wrappers rather than this directly. Every keyword argument
	// becomes an extension parameter.
	l.RegisterBuiltin("call_ksl_extension", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var name, namespace string
		// kwargs is passed as nil because UnpackPositionalArgs rejects any
		// keyword argument, and here they are all extension parameters.
		if err := starlark.UnpackPositionalArgs("call_ksl_extension", args, nil, 2, &name, &namespace); err != nil {
			return nil, err
		}

		if name == "" {
			return nil, fmt.Errorf("call_ksl_extension: name is required")
		}
		// The extension lives in a namespace this interpreter never emits, so
		// there is nothing for an empty namespace to fall back to.
		if namespace == "" {
			return nil, fmt.Errorf("call_ksl_extension: namespace is required")
		}

		params := make(map[string]string, len(kwargs))
		for _, kwarg := range kwargs {
			key, err := convert_to_string(kwarg[0])
			if err != nil {
				return nil, fmt.Errorf("call_ksl_extension: %w", err)
			}

			value, err := convert_to_string(kwarg[1])
			if err != nil {
				return nil, fmt.Errorf("call_ksl_extension: parameter %s must be a string: %w", key, err)
			}

			params[key] = value
		}

		l.recordExtensionReference(name, namespace, params)
		return starlark.None, nil
	})
}
