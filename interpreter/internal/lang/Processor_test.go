package lang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/util"
	"github.com/stretchr/testify/assert"
)

func setupProcessorWithKessel(t *testing.T, reader *inmemorySourceFileReader) *Processor {
	t.Helper()

	if err := addRealSchemaFile(reader, "kessel.star"); err != nil {
		t.Fatalf("failed to add kessel.star: %v", err)
	}

	loader := newLoaderForReader("schema", reader)
	return NewProcessor(loader)
}

func processAndVisit(t *testing.T, processor *Processor) *util.SpyVisitor {
	t.Helper()

	spy := util.NewSpyVisitor()
	if err := processor.Process(spy); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	return spy
}

func TestProcessorMergesCommonAndReporterFields(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/common_representation.star", []byte(`
load("kessel.star", "field", "text")

host = {
    "workspace_id": field(type=text(), required=True),
}
`))

	reader.AddFile("host/reporters/hbi/host.star", []byte(`
load("kessel.star", "resource", "field", "uuid")
load("host/common_representation.star", common="host")

host = resource(reporter="hbi", id_type=uuid(), common=common, fields={
    "insights_id": field(type=uuid())
})
`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `{
		"host": {
			"common": {"fields": [{"name": "workspace_id", "required": true, "type": {"kind": "text", "minLength": null, "maxLength": null, "regex": null}}], "relations": []},
			"reporters": {
				"hbi": {"fields": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}], "relations": []}
			}
		}
	}`)
}

func TestProcessorCommonOnlyFileProducesNoResources(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/common_representation.star", []byte(`
load("kessel.star", "field", "text")

host = {
    "workspace_id": field(type=text(), required=True),
}
`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `{}`)
}

func TestProcessorDuplicateReporterReturnsError(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/reporters/hbi/host.star", []byte(`
load("kessel.star", "resource", "field", "uuid")

host = resource("hbi", id_type=uuid(), fields={
    "insights_id": field(type=uuid()),
})
`))

	reader.AddFile("host/reporters/hbi/duplicate.star", []byte(`
load("kessel.star", "resource", "field", "uuid")

host = resource("hbi", id_type=uuid(), fields={
    "satellite_id": field(type=uuid()),
})
`))

	spy := util.NewSpyVisitor()
	err := processor.Process(spy)

	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "registered more than once")
}

func TestProcessorSkipsLibraryModules(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `{}`)
}

func TestProcessorMultipleReportersMerge(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/common_representation.star", []byte(`
load("kessel.star", "field", "text")

host = {
    "workspace_id": field(type=text(), required=True),
}
`))

	reader.AddFile("host/reporters/hbi/host.star", []byte(`
load("kessel.star", "resource", "field", "uuid")
load("host/common_representation.star", common="host")

host = resource("hbi", id_type=uuid(), common=common, fields={
    "insights_id": field(type=uuid()),
})
`))

	reader.AddFile("host/reporters/acm/host.star", []byte(`
load("kessel.star", "resource", "field", "text", "uuid")
load("host/common_representation.star", common="host")

host = resource("acm", id_type=uuid(), common=common, fields={
    "cluster_id": field(type=text(), required=True),
})
`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `{
		"host": {
			"common": {"fields": [{"name": "workspace_id", "required": true, "type": {"kind": "text", "minLength": null, "maxLength": null, "regex": null}}], "relations": []},
			"reporters": {
				"acm": {"fields": [{"name": "cluster_id", "required": true, "type": {"kind": "text", "minLength": null, "maxLength": null, "regex": null}}], "relations": []},
				"hbi": {"fields": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}], "relations": []}
			}
		}
	}`)
}

func TestProcessorProcessesDependencyModuleAfterLoadCaching(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/reporters/rbac/host.star", []byte(`
load("kessel.star", "resource", "field", "text", "uuid")

host = resource("rbac", id_type=uuid(), fields={
    "role": field(type=text(), required=True),
})
`))

	reader.AddFile("host/reporters/hbi/host.star", []byte(`
load("kessel.star", "resource", "field", "uuid")
load("host/reporters/rbac/host.star", rbac_host="host")

host = resource("hbi", id_type=uuid(), fields={
    "insights_id": field(type=uuid()),
})
`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `{
		"host": {
			"common": null,
			"reporters": {
				"hbi": {"fields": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}], "relations": []},
				"rbac": {"fields": [{"name": "role", "required": true, "type": {"kind": "text", "minLength": null, "maxLength": null, "regex": null}}], "relations": []}
			}
		}
	}`)
}

func TestAssignableResourceReference(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/assignable_resource_reference.star", []byte(`
load("kessel.star", "atMostOne", "resource")
other = resource("test", {})

this_resource = resource("test", {
	"other": atMostOne(other)
	})
`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `{
	"this_resource": {
		"common": null,
		"reporters": {
			"test": [{"name": "other", "required": false, "type": {"kind": "assignable", "typeNamespace": "test", "typeName": "other", "cardinality": "AtMostOne"}}]
		}
	}
}

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

func addRealSchemaFile(reader *inmemorySourceFileReader, path string) error {
	contents, err := os.ReadFile(filepath.Join("../../../schema/", path))
	if err != nil {
		return err
	}
	return reader.AddFile(path, contents)
}
