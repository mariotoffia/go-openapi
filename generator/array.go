package generator

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mariotoffia/go-openapi/generator/gentypes"
)

func HandleArray(ctx *GeneratorContext, arrayId *gentypes.ComponentReference, items *openapi3.SchemaRef) (*gentypes.ComponentDefinition, error) {
	if ctx.resolver.ResolveComponent(arrayId) != nil {
		return nil, nil
	}

	if IsReference(items) || items.Value.Type == "object" {
		// Array is ref or an array of objects -> create as reference
		ref, err := CreateComponentFromReference(ctx, arrayId, items)
		if err != nil {
			return nil, err
		}

		return ref, nil
	}

	// No object, this is a definition of a primitive type
	def, err := CreateComponentFromDefinition(ctx, arrayId, items.Value)
	if err != nil {
		return nil, err
	}

	return def, nil
}
