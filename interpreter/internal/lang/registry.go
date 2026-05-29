package lang

type ResourceRegistry struct {
	resources []ResourceDefinition
}

func NewResourceRegistry() *ResourceRegistry {
	return &ResourceRegistry{}
}

func (r *ResourceRegistry) Register(def ResourceDefinition) {
	r.resources = append(r.resources, def)
}

func (r *ResourceRegistry) Resources() []ResourceDefinition {
	return r.resources
}
