package lang

import (
	"go.starlark.net/starlarkstruct"
)

func registerDefaultBuiltins(l *Loader, registry *ResourceRegistry) {
	l.RegisterBuiltin("struct", starlarkstruct.Make)
	l.RegisterBuiltin("resource", resourceBuiltin(registry))
}
