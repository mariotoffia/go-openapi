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

func (ctx *GeneratorContext) GetSpecification() *gentypes.OpenAPISpecificationDefinition {
	return &ctx.specification
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

	if ref.Ref == "" {
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

// Contains will check if a string is in a slice of strings.
func Contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// RemoveSchemaRef will remove a _ref_ from the _references_ slice if found.
func RemoveSchemaRef(references openapi3.SchemaRefs, ref *openapi3.SchemaRef) openapi3.SchemaRefs {
	for j := range references {
		if references[j] == ref {
			return append(references[:j], references[j+1:]...)
		}
	}

	return references
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
			type_ref = gentypes.FromSchemaRef(ref, ctx.settings.spec_root)
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
			type_ref = gentypes.FromSchemaRef(ref, ctx.settings.spec_root)
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
		if strings.HasPrefix(path, ctx.settings.spec_root) {
			// Both are rooted - if longest is spec then it is spec rooted
			return len(ctx.settings.spec_root) > len(ctx.settings.model_root)
		}

		return false
	}

	return strings.HasPrefix(path, ctx.settings.spec_root)
}
