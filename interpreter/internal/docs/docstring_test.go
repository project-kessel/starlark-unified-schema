package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDedent(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		want string
	}{
		{
			name: "empty",
			doc:  "",
			want: "",
		},
		{
			name: "single line",
			doc:  "Returns a thing.",
			want: "Returns a thing.",
		},
		{
			// The opening line starts after the quotes, so it carries no indent
			// of its own and must not drag the common prefix down to zero.
			name: "first line is not indented",
			doc:  "Summary.\n\n    Detail.\n    ",
			want: "Summary.\n\nDetail.",
		},
		{
			name: "nested indentation is preserved",
			doc:  "Summary.\n\n    Args:\n      x: a value.\n    ",
			want: "Summary.\n\nArgs:\n  x: a value.",
		},
		{
			// Module docstrings are written at column 0, so the common prefix is
			// empty and every line must survive untouched. Stripping more than the
			// prefix would flatten nested lists and fenced examples.
			name: "an empty common prefix strips nothing",
			doc:  "Summary.\n\n- a bullet\n  wrapped onto a second line\n\n```python\nf(\n    x,\n)\n```\n",
			want: "Summary.\n\n- a bullet\n  wrapped onto a second line\n\n```python\nf(\n    x,\n)\n```",
		},
		{
			name: "relative indentation survives a non-empty prefix",
			doc:  "Summary.\n\n    ```python\n    f(\n        x,\n    )\n    ```\n    ",
			want: "Summary.\n\n```python\nf(\n    x,\n)\n```",
		},
		{
			name: "blank lines do not constrain the prefix",
			doc:  "Summary.\n\n\n        Detail.\n        ",
			want: "Summary.\n\n\nDetail.",
		},
		{
			name: "tabs expand to spaces",
			doc:  "Summary.\n\n\tDetail.\n\t",
			want: "Summary.\n\nDetail.",
		},
		{
			name: "trailing whitespace is stripped",
			doc:  "Summary.   \n\n    Detail.   \n    ",
			want: "Summary.\n\nDetail.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, Dedent(test.doc))
		})
	}
}

// Dedent runs on whatever go.starlark.net hands back, so pin the shape it
// actually produces rather than only hand-written strings.
func TestDedentRealDocstring(t *testing.T) {
	functions := loadFunctions(t, "def f(x):\n"+
		"    \"\"\"Summary.\n"+
		"\n"+
		"    Args:\n"+
		"      x: a value.\n"+
		"    \"\"\"\n"+
		"    return x\n")

	fn := functions["f"]
	require.NotNil(t, fn)
	assert.Equal(t, "Summary.\n\nArgs:\n  x: a value.", Dedent(fn.Doc()))
}

func TestFormatDoc(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		want string
	}{
		{
			name: "prose passes through",
			doc:  "Summary.\n\nMore detail, with `markdown`.",
			want: "Summary.\n\nMore detail, with `markdown`.",
		},
		{
			name: "args become a bullet list",
			doc:  "Summary.\n\nArgs:\n  reporter: the reporter.\n  id_type: the identifier.",
			want: "Summary.\n\n**Args:**\n\n- `reporter`: the reporter.\n- `id_type`: the identifier.",
		},
		{
			name: "multi-line arg descriptions fold into the entry",
			doc:  "Summary.\n\nArgs:\n  reporter: the reporter,\n    continued here.",
			want: "Summary.\n\n**Args:**\n\n- `reporter`: the reporter, continued here.",
		},
		{
			name: "varargs entries are recognized",
			doc:  "Summary.\n\nArgs:\n  *refs: more references.",
			want: "Summary.\n\n**Args:**\n\n- `*refs`: more references.",
		},
		{
			name: "returns is bolded",
			doc:  "Summary.\n\nReturns:\n  A struct.",
			want: "Summary.\n\n**Returns:**\n\n  A struct.",
		},
		{
			name: "an unindented line ends the args block",
			doc:  "Summary.\n\nArgs:\n  x: a value.\n\nTrailing prose.",
			want: "Summary.\n\n**Args:**\n\n- `x`: a value.\n\nTrailing prose.",
		},
		{
			// The permission-proxy explanation lives in resource's docstring and
			// needs a fenced example; indented code must not be parsed as args.
			name: "fenced block inside args passes through",
			doc:  "Summary.\n\nArgs:\n  x: a value.\n\n  ```python\n  f(a: b)\n  ```",
			want: "Summary.\n\n**Args:**\n\n- `x`: a value.\n\n  ```python\n  f(a: b)\n  ```",
		},
		{
			name: "fenced block after a section passes through",
			doc:  "Summary.\n\nExample:\n\n```python\nresource(a: b)\n```",
			want: "Summary.\n\n**Example:**\n\n```python\nresource(a: b)\n```",
		},
		{
			name: "a colon in prose is not an arg entry",
			doc:  "Args:\n  note that: this is prose.",
			want: "**Args:**\n\nnote that: this is prose.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, FormatDoc(test.doc))
		})
	}
}

func TestFormatDocEmpty(t *testing.T) {
	assert.Equal(t, "", FormatDoc(Dedent("")))
}
