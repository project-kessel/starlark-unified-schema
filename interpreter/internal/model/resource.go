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
