package generator

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
	"github.com/mariotoffia/go-openapi/generator/gentypes"
)

func HandleProperties(
	ctx *GeneratorContext,
	td *gentypes.TypeDefinition,
	component *gentypes.ComponentDefinition,
	def *openapi3.Schema) error {

	property_objects := openapi3.Schemas{}
	var err error

	for propertyName := range def.Properties {

		property := def.Properties[propertyName]

		if property.Ref == "" {
			property_objects[propertyName] = property
			continue
		}

		// Reference to other type.
		ref := ctx.resolver.ResolveComponent(&component.ID)
		if ref == nil {
			if ref, err = CreateComponentFromReference(ctx, &component.ID, property); err != nil {
				return err
			}
		}

		td.Properties = append(td.Properties, gentypes.Property{
			ComponentDefinition: *ref,
			Required:            Contains(def.Required, propertyName),
			PropertyName:        propertyName,
		})
	}

	// Handle Inline Properties
	for propertyName := range property_objects {
		property := property_objects[propertyName]

		property_id := component.ID.NewWithChangedTypeName(
			component.ID.TypeName + "_" + strcase.ToCamel(propertyName),
		)

		if property.Value.Type == "object" {
			// Create a new component for the property
			ref, err := CreateComponentFromReference(ctx, property_id, property)
			if err != nil {
				return err
			}

			td.Properties = append(td.Properties, gentypes.Property{
				ComponentDefinition: *ref,
				Required:            Contains(def.Required, propertyName),
				PropertyName:        propertyName,
			})

			continue
		}

		// Handle all other types except arrays
		if property.Value.Type != "array" {
			td.Properties = append(td.Properties, gentypes.Property{
				ComponentDefinition: gentypes.ComponentDefinition{
					ID: *property_id,
					Definition: &gentypes.TypeDefinition{
						ID:          *property_id,
						GoPackage:   td.GoPackage,
						Schema:      property.Value,
						Composition: []gentypes.Composition{},
						Properties:  []gentypes.Property{},
					},
				},
				Required:     Contains(def.Required, propertyName),
				PropertyName: propertyName,
			})

		}

		if property.Value.Items == nil {
			continue
		}

		// Array
		property_array_id := property_id.NewWithChangedTypeName(
			property_id.TypeName + "_" + strcase.ToCamel(propertyName) + "Array",
		)

		// Array is ref
		if property.Value.Items.Ref != "" {
			ref, err := CreateComponentFromReference(ctx, property_array_id, property.Value.Items)
			if err != nil {
				return err
			}

			td.Properties = append(td.Properties, gentypes.Property{
				ComponentDefinition: *ref,
				Required:            Contains(def.Required, propertyName),
				PropertyName:        propertyName,
			})

			continue
		}

		// The array is a array of object -> create as reference
		if property.Value.Type == "object" {
			ref, err := CreateComponentFromReference(ctx, property_array_id, property.Value.Items)
			if err != nil {
				return err
			}

			td.Properties = append(td.Properties, gentypes.Property{
				ComponentDefinition: *ref,
				Required:            Contains(def.Required, propertyName),
				PropertyName:        propertyName,
			})

			continue
		}

		// No object, this is a definition of a primitive type
		ref, err := CreateComponentFromDefinition(ctx, property_id, property.Value)
		if err != nil {
			return err
		}

		td.Properties = append(td.Properties, gentypes.Property{
			ComponentDefinition: *ref, // TODO: This should be a new component with a ref instead?
			Required:            Contains(def.Required, propertyName),
			PropertyName:        propertyName,
		})
	}

	return nil
}