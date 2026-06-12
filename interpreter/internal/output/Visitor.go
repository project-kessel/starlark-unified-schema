package output

type Visitor interface {
	VisitAnd(left any, right any) any
	VisitOr(left any, right any) any
	VisitUnless(left any, right any) any
	VisitRelationExpression(name string) any
	VisitSubRelationExpression(name string, sub string) any
	VisitAssignableExpression(typeNamespace string, typeName string, cardinality string) any

	BeginRelation(name string)
	// Construct relation expression
	VisitRelation(name string, body any) any

	BeginType(namespace string, name string)
	// Construct type expression
	VisitType(namespace string, name string, relations []any) any

	GetOutput() ([]OutputEntry, error)
}
