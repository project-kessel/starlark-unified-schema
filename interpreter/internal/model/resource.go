package model

type Resource struct {
	Name      string
	Common    []Field
	Reporters map[string][]Field
}

type Field struct {
	Name        string
	Required    bool
	Description *string
	Type        DataType
}

type DataType struct {
	Kind       string
	MinLength  *int
	MaxLength  *int
	Regex      *string
	Min        *int
	Max        *int
	Inner      *DataType
	Members    []DataType
	Items      *DataType
	Properties []Field
	Required   []string
	Values     []string
}

func (r *Resource) Accept(v SchemaVisitor) any {
	common := make([]any, len(r.Common))
	for i := range r.Common {
		common[i] = r.Common[i].Accept(v)
	}

	reporters := map[string][]any{}
	for name, fields := range r.Reporters {
		group := make([]any, len(fields))
		for i := range fields {
			group[i] = fields[i].Accept(v)
		}
		reporters[name] = group
	}

	return v.VisitResource(r, common, reporters)
}

func (f *Field) Accept(v SchemaVisitor) any {
	typeResult := f.Type.Accept(v)
	return v.VisitField(f, typeResult)
}

func (dt *DataType) Accept(v SchemaVisitor) any {
	var children []any

	switch dt.Kind {
	case "nullable":
		children = []any{dt.Inner.Accept(v)}
	case "union":
		children = make([]any, len(dt.Members))
		for i := range dt.Members {
			children[i] = dt.Members[i].Accept(v)
		}
	case "array":
		children = []any{dt.Items.Accept(v)}
	case "object":
		children = make([]any, len(dt.Properties))
		for i := range dt.Properties {
			children[i] = dt.Properties[i].Accept(v)
		}
	}

	return v.VisitDataType(dt, children)
}
