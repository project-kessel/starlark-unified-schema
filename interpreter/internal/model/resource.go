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

func (r *Resource) Accept(v SchemaVisitor) error {
	common := make([]any, len(r.Common))
	for i := range r.Common {
		result, err := r.Common[i].Accept(v)
		if err != nil {
			return err
		}
		common[i] = result
	}

	reporters := map[string][]any{}
	for name, fields := range r.Reporters {
		group := make([]any, len(fields))
		for i := range fields {
			result, err := fields[i].Accept(v)
			if err != nil {
				return err
			}
			group[i] = result
		}
		reporters[name] = group
	}

	_, err := v.VisitResource(r, common, reporters)
	return err
}

func (f *Field) Accept(v SchemaVisitor) (any, error) {
	typeResult, err := f.Type.Accept(v)
	if err != nil {
		return nil, err
	}
	return v.VisitField(f, typeResult)
}

func (dt *DataType) Accept(v SchemaVisitor) (any, error) {
	var children []any

	switch dt.Kind {
	case "nullable":
		inner, err := dt.Inner.Accept(v)
		if err != nil {
			return nil, err
		}
		children = []any{inner}
	case "union":
		children = make([]any, len(dt.Members))
		for i := range dt.Members {
			result, err := dt.Members[i].Accept(v)
			if err != nil {
				return nil, err
			}
			children[i] = result
		}
	case "array":
		items, err := dt.Items.Accept(v)
		if err != nil {
			return nil, err
		}
		children = []any{items}
	case "object":
		children = make([]any, len(dt.Properties))
		for i := range dt.Properties {
			result, err := dt.Properties[i].Accept(v)
			if err != nil {
				return nil, err
			}
			children[i] = result
		}
	}

	return v.VisitDataType(dt, children)
}
