package lang

import (
	"fmt"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func VisitResources(resources []ResourceDefinition, visitor output.SchemaVisitor) error {
	for _, def := range resources {
		visitor.BeginType(def.Name)

		var commonFields []any
		if def.Common != nil {
			fields, err := visitFieldDict(def.Common, visitor)
			if err != nil {
				return fmt.Errorf("error visiting common fields of %s: %w", def.Name, err)
			}
			commonFields = fields
		}

		reporterGroups := map[string][]any{}
		for reporterName, reporterDict := range def.Reporters {
			fields, err := visitFieldDict(reporterDict, visitor)
			if err != nil {
				return fmt.Errorf("error visiting reporter %s fields of %s: %w", reporterName, def.Name, err)
			}
			reporterGroups[reporterName] = fields
		}

		visitor.VisitType(def.Name, commonFields, reporterGroups)
	}

	return nil
}

func visitFieldDict(fields *starlark.Dict, visitor output.SchemaVisitor) ([]any, error) {
	var result []any
	for _, item := range fields.Items() {
		key, ok := item[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("field dict key must be a string, got %s", item[0].Type())
		}
		fieldName := string(key)

		fieldStruct, ok := item[1].(*starlarkstruct.Struct)
		if !ok {
			return nil, fmt.Errorf("field %s: value must be a struct (from field()), got %s", fieldName, item[1].Type())
		}

		required, err := getBoolAttr("required", fieldStruct)
		if err != nil {
			return nil, fmt.Errorf("error getting required for field %s: %w", fieldName, err)
		}

		description := getOptionalStringAttr("description", fieldStruct)

		typeVal, err := fieldStruct.Attr("type")
		if err != nil {
			return nil, fmt.Errorf("error getting type for field %s: %w", fieldName, err)
		}

		dataType, err := visitDataType(typeVal, visitor)
		if err != nil {
			return nil, fmt.Errorf("error visiting data type for field %s: %w", fieldName, err)
		}

		result = append(result, visitor.VisitDataField(fieldName, required, description, dataType))
	}

	return result, nil
}

func visitDataType(v starlark.Value, visitor output.SchemaVisitor) (any, error) {
	typeStruct, ok := v.(*starlarkstruct.Struct)
	if !ok {
		return nil, fmt.Errorf("expected struct for data type, got %s", v.Type())
	}

	kind, err := getStringAttr("kind", typeStruct)
	if err != nil {
		return nil, err
	}

	switch kind {
	case "text":
		minLength := getOptionalIntAttr("minLength", typeStruct)
		maxLength := getOptionalIntAttr("maxLength", typeStruct)
		regex := getOptionalStringAttr("regex", typeStruct)
		return visitor.VisitTextDataType(minLength, maxLength, regex), nil

	case "uuid":
		return visitor.VisitUUIDDataType(), nil

	case "numeric_id":
		min := getOptionalIntAttr("min", typeStruct)
		max := getOptionalIntAttr("max", typeStruct)
		return visitor.VisitNumericIDDataType(min, max), nil

	case "boolean":
		return visitor.VisitBooleanDataType(), nil

	case "date_time":
		return visitor.VisitDateTimeDataType(), nil

	case "enum":
		valuesVal, err := typeStruct.Attr("values")
		if err != nil {
			return nil, fmt.Errorf("error accessing 'values' on enum: %w", err)
		}
		values, err := extractStringList(valuesVal, "enum values")
		if err != nil {
			return nil, err
		}
		return visitor.VisitEnumDataType(values), nil

	case "nullable":
		innerVal, err := typeStruct.Attr("inner")
		if err != nil {
			return nil, fmt.Errorf("error accessing 'inner' on nullable: %w", err)
		}
		inner, err := visitDataType(innerVal, visitor)
		if err != nil {
			return nil, err
		}
		return visitor.VisitNullableDataType(inner), nil

	case "union":
		leftVal, err := typeStruct.Attr("left")
		if err != nil {
			return nil, fmt.Errorf("error accessing 'left' on union: %w", err)
		}
		left, err := visitDataType(leftVal, visitor)
		if err != nil {
			return nil, err
		}

		rightVal, err := typeStruct.Attr("right")
		if err != nil {
			return nil, fmt.Errorf("error accessing 'right' on union: %w", err)
		}
		right, err := visitDataType(rightVal, visitor)
		if err != nil {
			return nil, err
		}
		return visitor.VisitCompositeDataType([]any{left, right}), nil

	case "array":
		itemsVal, err := typeStruct.Attr("items")
		if err != nil {
			return nil, fmt.Errorf("error accessing 'items' on array: %w", err)
		}
		items, err := visitDataType(itemsVal, visitor)
		if err != nil {
			return nil, err
		}
		return visitor.VisitArrayDataType(items), nil

	case "object":
		propsVal, err := typeStruct.Attr("properties")
		if err != nil {
			return nil, fmt.Errorf("error accessing 'properties' on object: %w", err)
		}
		propsDict, ok := propsVal.(*starlark.Dict)
		if !ok {
			return nil, fmt.Errorf("object 'properties' must be a dict, got %s", propsVal.Type())
		}

		requiredVal, err := typeStruct.Attr("required")
		if err != nil {
			return nil, fmt.Errorf("error accessing 'required' on object: %w", err)
		}
		required, err := extractStringList(requiredVal, "object required")
		if err != nil {
			return nil, err
		}

		var properties []any
		for _, item := range propsDict.Items() {
			propKey, ok := item[0].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("object property key must be a string, got %s", item[0].Type())
			}
			propName := string(propKey)
			propType, err := visitDataType(item[1], visitor)
			if err != nil {
				return nil, fmt.Errorf("error visiting property %s: %w", propName, err)
			}
			properties = append(properties, visitor.VisitDataField(propName, false, nil, propType))
		}

		return visitor.VisitObjectDataType(properties, required), nil

	default:
		return nil, fmt.Errorf("unmatched data type kind: %s", kind)
	}
}
