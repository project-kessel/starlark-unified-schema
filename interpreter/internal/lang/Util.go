package lang

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func convert_to_string(v starlark.Value) (string, error) {
	if s, ok := v.(starlark.String); ok {
		return string(s), nil
	} else {
		return "", fmt.Errorf("unable to convert Starlark value of type %s to string", v.Type())
	}
}

func convert_to_callable(v starlark.Value) (starlark.Callable, error) {
	if c, ok := v.(starlark.Callable); ok {
		return c, nil
	} else {
		return nil, fmt.Errorf("unable to convert Starlark value of type %s to Callable", v.Type())
	}
}

func get_bool(name string, structure *starlarkstruct.Struct) (bool, error) {
	v, err := structure.Attr(name)
	if err != nil {
		return false, fmt.Errorf("error access member %s of struct %+v: %w", name, structure, err)
	}

	if b, ok := v.(starlark.Bool); ok {
		return bool(b), nil
	} else {
		return false, fmt.Errorf("unable to convert Starlark value of type %s to bool", v.Type())
	}
}

func get_string(name string, structure *starlarkstruct.Struct) (string, error) {
	v, err := structure.Attr(name)
	if err != nil {
		return "", fmt.Errorf("error accessing member %s of struct %+v: %w", name, structure, err)
	}

	return convert_to_string(v)
}

func get_optional_string(name string, structure *starlarkstruct.Struct) (*string, error) {
	v, err := structure.Attr(name)
	if err != nil {
		return nil, fmt.Errorf("error accessing member %s of struct %+v: %w", name, structure, err)
	}

	if _, ok := v.(starlark.NoneType); ok {
		return nil, nil
	}

	s, err := convert_to_string(v)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
