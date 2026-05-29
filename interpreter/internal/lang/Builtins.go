package lang

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func (l *Loader) registerDefaultBuiltins() {
	l.predeclared["struct"] = starlark.NewBuiltin("struct", starlarkstruct.Make)

	l.predeclared["resource"] = starlark.NewBuiltin("resource", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var name string
		if err := starlark.UnpackPositionalArgs("resource", args, nil, 1, &name); err != nil {
			return starlark.None, err
		}

		def := ResourceDefinition{
			Name:      name,
			Reporters: map[string]*starlark.Dict{},
		}

		for _, kv := range kwargs {
			key := string(kv[0].(starlark.String))
			switch key {
			case "common":
				dict, ok := kv[1].(*starlark.Dict)
				if !ok {
					return starlark.None, fmt.Errorf("resource %s: 'common' must be a dict, got %s", name, kv[1].Type())
				}
				def.Common = dict

			case "reporters":
				outerDict, ok := kv[1].(*starlark.Dict)
				if !ok {
					return starlark.None, fmt.Errorf("resource %s: 'reporters' must be a dict, got %s", name, kv[1].Type())
				}
				for _, item := range outerDict.Items() {
					rKey, ok := item[0].(starlark.String)
					if !ok {
						return starlark.None, fmt.Errorf("resource %s: reporter key must be a string, got %s", name, item[0].Type())
					}
					reporterName := string(rKey)
					reporterDict, ok := item[1].(*starlark.Dict)
					if !ok {
						return starlark.None, fmt.Errorf("resource %s: reporter '%s' must be a dict, got %s", name, reporterName, item[1].Type())
					}
					def.Reporters[reporterName] = reporterDict
				}

			default:
				return starlark.None, fmt.Errorf("resource %s: unexpected keyword argument '%s'", name, key)
			}
		}

		l.resources = append(l.resources, def)
		return starlark.None, nil
	})
}
