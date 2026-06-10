package output

type SchemaVisitor interface {
	BeginType(name string)
	VisitResource(typeName string, reporter string, commonFields []any, reporterFields []any) error

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
}
