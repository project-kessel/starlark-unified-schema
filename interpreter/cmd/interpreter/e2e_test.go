package main

import (
	"encoding/json"
	"testing"

	"github.com/project-kessel/ksl-schema-language/pkg/intermediate"
	"github.com/project-kessel/starlark-unified-schema/internal/lang"
	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"github.com/project-kessel/starlark-unified-schema/internal/output/ksil"
	"github.com/project-kessel/starlark-unified-schema/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	processor, reader := setupForTest(t)

	util.AddFile(t, reader, "principal.star", `
load("kessel.star", "resource", "uuid")
principal = resource("test", id_type=uuid())
`)

	util.AddFile(t, reader, "container.star", `
load("kessel.star", "resource", "at_most_one", "many", "self", "uuid")
load("principal.star", "principal")
container = resource("test", id_type=uuid(), fields={
	"parent": at_most_one(self()),
	"direct_allowed_users": many(principal)
}, permissions={
	"allowed_users": lambda c: c.direct_allowed_users.union(c.parent.allowed_users)
})`)

	util.AddFile(t, reader, "special_container.star", `
load("kessel.star", "resource", "wildcard", "boolean")
load("principal.star", "principal")
load("container.star", test_container="container")
container = resource("special", extends=test_container, fields={
	"direct_flag": wildcard(principal),
	"direct_boolean_flag": boolean(principal)
}, permissions={
	"flag": lambda r: r.direct_flag.union(r.parent.flag)
})
`)

	util.AddFile(t, reader, "common.star", `
load("kessel.star", "one")
load("container.star", "container")
res = {
	"container": one(container)
}
`)

	util.AddFile(t, reader, "res.star", `
load("kessel.star", "resource", "uuid", "field", "nullable", "union", "text")
load("common.star", common="res")
res = resource("test", common=common, id_type=uuid(), fields={
	"system_1_id": field(type=nullable(union(uuid(), text(regex="^\\d{10}$")))),
	"system_2_id": field(type=nullable(uuid())),
	"hostname": field(type=nullable(text(maxLength=255)))
}, permissions={
	"operation": lambda r: r.container.allowed_users
})

call_ksl_extension("test", "role_binding", "rbac", relation="admin")
`)

	testNS := "test"

	verifyOutputs(t, processor, map[string]*intermediate.Namespace{
		"test.json": {
			Name: "test",
			Types: []*intermediate.Type{
				{
					Name:      "principal",
					Relations: []*intermediate.Relation{},
				},
				{
					Name: "container",
					Relations: []*intermediate.Relation{
						{
							Name: "parent",
							Body: &intermediate.RelationBody{
								Kind: "self",
								Types: []*intermediate.TypeReference{{
									Namespace: "test",
									Name:      "container",
								}},
								Cardinality: "AtMostOne",
							},
						},
						{
							Name: "direct_allowed_users",
							Body: &intermediate.RelationBody{
								Kind: "self",
								Types: []*intermediate.TypeReference{{
									Namespace: "test",
									Name:      "principal",
								}},
								Cardinality: "Any",
							},
						},
						{
							Name: "allowed_users",
							Body: &intermediate.RelationBody{
								Kind: "union",
								Left: &intermediate.RelationBody{
									Kind:     "reference",
									Relation: "direct_allowed_users",
								},
								Right: &intermediate.RelationBody{
									Kind:        "nested_reference",
									Relation:    "parent",
									SubRelation: "allowed_users",
								},
							},
						},
					},
				},
				{
					Name: "res",
					Relations: []*intermediate.Relation{
						{
							Name: "container",
							Body: &intermediate.RelationBody{
								Kind:        "self",
								Types:       []*intermediate.TypeReference{{Namespace: "test", Name: "container"}},
								Cardinality: "ExactlyOne",
							},
						}, //Inherited from common representation
						{
							Name: "operation",
							Body: &intermediate.RelationBody{
								Kind:        "nested_reference",
								Relation:    "container",
								SubRelation: "allowed_users",
							},
						},
					},
				},
			},
			ExtensionReferences: []*intermediate.ExtensionReference{{
				Namespace: "rbac",
				Name:      "role_binding",
				Params:    map[string]string{"relation": "admin"},
			}},
		},
		"special.json": {
			Name: "special",
			ExtensionDefinitions: []*intermediate.ExtensionDefinition{
				{
					Name:       "container",
					Visibility: "internal",
					Params:     []string{},
					Types: []*intermediate.DynamicType{
						{
							Name:       &intermediate.DynamicName{Kind: "literal", Value: "container"},
							Namespace:  &testNS,
							Visibility: "public",
							Relations: []*intermediate.DynamicRelation{
								{
									Name: &intermediate.DynamicName{Kind: "literal", Value: "special_container_direct_flag"},
									Body: intermediate.DynamicRelationBody{
										Kind:        "self",
										Types:       []*intermediate.TypeReference{{Namespace: "test", Name: "principal", All: true}},
										Cardinality: "Any",
									},
								},
								{
									Name: &intermediate.DynamicName{Kind: "literal", Value: "special_container_direct_boolean_flag"},
									Body: intermediate.DynamicRelationBody{
										Kind:        "self",
										Types:       []*intermediate.TypeReference{{Namespace: "test", Name: "principal", All: true}},
										Cardinality: "Any",
									},
								},
								{
									Name: &intermediate.DynamicName{Kind: "literal", Value: "special_container_flag"},
									Body: intermediate.DynamicRelationBody{
										Kind: "union",
										Left: &intermediate.DynamicRelationBody{
											Kind:     "reference",
											Relation: &intermediate.DynamicName{Kind: "literal", Value: "special_container_direct_flag"},
										},
										Right: &intermediate.DynamicRelationBody{
											Kind:        "nested_reference",
											Relation:    &intermediate.DynamicName{Kind: "literal", Value: "parent"},
											SubRelation: &intermediate.DynamicName{Kind: "literal", Value: "special_container_flag"},
										},
									},
								},
							},
						},
					},
				},
			},
			ExtensionReferences: []*intermediate.ExtensionReference{{
				Namespace: "special",
				Name:      "container",
				Params:    map[string]string{},
			}},
		},
	}, map[string]util.JsonSchemaTestCase{
		"principal/common_representation.json": {
			Valid: []string{
				`{}`,
			},
		},
		"principal/reporters/test/principal.json": {
			Valid: []string{
				`{}`,
			},
		},
		"container/common_representation.json": {
			Valid: []string{
				`{}`,
			},
		},
		"container/reporters/test/container.json": {
			Valid: []string{
				`{
					"parent": "f6e29208-ab8d-11f1-b92a-6ae246604903", 
					"direct_allowed_users": ["122eacf4-ab8e-11f1-a632-6ae246604903", "19388cae-ab8e-11f1-8936-6ae246604903"]
				}`,
			},
		},
		"container/reporters/special/container.json": {
			Valid: []string{
				`{}`,
				`{"direct_flag": "test/principal:*"}`,
				`{"direct_boolean_flag": true}`,
				`{"direct_boolean_flag": false}`,
				`{"direct_flag": "test/principal:*", "direct_boolean_flag": true}`,
			},
			Invalid: []string{
				`{"direct_flag": true}`,
				`{"direct_flag": "test/other:*"}`,
				`{"direct_boolean_flag": "test/principal:*"}`,
				`{"direct_boolean_flag": "true"}`,
				`{"direct_boolean_flag": 1}`,
			},
		},
		"res/common_representation.json": {
			Valid: []string{
				`{
					"container": "f9c7da5e-ab8e-11f1-bffd-6ae246604903"
				}`,
			},
		},
		"res/reporters/test/res.json": {
			Valid: []string{
				`{
					"system_1_id": "1234567890",
					"system_2_id": "85f476d6-ab8f-11f1-9e40-6ae246604903",
					"hostname": "the-host.the-domain.tld"
				}`,
				`{
					"system_1_id": "b06faed0-ab8f-11f1-b830-6ae246604903",
					"system_2_id": null,
					"hostname": "the-host.the-domain.tld"
				}`,
			},
		},
	})
}

func TestShippedFeaturesWorkspaceWildcardInputs(t *testing.T) {
	processor, reader := setupForTest(t)
	for _, path := range []string{
		"service/reporters/features/service.star",
		"billing_account/reporters/features/billing_account.star",
		"workspace/reporters/rbac/workspace.star",
		"workspace/reporters/features/workspace.star",
	} {
		require.NoError(t, reader.AddRealSchemaFile(path))
	}

	const workspaceFile = "workspace/reporters/features/workspace.star"
	jsonSchema := output.NewJSONSchemaVisitor()
	require.NoError(t, processor.Process(jsonSchema, workspaceFile))
	util.VerifyJSONSchemaResults(t, jsonSchema, map[string]util.JsonSchemaTestCase{
		"workspace/common_representation.json": {
			Valid: []string{`{}`},
		},
		"workspace/reporters/features/workspace.json": {
			Valid: []string{
				`{}`,
				`{"desire_all_services": true, "ignore_inherited_desired_services": true, "ignore_inherited_paid_services": true}`,
				`{"desire_all_services": false, "ignore_inherited_desired_services": false, "ignore_inherited_paid_services": false}`,
			},
			Invalid: []string{
				`{"desire_all_services": "features/service:*"}`,
				`{"ignore_inherited_desired_services": "features/service:*"}`,
				`{"ignore_inherited_paid_services": "features/service:*"}`,
				`{"desire_all_services": 1}`,
				`{"ignore_inherited_desired_services": 1}`,
				`{"ignore_inherited_paid_services": 1}`,
			},
		},
	})

	ksilVisitor := ksil.NewKSILVisitor()
	require.NoError(t, processor.Process(ksilVisitor, workspaceFile))
	entries, err := ksilVisitor.Results()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "features.json", entries[0].Path)

	var namespace map[string]any
	require.NoError(t, json.Unmarshal(entries[0].Contents, &namespace))
	asMap := func(value any) map[string]any {
		result, ok := value.(map[string]any)
		require.True(t, ok)
		return result
	}
	asList := func(value any) []any {
		result, ok := value.([]any)
		require.True(t, ok)
		return result
	}
	asString := func(value any) string {
		result, ok := value.(string)
		require.True(t, ok)
		return result
	}

	definedExtensions := asList(namespace["defined_extensions"])
	require.Len(t, definedExtensions, 1)
	dynamicTypes := asList(asMap(definedExtensions[0])["types"])
	require.Len(t, dynamicTypes, 1)
	dynamicRelations := asList(asMap(dynamicTypes[0])["relations"])
	relationBodies := make(map[string]map[string]any, len(dynamicRelations))
	for _, relation := range dynamicRelations {
		relationMap := asMap(relation)
		relationName := asString(asMap(relationMap["name"])["value"])
		relationBodies[relationName] = asMap(relationMap["body"])
	}

	for _, name := range []string{
		"features_workspace_desire_all_services",
		"features_workspace_ignore_inherited_desired_services",
		"features_workspace_ignore_inherited_paid_services",
	} {
		body := relationBodies[name]
		require.NotNil(t, body, "missing wildcard relation %s", name)
		assert.Equal(t, "self", body["kind"])
		assert.Equal(t, "Any", body["cardinality"])
		types := asList(body["types"])
		require.Len(t, types, 1)
		assert.Equal(t, map[string]any{"namespace": "features", "name": "service", "all": true}, asMap(types[0]))
	}

	paidServices := relationBodies["features_workspace__paid_services"]
	assert.Equal(t, "union", paidServices["kind"])
	paidServicesExclusion := asMap(paidServices["right"])
	assert.Equal(t, "except", paidServicesExclusion["kind"])
	assert.Equal(t, "features_workspace_ignore_inherited_paid_services",
		asString(asMap(asMap(paidServicesExclusion["right"])["relation"])["value"]))

	desiredServices := relationBodies["features_workspace__desired_services"]
	assert.Equal(t, "union", desiredServices["kind"])
	desiredServicesDirect := asMap(desiredServices["left"])
	assert.Equal(t, "union", desiredServicesDirect["kind"])
	assert.Equal(t, "features_workspace_desire_all_services",
		asString(asMap(asMap(desiredServicesDirect["right"])["relation"])["value"]))
	desiredServicesExclusion := asMap(desiredServices["right"])
	assert.Equal(t, "except", desiredServicesExclusion["kind"])
	assert.Equal(t, "features_workspace_ignore_inherited_desired_services",
		asString(asMap(asMap(desiredServicesExclusion["right"])["relation"])["value"]))

	enabledServices := relationBodies["features_workspace_enabled_services"]
	assert.Equal(t, "intersect", enabledServices["kind"])
	assert.Equal(t, "features_workspace__paid_services",
		asString(asMap(asMap(enabledServices["left"])["relation"])["value"]))
	assert.Equal(t, "features_workspace__desired_services",
		asString(asMap(asMap(enabledServices["right"])["relation"])["value"]))
}

func setupForTest(t *testing.T) (*lang.Processor, *lang.InmemorySourceFileReader) {
	t.Helper()
	reader := lang.NewInMemorySourceFileReader("schema", "../../../schema")
	loader := lang.NewLoaderForReader("schema", reader)
	processor := lang.NewProcessor(loader)

	err := reader.AddRealSchemaFile("kessel.star")
	if err != nil {
		t.Log("error in test setup: ", err)
		t.FailNow()
	}

	return processor, reader
}

func verifyOutputs(t *testing.T, processor *lang.Processor, ksilModules map[string]*intermediate.Namespace, jsonSchemaExamples map[string]util.JsonSchemaTestCase) {
	t.Helper()

	jsonSchema := output.NewJSONSchemaVisitor()
	err := processor.Process(jsonSchema)
	if assert.NoError(t, err, "error processing jsonschema visitor") {
		util.VerifyJSONSchemaResults(t, jsonSchema, jsonSchemaExamples)
	}

	ksilVisitor := ksil.NewKSILVisitor()
	err = processor.Process(ksilVisitor)
	if assert.NoError(t, err, "error processing ksil visitor") {
		util.VerifyKSILResults(t, ksilVisitor, ksilModules)
	}
}
