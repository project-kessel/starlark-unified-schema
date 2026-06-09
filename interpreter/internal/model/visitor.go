package model

type SchemaVisitor interface {
	VisitResource(r *Resource, common []any, reporters map[string][]any) (any, error)
	VisitField(f *Field, dataType any) (any, error)
	VisitDataType(dt *DataType, children []any) (any, error)
}
