package gentypes

// ComponentDefinition is a `TypeDefinition` or reference to
// `TypeReference`.
type ComponentDefinition struct {
	// ID is the identity of the component itself.
	ID ComponentReference
	// Definition contains the definition of the component if
	// not just importing it via `Reference`.
	Definition *TypeDefinition
	// Reference contains the reference to another component that
	// will be the source of the definition or just a pointer
	// to the other component.
	//
	// If the former, then it will do a shallow copy of the
	// other component and set the `ID` to the `ID` of this
	// component.
	Reference *ComponentReference
}

// IsReference returns true when only the `Reference` is set and the
// `Definition` is `nil`.
func (cd *ComponentDefinition) IsReference() bool {
	return cd.Reference != nil && cd.Definition == nil
}

// IsDefinition returns true when only the `Definition` is set and the
// `Reference` is `nil`.
func (cd *ComponentDefinition) IsDefinition() bool {
	return cd.Definition != nil && cd.Reference == nil
}
