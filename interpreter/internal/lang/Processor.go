package lang

import (
	"fmt"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"go.starlark.net/starlark"
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

	// Iterate through module globals
	for globalName, globalValue := range globals {
		// Check if this global is a type (Dict)
		if typeDict, ok := globalValue.(*starlark.Dict); ok {
			// Get metadata for this type
			metadata, exists := p.metadata[typeDict]
			if !exists {
				return fmt.Errorf("no metadata found for type %s", globalName)
			}

			namespace := metadata.moduleName
			typeName := metadata.typeName

			// Begin type processing
			visitor.BeginType(namespace, typeName)

			// Process relations
			relations, err := p.processTypeRelations(typeDict, namespace, typeName, visitor)
			if err != nil {
				return fmt.Errorf("failed to process relations for type %s: %w", typeName, err)
			}

			// Visit the type
			visitor.VisitType(namespace, typeName, relations)
		}
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

func (p *Processor) processCardinalityRelation(relationDict *starlark.Dict, kind string, parentNamespace, parentTypeName string, visitor output.Visitor) (any, error) {
	// Get the "type" field
	typeValue, found, err := relationDict.Get(starlark.String("type"))
	if !found || err != nil {
		return nil, fmt.Errorf("cardinality relation missing 'type' field")
	}

	// Check if type is a Dict (reference to another type)
	if typeDict, ok := typeValue.(*starlark.Dict); ok {
		// Check if it's a selfType
		selfKindValue, found, _ := typeDict.Get(starlark.String("kind"))
		if found {
			selfKind, _ := convert_to_string(selfKindValue)
			if selfKind == "selfType" {
				// Self-reference: use parent type
				cardinality := mapCardinality(kind)
				return visitor.VisitAssignableExpression(parentNamespace, parentTypeName, cardinality), nil
			}
		}

		// Look up metadata for this type dict
		metadata, exists := p.metadata[typeDict]
		if !exists {
			return nil, fmt.Errorf("no metadata found for referenced type dict")
		}

		namespace := metadata.moduleName
		typeName := metadata.typeName

		// Map kind to cardinality string
		cardinality := mapCardinality(kind)

		// Call visitor
		return visitor.VisitAssignableExpression(namespace, typeName, cardinality), nil
	}

	return nil, fmt.Errorf("type field is not a dict, got: %T", typeValue)
}

func (p *Processor) processRelationValue(value starlark.Value, parentNamespace, parentTypeName string, visitor output.Visitor) (any, error) {
	// Relation values are dicts with "kind" and "type" fields
	valueDict, ok := value.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("relation value is not a dict, got: %T", value)
	}

	// Get the "kind" field
	kindValue, found, err := valueDict.Get(starlark.String("kind"))
	if !found || err != nil {
		return nil, fmt.Errorf("relation dict missing 'kind' field")
	}

	kind, err := convert_to_string(kindValue)
	if err != nil {
		return nil, fmt.Errorf("failed to convert kind to string: %w", err)
	}

	// Handle based on kind
	switch kind {
	case "atMostOne", "exactlyOne", "atLeastOne", "many", "boolean":
		return p.processCardinalityRelation(valueDict, kind, parentNamespace, parentTypeName, visitor)
	case "selfType":
		// Self-type: use parent type's namespace and name
		cardinality := mapCardinality(kind)
		return visitor.VisitAssignableExpression(parentNamespace, parentTypeName, cardinality), nil
	default:
		return nil, fmt.Errorf("unknown relation kind: %s", kind)
	}
}

func (p *Processor) processTypeRelations(typeDict *starlark.Dict, parentNamespace, parentTypeName string, visitor output.Visitor) ([]any, error) {
	relations := make([]any, 0)

	// Iterate through dict items
	for _, item := range typeDict.Items() {
		relationName := item[0] // Key (starlark.String)
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
