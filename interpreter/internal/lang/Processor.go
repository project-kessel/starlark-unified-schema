package lang

import (
	"fmt"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type Processor struct {
	thread   *starlark.Thread
	loader   *Loader
	metadata map[resourceType]meta
}

func NewProcessor(loader *Loader) *Processor {
	metadata := map[resourceType]meta{}
	p := &Processor{
		loader:   loader,
		thread:   &starlark.Thread{Name: "processor thread", Load: loader.Load},
		metadata: metadata,
	}

	loader.SetMetadata(metadata)

	return p
}

func (p *Processor) Process(visitor output.SchemaVisitor, files ...string) error {
	names := files
	if len(names) == 0 {
		var err error
		names, err = p.loader.GetAllModuleNames()
		if err != nil {
			return err
		}
	}

	for _, name := range names {
		if err := p.processModule(name, visitor); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) processModule(name string, visitor output.SchemaVisitor) error {
	globals, err := p.loader.Load(p.thread, name)
	if err != nil {
		return err
	}

	for varName, value := range globals {
		s, ok := value.(*starlarkstruct.Struct)
		if !ok {
			continue
		}
		kind, err := getStringAttr("kind", s)
		if err != nil || kind != "resource" {
			continue
		}

		reporter, err := getStringAttr("reporter", s)
		if err != nil {
			return fmt.Errorf("resource %s: %w", varName, err)
		}

		visitor.BeginType(varName)

		var commonMembers *output.Members
		commonDict, err := getDictAttr("common", s)
		if err != nil {
			return fmt.Errorf("error getting common members of %s: %w", varName, err)
		}
		commonMembers, err = p.visitMembers(commonDict, visitor)
		if err != nil {
			return fmt.Errorf("error visiting common members of %s: %w", varName, err)
		}

		var reporterMembers *output.Members
		if reporter != "" {
			fieldsDict, err := getDictAttr("fields", s)
			if err != nil {
				return fmt.Errorf("error getting reporter %s fields of %s: %w", reporter, varName, err)
			}
			reporterMembers, err = p.visitMembers(fieldsDict, visitor)
			if err != nil {
				return fmt.Errorf("error visiting reporter %s fields of %s: %w", reporter, varName, err)
			}
		}

		if err := visitor.VisitResource(varName, reporter, commonMembers, reporterMembers); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) visitMembers(fields *starlark.Dict, visitor output.SchemaVisitor) (*output.Members, error) {
	var dataFields []any
	var relationFields []any
	var permissions []any

	for _, item := range fields.Items() {
		key, ok := item[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("field dict key must be a string, got %s", item[0].Type())
		}
		fieldName := string(key)

		fieldStruct, ok := item[1].(*starlarkstruct.Struct)
		if !ok {
			continue
		}

		kind, err := getStringAttr("kind", fieldStruct)
		if err != nil {
			continue
		}

		switch kind {
		case "field":
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

			dataFields = append(dataFields, visitor.VisitDataField(fieldName, required, description, dataType))
		case "relation":
			cardinality, err := getStringAttr("cardinality", fieldStruct)
			if err != nil {
				return nil, fmt.Errorf("error getting cardinality for relation %s: %w", fieldName, err)
			}
			typeStruct, err := getStructAttr("type", fieldStruct)
			if err != nil {
				return nil, fmt.Errorf("error getting type for relation %s: %w", fieldName, err)
			}

			metadata, ok := p.metadata[typeStruct]
			if !ok {
				return nil, fmt.Errorf("metadata not found for type %s", typeStruct)
			}

			idType, err := visitDataType(metadata.idType, visitor)
			if err != nil {
				return nil, fmt.Errorf("error visiting id type for relation %s: %w", fieldName, err)
			}
			relationFields = append(relationFields, visitor.VisitRelation(fieldName, metadata.reporter, metadata.typeName, cardinality, idType))
		case "permission":
			bodyStruct, err := getStructAttr("body", fieldStruct)
			if err != nil {
				return nil, fmt.Errorf("error getting body for permission %s: %w", fieldName, err)
			}
			body, err := visitPermissionBody(bodyStruct, visitor)
			if err != nil {
				return nil, fmt.Errorf("error visiting permission body for permission %s: %w", fieldName, err)
			}
			permissions = append(permissions, visitor.VisitPermission(fieldName, body))
		default:
			return nil, fmt.Errorf("unmatched member kind: %s", kind)
		}
	}

	return &output.Members{
		DataFields:     dataFields,
		RelationFields: relationFields,
		Permissions:    permissions,
	}, nil
}

func visitPermissionBody(v starlark.Value, visitor output.SchemaVisitor) (any, error) {
	typeStruct, ok := v.(*starlarkstruct.Struct)
	if !ok {
		return nil, fmt.Errorf("expected struct for permission body, got %s", v.Type())
	}
	kind, err := getStringAttr("kind", typeStruct)
	if err != nil {
		return nil, err
	}

	switch kind {
	case "and":
		left, right, err := getBinaryPermissionBodyArgs(typeStruct, visitor)
		if err != nil {
			return nil, err
		}
		return visitor.VisitAnd(left, right), nil
	case "or":
		left, right, err := getBinaryPermissionBodyArgs(typeStruct, visitor)
		if err != nil {
			return nil, err
		}
		return visitor.VisitOr(left, right), nil
	case "unless":
		left, right, err := getBinaryPermissionBodyArgs(typeStruct, visitor)
		if err != nil {
			return nil, err
		}
		return visitor.VisitUnless(left, right), nil
	case "ref":
		name, err := getStringAttr("name", typeStruct)
		if err != nil {
			return nil, err
		}
		return visitor.VisitReferenceExpression(name), nil
	case "subref":
		name, err := getStringAttr("name", typeStruct)
		if err != nil {
			return nil, err
		}
		sub, err := getStringAttr("sub", typeStruct)
		if err != nil {
			return nil, err
		}
		return visitor.VisitSubReferenceExpression(name, sub), nil
	default:
		return nil, fmt.Errorf("unmatched permission body kind: %s", kind)
	}
}

func getBinaryPermissionBodyArgs(typeStruct *starlarkstruct.Struct, visitor output.SchemaVisitor) (left any, right any, err error) {
	leftVal, err := typeStruct.Attr("left")
	if err != nil {
		return nil, nil, fmt.Errorf("error accessing 'left' on binary permission body: %w", err)
	}
	rightVal, err := typeStruct.Attr("right")
	if err != nil {
		return nil, nil, fmt.Errorf("error accessing 'right' on binary permission body: %w", err)
	}

	left, err = visitPermissionBody(leftVal, visitor)
	if err != nil {
		return nil, nil, err
	}
	right, err = visitPermissionBody(rightVal, visitor)
	if err != nil {
		return nil, nil, err
	}
	return
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
