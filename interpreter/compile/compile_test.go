package compile

import (
	"os"
	"path/filepath"
	"testing"
)

// mockVisitor implements SchemaVisitor for testing
type mockVisitor struct {
	beginTypeCalls    []string
	visitResourceCall bool
	resultEntries     []OutputEntry
	resultError       error
}

func (m *mockVisitor) BeginType(name string) {
	m.beginTypeCalls = append(m.beginTypeCalls, name)
}

func (m *mockVisitor) VisitResource(typeName string, reporter string, commonMembers *Members, reporterMembers *Members, extendsResource *ResourceTypeReference) error {
	m.visitResourceCall = true
	return nil
}

func (m *mockVisitor) VisitDataField(name string, required bool, description *string, dataType any) any {
	return map[string]any{"name": name, "required": required}
}

func (m *mockVisitor) VisitTextDataType(minLength *int, maxLength *int, regex *string) any {
	return map[string]any{"kind": "text"}
}

func (m *mockVisitor) VisitUUIDDataType() any {
	return map[string]any{"kind": "uuid"}
}

func (m *mockVisitor) VisitNumericIDDataType(min *int, max *int) any {
	return map[string]any{"kind": "numeric_id"}
}

func (m *mockVisitor) VisitBooleanDataType() any {
	return map[string]any{"kind": "boolean"}
}

func (m *mockVisitor) VisitDateTimeDataType() any {
	return map[string]any{"kind": "date_time"}
}

func (m *mockVisitor) VisitEnumDataType(values []string) any {
	return map[string]any{"kind": "enum", "values": values}
}

func (m *mockVisitor) VisitNullableDataType(inner any) any {
	return map[string]any{"kind": "nullable", "inner": inner}
}

func (m *mockVisitor) VisitCompositeDataType(dataTypes []any) any {
	return map[string]any{"kind": "composite", "types": dataTypes}
}

func (m *mockVisitor) VisitArrayDataType(items any) any {
	return map[string]any{"kind": "array", "items": items}
}

func (m *mockVisitor) VisitObjectDataType(properties []any, required []string) any {
	return map[string]any{"kind": "object", "properties": properties}
}

func (m *mockVisitor) VisitAnd(left any, right any) any {
	return map[string]any{"kind": "and", "left": left, "right": right}
}

func (m *mockVisitor) VisitOr(left any, right any) any {
	return map[string]any{"kind": "or", "left": left, "right": right}
}

func (m *mockVisitor) VisitUnless(left any, right any) any {
	return map[string]any{"kind": "unless", "left": left, "right": right}
}

func (m *mockVisitor) VisitReferenceExpression(name string) any {
	return map[string]any{"kind": "reference", "name": name}
}

func (m *mockVisitor) VisitSubReferenceExpression(name string, sub string) any {
	return map[string]any{"kind": "subreference", "name": name, "sub": sub}
}

func (m *mockVisitor) VisitRelation(name string, reporter string, typeName string, cardinality string, idType any) any {
	return map[string]any{"name": name, "typeName": typeName}
}

func (m *mockVisitor) BeginPermission(name string) {}

func (m *mockVisitor) VisitPermission(name string, body any) any {
	return map[string]any{"name": name, "body": body}
}

func (m *mockVisitor) Results() ([]OutputEntry, error) {
	return m.resultEntries, m.resultError
}

// TestCompileRealSchema compiles the actual committed schema to verify the public
// API works end-to-end with real world complexity
func TestCompileRealSchema(t *testing.T) {
	schemaDir := "../../schema"
	if _, err := os.Stat(schemaDir); os.IsNotExist(err) {
		t.Skip("Schema directory not found (expected for isolated tests)")
	}

	files := map[string][]byte{}
	err := filepath.WalkDir(schemaDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".star" {
			return nil
		}
		rel, err := filepath.Rel(schemaDir, path)
		if err != nil {
			return err
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = contents
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to read schema files: %v", err)
	}

	visitor := &mockVisitor{}
	err = Compile(files, visitor)
	if err != nil {
		t.Fatalf("Compile failed on real schema: %v", err)
	}

	// Real schema has multiple resource types
	if len(visitor.beginTypeCalls) < 3 {
		t.Errorf("Expected multiple resource types in real schema, got %d", len(visitor.beginTypeCalls))
	}

	if !visitor.visitResourceCall {
		t.Error("Expected VisitResource to be called")
	}
}

// TestCompileMissingPrelude verifies error handling when kessel.star is missing
func TestCompileMissingPrelude(t *testing.T) {
	files := map[string][]byte{
		"simple.star": []byte(`
load("kessel.star", "resource")
simple_resource = resource(reporter="test")
`),
	}

	visitor := &mockVisitor{}
	err := Compile(files, visitor)
	if err == nil {
		t.Fatal("Expected error when kessel.star is missing")
	}
}

// TestCompileInvalidStarlark verifies error handling for syntax errors
func TestCompileInvalidStarlark(t *testing.T) {
	files := map[string][]byte{
		"kessel.star": []byte(`def resource(): pass`),
		"invalid.star": []byte(`this is not valid starlark syntax ][{`),
	}

	visitor := &mockVisitor{}
	err := Compile(files, visitor)
	if err == nil {
		t.Fatal("Expected error for invalid Starlark syntax")
	}
}

// TestCompileIgnoresNonStarFiles verifies that only .star files are processed
func TestCompileIgnoresNonStarFiles(t *testing.T) {
	schemaDir := "../../schema"
	kesselStar, err := os.ReadFile(filepath.Join(schemaDir, "kessel.star"))
	if err != nil {
		t.Skip("kessel.star not available")
	}

	files := map[string][]byte{
		"kessel.star": kesselStar,
		"test.star": []byte(`
load("kessel.star", "resource", "uuid")
test_resource = resource(reporter="test", id_type=uuid())
`),
		"README.md":   []byte(`# This should be ignored`),
		"config.json": []byte(`{"ignored": true}`),
		"helper.py":   []byte(`print("not starlark")`),
	}

	visitor := &mockVisitor{}
	err = Compile(files, visitor)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Only test.star should produce a resource
	if len(visitor.beginTypeCalls) != 1 {
		t.Errorf("Expected 1 resource, got %d (non-.star files should be ignored)", len(visitor.beginTypeCalls))
	}

	if visitor.beginTypeCalls[0] != "test_resource" {
		t.Errorf("Expected resource 'test_resource', got %s", visitor.beginTypeCalls[0])
	}
}

// TestCompileDeterministicOrder verifies files are processed in sorted order
func TestCompileDeterministicOrder(t *testing.T) {
	schemaDir := "../../schema"
	kesselStar, err := os.ReadFile(filepath.Join(schemaDir, "kessel.star"))
	if err != nil {
		t.Skip("kessel.star not available")
	}

	files := map[string][]byte{
		"kessel.star": kesselStar,
		"z_last.star": []byte(`
load("kessel.star", "resource", "uuid")
z_resource = resource(reporter="z", id_type=uuid())
`),
		"a_first.star": []byte(`
load("kessel.star", "resource", "uuid")
a_resource = resource(reporter="a", id_type=uuid())
`),
		"m_middle.star": []byte(`
load("kessel.star", "resource", "uuid")
m_resource = resource(reporter="m", id_type=uuid())
`),
	}

	visitor := &mockVisitor{}
	err = Compile(files, visitor)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Files should be processed in sorted order
	if len(visitor.beginTypeCalls) != 3 {
		t.Fatalf("Expected 3 resources, got %d", len(visitor.beginTypeCalls))
	}

	expected := []string{"a_resource", "m_resource", "z_resource"}
	for i, name := range expected {
		if visitor.beginTypeCalls[i] != name {
			t.Errorf("Position %d: expected %s, got %s", i, name, visitor.beginTypeCalls[i])
		}
	}
}
