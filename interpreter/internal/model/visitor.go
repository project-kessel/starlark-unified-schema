package model

type SchemaVisitor interface {
	VisitResource(resource Resource) error
}
