package lang

import (
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const kesselStarContent = `
def resource(reporter, common={}, fields={}):
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

func setupProcessorWithKessel(t *testing.T, reader *InMemorySourceFileReader) *Processor {
	t.Helper()

	if err := reader.AddFile("kessel.star", []byte(kesselStarContent)); err != nil {
		t.Fatalf("failed to add kessel.star: %v", err)
	}

	loader := NewLoaderForReader("schema", reader)
	return NewProcessor(loader)
}

func mustProcessAll(t *testing.T, processor *Processor) []model.Resource {
	t.Helper()
	resources, err := processor.ProcessAll()
	require.NoError(t, err)
	return resources
}

func TestProcessorMergesCommonAndReporterFields(t *testing.T) {
	reader := NewInMemorySourceFileReader("schema")
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

host = resource("hbi", common, {
    "insights_id": field(type=uuid()),
})
`))

	resources := mustProcessAll(t, processor)

	require.Len(t, resources, 1)
	assert.Equal(t, "host", resources[0].Name)
	assert.Equal(t, []model.Field{
		{Name: "workspace_id", Required: true, Type: model.DataType{Kind: "text"}},
	}, resources[0].Common)
	assert.Equal(t, map[string][]model.Field{
		"hbi": {{Name: "insights_id", Type: model.DataType{Kind: "uuid"}}},
	}, resources[0].Reporters)
}

func TestProcessorCommonOnlyFileProducesNoResources(t *testing.T) {
	reader := NewInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/common_representation.star", []byte(`
load("kessel.star", "field", "text")

host = {
    "workspace_id": field(type=text(), required=True),
}
`))

	resources := mustProcessAll(t, processor)

	assert.Empty(t, resources)
}

func TestProcessorDuplicateReporterReturnsError(t *testing.T) {
	reader := NewInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/reporters/hbi/host.star", []byte(`
load("kessel.star", "resource", "field", "uuid")

host = resource("hbi", fields={
    "insights_id": field(type=uuid()),
})
`))

	reader.AddFile("host/reporters/hbi/duplicate.star", []byte(`
load("kessel.star", "resource", "field", "uuid")

host = resource("hbi", fields={
    "satellite_id": field(type=uuid()),
})
`))

	_, err := processor.ProcessAll()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "registered more than once")
}

func TestProcessorSkipsLibraryModules(t *testing.T) {
	reader := NewInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	resources := mustProcessAll(t, processor)

	assert.Empty(t, resources)
}

func TestProcessorRejectsNonStructFieldEntry(t *testing.T) {
	reader := NewInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/reporters/hbi/host.star", []byte(`
load("kessel.star", "resource", "field", "uuid")

host = resource("hbi", fields={
    "insights_id": "not a field struct",
})
`))

	_, err := processor.ProcessAll()

	require.Error(t, err)
	assert.Contains(t, err.Error(), `field insights_id: expected struct`)
}

func TestProcessorRejectsWrongFieldKind(t *testing.T) {
	reader := NewInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("host/reporters/hbi/host.star", []byte(`
load("kessel.star", "resource", "field", "uuid")

host = resource("hbi", fields={
    "insights_id": struct(kind="uuid"),
})
`))

	_, err := processor.ProcessAll()

	require.Error(t, err)
	assert.Contains(t, err.Error(), `field insights_id: expected kind "field"`)
}

func TestProcessorMultipleReportersMerge(t *testing.T) {
	reader := NewInMemorySourceFileReader("schema")
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

host = resource("hbi", common, {
    "insights_id": field(type=uuid()),
})
`))

	reader.AddFile("host/reporters/acm/host.star", []byte(`
load("kessel.star", "resource", "field", "text")
load("host/common_representation.star", common="host")

host = resource("acm", common, {
    "cluster_id": field(type=text(), required=True),
})
`))

	resources := mustProcessAll(t, processor)

	require.Len(t, resources, 1)
	assert.Equal(t, "host", resources[0].Name)
	assert.Equal(t, []model.Field{
		{Name: "workspace_id", Required: true, Type: model.DataType{Kind: "text"}},
	}, resources[0].Common)
	assert.Equal(t, map[string][]model.Field{
		"hbi": {{Name: "insights_id", Type: model.DataType{Kind: "uuid"}}},
		"acm": {{Name: "cluster_id", Required: true, Type: model.DataType{Kind: "text"}}},
	}, resources[0].Reporters)
}
