package lang

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func getStringAttr(name string, s *starlarkstruct.Struct) (string, error) {
	v, err := s.Attr(name)
	if err != nil {
		return "", fmt.Errorf("error accessing member %s of struct %+v: %w", name, s, err)
	}

	str, ok := v.(starlark.String)
	if !ok {
		return "", fmt.Errorf("expected string for %s, got %s", name, v.Type())
	}
	return string(str), nil
}

func getBoolAttr(name string, s *starlarkstruct.Struct) (bool, error) {
	v, err := s.Attr(name)
	if err != nil {
		return false, fmt.Errorf("error accessing member %s of struct %+v: %w", name, s, err)
	}

	b, ok := v.(starlark.Bool)
	if !ok {
		return false, fmt.Errorf("expected bool for %s, got %s", name, v.Type())
	}
	return bool(b), nil
}

func getOptionalStringAttr(name string, s *starlarkstruct.Struct) *string {
	v, err := s.Attr(name)
	if err != nil {
		return nil
	}

	if _, ok := v.(starlark.NoneType); ok {
		return nil
	}

	if str, ok := v.(starlark.String); ok {
		result := string(str)
		return &result
	}
	return nil
}

func getOptionalIntAttr(name string, s *starlarkstruct.Struct) *int {
	v, err := s.Attr(name)
	if err != nil {
		return nil
	}

	if _, ok := v.(starlark.NoneType); ok {
		return nil
	}

	i, ok := v.(starlark.Int)
	if !ok {
		return nil
	}
	n, _ := i.Int64()
	result := int(n)
	return &result
}

func extractStringList(v starlark.Value, context string) ([]string, error) {
	list, ok := v.(*starlark.List)
	if !ok {
		return nil, fmt.Errorf("%s must be a list, got %s", context, v.Type())
	}
	result := make([]string, list.Len())
	for i := 0; i < list.Len(); i++ {
		s, ok := list.Index(i).(starlark.String)
		if !ok {
			return nil, fmt.Errorf("%s entry at index %d must be a string, got %s", context, i, list.Index(i).Type())
		}
		result[i] = string(s)
	}
	return result, nil
}
