package lang

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"github.com/stretchr/testify/assert"
)

func TestAssignableResourceReference(t *testing.T) {
	assertSourceMatchesGolden(t,
		`
load("kessel.star", "atMostOne", "resource_type")
other = resource_type({})

resource = resource_type({
	"other": atMostOne(other)
	})
`,
		`
[{
	"kind":"type",
	"namespace":"test",
	"name":"other",
	"relations":[]
},
{
	"kind":"type", 
	"namespace":"test", 
	"name":"resource", 
	"relations":[{
		"kind":"relation",
		"name":"other",
		"body": {
			"kind":"assignable",
			"typeNamespace":"test",
			"typeName":"other",
			"cardinality":"AtMostOne"
		}
	}]
}]`)
}

func TestAssignableSelfReference(t *testing.T) {
	assertSourceMatchesGolden(t,
		`
load("kessel.star", "atMostOne", "self", "resource_type")

resource = resource_type({
	"parent": atMostOne(self())
	})
`,
		`
[{
	"kind":"type", 
	"namespace":"test", 
	"name":"resource", 
	"relations":[{
		"kind":"relation",
		"name":"parent",
		"body": {
			"kind":"assignable",
			"typeNamespace":"test",
			"typeName":"resource",
			"cardinality":"AtMostOne"
		}
	}]
}]`)
}

func TestPassthroughPermission(t *testing.T) {
	assertSourceMatchesGolden(t,
		`
load("kessel.star", "self", "boolean", "permissions", "resource_type")
resource = resource_type({
	"relation": boolean(self())
})

permissions(resource, {
	"permission": lambda r: r.relation
})
`,
		`
[{
	"kind":"type", 
	"namespace":"test", 
	"name":"resource", 
	"relations":[{
		"kind": "relation",
		"name": "permission",
		"body": {
			"kind":"ref",
			"name": "relation"
		}
	},
	{
		"kind":"relation",
		"name":"relation",
		"body": {
			"kind":"assignable",
			"typeNamespace":"test",
			"typeName":"resource",
			"cardinality":"Boolean"
		}
	}]
}]`)
}

func TestPermissionWithBinaryLogic(t *testing.T) {
	assertSourceMatchesGolden(t,
		`
load("kessel.star", "self", "boolean", "permissions", "resource_type")
resource = resource_type({
	"left": boolean(self()),
	"right": boolean(self())
})

permissions(resource, {
	"permission": lambda r: r.left.union(r.right)
})
`,
		`
[{
	"kind":"type", 
	"namespace":"test", 
	"name":"resource", 
	"relations":[{
		"kind":"relation",
		"name":"left",
		"body": {
			"kind":"assignable",
			"typeNamespace":"test",
			"typeName":"resource",
			"cardinality":"Boolean"
		}
	},
	{
		"kind": "relation",
		"name": "permission",
		"body": {
			"kind":"or",
			"left": {
				"kind":"ref",
				"name": "left"
			},
			"right": {
				"kind":"ref",
				"name": "right"
			}
		}
	},
	{
		"kind":"relation",
		"name":"right",
		"body": {
			"kind":"assignable",
			"typeNamespace":"test",
			"typeName":"resource",
			"cardinality":"Boolean"
		}
	}]
}]`)
}

func TestSubRefPermissionAcrossTypes(t *testing.T) {
	assertSourceMatchesGolden(t,
		`
load("kessel.star", "self", "boolean", "permissions", "resource_type", "atMostOne")
container = resource_type({
	"flag": boolean(self())
})

resource = resource_type({
	"container": atMostOne(container)
})

permissions(resource, {
	"permission": lambda r: r.container.flag
})
`,
		`
[{
	"kind":"type",
	"namespace":"test",
	"name":"container",
	"relations":[{
		"kind":"relation",
		"name":"flag",
		"body": {
			"kind":"assignable",
			"typeNamespace":"test",
			"typeName":"container",
			"cardinality":"Boolean"
		}
	}]
},
{
	"kind":"type",
	"namespace":"test",
	"name":"resource",
	"relations":[{
		"kind":"relation",
		"name":"container",
		"body": {
			"kind":"assignable",
			"typeNamespace":"test",
			"typeName":"container",
			"cardinality":"AtMostOne"
		}
	},
	{
		"kind":"relation",
		"name":"permission",
		"body": {
			"kind":"subref",
			"name":"container",
			"sub":"flag"
		}
	}]
}]`)
}

func TestRecursivePermission(t *testing.T) {
	assertSourceMatchesGolden(t,
		`
load("kessel.star", "self", "boolean", "permissions", "resource_type", "atMostOne")
resource = resource_type({
	"parent": atMostOne(self()),
	"flag": boolean(self())
})

permissions(resource, {
	"permission": lambda r: r.flag.union(r.parent.permission)
})
`,
		`
[{
	"kind":"type", 
	"namespace":"test", 
	"name":"resource", 
	"relations":[{
		"kind":"relation",
		"name":"flag",
		"body": {
			"kind":"assignable",
			"typeNamespace":"test",
			"typeName":"resource",
			"cardinality":"Boolean"
		}
	},
	{
		"kind":"relation",
		"name":"parent",
		"body": {
			"kind":"assignable",
			"typeNamespace":"test",
			"typeName":"resource",
			"cardinality":"AtMostOne"
		}
	},
	{
		"kind": "relation",
		"name": "permission",
		"body": {
			"kind":"or",
			"left": {
				"kind":"ref",
				"name": "flag"
			},
			"right": {
				"kind":"subref",
				"name": "parent",
				"sub": "permission"
			}
		}
	}]
}]`)
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
