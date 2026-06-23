package output

type Members struct {
	DataFields     []any
	RelationFields []any
	Permissions    []any
}

type SchemaVisitor interface {
	BeginType(name string)
	VisitResource(typeName string, reporter string, commonMembers *Members, reporterMembers *Members) error

	VisitDataField(name string, required bool, description *string, dataType any) any

	VisitTextDataType(minLength *int, maxLength *int, regex *string) any
	VisitUUIDDataType() any
	VisitNumericIDDataType(min *int, max *int) any
	VisitBooleanDataType() any
	VisitDateTimeDataType() any
	VisitEnumDataType(values []string) any
	VisitNullableDataType(inner any) any
	VisitCompositeDataType(dataTypes []any) any
	VisitArrayDataType(items any) any
	VisitObjectDataType(properties []any, required []string) any

	VisitAnd(left any, right any) any
	VisitOr(left any, right any) any
	VisitUnless(left any, right any) any
	VisitReferenceExpression(name string) any
	VisitSubReferenceExpression(name string, sub string) any

	VisitRelation(name string, reporter string, typeName string, cardinality string, idType any) any

	BeginPermission(name string)
	VisitPermission(name string, body any) any

	Results() ([]OutputEntry, error)
}
