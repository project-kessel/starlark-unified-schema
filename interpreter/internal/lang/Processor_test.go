package lang

import (
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"github.com/stretchr/testify/assert"
)

const kesselStarContent = `
def resource(reporter="", common={}, fields={}):
    return struct(kind="resource", reporter=reporter, common=common, fields=fields)

def field(type, required=False, description=None):
    return struct(kind="field", type=type, required=required, description=description)

def text(minLength=None, maxLength=None, regex=None):
    return struct(kind="text", minLength=minLength, maxLength=maxLength, regex=regex)

def uuid():
    return struct(kind="uuid")

def nullable(inner):
    return struct(kind="nullable", inner=inner)

def union(left, right):
    return struct(kind="union", left=left, right=right)
`

func setupProcessorWithKessel(t *testing.T, reader *inmemorySourceFileReader) *Processor {
	t.Helper()

	if err := reader.AddFile("kessel.star", []byte(kesselStarContent)); err != nil {
		t.Fatalf("failed to add kessel.star: %v", err)
	}

	loader := newLoaderForReader("schema", reader)
	return NewProcessor(loader)
}

func processAndVisit(t *testing.T, processor *Processor) *output.SpyVisitor {
	t.Helper()

	if err := processor.ProcessAllModules(); err != nil {
		t.Fatalf("ProcessAllModules failed: %v", err)
	}

	spy := output.NewSpyVisitor()
	if err := processor.Visit(spy); err != nil {
		t.Fatalf("Visit failed: %v", err)
	}

	return spy
}

func TestProcessorMergesCommonAndReporterFields(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/common_representation.star", []byte(`
load("kessel.star", "resource", "field", "text")

host = resource(
    common = {
        "workspace_id": field(type=text(), required=True),
    },
)`))

	reader.AddFile("host/reporters/hbi/host.star", []byte(`
load("kessel.star", "resource", "field", "uuid")

host = resource(reporter="hbi",
    fields = {
        "insights_id": field(type=uuid()),
    },
)`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `{
		"host": {
			"common": [{"name": "workspace_id", "required": true, "type": {"kind": "text", "minLength": null, "maxLength": null, "regex": null}}],
			"reporters": {
				"hbi": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}]
			}
		}
	}`)
}

func TestProcessorCommonOnlyResource(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/common_representation.star", []byte(`
load("kessel.star", "resource", "field", "text")

host = resource(
    common = {
        "workspace_id": field(type=text(), required=True),
    },
)`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `{
		"host": {
			"common": [{"name": "workspace_id", "required": true, "type": {"kind": "text", "minLength": null, "maxLength": null, "regex": null}}],
			"reporters": {}
		}
	}`)
}

func TestProcessorDuplicateReporterReturnsError(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/reporters/hbi/host.star", []byte(`
load("kessel.star", "resource", "field", "uuid")

host = resource(reporter="hbi",
    fields = {
        "insights_id": field(type=uuid()),
    },
)`))

	reader.AddFile("host/reporters/hbi/duplicate.star", []byte(`
load("kessel.star", "resource", "field", "uuid")

host = resource(reporter="hbi",
    fields = {
        "satellite_id": field(type=uuid()),
    },
)`))

	err := processor.ProcessAllModules()

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

	reader.AddFile("host/reporters/hbi/host.star", []byte(`
load("kessel.star", "resource", "field", "uuid")

host = resource(reporter="hbi",
    fields = {
        "insights_id": field(type=uuid()),
    },
)`))

	reader.AddFile("host/reporters/acm/host.star", []byte(`
load("kessel.star", "resource", "field", "text")

host = resource(reporter="acm",
    fields = {
        "cluster_id": field(type=text(), required=True),
    },
)`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `{
		"host": {
			"common": null,
			"reporters": {
				"acm": [{"name": "cluster_id", "required": true, "type": {"kind": "text", "minLength": null, "maxLength": null, "regex": null}}],
				"hbi": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}]
			}
		}
	}`)
}
