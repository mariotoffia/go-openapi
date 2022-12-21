package generator

import (
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

	// Extract *all* inline objects first
	var err error

	for {

		inline_objects := openapi3.SchemaRefs{}

		for i := range def.OneOf {
			compose_type := def.OneOf[i]

			if IsDefinition(compose_type) {
				inline_objects = append(inline_objects, compose_type)
				continue
			}
		}

		if len(inline_objects) == 0 {
			break
		}

		// Remove all inline objects in td.Schema.OneOf
		for i := range inline_objects {
			td.Schema.OneOf = RemoveSchemaRef(td.Schema.OneOf, inline_objects[i])
		}

		// Merge the inline objects into the main schema.
		for i := range inline_objects {
			c := inline_objects[i]
			if td.Schema, err = MergeSchemaObjects(ctx, td.Schema, c.Value); err != nil {
				return err
			}
		}
	}

	mapping_table := CreateMappingTable(ctx, componentId, def)

	// Add itself as a discriminator component
	if len(def.Properties) > 0 {

		td.DiscriminatorComponents = append(td.DiscriminatorComponents, gentypes.DiscriminatorComponent{
			ComponentDefinition: gentypes.ComponentDefinition{
				ID:        *componentId.NewWithAppendTypeName("Polymorphic"),
				Reference: componentId,
			},
			Discriminator: td.Schema.Discriminator.PropertyName,
			MapFrom:       "TODO: MapFrom",
		})
	}

	// Handle all references (all inline objects have been removed)
	for i := range def.OneOf {
		ref := ResolveReferenceAndSwitchIfNeeded(ctx, componentId, def.OneOf[i])
		// Make sure the the _ref_ is created
		if ctx.resolver.ResolveComponent(ref) == nil {
			if _, err = CreateComponentFromReference(ctx, ref, def.OneOf[i]); err != nil {
				return err
			}
		}

		td.DiscriminatorComponents = append(td.DiscriminatorComponents, gentypes.DiscriminatorComponent{
			ComponentDefinition: gentypes.ComponentDefinition{
				ID:        *ref.NewWithAppendTypeName("Polymorphic"),
				Reference: ref,
			},
			Discriminator: td.Schema.Discriminator.PropertyName,
			MapFrom:       "TODO: MapFrom",
		})
	}

	// TODO: If inline (type=object && len(properties) > 0...) add discriminator to self.

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
			if map_ref.Equal(ref) {
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
