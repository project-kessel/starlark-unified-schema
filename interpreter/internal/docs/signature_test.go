package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// loadFunctions executes src and returns its global functions by name.
func loadFunctions(t *testing.T, src string) map[string]*starlark.Function {
	t.Helper()

	globals, err := starlark.ExecFileOptions(&syntax.FileOptions{}, &starlark.Thread{}, "fixture.star", src, nil)
	require.NoError(t, err)

	functions := map[string]*starlark.Function{}
	for name, value := range globals {
		if fn, ok := value.(*starlark.Function); ok {
			functions[name] = fn
		}
	}
	return functions
}

func TestSignature(t *testing.T) {
	// One fixture covering every parameter shape Starlark allows, so the index
	// arithmetic in Signature is exercised end to end rather than mocked.
	functions := loadFunctions(t, `
def nullary(): pass
def required_only(a, b): pass
def defaults(none=None, no=False, num=5, s="x", list=[], dict={}): pass
def varargs(root, *refs): pass
def kwargs(name, **params): pass
def bare_star_kwonly(a, *, b): pass
def bare_star_kwonly_default(a, *, b=1): pass
def varargs_then_kwonly(a, *args, b=1): pass
def everything(a, b=2, *args, c, d=4, **kwargs): pass
`)

	tests := []struct {
		name string
		want string
	}{
		{"nullary", "nullary()"},
		{"required_only", "required_only(a, b)"},
		{"defaults", `defaults(none=None, no=False, num=5, s="x", list=[], dict={})`},
		{"varargs", "varargs(root, *refs)"},
		{"kwargs", "kwargs(name, **params)"},
		{"bare_star_kwonly", "bare_star_kwonly(a, *, b)"},
		{"bare_star_kwonly_default", "bare_star_kwonly_default(a, *, b=1)"},
		{"varargs_then_kwonly", "varargs_then_kwonly(a, *args, b=1)"},
		{"everything", "everything(a, b=2, *args, c, d=4, **kwargs)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fn := functions[test.name]
			require.NotNil(t, fn, "fixture has no %s", test.name)
			assert.Equal(t, test.want, Signature(test.name, fn))
		})
	}
}

// Signature renders the global a function is bound to, not the name it was
// defined under, so a heading and its signature can never disagree.
func TestSignatureUsesBindingName(t *testing.T) {
	functions := loadFunctions(t, `
def original(a): pass
alias = original
`)

	assert.Equal(t, "alias(a)", Signature("alias", functions["alias"]))
}
