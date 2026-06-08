package lang

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
)

func TestTemp(t *testing.T) {
	assertSourceMatchesGolden(t,
		`
load("kessel.star", "atMostOne")
other = {}

resource = {
	"other": atMostOne(other)
}
`,
		`{}`)
}

func assertSourceMatchesGolden(t *testing.T, source string, golden string) {
	visitor, reader, processor := createDefaultVisitorReaderAndProcessor()
	addRealSchemaFile(reader, "kessel.star")

	reader.AddFile("test.star", []byte(source))

	processor.ProcessModule("test.star", visitor)
	visitor.AssertJSON(t, golden)
}

func createDefaultVisitorReaderAndProcessor() (*output.SpyVisitor, *inmemorySourceFileReader, *Processor) {
	visitor := output.NewSpyVisitor()
	reader := newInMemorySourceFileReader("schema")
	loader := newLoaderForReader("schema", reader)
	processor := NewProcessor(loader)
	return visitor, reader, processor
}

func addRealSchemaFile(reader *inmemorySourceFileReader, path string) error {
	contents, err := os.ReadFile(filepath.Join("../../../schema/", path))
	if err != nil {
		return err
	}
	return reader.AddFile(filepath.Join(reader.path, path), contents)
}
