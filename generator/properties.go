package generator

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
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
		property_id := component.ID.NewWithAppendTypeName(propertyName)

		ref := ResolveReferenceAndSwitchIfNeeded(ctx, &component.ID, property)

		if ctx.resolver.ResolveComponent(&component.ID) != nil {
			if _, err = CreateComponentFromReference(ctx, ref, property); err != nil {
				return err
			}
		}

		td.Properties = append(td.Properties, gentypes.Property{
			ComponentDefinition: gentypes.ComponentDefinition{
				ID:        *property_id,
				Reference: ref,
			},
			Required:     ContainsString(def.Required, propertyName),
			PropertyName: propertyName,
		})
	}

	// Handle Inline Properties
	for propertyName := range property_objects {
		property := property_objects[propertyName]

		property_id := component.ID.NewWithAppendTypeName(propertyName)

		if property.Value.Type == "object" {
			// Create a new component for the property
			ref, err := CreateComponentFromReference(ctx, property_id, property)
			if err != nil {
				return err
			}

			td.Properties = append(td.Properties, gentypes.Property{
				ComponentDefinition: *ref,
				Required:            ContainsString(def.Required, propertyName),
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
				Required:     ContainsString(def.Required, propertyName),
				PropertyName: propertyName,
			})

		}

		if property.Value.Items == nil {
			continue
		}

		// Array
		property_array_id := property_id.NewWithAppendTypeName("Array")
		if ctx.resolver.ResolveComponent(property_array_id) != nil {
			return fmt.Errorf("array already defined: %s", property_array_id)
		}

		ref, err := HandleArray(ctx, property_array_id, property.Value.Items)
		if err != nil {
			return err
		}

		td.Properties = append(td.Properties, gentypes.Property{
			ComponentDefinition: *ref,
			Required:            ContainsString(def.Required, propertyName),
			PropertyName:        propertyName,
		})
	}

	return nil
}
