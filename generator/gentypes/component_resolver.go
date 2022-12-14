package gentypes

type ReferenceResolverImpl struct {
	components map[string]*ComponentDefinition
}

func NewReferenceResolver() *ReferenceResolverImpl {
	return &ReferenceResolverImpl{
		components: make(map[string]*ComponentDefinition),
	}
}

func (r *ReferenceResolverImpl) RegisterComponent(td *ComponentDefinition) {
	r.components[td.ID.String()] = td
}

func (r *ReferenceResolverImpl) ResolveComponent(ref *ComponentReference) *ComponentDefinition {
	return r.components[ref.String()]
}

func (r *ReferenceResolverImpl) Components() map[string]*ComponentDefinition {
	return r.components
}
