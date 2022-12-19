package generator

import (
	"context"
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

func (ctx *GeneratorContext) GetSpecification() *gentypes.OpenAPISpecificationDefinition {
	return &ctx.specification
}

func (ctx *GeneratorContext) GetResolver() *gentypes.ReferenceResolverImpl {
	return &ctx.resolver
}

func (ctx *GeneratorContext) ResolveTypeDefinition(ref *gentypes.ComponentReference) *gentypes.TypeDefinition {
	for {
		component := ctx.resolver.ResolveComponent(ref)
		if component == nil {
			return nil
		}

		if component.Definition != nil {
			return component.Definition
		}

		if component.Reference == nil {
			return nil
		}

		ref = component.Reference
	}
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

func (gen *Generator) Generate(ctx *GeneratorContext) error {
	ctx.settings = gen.settings
	ctx.resolver = *gentypes.NewReferenceResolver()

	ctx.specification = gentypes.OpenAPISpecificationDefinition{
		Components: map[string]*gentypes.ComponentDefinition{},
	}

	tempSpecFile, err := prepareSpecificationFile(ctx)

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

	ProcessSpecification(ctx, doc.Components.Schemas)

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
			fmt.Sprintf(
				"%s#/components/schemas/%s",
				filepath.Base(ctx.settings.spec), componentName), ctx.settings.spec_root,
		)

		comp, err := CreateComponentFromReference(ctx, id, v)
		if err != nil {
			return err
		}

		if comp != nil {
			ctx.specification.Components[componentName] = comp
		}
	}

	return nil
}

// CreateComponentFromReference will create the `ComponentDefinition` and return it. If there are sub-components such
// as references or `allOf`, `oneOf` or `anyOf` they will be created as well.
//
// The types and components will be registered in the resolver so no duplicates will be created.
func CreateComponentFromReference(
	ctx *GeneratorContext,
	componentId *gentypes.ComponentReference,
	ref *openapi3.SchemaRef) (*gentypes.ComponentDefinition, error) {
	// Been here before?
	if ctx.resolver.ResolveComponent(componentId) != nil {
		return nil, nil
	}

	if IsDefinition(ref) {
		return CreateComponentFromDefinition(
			ctx, componentId, ref.Value,
		)
	}
	// This is a reference
	component := &gentypes.ComponentDefinition{
		ID:         *componentId,
		Reference:  ResolveReferenceAndSwitchIfNeeded(ctx, componentId, ref),
		Definition: nil,
	}

	cached_type := ctx.resolver.ResolveComponent(component.Reference)

	if cached_type != nil {
		return component, nil
	}

	CreateComponentFromDefinition(ctx, component.Reference, ref.Value)
	return component, nil
}

func CreateComponentFromDefinition(
	ctx *GeneratorContext,
	componentId *gentypes.ComponentReference,
	def *openapi3.Schema) (*gentypes.ComponentDefinition, error) {
	// Been here before?
	if ctx.resolver.ResolveComponent(componentId) != nil {
		return nil, nil
	}

	td := gentypes.TypeDefinition{
		ID:          *componentId,
		GoPackage:   ResolveGoPackage(ctx, componentId),
		Schema:      def,
		Composition: []gentypes.Composition{},
		Properties:  []gentypes.Property{},
	}

	component := &gentypes.ComponentDefinition{
		ID:         *componentId,
		Reference:  nil,
		Definition: &td,
	}

	// Register the component so it will be resolved if cyclic references
	ctx.resolver.RegisterComponent(component)

	// Composition
	if err := HandleComposition(ctx, &td, component, def); err != nil {
		return nil, err
	}

	// Handle Reference Properties
	if err := HandleProperties(ctx, &td, component, def); err != nil {
		return nil, err
	}

	// TODO: Chase down anyOf, oneOf

	return component, nil
}
