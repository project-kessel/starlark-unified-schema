package lang

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"github.com/stretchr/testify/assert"
)

func TestTemp(t *testing.T) {
	assertSourceMatchesGolden(t,
		`
load("kessel.star", "atMostOne", "resource_type")
other = resource_type({})

resource = resource_type({
	"other": atMostOne(other)
	})
`,
		`
{
	"kind":"type", 
	"namespace":"test", 
	"name":"resource", 
	"relations":{
		"kind":"relation", 
		"name":"other", 
		"body": {
			"kind":"assignable",
			"typeNamespace":"test",
			"typeName":"other",
			"cardinality":"AtMostOne"
		}
	}
}`)
}

func assertSourceMatchesGolden(t *testing.T, source string, golden string) {
	visitor, reader, processor := createDefaultVisitorReaderAndProcessor()
	addRealSchemaFile(reader, "kessel.star")

	reader.AddFile("test.star", []byte(source))

	err := processor.ProcessModule("test.star", visitor)
	if !assert.NoError(t, err) {
		return
	}
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
	return reader.AddFile(path, contents)
}
