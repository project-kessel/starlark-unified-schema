package domain

type SchemaVisitor interface {
	VisitResource(resource Resource) error
}
