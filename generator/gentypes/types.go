package gentypes

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// OpenAPISpecificationDefinition is the open api specification.
type OpenAPISpecificationDefinition struct {
	// Components are the component schemas where
	// the name is the key and the value is the
	// component definition.
	Components map[string]*ComponentDefinition
}

// TypeDefinition is the concrete definition of a `ComponentDefinition`
// where the actual type is defined.
type TypeDefinition struct {
	// ID is the identity of this type. It is usually the
	// same as the `ComponentDefinition.ID`
	ID ComponentReference
	// GoPackage returns the fully qualified go package name of
	// the generated go file for this type.
	GoPackage string
	// Schema returns the schema that created this `TypeDefinition`.
	Schema *openapi3.Schema
	// Composition returns a set of types that this `TypeDefinition`
	// inherits from. This is the _OpenAPI_ `allOf` keyword.
	Composition []Composition
	// Properties returns the set of properties that this `TypeDefinition`
	// has. This is the _OpenAPI_ `properties` keyword.
	Properties []Property
}

type Composition struct {
	ComponentDefinition
	// Inline will inline even if it is a reference in the `OpenAPI` specification.
	Inline bool
}

type Property struct {
	ComponentDefinition
	// Required specified that this is required
	Required bool
	// PropertyName is the name of the property
	PropertyName string
}
