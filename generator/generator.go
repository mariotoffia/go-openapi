package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
			fmt.Sprintf("%s#/components/schemas/%s", filepath.Base(ctx.settings.spec), componentName),
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
		// This is a inline type
		td := gentypes.TypeDefinition{
			ID:          component.ID,
			GoPackage:   ResolveGoPackage(ctx, &component.ID),
			Schema:      ref.Value,
			Composition: []gentypes.Composition{},
			Properties:  []gentypes.Property{},
		}

		component.Definition = &td

	} else {
		// This is a reference
		component.Reference = ResolveReferenceAndSwitchIfNeeded(ctx, &component.ID, ref)
		cached_type := ctx.resolver.ResolveComponent(component.Reference)

		if cached_type != nil {
			return component, nil
		}

	}

	// TODO: If component contains definition -> Chase down properties, allOf, ...
	// TODO: If component contains reference -> create type and put that into new component
	// TODO: If the ref component.TypeName != component.ID.TypeName -> create a additional
	// new type by shallow copy the ref component and change the type name.

	return component, nil
}

// ResolveReferenceAndSwitchIfNeeded will create a `ComponentReference`. If the _ref_ is under the specification root path
// then it will be used. If it is under model root path it will use that instead to create the `ComponentReference`
// as root path. If specification and module is on the same root path, the the longest path will be used as root path.
func ResolveReferenceAndSwitchIfNeeded(
	ctx *GeneratorContext,
	componentId *gentypes.ComponentReference,
	ref *openapi3.SchemaRef,
) *gentypes.ComponentReference {
	// Create the fully qualified path to the reference
	ref_path := filepath.Join(componentId.RootPath, componentId.Path, ref.Ref)
	ref_path = filepath.Clean(ref_path)

	var type_ref *gentypes.ComponentReference

	if IsSpecificationRooted(ctx, ref_path) {
		if IsSpecificationRef(ctx, componentId) {
			// Specification rooted component id -> no switch
			type_ref = gentypes.FromSchemaRef(ref, ctx.settings.spec)
		} else {
			// Switch root to module root
			type_ref = gentypes.FromSchemaRef(ref, ctx.settings.model_root)
		}
	} else {
		if !IsSpecificationRef(ctx, componentId) {
			// Module rooted component id -> no switch
			type_ref = gentypes.FromSchemaRef(ref, ctx.settings.model_root)
		} else {
			// Switch root to spec root
			type_ref = gentypes.FromSchemaRef(ref, ctx.settings.spec)
		}
	}

	return type_ref
}

func ResolveGoPackage(ctx *GeneratorContext, ref *gentypes.ComponentReference) string {

	if IsSpecificationRef(ctx, ref) {
		return ref.ToGoPackage(ctx.settings.spec_package)
	}

	return ref.ToGoPackage(ctx.settings.model_package)
}

func IsSpecificationRef(ctx *GeneratorContext, ref *gentypes.ComponentReference) bool {
	return IsSpecificationRooted(ctx, ref.RootPath)
}

func IsSpecificationRooted(ctx *GeneratorContext, path string) bool {

	if strings.HasPrefix(path, ctx.settings.model_root) {
		if strings.HasPrefix(path, ctx.settings.spec) {
			// Both are rooted - if longest is spec then it is spec rooted
			return len(ctx.settings.spec) > len(ctx.settings.model_root)
		}

		return false
	}

	return strings.HasPrefix(path, ctx.settings.spec)
}
