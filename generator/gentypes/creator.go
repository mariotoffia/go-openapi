package gentypes

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

type GoFileStrategy string

const (
	// GoFileStrategyOpenAPI will follow the file strategy that the OpenAPI author did
	GoFileStrategyOpenAPI GoFileStrategy = "openapi"
	// GoFileStrategyOnePerType will create one go file per type
	GoFileStrategyOnePerType GoFileStrategy = "one-per-type"
	// GoFileStrategyUseOverride will use the override file name in the `CreateConfig` and fail
	// if not set.
	GoFileStrategyUseOverride GoFileStrategy = "use-override"
)

// CreateConfig is a configuration for the `CreateTypeFromSchema` function.
type CreateConfig struct {
	// RootPath is *mandatory* and is gotten from the `generator.Settings`.
	RootPath string
	// If the _Name_ is not empty, it will be used as the name of the type.
	Name string
	// If the _Path_ is not empty, it will be used as the path of the type.
	Path string
	// GoRootPackage is *mandatory* if not _GoPackage_ is set. It is the root package
	// for the types to be generated.
	GoRootPackage string
	// If the _GoPackage_ is not empty, it will be used as the go package of the type.
	GoPackage string
	// If the _GoFile_ is not empty, it will be used as the go file of the type.
	GoFile string
	// The GoFileStrategy is the strategy to use when creating the go file name.
	// If not set it will default to `GoFileStrategyOpenAPI`.
	GoFileStrategy GoFileStrategy
	// openAPIFile is the filename of the openapi file.
	openAPIFile string
}

// CreateType will create a `TypeDefinitionImpl` that represents a plain component from the given `SchemaRef`.
func CreateType(ref *openapi3.SchemaRef, config CreateConfig) (*TypeDefinition, error) {
	if ref == nil {
		return nil, fmt.Errorf("ref is nil")
	}

	if err := PrepareConfigForCreation(ref, &config); err != nil {
		return nil, err
	}

	td := &TypeDefinition{
		GoPackage: config.GoPackage,
		Schema:    ref.Value,
	}

	return td, nil
}

// CreateComposite will create a `TypeDefinitionImpl` that represents a `allOf` from the given `SchemaRef`.
func CreateComposite(ref *openapi3.SchemaRef, config CreateConfig) (*Composition, error) {
	if ref == nil {
		return nil, fmt.Errorf("ref is nil")
	}

	if err := PrepareConfigForCreation(ref, &config); err != nil {
		return nil, err
	}

	td := &Composition{
		ComponentDefinition: ComponentDefinition{
			Definition: &TypeDefinition{
				GoPackage: config.GoPackage,
				Schema:    ref.Value,
			}},
		Inline: ref.Ref == "",
	}

	return td, nil

}

// CreateProperty will create a `TypeDefinitionImpl` that represents a property onto a component.
func CreateProperty(ref *openapi3.SchemaRef, property string, config CreateConfig) (*Property, error) {
	if ref == nil {
		return nil, fmt.Errorf("ref is nil")
	}

	if err := PrepareConfigForCreation(ref, &config); err != nil {
		return nil, err
	}

	var isRequired bool
	for _, required := range ref.Value.Required {
		if required == property {
			isRequired = true
			break
		}
	}

	td := &Property{
		ComponentDefinition: ComponentDefinition{
			Definition: &TypeDefinition{
				GoPackage: config.GoPackage,
				Schema:    ref.Value,
			}},
		Required:     isRequired,
		PropertyName: property,
	}

	return td, nil

}

func RenameExtensionOnFile(filename, new_extension string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext) + new_extension
}

// PrepareConfigForCreation will prepare and validate the config to be used
// when invoking a creation function.
func PrepareConfigForCreation(ref *openapi3.SchemaRef, config *CreateConfig) error {
	if config == nil || config.RootPath == "" {
		return fmt.Errorf("config is nil or root path is empty")
	}

	if config.GoFileStrategy == "" {
		config.GoFileStrategy = GoFileStrategyOpenAPI
	}

	type_ref := FromSchemaRef(ref, config.RootPath)
	config.openAPIFile = type_ref.Module

	if config.GoFile == "" {
		switch config.GoFileStrategy {
		case GoFileStrategyOpenAPI:
			config.GoFile = RenameExtensionOnFile(config.openAPIFile, ".go")
		case GoFileStrategyOnePerType:
			config.GoFile = RenameExtensionOnFile(strcase.ToSnake(type_ref.TypeName), ".go")
		case GoFileStrategyUseOverride:
			return fmt.Errorf("override file name not set in config")
		default:
			return fmt.Errorf("unknown go file strategy: " + string(config.GoFileStrategy))
		}
	}

	if config.Name == "" {
		config.Name = type_ref.TypeName
	}

	config.Name = strcase.ToCamel(config.Name)

	if config.Path == "" {
		config.Path = type_ref.Path
	}

	if config.Path == "." {
		config.Path = ""
	}

	if config.GoPackage == "" && config.GoRootPackage == "" {
		return fmt.Errorf("neither go package or go root package is set")
	}

	if config.GoPackage == "" {
		if config.Path == "" {
			config.GoPackage = config.GoRootPackage
		} else {
			config.GoPackage = strings.TrimSuffix(config.GoRootPackage, "/") + "/" + strings.ToLower(config.Path)
		}
	}

	return nil
}
