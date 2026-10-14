package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Bare invocation means "compile everything" and is relied on by the Makefile,
// CI, and .vscode/launch.json, so only an exact first-argument match may
// divert to the docs subcommand.
func TestIsDocsInvocation(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"bare invocation compiles", []string{"interpreter"}, false},
		{"src flag compiles", []string{"interpreter", "-src", "schema"}, false},
		{"explicit files compile", []string{"interpreter", "service/reporters/features/service.star"}, false},
		{"a file named docs.star compiles", []string{"interpreter", "docs.star"}, false},
		{"docs subcommand", []string{"interpreter", "docs"}, true},
		{"docs subcommand with flags", []string{"interpreter", "docs", "-o", "/tmp/API.md"}, true},
		{"docs not in first position compiles", []string{"interpreter", "-src", "schema", "docs"}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, isDocsInvocation(test.args))
		})
	}
}
