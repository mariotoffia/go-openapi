package generator

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mariotoffia/go-openapi/generator/gentypes"
)

// HandleDiscriminatorBasedPolymorphism will handle when a discriminator and mapping
// is provided to determine what type to use. The is the _oneOf_ schema element.
func HandleDiscriminatorBasedPolymorphism(
	ctx *GeneratorContext,
	td *gentypes.TypeDefinition,
	componentId *gentypes.ComponentReference,
	def *openapi3.Schema) error {

	if len(def.OneOf) == 0 {
		return nil
	}

	if td.Schema.Discriminator == nil || td.Schema.Discriminator.PropertyName == "" {
		return nil
	}

	if len(td.Schema.Properties) > 0 {
		return fmt.Errorf("discriminator not supported on object with properties (component: %s)", componentId)
	}

	if len(td.Schema.AllOf) > 0 {
		return fmt.Errorf("discriminator not supported on object with allOf (component: %s)", componentId)
	}

	if len(td.Schema.AnyOf) > 0 {
		return fmt.Errorf("discriminator not supported on object with anyOf (component: %s)", componentId)
	}

	for i := range def.OneOf {
		compose_type := def.OneOf[i]

		if IsDefinition(compose_type) {
			return fmt.Errorf("discriminator object with inline object not supported (component: %s)", componentId)
		}
	}

	mapping_table := CreateMappingTable(ctx, componentId, def)

	get_map_from := func(ref *gentypes.ComponentReference) string {
		for from, map_ref := range mapping_table {
			if ref.Equal(&map_ref) {
				return from
			}
		}

		return ""
	}

	for i := range def.OneOf {
		ref := ResolveReferenceAndSwitchIfNeeded(ctx, componentId, def.OneOf[i])
		// Make sure the the _ref_ is created
		if ctx.resolver.ResolveComponent(ref) == nil {
			if _, err := CreateComponentFromReference(ctx, ref, def.OneOf[i]); err != nil {
				return err
			}
		}

		td.DiscriminatorComponents = append(td.DiscriminatorComponents, gentypes.DiscriminatorComponent{
			ComponentDefinition: gentypes.ComponentDefinition{
				ID:        *ref,
				Reference: ref,
			},
			Discriminator: td.Schema.Discriminator.PropertyName,
			MapFrom:       get_map_from(ref),
		})
	}

	return nil
}

// CreateMappingTable makes sure that a complete mapping table is returned, it will fill in the entries
// that have not been defined in the mapping table. It does so by using the type name of the component
// as key and the reference as value.
//
// The _componentId_ is the component that the mapping table is created for (and the _schema_ represents).
//
// NOTE: Some entries may have a namespace since referenced in same module (file).
//
// Entries that have been added will be resolved so they too have a `ComponentReference` and not just
// a `SchemaRef`.
func CreateMappingTable(
	ctx *GeneratorContext,
	componentId *gentypes.ComponentReference,
	schema *openapi3.Schema) map[string]gentypes.ComponentReference {

	mapping := map[string]gentypes.ComponentReference{}

	if schema.Discriminator == nil || schema.Discriminator.PropertyName == "" {
		return mapping
	}

	for name, reference := range schema.Discriminator.Mapping {

		mapping[name] = *ResolveReferenceAndSwitchIfNeeded(
			ctx, componentId, &openapi3.SchemaRef{Ref: reference},
		)

	}

	finder := func(ref *gentypes.ComponentReference) bool {
		for _, map_ref := range mapping {
			if ref.Equal(&map_ref) {
				return true
			}
		}

		return false
	}

	// Add missing mappings
	if len(mapping) < len(schema.OneOf) {
		for i := range schema.OneOf {

			ref := ResolveReferenceAndSwitchIfNeeded(ctx, componentId, schema.OneOf[i])
			if _, ok := mapping[ref.TypeName]; ok {
				continue
			}

			if !finder(ref) {
				mapping[ref.TypeName] = *ref
			}
		}
	}

	return mapping
}
