package docs

import (
	"strings"

	"go.starlark.net/starlark"
)

// Signature reconstructs Starlark call syntax for fn, e.g.
//
//	resource(reporter, id_type=None, common={}, fields={}, permissions={})
//	any(root, *refs)
//	f(a, *, b=1)
//
// name is the global the function is bound to, which is what a schema author
// calls it by. For a def that is fn.Name(), but a global can also hold a
// function defined under some other name.
func Signature(name string, fn *starlark.Function) string {
	// Parameter index layout, per go.starlark.net/starlark/value.go:781 and the
	// argument binding in eval.go: **kwargs is last, then *args, and the
	// remaining [0, n) are ordinary parameters whose final NumKwonlyParams()
	// entries are keyword-only.
	n := fn.NumParams()

	kwargsIndex := -1
	if fn.HasKwargs() {
		n--
		kwargsIndex = n
	}

	varargsIndex := -1
	if fn.HasVarargs() {
		n--
		varargsIndex = n
	}

	firstKwonly := n - fn.NumKwonlyParams()

	parts := make([]string, 0, fn.NumParams()+1)
	for i := 0; i < n; i++ {
		if i == firstKwonly {
			// Keyword-only parameters are introduced by *args, or by a bare "*"
			// when the function takes no varargs.
			if varargsIndex >= 0 {
				parts = append(parts, "*"+paramName(fn, varargsIndex))
			} else {
				parts = append(parts, "*")
			}
		}

		part := paramName(fn, i)
		if dflt := fn.ParamDefault(i); dflt != nil {
			part += "=" + renderDefault(dflt)
		}
		parts = append(parts, part)
	}

	// With no keyword-only parameters the loop above never reaches firstKwonly,
	// so *args still has to be emitted.
	if varargsIndex >= 0 && fn.NumKwonlyParams() == 0 {
		parts = append(parts, "*"+paramName(fn, varargsIndex))
	}

	if kwargsIndex >= 0 {
		parts = append(parts, "**"+paramName(fn, kwargsIndex))
	}

	return name + "(" + strings.Join(parts, ", ") + ")"
}

func paramName(fn *starlark.Function, i int) string {
	name, _ := fn.Param(i)
	return name
}

// renderDefault prints a default as Starlark source. Value.String() already
// emits literal syntax for every type that can appear as a default in this DSL:
// None, True/False, 5, "x", [], {}.
func renderDefault(v starlark.Value) string {
	s := v.String()
	// Values with no literal syntax print as <function f> and friends. Eliding
	// is better than emitting something a reader cannot copy into a schema.
	if strings.HasPrefix(s, "<") {
		return "..."
	}
	return s
}
