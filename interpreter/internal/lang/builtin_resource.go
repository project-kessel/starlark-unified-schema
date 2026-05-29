package lang

import (
	"fmt"

	"go.starlark.net/starlark"
)

func resourceBuiltin(registry *ResourceRegistry) func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var name string
		var common *starlark.Dict
		var reporters *starlark.Dict
		if err := starlark.UnpackArgs("resource", args, kwargs,
			"name", &name,
			"common?", &common,
			"reporters?", &reporters,
		); err != nil {
			return starlark.None, err
		}

		def := ResourceDefinition{Name: name}

		if common != nil {
			def.Common = common
		}

		if reporters != nil {
			reportersMap, err := buildReportersMap(name, reporters)
			if err != nil {
				return starlark.None, err
			}
			def.Reporters = reportersMap
		}

		registry.Register(def)
		return starlark.None, nil
	}
}

func buildReportersMap(resourceName string, reporters *starlark.Dict) (map[string]*starlark.Dict, error) {
	result := map[string]*starlark.Dict{}
	for _, item := range reporters.Items() {
		rKey, ok := item[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("resource %s: reporter key must be a string, got %s", resourceName, item[0].Type())
		}
		reporterName := string(rKey)

		reporterDict, ok := item[1].(*starlark.Dict)
		if !ok {
			return nil, fmt.Errorf("resource %s: reporter '%s' must be a dict, got %s", resourceName, reporterName, item[1].Type())
		}
		result[reporterName] = reporterDict
	}
	return result, nil
}
