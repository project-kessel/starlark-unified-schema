package lang

import (
	"fmt"
	"sort"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type resourceType *starlark.Dict

type Processor struct {
	thread   *starlark.Thread
	loader   *Loader
	metadata map[resourceType]meta
}

func NewProcessor(loader *Loader) *Processor {
	m := map[resourceType]meta{}

	p := &Processor{
		loader:   loader,
		thread:   &starlark.Thread{Name: "processor thread", Load: loader.Load},
		metadata: m,
	}

	loader.SetMetadata(m)

	return p
}

func (p *Processor) ProcessModule(name string, visitor output.Visitor) error {
	globals, err := p.loader.Load(p.thread, name)
	if err != nil {
		return err
	}

	// Collect type dicts and sort by name for deterministic output
	type typeEntry struct {
		name     string
		typeDict *starlark.Dict
	}

	var types []typeEntry
	for globalName, globalValue := range globals {
		if typeDict, ok := globalValue.(*starlark.Dict); ok {
			metadata, exists := p.metadata[typeDict]
			if !exists {
				return fmt.Errorf("no metadata found for type %s", globalName)
			}
			types = append(types, typeEntry{name: metadata.typeName, typeDict: typeDict})
		}
	}

	// Sort types alphabetically by name for deterministic output
	sort.Slice(types, func(i, j int) bool {
		return types[i].name < types[j].name
	})

	// Process types in sorted order
	for _, entry := range types {
		metadata := p.metadata[entry.typeDict]
		namespace := metadata.moduleName
		typeName := metadata.typeName

		// Begin type processing
		visitor.BeginType(namespace, typeName)

		// Process relations
		relations, err := p.processTypeRelations(entry.typeDict, namespace, typeName, visitor)
		if err != nil {
			return fmt.Errorf("failed to process relations for type %s: %w", typeName, err)
		}

		// Visit the type
		visitor.VisitType(namespace, typeName, relations)
	}

	return nil
}

func (p *Processor) ProcessAllModules(visitor output.Visitor) error {
	names, err := p.loader.GetAllModuleNames()
	if err != nil {
		return err
	}

	for _, name := range names {
		err := p.ProcessModule(name, visitor)
		if err != nil {
			return err
		}
	}

	return nil
}

type meta struct {
	moduleName string
	typeName   string
}

func mapCardinality(kind string) string {
	switch kind {
	case "atMostOne":
		return "AtMostOne"
	case "exactlyOne":
		return "ExactlyOne"
	case "atLeastOne":
		return "AtLeastOne"
	case "many":
		return "Many"
	case "boolean":
		return "Boolean"
	default:
		return kind
	}
}

func getAttr(value starlark.Value, attrName string) (starlark.Value, bool, error) {
	if dict, ok := value.(*starlark.Dict); ok {
		val, found, err := dict.Get(starlark.String(attrName))
		return val, found, err
	} else if strct, ok := value.(*starlarkstruct.Struct); ok {
		val, err := strct.Attr(attrName)
		if err != nil {
			return starlark.None, false, nil
		}
		return val, true, nil
	}
	return nil, false, fmt.Errorf("value is not a dict or struct, got: %T", value)
}

func sortStarlarkDictItems(items []starlark.Tuple) {
	sort.Slice(items, func(i, j int) bool {
		nameI, _ := convert_to_string(items[i][0])
		nameJ, _ := convert_to_string(items[j][0])
		return nameI < nameJ
	})
}

func (p *Processor) processCardinalityRelation(relationValue starlark.Value, kind string, parentNamespace, parentTypeName string, visitor output.Visitor) (any, error) {
	typeValue, found, err := getAttr(relationValue, "type")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("cardinality relation missing 'type' field")
	}

	// Check if it's a selfType
	selfKindValue, found, _ := getAttr(typeValue, "kind")
	if found {
		selfKind, _ := convert_to_string(selfKindValue)
		if selfKind == "selfType" {
			cardinality := mapCardinality(kind)
			return visitor.VisitAssignableExpression(parentNamespace, parentTypeName, cardinality), nil
		}
	}

	// For now, assume it's a reference to another type via metadata
	// This will need enhancement for more complex type references
	if typeDict, ok := typeValue.(*starlark.Dict); ok {
		metadata, exists := p.metadata[typeDict]
		if !exists {
			return nil, fmt.Errorf("no metadata found for referenced type dict")
		}

		namespace := metadata.moduleName
		typeName := metadata.typeName
		cardinality := mapCardinality(kind)

		return visitor.VisitAssignableExpression(namespace, typeName, cardinality), nil
	}

	return nil, fmt.Errorf("type field has unsupported structure: %T", typeValue)
}

func (p *Processor) processRefExpression(refValue starlark.Value, visitor output.Visitor) (any, error) {
	nameValue, found, err := getAttr(refValue, "name")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("ref expression missing 'name' field")
	}

	name, err := convert_to_string(nameValue)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ref name to string: %w", err)
	}

	return visitor.VisitRelationExpression(name), nil
}

func (p *Processor) processSubrefExpression(subrefValue starlark.Value, visitor output.Visitor) (any, error) {
	nameValue, found, err := getAttr(subrefValue, "name")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("subref expression missing 'name' field")
	}

	name, err := convert_to_string(nameValue)
	if err != nil {
		return nil, fmt.Errorf("failed to convert subref name to string: %w", err)
	}

	subValue, found, err := getAttr(subrefValue, "sub")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("subref expression missing 'sub' field")
	}

	sub, err := convert_to_string(subValue)
	if err != nil {
		return nil, fmt.Errorf("failed to convert subref sub to string: %w", err)
	}

	return visitor.VisitSubRelationExpression(name, sub), nil
}

func (p *Processor) processBinaryExpression(value starlark.Value, operatorName string, parentNamespace, parentTypeName string, visitor output.Visitor, visitorFunc func(any, any) any) (any, error) {
	leftValue, found, err := getAttr(value, "left")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("%s expression missing 'left' field", operatorName)
	}

	left, err := p.processRelationValue(leftValue, parentNamespace, parentTypeName, visitor)
	if err != nil {
		return nil, fmt.Errorf("failed to process %s left operand: %w", operatorName, err)
	}

	rightValue, found, err := getAttr(value, "right")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("%s expression missing 'right' field", operatorName)
	}

	right, err := p.processRelationValue(rightValue, parentNamespace, parentTypeName, visitor)
	if err != nil {
		return nil, fmt.Errorf("failed to process %s right operand: %w", operatorName, err)
	}

	return visitorFunc(left, right), nil
}

func (p *Processor) processOrExpression(orValue starlark.Value, parentNamespace, parentTypeName string, visitor output.Visitor) (any, error) {
	return p.processBinaryExpression(orValue, "or", parentNamespace, parentTypeName, visitor, visitor.VisitOr)
}

func (p *Processor) processAndExpression(andValue starlark.Value, parentNamespace, parentTypeName string, visitor output.Visitor) (any, error) {
	return p.processBinaryExpression(andValue, "and", parentNamespace, parentTypeName, visitor, visitor.VisitAnd)
}

func (p *Processor) processUnlessExpression(unlessValue starlark.Value, parentNamespace, parentTypeName string, visitor output.Visitor) (any, error) {
	return p.processBinaryExpression(unlessValue, "unless", parentNamespace, parentTypeName, visitor, visitor.VisitUnless)
}

func (p *Processor) processRelationValue(value starlark.Value, parentNamespace, parentTypeName string, visitor output.Visitor) (any, error) {
	// Get kind first
	kindValue, found, err := getAttr(value, "kind")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("relation value missing 'kind' field")
	}

	kind, err := convert_to_string(kindValue)
	if err != nil {
		return nil, fmt.Errorf("failed to convert kind to string: %w", err)
	}

	// Handle based on kind
	switch kind {
	case "atMostOne", "exactlyOne", "atLeastOne", "many", "boolean":
		return p.processCardinalityRelation(value, kind, parentNamespace, parentTypeName, visitor)
	case "selfType":
		cardinality := mapCardinality(kind)
		return visitor.VisitAssignableExpression(parentNamespace, parentTypeName, cardinality), nil
	case "ref":
		return p.processRefExpression(value, visitor)
	case "subref":
		return p.processSubrefExpression(value, visitor)
	case "or":
		return p.processOrExpression(value, parentNamespace, parentTypeName, visitor)
	case "and":
		return p.processAndExpression(value, parentNamespace, parentTypeName, visitor)
	case "unless":
		return p.processUnlessExpression(value, parentNamespace, parentTypeName, visitor)
	default:
		return nil, fmt.Errorf("unknown relation kind: %s", kind)
	}
}

func (p *Processor) processTypeRelations(typeDict *starlark.Dict, parentNamespace, parentTypeName string, visitor output.Visitor) ([]any, error) {
	relations := make([]any, 0)

	// Get dict items and sort by name for deterministic output
	items := typeDict.Items()
	sortStarlarkDictItems(items)

	// Iterate through sorted dict items
	for _, item := range items {
		relationName := item[0]  // Key (starlark.String)
		relationValue := item[1] // Value (starlark.Dict)

		// Convert relation name to string
		nameStr, err := convert_to_string(relationName)
		if err != nil {
			return nil, fmt.Errorf("failed to convert relation name to string: %w", err)
		}

		// Begin relation
		visitor.BeginRelation(nameStr)

		// Process the relation value dict
		body, err := p.processRelationValue(relationValue, parentNamespace, parentTypeName, visitor)
		if err != nil {
			return nil, fmt.Errorf("failed to process relation %s: %w", nameStr, err)
		}

		// Visit relation and collect result
		result := visitor.VisitRelation(nameStr, body)
		relations = append(relations, result)
	}

	return relations, nil
}
