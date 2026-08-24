package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

func TestJSONSchemaVisitorRequiredFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("required_fields")
	required := v.VisitDataField("required", true, nil, v.VisitTextDataType(nil, nil, nil))
	optional := v.VisitDataField("optional", false, nil, v.VisitTextDataType(nil, nil, nil))
	err := v.VisitResource("required_fields", "test", nil, &Members{
		DataFields: []any{required, optional},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"required_fields/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"required_fields/reporters/test/required_fields.json": {
			valid: []string{
				`{"required": "stuff"}`,
				`{"required": "stuff", "optional": "stuff"}`,
			},
			invalid: []string{
				`{}`,
				`{"optional": "stuff"}`,
			},
		},
	})
}

func TestJSONSchemaVisitorCommonFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("with_common")
	common := v.VisitDataField("common", false, nil, v.VisitTextDataType(nil, nil, nil))
	err := v.VisitResource("with_common", "test", &Members{
		DataFields: []any{common},
	}, &Members{}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"with_common/common_representation.json": {
			valid: []string{
				`{}`,
				`{"common": "data"}`,
			},
			invalid: []string{
				`{"other": "foo"}`,
			},
		},
		"with_common/reporters/test/with_common.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
	})
}

func TestJSONSchemaVisitorUUIDFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("with_uuid")
	uuid := v.VisitDataField("uuid", false, nil, v.VisitUUIDDataType())
	err := v.VisitResource("with_uuid", "test", nil, &Members{
		DataFields: []any{uuid},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"with_uuid/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"with_uuid/reporters/test/with_uuid.json": {
			valid:   []string{`{"uuid": "9bcd4eec-9d9d-11f1-8f44-6ae246604903"}`},
			invalid: []string{`{"uuid": "9bcd4eec_9d9d_11f1_8f44_6ae246604903"}`},
		},
	})
}

func TestJSONSchemaVisitorBooleanFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("with_flag")
	flag := v.VisitDataField("flag", false, nil, v.VisitBooleanDataType())
	err := v.VisitResource("with_flag", "test", nil, &Members{
		DataFields: []any{flag},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"with_flag/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"with_flag/reporters/test/with_flag.json": {
			valid: []string{
				`{"flag": true}`,
				`{"flag": false}`,
			},
			invalid: []string{
				`{"flag": "True"}`,
				`{"flag": "true"}`,
				`{"flag": "False"}`,
				`{"flag": "false"}`,
				`{"flag": 0}`,
				`{"flag": 1}`,
			},
		},
	})
}

func TestJSONSchemaVisitorDateFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("with_timestamp")
	timestamp := v.VisitDataField("timestamp", false, nil, v.VisitDateTimeDataType())
	err := v.VisitResource("with_timestamp", "test", nil, &Members{
		DataFields: []any{timestamp},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"with_timestamp/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"with_timestamp/reporters/test/with_timestamp.json": {
			valid: []string{
				`{"timestamp": "2018-11-13T20:20:39+00:00"}`,
				`{"timestamp": "2018-11-13"}`,
			},
			invalid: []string{
				`{"timestamp": "2018/11/13"}`,
			},
		},
	})
}

func TestJSONSchemaVisitorEnumFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("with_enum")
	enum := v.VisitDataField("enum", false, nil, v.VisitEnumDataType([]string{"apple", "orange"}))
	err := v.VisitResource("with_enum", "test", nil, &Members{
		DataFields: []any{enum},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"with_enum/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"with_enum/reporters/test/with_enum.json": {
			valid: []string{
				`{"enum": "apple"}`,
				`{"enum": "orange"}`,
			},
			invalid: []string{
				`{"enum": "ball"}`,
			},
		},
	})
}

func TestJSONSchemaVisitorNullableFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("with_nullable")
	non_nullable := v.VisitDataField("non_nullable", false, nil, v.VisitTextDataType(nil, nil, nil))
	nullable := v.VisitDataField("nullable", false, nil, v.VisitNullableDataType(v.VisitTextDataType(nil, nil, nil)))
	err := v.VisitResource("with_nullable", "test", nil, &Members{
		DataFields: []any{nullable, non_nullable},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"with_nullable/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"with_nullable/reporters/test/with_nullable.json": {
			valid: []string{
				`{}`,
				`{"nullable": null}`,
			},
			invalid: []string{
				`{"non_nullable": null}`,
			},
		},
	})
}

func TestJSONSchemaVisitorCompositeTypeFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("with_composite")
	regex := `^\d{5}$`
	//Either 5 digits or a uuid
	composite := v.VisitDataField("composite", false, nil, v.VisitCompositeDataType([]any{v.VisitTextDataType(nil, nil, &regex), v.VisitUUIDDataType()}))

	err := v.VisitResource("with_composite", "test", nil, &Members{
		DataFields: []any{composite},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"with_composite/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"with_composite/reporters/test/with_composite.json": {
			valid: []string{
				`{}`,
				`{"composite": "12345"}`,
				`{"composite": "9bcd4eec-9d9d-11f1-8f44-6ae246604903"}`,
			},
			invalid: []string{
				`{"composite": "9bcd4eec"}`,
			},
		},
	})
}

func TestJSONSchemaVisitorArrayFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("with_array")
	array := v.VisitDataField("array", false, nil, v.VisitArrayDataType(v.VisitTextDataType(nil, nil, nil)))

	err := v.VisitResource("with_array", "test", nil, &Members{
		DataFields: []any{array},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"with_array/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"with_array/reporters/test/with_array.json": {
			valid: []string{
				`{}`,
				`{"array": []}`,
				`{"array": ["test"]}`,
				`{"array": ["test", "test2"]}`,
			},
			invalid: []string{
				`{"array": "test"}`,
				`{"array": [1]}`,
				`{"array": [1, 2]}`,
			},
		},
	})
}

func TestJSONSchemaVisitorObjectFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("with_obj")
	obj := v.VisitDataField("obj", false, nil, v.VisitObjectDataType(
		[]any{
			v.VisitDataField("required", true, nil, v.VisitTextDataType(nil, nil, nil)),
			v.VisitDataField("optional", false, nil, v.VisitTextDataType(nil, nil, nil)),
		}, []string{"required"}))

	err := v.VisitResource("with_obj", "test", nil, &Members{
		DataFields: []any{obj},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"with_obj/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"with_obj/reporters/test/with_obj.json": {
			valid: []string{
				`{}`,
				`{"obj": {"required": "test"}}`,
				`{"obj": {"required": "test", "optional": "test2"}}`,
			},
			invalid: []string{
				`{"obj": {}}`,
			},
		},
	})
}

func TestJSONSchemaVisitorRelationFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	v.BeginType("with_relation")
	at_most_one := v.VisitRelation("at_most_one", "test", "other", "AtMostOne", v.VisitUUIDDataType())
	one := v.VisitRelation("one", "test", "other", "ExactlyOne", v.VisitUUIDDataType())
	at_least_one := v.VisitRelation("at_least_one", "test", "other", "AtLeastOne", v.VisitUUIDDataType())
	many := v.VisitRelation("many", "test", "other", "Many", v.VisitUUIDDataType())
	wildcard := v.VisitRelation("wildcard", "test", "other", "All", v.VisitUUIDDataType())
	err := v.VisitResource("with_relation", "test", nil, &Members{
		RelationFields: []any{at_most_one, one, at_least_one, many, wildcard},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"with_relation/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"with_relation/reporters/test/with_relation.json": {
			valid: []string{
				`{
					"one": "9bcd4eec-9d9d-11f1-8f44-6ae246604903",
					"at_least_one": ["9bcd4eec-9d9d-11f1-8f44-6ae246604903"]
				}`,
				`{
					"one": "9bcd4eec-9d9d-11f1-8f44-6ae246604903",
					"at_least_one": ["9bcd4eec-9d9d-11f1-8f44-6ae246604903", "1b10a5d6-9da9-11f1-b8d3-6ae246604903"]
				}`,
				`{
					"one": "9bcd4eec-9d9d-11f1-8f44-6ae246604903",
					"at_least_one": ["9bcd4eec-9d9d-11f1-8f44-6ae246604903"],
					"at_most_one": "9bcd4eec-9d9d-11f1-8f44-6ae246604903"
				}`,
				`{
					"one": "9bcd4eec-9d9d-11f1-8f44-6ae246604903",
					"at_least_one": ["9bcd4eec-9d9d-11f1-8f44-6ae246604903"],
					"many": ["9bcd4eec-9d9d-11f1-8f44-6ae246604903"]
				}`,
				`{
					"one": "9bcd4eec-9d9d-11f1-8f44-6ae246604903",
					"at_least_one": ["9bcd4eec-9d9d-11f1-8f44-6ae246604903"],
					"many": ["9bcd4eec-9d9d-11f1-8f44-6ae246604903", "1b10a5d6-9da9-11f1-b8d3-6ae246604903"]
				}`,
				`{
					"one": "9bcd4eec-9d9d-11f1-8f44-6ae246604903",
					"at_least_one": ["9bcd4eec-9d9d-11f1-8f44-6ae246604903"],
					"wildcard": "test/other:*"
				}`,
			},
			invalid: []string{
				`{}`,
				`{"one": "9bcd4eec-9d9d-11f1-8f44-6ae246604903"}`,
				`{"at_least_one": ["9bcd4eec-9d9d-11f1-8f44-6ae246604903"]}`,
				`{
					"one": "9bcd4eec-9d9d-11f1-8f44-6ae246604903",
					"at_least_one": ["9bcd4eec-9d9d-11f1-8f44-6ae246604903"],
					"wildcard": "test/some:*"
				}`,
				`{
					"one": "9bcd4eec-9d9d-11f1-8f44-6ae246604903",
					"at_least_one": []
				}`,
			},
		},
	})
}

func TestJSONSchemaVisitorNumericFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	limit := 5

	v.BeginType("number_holder")
	small := v.VisitDataField("small", false, nil, v.VisitNumericIDDataType(nil, &limit))
	big := v.VisitDataField("big", false, nil, v.VisitNumericIDDataType(&limit, nil))
	err := v.VisitResource("number_holder", "test", nil, &Members{
		DataFields: []any{small, big},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"number_holder/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"number_holder/reporters/test/number_holder.json": {
			valid: []string{
				`{}`,
				`{"big": 6}`,
				`{"small": 4}`,
				`{"small": 4, "big": 6}`,
			},
			invalid: []string{
				`{"other": "foo"}`,
				`{"big": 4}`,
				`{"big": "6"}`,
				`{"small": "4"}`,
				`{"small": 6}`,
			},
		},
	})
}

func TestJSONSchemaVisitorTextFields(t *testing.T) {
	v := NewJSONSchemaVisitor()

	length := 5
	regex := `^\d{3}$`

	v.BeginType("text_holder")
	short := v.VisitDataField("short", false, nil, v.VisitTextDataType(nil, &length, nil))
	long := v.VisitDataField("long", false, nil, v.VisitTextDataType(&length, nil, nil))
	three_digit := v.VisitDataField("three_digit", false, nil, v.VisitTextDataType(nil, nil, &regex))

	err := v.VisitResource("text_holder", "test", nil, &Members{
		DataFields: []any{short, long, three_digit},
	}, nil)
	if !assert.NoError(t, err, "error visiting resource") {
		return
	}

	verifyJSONSchemaResults(t, v, map[string]jsonSchemaTestCase{
		"text_holder/common_representation.json": {
			valid:   []string{`{}`},
			invalid: []string{`{"other": "foo"}`},
		},
		"text_holder/reporters/test/text_holder.json": {
			valid: []string{
				`{}`,
				`{"short": "abc"}`,
				`{"long": "abcde"}`,
				`{"three_digit": "123"}`,
				`{"short": "abc", "long": "abcde", "three_digit": "123"}`,
			},
			invalid: []string{
				`{"short": "abcdef"}`,
				`{"long": "abc"}`,
				`{"three_digit": ""}`,
				`{"three_digit": "12"}`,
				`{"three_digit": "1234"}`,
				`{"three_digit": "I23"}`,
			},
		},
	})
}

func verifyJSONSchemaResults(t *testing.T, v *JSONSchemaVisitor, examples map[string]jsonSchemaTestCase) {
	t.Helper()

	results, err := v.Results()
	if !assert.NoError(t, err, "error getting results") {
		return
	}

	resultPaths := make(map[string]bool, len(results))
	for _, entry := range results {
		resultPaths[entry.Path] = true
	}
	for path := range examples {
		assert.True(t, resultPaths[path], "no result entry found for example path: %s", path)
	}

	for _, entry := range results {
		example, ok := examples[entry.Path]
		if !assert.True(t, ok, "no test case found for path: %s", entry.Path) {
			return
		}

		schema := gojsonschema.NewBytesLoader(entry.Contents)

		if example.valid != nil {
			for _, valid := range example.valid {
				data := gojsonschema.NewStringLoader(valid)
				result, err := gojsonschema.Validate(schema, data)

				if !assert.NoError(t, err, "error validating valid json") {
					return
				}

				assert.True(t, result.Valid(), "json was unexpectedly invalid. path=%s, contents=%s\n\nErrors: %+v", entry.Path, valid, result.Errors())
			}
		}

		if example.invalid != nil {
			for _, invalid := range example.invalid {
				data := gojsonschema.NewStringLoader(invalid)
				result, err := gojsonschema.Validate(schema, data)

				if !assert.NoError(t, err, "error validating invalid json") {
					return
				}

				assert.False(t, result.Valid(), "json was unexpectedly valid. path=%s, contents=%s", entry.Path, invalid)
			}
		}
	}
}

type jsonSchemaTestCase struct {
	valid   []string
	invalid []string
}
