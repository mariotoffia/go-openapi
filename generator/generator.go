package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mariotoffia/go-openapi/generator/gentypes"
)

// GeneratorContext is used when generating the types.
type GeneratorContext struct {
	settings      Settings
	resolver      gentypes.ReferenceResolverImpl
	specification gentypes.OpenAPISpecificationDefinition
}
type Generator struct {
	// settings is only used to clone into `GeneratorContext`
	settings Settings
}

func NewGenerator(settings Settings) *Generator {
	gen := &Generator{settings: settings}

	if len(gen.settings.inclusion) == 0 && gen.settings.spec == "" {
		// Default inclusion when no spec is provided
		gen.settings.inclusion = []Include{{Path: ".", Glob: "*.yaml"}}
	}

	if gen.settings.loader == nil {
		// Default loader
		gen.settings.loader = &openapi3.Loader{
			Context:               context.Background(),
			IsExternalRefsAllowed: true,
		}
	}

	return gen
}

func (gen *Generator) Generate(ctx GeneratorContext) error {
	ctx.settings = gen.settings
	ctx.resolver = *gentypes.NewReferenceResolver()

	ctx.specification = gentypes.OpenAPISpecificationDefinition{
		Components: map[string]*gentypes.ComponentDefinition{},
	}

	tempSpecFile, err := prepareSpecificationFile(&ctx)

	// Remove temp spec file
	defer func() {
		if tempSpecFile != "" {
			os.Remove(tempSpecFile)
		}
	}()

	// This is after the defer, so that the temp file is removed
	// in any case.
	if err != nil {
		return err
	}

	// Load the specification
	doc, err := ctx.settings.loader.LoadFromFile(ctx.settings.spec)

	if err != nil {
		return err
	}

	err = doc.Validate(ctx.settings.loader.Context)
	if err != nil {
		return err
	}

	ProcessSpecification(&ctx, doc.Components.Schemas)

	return nil
}

func ProcessSpecification(ctx *GeneratorContext, schemas map[string]*openapi3.SchemaRef) error {
	// Iterate all components and create them
	for componentName, v := range schemas {
		// Is it our synthetic component?
		if componentName == "PackageInfo" &&
			v.Value != nil &&
			v.Value.Description == "__go-openapi-gen:remove" {
			// Skip dummy
			continue
		}

		id := gentypes.FromRefString(
			fmt.Sprintf("%s#/components/schemas/%s", ctx.settings.spec, componentName),
			filepath.Dir(ctx.settings.spec),
		)

		comp, err := CreateComponent(ctx, id, v)
		if err != nil {
			return err
		}

		if comp != nil {
			ctx.specification.Components[componentName] = comp
		}
	}

	return nil
}

// CreateComponent will create the `ComponentDefinition` and return it. If there are sub-components such
// as references or `allOf`, `oneOf` or `anyOf` they will be created as well.
//
// The types and components will be registered in the resolver so no duplicates will be created.
func CreateComponent(ctx *GeneratorContext, componentId *gentypes.ComponentReference, ref *openapi3.SchemaRef) (*gentypes.ComponentDefinition, error) {
	// Been here before?
	if ctx.resolver.ResolveComponent(componentId) != nil {
		return nil, nil
	}

	component := &gentypes.ComponentDefinition{
		ID:         *componentId,
		Reference:  nil,
		Definition: nil,
	}

	if ref.Ref == "" {
		// This is a inline type, we need to create it.
		td := gentypes.TypeDefinition{
			ID:          component.ID,
			GoPackage:   component.ID.ToGoPackage(ctx.settings.spec_package),
			Schema:      ref.Value,
			Composition: []gentypes.Composition{},
			Properties:  []gentypes.Property{},
		}

		component.Definition = &td
	} else {
		// This is a reference, we need to resolve it.
		type_ref := gentypes.FromComponentRefAndSchemaRef(&component.ID, ref)
		cached_type := ctx.resolver.ResolveComponent(type_ref)

		if cached_type != nil {
			return nil, nil
		}

	}

	config := gentypes.CreateConfig{
		RootPath:       ctx.settings.models,
		GoRootPackage:  ctx.settings.model_package,
		Name:           componentName,
		GoFileStrategy: gentypes.GoFileStrategyOnePerType,
	}

	if ref.Ref != "" {
		type_ref := gentypes.FromSchemaRef(ref, ctx.settings.models)
		cached_type := ctx.resolver.ResolveReference(type_ref.String())

		if cached_type != nil {
			// Already cached
			// TODO: If override types (then we need to create a new type) - need to upgrade kin-openapi to handle that
			return &gentypes.ComponentDefinition{
				Reference: type_ref,
			}, nil
		}

	}

	// Create type and initialize it.
	td, err := gentypes.CreateType(ref, config)
	if err != nil {
		return nil, err
	}

	// Register the final type
	ctx.resolver.RegisterType(td, ctx.settings.models)

	var registerComponent, returnComponent *gentypes.ComponentDefinition
	if ref.Ref != "" {
		// Return the reference
		returnComponent = &gentypes.ComponentDefinition{
			Reference: gentypes.FromSchemaRef(ref, ctx.settings.models),
		}

		// TODO: If override types (then we need to create a new type) - need to upgrade kin-openapi to handle that
		registerComponent = &gentypes.ComponentDefinition{
			Definition: td,
		}

	} else {
		// Type Definition (not a reference)
		registerComponent = &gentypes.ComponentDefinition{
			Definition: td,
		}

		returnComponent = registerComponent
	}

	ctx.resolver.RegisterComponent(registerComponent, path)
	// Now that we have registered it, we may safely continue resolving
	// the type.

	// TODO: Handle Properties
	for propertyName := range ref.Value.Properties {
		prop := ref.Value.Properties[propertyName]
		ref := prop.Ref

		if ref == "" {
			ref = propertyName
		}

		CreateComponent(ctx, "componentName", filepath.Join(path, ref), prop)

	}

	// TODO: Handle Composites (allOf)

	// TODO: Handle oneOf and anyOf

	return returnComponent, nil
}

// HandleSpecRef is called for each component found in the specification
// and not in a separate module file.
func HandleComponentSpecRef(ctx *GeneratorContext, componentName string, v *openapi3.SchemaRef) error {

	config := gentypes.CreateConfig{
		RootPath:       ctx.settings.models,
		GoRootPackage:  ctx.settings.model_package,
		Name:           componentName,
		GoFileStrategy: gentypes.GoFileStrategyOnePerType,
	}

	td, err := gentypes.CreateType(v, config)

	if err != nil {
		panic(err)
	}

	// Composite
	if len(v.Value.AllOf) > 0 {
		composites, schema, err := HandleComposite(config, td, v.Value.AllOf)
		if err != nil {
			return err
		}

		// Composite Properties
		for i := range composites {
			fmt.Println(i)
			/*
				properties, err := HandleProperties(ctx, config, &composites[i].TypeDefinition)
				if err != nil {
					return err
				}

				composites[i].Properties = properties
			*/
		}

		td.Composition = composites
		td.Schema = schema

	}

	// Properties
	properties, err := HandleProperties(ctx, config, td)
	if err != nil {
		return err
	}

	td.Properties = properties

	// TODO: Handle oneOf, anyOf, not(???)

	// TODO: Add to resolver so we don't create types twice

	data, _ := json.Marshal(td)
	fmt.Println(string(data))
	return nil
}

func HandleProperties(
	ctx *GeneratorContext,
	config gentypes.CreateConfig,
	td *gentypes.TypeDefinition) (properties []gentypes.Property, err error) {

	if len(td.Schema.Properties) == 0 {
		return
	}

	for k := range td.Schema.Properties {
		cfg := config
		property, err := gentypes.CreateProperty(td.Schema.Properties[k], k, cfg)
		if err != nil {
			return nil, err
		}

		properties = append(properties, *property)
	}

	return
}
