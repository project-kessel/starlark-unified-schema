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
			s, ok := v.(starlark.String)
			if !ok {
				return starlark.None, fmt.Errorf("println: expected string, got %s", v.Type())
			}
			chunks = append(chunks, string(s))
		}

		fmt.Println(chunks)
		return starlark.None, nil
	})
}
