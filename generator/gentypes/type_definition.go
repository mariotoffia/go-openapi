package gentypes

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// TypeDefinition is the root definition of a type.
type TypeDefinition interface {
	// GetRootPath returns the root path where all other paths
	// are relative to.
	GetRootPath() string
	// Path returns the relative path (to _rootPath_) to the type.
	GetPath() string
	// Name returns the name of the type.
	GetName() string
	// OpenAPIFile returns the filename of the openapi file
	GetOpenAPIFile() string
	// GoFile returns the filename of the generated or to be
	// generated go file.
	//
	// NOTE: That many `TypeDefinition` can point to the same
	// go_file including same _path_.
	GetGoFile() string
	// GoPackage returns the fully qualified go package name of
	// the generated go file for this type.
	GetGoPackage() string
	// Schema returns the schema that created this `TypeDefinition`.
	GetSchema() *openapi3.Schema
	// Inherits returns a set of types that this `TypeDefinition`
	// inherits from. This is the _OpenAPI_ `allOf` keyword.
	GetInherits() []TypeInheritanceDefinition
	// Properties returns the set of properties that this `TypeDefinition`
	// has. This is the _OpenAPI_ `properties` keyword.
	GetProperties() []TypePropertyDefinition
}

// TypeInheritanceDefinition is a type that is part of an inheritance.
type TypeInheritanceDefinition interface {
	TypeDefinition
	// Inline returns `true` if the type is inlined, otherwise
	// it may be a pointer to the type.
	//
	// NOTE: This is automatically set to `true` if the _allOf_ entry is
	// expressing a inline type definition. This is only controllable
	// when a _allOf_ entry is expressing a_$ref_.
	GetInline() bool
}

// TypePropertyDefinition is a property of a type.
type TypePropertyDefinition interface {
	TypeDefinition
	// Required returns `true` if the property is required.
	GetRequired() bool
	// inline is set to `true` if the type is inlined, otherwise
	// it may be a pointer to the type.
	GetInline() bool
}

// TypeDefinitionImpl is the root definition of a type.
type TypeDefinitionImpl struct {
	RootPath    string
	Path        string
	Name        string
	OpenAPIFile string
	GoFile      string
	GoPackage   string
	Schema      *openapi3.Schema
	Inherits    []TypeInheritanceDefinitionImpl
	Properties  []TypePropertyDefinitionImpl
}

// TypeInheritanceDefinitionImpl is a type that is part of an inheritance.
type TypeInheritanceDefinitionImpl struct {
	TypeDefinitionImpl
	Inline bool
}

// TypePropertyDefinitionImpl is a property of a type.
type TypePropertyDefinitionImpl struct {
	TypeDefinitionImpl
	Required bool
	Inline   bool
}

func (td TypeDefinitionImpl) GetRootPath() string {
	return td.RootPath
}
func (td TypeDefinitionImpl) GetPath() string {
	return td.Path
}
func (td TypeDefinitionImpl) GetName() string {
	return td.Name
}
func (td TypeDefinitionImpl) GetOpenAPIFile() string {
	return td.OpenAPIFile
}
func (td TypeDefinitionImpl) GetGoFile() string {
	return td.GoFile
}
func (td TypeDefinitionImpl) GetGoPackage() string {
	return strings.TrimSuffix(td.GoPackage, "/")
}
func (td TypeDefinitionImpl) GetSchema() *openapi3.Schema {
	return td.Schema
}
func (td TypeDefinitionImpl) GetInherits() []TypeInheritanceDefinition {
	res := make([]TypeInheritanceDefinition, len(td.Inherits))
	for i, v := range td.Inherits {
		res[i] = v
	}
	return res
}
func (td TypeDefinitionImpl) GetProperties() []TypePropertyDefinition {
	res := make([]TypePropertyDefinition, len(td.Properties))
	for i, v := range td.Properties {
		res[i] = v
	}
	return res
}

func (td TypeInheritanceDefinitionImpl) GetInline() bool {
	return td.Inline
}

func (td TypePropertyDefinitionImpl) GetRequired() bool {
	return td.Required
}
func (td TypePropertyDefinitionImpl) GetInline() bool {
	return td.Inline
}
