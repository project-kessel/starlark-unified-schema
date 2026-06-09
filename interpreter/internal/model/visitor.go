package model

type SchemaVisitor interface {
	VisitResource(r *Resource, common []any, reporters map[string][]any) any
	VisitField(f *Field, dataType any) any
	VisitDataType(dt *DataType, children []any) any
}
