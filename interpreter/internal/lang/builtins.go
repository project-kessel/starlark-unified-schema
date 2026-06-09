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
			s, err := convertToString(v)
			if err != nil {
				return starlark.None, err
			}

			chunks = append(chunks, s)
		}

		fmt.Println(chunks)
		return starlark.None, nil
	})
}
