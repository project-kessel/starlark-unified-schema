package output

import (
	"bytes"
	"fmt"

	"github.com/project-kessel/ksl-schema-language/pkg/intermediate"
)

type KSILVisitor struct {
	namespaces map[string]*intermediate.Namespace
}

func NewKSILVisitor() *KSILVisitor {
	return &KSILVisitor{
		namespaces: make(map[string]*intermediate.Namespace),
	}
}

func (k *KSILVisitor) VisitAnd(left any, right any) any {
	return &intermediate.RelationBody{
		Kind:  "intersect",
		Left:  left.(*intermediate.RelationBody),
		Right: right.(*intermediate.RelationBody),
	}
}

func (k *KSILVisitor) VisitOr(left any, right any) any {
	return &intermediate.RelationBody{
		Kind:  "union",
		Left:  left.(*intermediate.RelationBody),
		Right: right.(*intermediate.RelationBody),
	}
}

func (k *KSILVisitor) VisitUnless(left any, right any) any {
	return &intermediate.RelationBody{
		Kind:  "except",
		Left:  left.(*intermediate.RelationBody),
		Right: right.(*intermediate.RelationBody),
	}
}

func (k *KSILVisitor) VisitRelationExpression(name string) any {
	return &intermediate.RelationBody{
		Kind:     "reference",
		Relation: name,
	}
}

func (k *KSILVisitor) VisitSubRelationExpression(name string, sub string) any {
	return &intermediate.RelationBody{
		Kind:        "nested_reference",
		Relation:    name,
		SubRelation: sub,
	}
}

func (k *KSILVisitor) VisitAssignableExpression(typeNamespace string, typeName string, cardinality string) any {
	if cardinality == "Many" { //Convert to legacy cardinality
		cardinality = "Any"
	}

	return &intermediate.RelationBody{
		Kind:        "self",
		Types:       []*intermediate.TypeReference{{Namespace: typeNamespace, Name: typeName}},
		Cardinality: cardinality,
	}
}

func (k *KSILVisitor) BeginRelation(name string) {
	// No-op: similar to SpyVisitor, this is just a marker for context
}

// Construct relation expression
func (k *KSILVisitor) VisitRelation(name string, body any) any {
	return &intermediate.Relation{
		Name: name,
		Body: body.(*intermediate.RelationBody),
	}
}

func (k *KSILVisitor) BeginType(namespace string, name string) {
	// Ensure the namespace exists
	if _, exists := k.namespaces[namespace]; !exists {
		k.namespaces[namespace] = &intermediate.Namespace{
			Name:  namespace,
			Types: []*intermediate.Type{},
		}
	}
}

// Construct type expression
func (k *KSILVisitor) VisitType(namespace string, name string, relations []any) any {
	// Convert relations from []any to []*intermediate.Relation
	typedRelations := make([]*intermediate.Relation, len(relations))
	for i, rel := range relations {
		typedRelations[i] = rel.(*intermediate.Relation)
	}

	typeObj := &intermediate.Type{
		Name:      name,
		Relations: typedRelations,
	}

	// Add to the appropriate namespace
	ns := k.namespaces[namespace]
	ns.Types = append(ns.Types, typeObj)

	return typeObj
}

func (k *KSILVisitor) GetOutput() ([]OutputEntry, error) {
	var outputs []OutputEntry

	for nsName, namespace := range k.namespaces {
		var buf bytes.Buffer

		if err := intermediate.Store(namespace, &buf); err != nil {
			return nil, fmt.Errorf("error serializing namespace %s: %w", nsName, err)
		}

		outputs = append(outputs, OutputEntry{
			Path:     nsName + ".json",
			Contents: buf.Bytes(),
		})
	}

	return outputs, nil
}
