package generator

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mariotoffia/go-openapi/generator/gentypes"
)

func HandleComposition(
	ctx *GeneratorContext,
	td *gentypes.TypeDefinition,
	component *gentypes.ComponentDefinition,
	def *openapi3.Schema) error {

	if len(def.AllOf) == 0 {
		return nil
	}

	// Extract *all* inline objects first
	var err error

	for {

		inline_objects := openapi3.SchemaRefs{}

		for i := range def.AllOf {
			compose_type := def.AllOf[i]

			if compose_type.Ref == "" {
				inline_objects = append(inline_objects, compose_type)
				continue
			}
		}

		if len(inline_objects) == 0 {
			break
		}

		// Remove all inline objects in td.Schema.AllOf
		for i := range inline_objects {
			td.Schema.AllOf = RemoveSchemaRef(td.Schema.AllOf, inline_objects[i])
		}

		// Merge the inline objects into the main schema.
		for i := range inline_objects {
			c := inline_objects[i]
			if td.Schema, err = MergeSchemaObjects(ctx, td.Schema, c.Value); err != nil {
				return err
			}
		}

	}

	// Handle all references (all inline objects have been removed)
	for i := range def.AllOf {
		compose_type := def.AllOf[i]
		// Reference to other type.
		ref := ctx.resolver.ResolveComponent(&component.ID)
		if ref == nil {
			if ref, err = CreateComponentFromReference(ctx, &component.ID, compose_type); err != nil {
				return err
			}
		}

		td.Composition = append(td.Composition, gentypes.Composition{
			ComponentDefinition: *ref,
			Inline:              false,
		})
	}

	return nil
}
