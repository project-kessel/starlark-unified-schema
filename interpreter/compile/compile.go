package compile

import (
	"github.com/project-kessel/starlark-unified-schema/internal/lang"
)

// Compile parses the given in-memory schema files (path -> source) and drives
// the supplied visitor with the parsed schema definitions.
//
// The files map must include the kessel.star prelude, as schema modules load it
// via load("kessel.star", ...). Modules are processed in sorted path order for
// deterministic output.
//
// This function touches no filesystem and runs unchanged under GOOS=js/wasm,
// making it suitable for browser-based schema processing.
//
// Returns an error if parsing fails or if the visitor returns an error during
// processing.
func Compile(files map[string][]byte, visitor SchemaVisitor) error {
	// Adapt the public visitor to the internal interface.
	adapter := &visitorAdapter{visitor: visitor}

	// Use the internal compile machinery.
	return lang.CompileInMemory(files, adapter)
}
