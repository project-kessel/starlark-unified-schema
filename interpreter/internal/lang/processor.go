package lang

import (
	"fmt"

	"github.com/project-kessel/starlark-unified-schema/internal/model"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type mergedResource struct {
	name      string
	common    *starlark.Dict
	reporters map[string]*starlark.Dict
}

type Processor struct {
	loader    *Loader
	thread    *starlark.Thread
	resources map[string]*mergedResource
	order     []string
}

func NewProcessor(loader *Loader) *Processor {
	return &Processor{
		loader:    loader,
		thread:    &starlark.Thread{Name: "processor thread", Load: loader.Load},
		resources: map[string]*mergedResource{},
	}
}

func (p *Processor) ProcessAll() ([]model.Resource, error) {
	names, err := p.loader.GetAllModuleNames()
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		if err := p.processModule(name); err != nil {
			return nil, err
		}
	}

	return p.buildResources()
}

func (p *Processor) processModule(name string) error {
	alreadyLoaded := p.loader.IsLoaded(name)

	globals, err := p.loader.Load(p.thread, name)
	if err != nil {
		return err
	}

	if alreadyLoaded {
		return nil
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

		merged, exists := p.resources[varName]
		if !exists {
			merged = &mergedResource{
				name:      varName,
				reporters: map[string]*starlark.Dict{},
			}
			p.resources[varName] = merged
			p.order = append(p.order, varName)
		}

		commonVal, err := s.Attr("common")
		if err == nil {
			if commonDict, ok := commonVal.(*starlark.Dict); ok {
				if merged.common == nil || merged.common.Len() == 0 {
					merged.common = commonDict
				}
			}
		}

		if reporter == "" {
			continue
		}

		fieldsVal, err := s.Attr("fields")
		if err == nil {
			if fieldsDict, ok := fieldsVal.(*starlark.Dict); ok {
				if _, exists := merged.reporters[reporter]; exists {
					return fmt.Errorf("resource %s: reporter '%s' registered more than once", varName, reporter)
				}
				merged.reporters[reporter] = fieldsDict
			}
		}
	}

	return nil
}

func (p *Processor) buildResources() ([]model.Resource, error) {
	resources := make([]model.Resource, 0, len(p.order))

	for _, name := range p.order {
		merged := p.resources[name]

		var commonFields []model.Field
		if merged.common != nil {
			fields, err := extractFields(merged.common)
			if err != nil {
				return nil, fmt.Errorf("error extracting common fields of %s: %w", name, err)
			}
			commonFields = fields
		}

		reporters := map[string][]model.Field{}
		for reporterName, dict := range merged.reporters {
			fields, err := extractFields(dict)
			if err != nil {
				return nil, fmt.Errorf("error extracting reporter %s fields of %s: %w", reporterName, name, err)
			}
			reporters[reporterName] = fields
		}

		resources = append(resources, model.Resource{
			Name:      name,
			Common:    commonFields,
			Reporters: reporters,
		})
	}

	return resources, nil
}

func extractFields(dict *starlark.Dict) ([]model.Field, error) {
	var fields []model.Field
	for _, item := range dict.Items() {
		key, ok := item[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("field dict key must be a string, got %s", item[0].Type())
		}
		fieldName := string(key)

		fieldStruct, ok := item[1].(*starlarkstruct.Struct)
		if !ok {
			return nil, fmt.Errorf("field %s: expected struct, got %s", fieldName, item[1].Type())
		}

		kind, err := getStringAttr("kind", fieldStruct)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", fieldName, err)
		}

		if kind != "field" {
			return nil, fmt.Errorf("field %s: expected kind %q, got %q", fieldName, "field", kind)
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

		dataType, err := extractDataType(typeVal)
		if err != nil {
			return nil, fmt.Errorf("error extracting data type for field %s: %w", fieldName, err)
		}

		fields = append(fields, model.Field{
			Name:        fieldName,
			Required:    required,
			Description: description,
			Type:        dataType,
		})
	}

	return fields, nil
}

func extractDataType(v starlark.Value) (model.DataType, error) {
	typeStruct, ok := v.(*starlarkstruct.Struct)
	if !ok {
		return model.DataType{}, fmt.Errorf("expected struct for data type, got %s", v.Type())
	}

	kind, err := getStringAttr("kind", typeStruct)
	if err != nil {
		return model.DataType{}, err
	}

	switch kind {
	case "text":
		return model.DataType{
			Kind:      "text",
			MinLength: getOptionalIntAttr("minLength", typeStruct),
			MaxLength: getOptionalIntAttr("maxLength", typeStruct),
			Regex:     getOptionalStringAttr("regex", typeStruct),
		}, nil

	case "uuid":
		return model.DataType{Kind: "uuid"}, nil

	case "numeric_id":
		return model.DataType{
			Kind: "numeric_id",
			Min:  getOptionalIntAttr("min", typeStruct),
			Max:  getOptionalIntAttr("max", typeStruct),
		}, nil

	case "boolean":
		return model.DataType{Kind: "boolean"}, nil

	case "date_time":
		return model.DataType{Kind: "date_time"}, nil

	case "enum":
		valuesVal, err := typeStruct.Attr("values")
		if err != nil {
			return model.DataType{}, fmt.Errorf("error accessing 'values' on enum: %w", err)
		}
		values, err := extractStringList(valuesVal, "enum values")
		if err != nil {
			return model.DataType{}, err
		}
		return model.DataType{Kind: "enum", Values: values}, nil

	case "nullable":
		innerVal, err := typeStruct.Attr("inner")
		if err != nil {
			return model.DataType{}, fmt.Errorf("error accessing 'inner' on nullable: %w", err)
		}
		inner, err := extractDataType(innerVal)
		if err != nil {
			return model.DataType{}, err
		}
		return model.DataType{Kind: "nullable", Inner: &inner}, nil

	case "union":
		leftVal, err := typeStruct.Attr("left")
		if err != nil {
			return model.DataType{}, fmt.Errorf("error accessing 'left' on union: %w", err)
		}
		left, err := extractDataType(leftVal)
		if err != nil {
			return model.DataType{}, err
		}

		rightVal, err := typeStruct.Attr("right")
		if err != nil {
			return model.DataType{}, fmt.Errorf("error accessing 'right' on union: %w", err)
		}
		right, err := extractDataType(rightVal)
		if err != nil {
			return model.DataType{}, err
		}
		return model.DataType{Kind: "union", Members: []model.DataType{left, right}}, nil

	case "array":
		itemsVal, err := typeStruct.Attr("items")
		if err != nil {
			return model.DataType{}, fmt.Errorf("error accessing 'items' on array: %w", err)
		}
		items, err := extractDataType(itemsVal)
		if err != nil {
			return model.DataType{}, err
		}
		return model.DataType{Kind: "array", Items: &items}, nil

	case "object":
		propsVal, err := typeStruct.Attr("properties")
		if err != nil {
			return model.DataType{}, fmt.Errorf("error accessing 'properties' on object: %w", err)
		}
		propsDict, ok := propsVal.(*starlark.Dict)
		if !ok {
			return model.DataType{}, fmt.Errorf("object 'properties' must be a dict, got %s", propsVal.Type())
		}

		requiredVal, err := typeStruct.Attr("required")
		if err != nil {
			return model.DataType{}, fmt.Errorf("error accessing 'required' on object: %w", err)
		}
		required, err := extractStringList(requiredVal, "object required")
		if err != nil {
			return model.DataType{}, err
		}

		var properties []model.Field
		for _, item := range propsDict.Items() {
			propKey, ok := item[0].(starlark.String)
			if !ok {
				return model.DataType{}, fmt.Errorf("object property key must be a string, got %s", item[0].Type())
			}
			propName := string(propKey)
			propType, err := extractDataType(item[1])
			if err != nil {
				return model.DataType{}, fmt.Errorf("error extracting property %s: %w", propName, err)
			}
			properties = append(properties, model.Field{Name: propName, Type: propType})
		}

		return model.DataType{Kind: "object", Properties: properties, Required: required}, nil

	default:
		return model.DataType{}, fmt.Errorf("unmatched data type kind: %s", kind)
	}
}
