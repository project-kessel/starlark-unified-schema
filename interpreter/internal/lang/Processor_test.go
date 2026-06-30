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
			"common": {"fields": [{"name": "workspace_id", "required": true, "type": {"kind": "text"}}]},
			"reporters": {
				"hbi": {"fields": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}]}
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
		"common": {"fields": [{"name": "workspace_id", "required": true, "type": {"kind": "text"}}]},
		"reporters": {
			"acm": {"fields": [{"name": "cluster_id", "required": true, "type": {"kind": "text"}}]},
			"hbi": {"fields": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}]}}
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
		"common": {},
		"reporters": {
			"hbi": {"fields": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}]},
			"rbac": {"fields": [{"name": "role", "required": true, "type": {"kind": "text"}}]}
		}
	}
}`)
}

func TestAssignableResourceReference(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/assignable_resource_reference.star", []byte(`
load("kessel.star", "at_most_one", "resource", "uuid")
other = resource("test", id_type=uuid())

this_resource = resource("test", id_type=uuid(), fields={
	"other": at_most_one(other)
	})
`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `
{
	"other": {
		"common": {},
		"reporters": {
			"test": {}
		}
	},
	"this_resource": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "other", "cardinality": "AtMostOne", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "other"}
				]
			}
		}
	}
}`)
}

func TestAssignableSelfReference(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/assignable_self_reference.star", []byte(`
load("kessel.star", "at_most_one", "self", "resource", "uuid")

this_resource = resource("test", id_type=uuid(), fields={
	"parent": at_most_one(self())
	})
`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `
{
	"this_resource": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "parent", "cardinality": "AtMostOne", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"}
				]
			}
		}
	}
}`)
}

func TestPermissionLogicUnionIntersectExclude(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/relation_logic_union_intersect_exclude.star", []byte(`
load("kessel.star", "self", "wildcard", "resource", "uuid")

this_resource = resource("test", id_type=uuid(), fields={
	"relation1": wildcard(self()),
	"relation2": wildcard(self())
}, permissions={
	"union_perm": lambda r: r.relation1.union(r.relation2),
	"intersect_perm": lambda r: r.relation1.intersect(r.relation2),
	"exclude_perm": lambda r: r.relation1.exclude(r.relation2)
})`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `
{
	"this_resource": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "relation1", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"},
					{"kind": "relation", "name": "relation2", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"}
				],
				"permissions": [
					{
						"kind": "permission",
						"name": "union_perm",
						"body": {
							"kind": "or",
							"left": {"kind": "reference", "name": "relation1"},
							"right": {"kind": "reference", "name": "relation2"}
						}
					},
					{
						"kind": "permission",
						"name": "intersect_perm",
						"body": {
							"kind": "and",
							"left": {"kind": "reference", "name": "relation1"},
							"right": {"kind": "reference", "name": "relation2"}
						}
					},
					{
						"kind": "permission",
						"name": "exclude_perm",
						"body": {
							"kind": "unless",
							"left": {"kind": "reference", "name": "relation1"},
							"right": {"kind": "reference", "name": "relation2"}
						}
					}
				]
			}
		}
	}
}`)
}

func TestAnyPermission(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/any_permission.star", []byte(`
load("kessel.star", "self", "wildcard", "resource", "uuid", "any")

this_resource = resource("test", id_type=uuid(), fields={
	"r1": wildcard(self()),
	"r2": wildcard(self()),
	"r3": wildcard(self())
}, permissions={
	"any_one": lambda r: any(r.r1),
	"any_two": lambda r: any(r.r1, r.r2),
	"any_three": lambda r: any(r.r1, r.r2, r.r3)
})`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `
{
	"this_resource": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "r1", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"},
					{"kind": "relation", "name": "r2", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"},
					{"kind": "relation", "name": "r3", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"}
				],
				"permissions": [
					{
						"kind": "permission",
						"name": "any_one",
						"body": {"kind": "reference", "name": "r1"}
					},
					{
						"kind": "permission",
						"name": "any_two",
						"body": {
							"kind": "or",
							"left": {"kind": "reference", "name": "r1"},
							"right": {"kind": "reference", "name": "r2"}
						}
					},
					{
						"kind": "permission",
						"name": "any_three",
						"body": {
							"kind": "or",
							"left": {
								"kind": "or",
								"left": {"kind": "reference", "name": "r1"},
								"right": {"kind": "reference", "name": "r2"}
							},
							"right": {"kind": "reference", "name": "r3"}
						}
					}
				]
			}
		}
	}
}`)
}

func TestAllPermission(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/all_permission.star", []byte(`
load("kessel.star", "self", "wildcard", "resource", "uuid", "all")

this_resource = resource("test", id_type=uuid(), fields={
	"r1": wildcard(self()),
	"r2": wildcard(self()),
	"r3": wildcard(self())
}, permissions={
	"all_one": lambda r: all(r.r1),
	"all_two": lambda r: all(r.r1, r.r2),
	"all_three": lambda r: all(r.r1, r.r2, r.r3)
})`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `
{
	"this_resource": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "r1", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"},
					{"kind": "relation", "name": "r2", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"},
					{"kind": "relation", "name": "r3", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"}
				],
				"permissions": [
					{
						"kind": "permission",
						"name": "all_one",
						"body": {"kind": "reference", "name": "r1"}
					},
					{
						"kind": "permission",
						"name": "all_two",
						"body": {
							"kind": "and",
							"left": {"kind": "reference", "name": "r1"},
							"right": {"kind": "reference", "name": "r2"}
						}
					},
					{
						"kind": "permission",
						"name": "all_three",
						"body": {
							"kind": "and",
							"left": {
								"kind": "and",
								"left": {"kind": "reference", "name": "r1"},
								"right": {"kind": "reference", "name": "r2"}
							},
							"right": {"kind": "reference", "name": "r3"}
						}
					}
				]
			}
		}
	}
}`)
}

func TestPassthroughPermission(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/passthrough_permission.star", []byte(`
load("kessel.star", "self", "wildcard", "resource", "uuid")
this_resource = resource("test", id_type=uuid(), fields={
	"relation": wildcard(self())
}, permissions={
	"permission": lambda r: r.relation
})`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `
{
	"this_resource": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "relation", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"}
				],
				"permissions": [
					{"kind": "permission", "name": "permission", "body": {"kind": "reference", "name": "relation"}}
				]
			}
		}
	}
}`)
}

func TestPermissionWithBinaryLogic(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/permission_with_binary_logic.star", []byte(`
load("kessel.star", "self", "wildcard", "resource", "uuid")
this_resource = resource("test", id_type=uuid(),
fields={
	"left": wildcard(self()),
	"right": wildcard(self())
}, permissions={
	"permission": lambda r: r.left.union(r.right)
})`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `
{
	"this_resource": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "left", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"},
					{"kind": "relation", "name": "right", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"}
				],
				"permissions": [
					{
					"kind": "permission", 
					"name": "permission", 
					"body": {
						"kind": "or", "left": {"kind": "reference", "name": "left"}, "right": {"kind": "reference", "name": "right"}}
					}
				]
			}
		}
	}
}`)
}

func TestPermissionCallingPermission(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/permission_calling_permission.star", []byte(`
load("kessel.star", "self", "at_most_one", "resource", "uuid")
this_resource = resource("test", id_type=uuid(), 
fields={
	"relation": at_most_one(self())
}, permissions={
	"inner": lambda r: r.relation,
	"outer": lambda r: r.inner,
})`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `
{
	"this_resource": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "relation", "cardinality": "AtMostOne", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"}
				],
				"permissions": [
					{"kind": "permission", "name": "inner", "body": {"kind": "reference", "name": "relation"}},
					{"kind": "permission", "name": "outer", "body": {"kind": "reference", "name": "inner"}}
				]
			}
		}
	}
}`)
}

func TestSubRefPermissionAcrossTypes(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/subref_permission_across_types.star", []byte(`
load("kessel.star", "self", "wildcard", "resource", "uuid", "at_most_one")
container = resource("test", id_type=uuid(), fields={
	"flag": wildcard(self())
})

this_resource = resource("test", id_type=uuid(), fields={
	"container": at_most_one(container)
}, permissions={
	"permission": lambda r: r.container.flag
})`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `
{
	"container": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "flag", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "container"}
				]
			}
		}
	},
	"this_resource": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "container", "cardinality": "AtMostOne", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "container"}
				],
				"permissions": [
					{"kind": "permission", "name": "permission", "body": {"kind": "subreference", "name": "container", "sub": "flag"}}
				]
			}
		}
	}
}`)
}

func TestRecursivePermission(t *testing.T) {
	reader := newInMemorySourceFileReader("schema")
	processor := setupProcessorWithKessel(t, reader)

	reader.AddFile("test/recursive_permission.star", []byte(`
load("kessel.star", "self", "wildcard", "resource", "uuid", "at_most_one")
this_resource = resource("test", id_type=uuid(), fields={
	"parent": at_most_one(self()),
	"flag": wildcard(self())
}, permissions={
	"permission": lambda r: r.flag.union(r.parent.permission)
})`))

	spy := processAndVisit(t, processor)

	spy.AssertJSON(t, `
{
	"this_resource": {
		"common": {},
		"reporters": {
			"test": {
				"relations": [
					{"kind": "relation", "name": "parent", "cardinality": "AtMostOne", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"},
					{"kind": "relation", "name": "flag", "cardinality": "All", "dataType": {"kind": "uuid"}, "reporter": "test", "typeName": "this_resource"}
				],
				"permissions": [
					{"kind": "permission", "name": "permission", 
						"body": {
							"kind": "or", 
							"left": {
								"kind": "reference", 
								"name": "flag"
							}, 
							"right": {
								"kind": "subreference", 
								"name": "parent", 
								"sub": "permission"
							}
						}
					}
				]
			}
		}
	}
}`)
}

var loadedRealSchemaFiles = map[string][]byte{}

func addRealSchemaFile(reader *inmemorySourceFileReader, path string) error {
	if contents, ok := loadedRealSchemaFiles[path]; ok {
		return reader.AddFile(path, contents)
	}
	contents, err := os.ReadFile(filepath.Join("../../../schema/", path))
	if err != nil {
		return err
	}
	loadedRealSchemaFiles[path] = contents
	return reader.AddFile(path, contents)
}
