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
	const ksl_extension_backward_compatibility_call_name = "call_ksl_extension"
	l.RegisterBuiltin(ksl_extension_backward_compatibility_call_name, func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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
