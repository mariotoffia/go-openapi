package generator

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// AddSchemaToObject will add a schema to an object schema.
type ExtendConfig struct {
	// MergeDocs is set to `true` it will add documentation onto each other (if _extending_ contains description).
	MergeDocs bool
	// MergeEnum indicates wether a enum should be replaced or merge with _target_.
	//
	// NOTE: It will use the equality Operator to determine if elements are equal or not.
	MergeEnum bool
	// MergeExtensions will stipulate if the extending schema will completely overwrite or merge its
	// extensions onto target schema.
	MergeExtensions bool
}

// ExtendSchema will extend the _target_ schema with the _extending_ and the _target_ schema is returned.
//
// CAUTION: This will not take into account that the _extending_ schema type is not set. If set, it will make
// sure that it matches the _target_ schema type.
//
// Extend Support:
//
// * Description
// * ExternalDocs
// * Example
// * Enum
// * Extensions
// * Default
// * Pattern (note it will not affect the `openapi3.Schema.compiledPattern` !!!)
func ExtendSchema(
	ctx *GeneratorContext,
	target openapi3.Schema,
	extending *openapi3.Schema, config *ExtendConfig) (*openapi3.Schema, error) {

	if target.Type == "" {
		return nil, fmt.Errorf("target schema type is not set")
	}

	if extending.Type != "" && (extending.Type != target.Type) {
		return nil, fmt.Errorf("when extending schema has a type, it must conform to target type")
	}

	if config == nil {
		config = &ExtendConfig{
			MergeDocs: false,
		}
	}

	if len(extending.Description) > 0 {
		if config.MergeDocs {
			target.Description = MergeStrings(target.Description, extending.Description)
		} else {
			target.Description = extending.Description
		}
	}

	if extending.Example != nil {
		target.Example = extending.Example
	}

	if len(extending.Enum) > 0 {
		if config.MergeEnum && len(target.Enum) > 0 {

			var add []any
			for i := range extending.Enum {
				if !ContainsInterface(target.Enum, extending.Enum[i]) {
					add = append(add, extending.Enum[i])
				}
			}

			if len(add) > 0 {
				target.Enum = append(target.Enum, add...)
			}

		} else {
			target.Enum = extending.Enum
		}
	}

	if len(extending.Extensions) > 0 {
		if config.MergeExtensions {

			if target.Extensions == nil {
				target.Extensions = map[string]any{}
			}

			for k := range extending.Extensions {
				target.Extensions[k] = extending.Extensions[k]
			}

		} else {
			target.Extensions = extending.Extensions
		}
	}

	if extending.Default != nil {
		target.Default = extending.Default
	}

	if extending.ExternalDocs != nil {
		target.ExternalDocs = extending.ExternalDocs
	}

	if len(extending.Pattern) > 0 {
		target.Pattern = extending.Pattern
	}

	return &target, nil
}

// MergeSchemaObjects will merge two schema of type object into one schema.
//
// CAUTION: It will not merge a object schema with a non-object schema.
func MergeSchemaObjects(ctx *GeneratorContext, to *openapi3.Schema, from *openapi3.Schema) (*openapi3.Schema, error) {

	merged := *to

	if from.Type != "object" && to.Type != "object" {
		return nil, fmt.Errorf("cannot merge schemas that are not type of object")
	}

	if len(from.OneOf) > 0 {
		for i := range from.OneOf {
			if !ContainsSchemaRef(to.OneOf, from.OneOf[i]) {
				merged.OneOf = append(merged.OneOf, from.OneOf[i])
			}
		}
	}

	if len(from.AllOf) > 0 {
		for i := range from.AllOf {
			if !ContainsSchemaRef(to.AllOf, from.AllOf[i]) {
				merged.AllOf = append(merged.AllOf, from.AllOf[i])
			}
		}
	}

	if len(from.AnyOf) > 0 {
		for i := range from.AnyOf {
			if !ContainsSchemaRef(to.AnyOf, from.AnyOf[i]) {
				merged.AnyOf = append(merged.AnyOf, from.AnyOf[i])
			}
		}
	}

	if from.Not != nil {
		if merged.Not == nil {
			// Premiere the to Not and drop the from Not
			merged.Not = from.Not
		}
	}

	merged.Title = MergeStrings(to.Title, from.Title)
	merged.Description = MergeStrings(to.Description, from.Description)
	merged.Enum = append(merged.Enum, from.Enum...)

	if to.Default == nil {
		merged.Default = from.Default
	}

	if to.Example == nil {
		merged.Example = from.Example
	}

	if to.ExternalDocs == nil {
		merged.ExternalDocs = from.ExternalDocs
	}

	if !to.UniqueItems {
		merged.UniqueItems = from.UniqueItems
	}

	if !to.ExclusiveMin {
		merged.ExclusiveMin = from.ExclusiveMin
	}

	if !to.ExclusiveMax {
		merged.ExclusiveMax = from.ExclusiveMax
	}

	if !to.Nullable {
		merged.Nullable = from.Nullable
	}

	if !to.ReadOnly {
		merged.ReadOnly = from.ReadOnly
	}

	if !to.WriteOnly {
		merged.WriteOnly = from.WriteOnly
	}

	if !to.AllowEmptyValue {
		merged.AllowEmptyValue = from.AllowEmptyValue
	}

	if !to.Deprecated {
		merged.Deprecated = from.Deprecated
	}

	if to.XML == nil {
		merged.XML = from.XML
	}

	if to.Min == nil {
		merged.Min = from.Min
	}

	if to.Max == nil {
		merged.Max = from.Max
	}

	if to.MultipleOf == nil {
		merged.MultipleOf = from.MultipleOf
	}

	if to.MinLength == 0 {
		merged.MinLength = from.MinLength
	}

	if to.MaxLength == nil {
		merged.MaxLength = from.MaxLength
	}

	if to.Pattern == "" {
		merged.Pattern = from.Pattern
	}

	if to.MinItems == 0 {
		merged.MinItems = from.MinItems
	}

	if to.MaxItems == nil {
		merged.MaxItems = from.MaxItems
	}

	if to.Items == nil {
		merged.Items = from.Items
	}

	merged.MinProps += from.MinProps

	if from.MaxProps != nil {
		if merged.MaxProps == nil {
			merged.MaxProps = from.MaxProps
		} else {
			*merged.MaxProps += *from.MaxProps
		}
	} else {
		merged.MaxProps = from.MaxProps
	}

	if to.AdditionalPropertiesAllowed == nil {
		merged.AdditionalPropertiesAllowed = from.AdditionalPropertiesAllowed
		merged.AdditionalProperties = from.AdditionalProperties
		merged.Discriminator = from.Discriminator
	}

	if len(from.Required) > 0 {
		if len(merged.Required) > 0 {
			for i := range from.Required {
				if !ContainsString(merged.Required, from.Required[i]) {
					merged.Required = append(merged.Required, from.Required[i])
				}
			}
		} else {
			merged.Required = from.Required
		}
	}

	if len(from.Properties) > 0 {
		if len(to.Properties) == 0 {
			merged.Properties = from.Properties
		} else {

			for k := range from.Properties {

				if _, ok := to.Properties[k]; !ok {
					// Premiere the to property and drop from property
					merged.Properties[k] = from.Properties[k]
				}
			}
		}
	}

	return &merged, nil

}
