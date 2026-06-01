package output

import "testing"

func TestTemp(t *testing.T) {
	spy := NewSpyVisitor()

	spy.TempAdd("foo", "bar")

	spy.AssertJSON(t, `{"foo": "bar"}`)
}
