package lang

import "go.starlark.net/starlark"

type ResourceDefinition struct {
	Name      string
	Common    *starlark.Dict
	Reporters map[string]*starlark.Dict
}
