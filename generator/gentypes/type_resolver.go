package gentypes

type ReferenceResolverImpl struct {
	open_api map[string]*TypeDefinitionImpl
	go_types map[string]*TypeDefinitionImpl
}

func (r *ReferenceResolverImpl) RegisterType(td *TypeDefinitionImpl) {
	if td.GoPackage != "" && td.Name != "" {
		r.go_types[td.GoPackage+"/"+td.Name] = td
	}

	if td.OpenAPIFile != "" && td.Path != "" && td.Name != "" && td.RootPath != "" {
		r.open_api[ToOpenAPIReference(td)] = td
	}
}

// ResolveReference resolves a type based on the open api reference or golang
// fully qualified package.
//
// NOTE: The open api reference should be formatted by the `ToOpenAPIReference()`
// function. That is path/file#/name. Where path is normalized against the root
// path.
func (r *ReferenceResolverImpl) ResolveReference(ref string) *TypeDefinitionImpl {
	if td, ok := r.open_api[ref]; ok {
		return td
	}

	if td, ok := r.go_types[ref]; ok {
		return td
	}

	return nil
}
