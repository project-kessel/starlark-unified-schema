package lang

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
)

// CompileInMemory processes in-memory schema files with the given visitor.
// This is the generic form used by the public compile package.
func CompileInMemory(files map[string][]byte, visitor output.SchemaVisitor) error {
	reader := newInMemorySourceFileReader("schema")

	names := make([]string, 0, len(files))
	for name, contents := range files {
		if err := reader.AddFile(name, contents); err != nil {
			return fmt.Errorf("adding %s: %w", name, err)
		}
		if filepath.Ext(name) == ".star" {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	loader := newLoaderForReader("schema", reader)
	processor := NewProcessor(loader)

	if err := processor.Process(visitor, names...); err != nil {
		return err
	}

	return nil
}
