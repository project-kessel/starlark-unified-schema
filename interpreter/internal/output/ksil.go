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

func (k *KSILVisitor) BeginRelation(name string) {

}

func (k *KSILVisitor) BeginType(name string) {}

func (k *KSILVisitor) VisitResource(typeName string, reporter string, commonMembers *Members, reporterMembers *Members) error {
	if _, exists := k.namespaces[reporter]; !exists {
		k.namespaces[reporter] = &intermediate.Namespace{
			Name:  reporter,
			Types: []*intermediate.Type{},
		}
	}

	// Convert relations from []any to []*intermediate.Relation
	typedRelations := make([]*intermediate.Relation, 0,
		len(commonMembers.RelationFields)+len(reporterMembers.RelationFields)+
			len(commonMembers.Permissions)+len(reporterMembers.Permissions))

	for _, rel := range commonMembers.RelationFields {
		typedRelations = append(typedRelations, rel.(*intermediate.Relation))
	}
	for _, rel := range reporterMembers.RelationFields {
		typedRelations = append(typedRelations, rel.(*intermediate.Relation))
	}
	for _, perm := range commonMembers.Permissions {
		typedRelations = append(typedRelations, perm.(*intermediate.Relation))
	}
	for _, perm := range reporterMembers.Permissions {
		typedRelations = append(typedRelations, perm.(*intermediate.Relation))
	}

	typeObj := &intermediate.Type{
		Name:      typeName,
		Relations: typedRelations,
	}

	// Add to the appropriate namespace
	ns := k.namespaces[reporter]
	ns.Types = append(ns.Types, typeObj)

	return nil
}

func (k *KSILVisitor) VisitReferenceExpression(name string) any {
	return &intermediate.RelationBody{
		Kind:     "reference",
		Relation: name,
	}
}

func (k *KSILVisitor) VisitSubReferenceExpression(name string, sub string) any {
	return &intermediate.RelationBody{
		Kind:        "nested_reference",
		Relation:    name,
		SubRelation: sub,
	}
}

func (k *KSILVisitor) BeginPermission(name string) {

}

func (k *KSILVisitor) VisitPermission(name string, body any) any {
	return &intermediate.Relation{
		Name: name,
		Body: body.(*intermediate.RelationBody),
	}
}

func (k *KSILVisitor) Results() ([]OutputEntry, error) {
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

func (k *KSILVisitor) VisitRelation(name string, reporter string, typeName string, cardinality string, idType any) any {
	if cardinality == "Many" { //Convert to legacy cardinality
		cardinality = "Any"
	}

	body := &intermediate.RelationBody{
		Kind:        "self",
		Types:       []*intermediate.TypeReference{{Namespace: reporter, Name: typeName}},
		Cardinality: cardinality,
	}

	return &intermediate.Relation{
		Name: name,
		Body: body,
	}
}

func (k *KSILVisitor) VisitDataField(name string, required bool, description *string, dataType any) any {
	return nil
}

func (k *KSILVisitor) VisitTextDataType(minLength *int, maxLength *int, regex *string) any {
	return nil
}

func (k *KSILVisitor) VisitUUIDDataType() any {
	return nil
}

func (k *KSILVisitor) VisitNumericIDDataType(min *int, max *int) any {
	return nil
}

func (k *KSILVisitor) VisitBooleanDataType() any {
	return nil
}

func (k *KSILVisitor) VisitDateTimeDataType() any {
	return nil
}

func (k *KSILVisitor) VisitEnumDataType(values []string) any {
	return nil
}

func (k *KSILVisitor) VisitNullableDataType(inner any) any {
	return nil
}

func (k *KSILVisitor) VisitCompositeDataType(dataTypes []any) any {
	return nil
}

func (k *KSILVisitor) VisitArrayDataType(items any) any {
	return nil
}

func (k *KSILVisitor) VisitObjectDataType(properties []any, required []string) any {
	return nil
}
