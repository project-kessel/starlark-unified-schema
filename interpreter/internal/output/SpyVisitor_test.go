package output

import "testing"

func TestTemp(t *testing.T) {
	spy := NewSpyVisitor()

	spy.BeginType("test", "resource")
	spy.VisitType("test", "resource", []any{})

	spy.AssertJSON(t,
		`
[{
	"kind":"type",
	"name":"resource",
	"namespace": "test",
	"relations":[]
}]`)
}
