package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
	"github.com/mariotoffia/go-openapi/generator/gentypes"
)

// AddSchemaToObject will add a schema to an object schema.
func AddSchemaToObject(to *openapi3.Schema, from *openapi3.SchemaRef, required bool) error {

	if to.Type != "object" {
		return fmt.Errorf("cannot add schema to non-object schema")
	}

	if from.Value.Type == "object" {
		return fmt.Errorf("cannot add object schema to object schema")
	}

	if to.Properties == nil {
		to.Properties = openapi3.Schemas{
			from.Value.Title: from,
		}
	} else {
		if _, ok := to.Properties[from.Value.Title]; ok {
			return fmt.Errorf("cannot add schema to object schema, already exists")
		}

		to.Properties[from.Value.Title] = from
	}

	if required {

		if to.Required == nil {
			to.Required = []string{}
		}
	}

	return nil
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
		merged.OneOf = append(merged.OneOf, from.OneOf...)
	}

	if len(from.AllOf) > 0 {
		merged.AllOf = append(merged.AllOf, from.AllOf...)
	}

	if len(from.AnyOf) > 0 {
		merged.AnyOf = append(merged.AnyOf, from.AnyOf...)
	}

	if from.Not != nil {
		m, err := MergeSchemaRef(ctx, merged.Not, from.Not)
		if err != nil {
			return nil, err
		}

		merged.Not = m
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

	merged.Required = append(merged.Required, from.Required...)

	if len(from.Properties) > 0 {
		if len(to.Properties) == 0 {
			merged.Properties = from.Properties
		} else {

			for k := range from.Properties {

				if _, ok := to.Properties[k]; ok {
					m, err := MergeSchemaRef(ctx, to.Properties[k], from.Properties[k])

					if err != nil {
						return nil, err
					}

					merged.Properties[k] = m

				} else {
					merged.Properties[k] = from.Properties[k]
				}
			}
		}
	}

	return &merged, nil

}

func MergeStrings(to, from string) string {

	if to == "" {
		return from
	}

	if from == "" {
		return to
	}

	return fmt.Sprintf("%s\n\n%s", to, from)

}
func MergeSchemaRef(ctx *GeneratorContext, to *openapi3.SchemaRef, from *openapi3.SchemaRef) (*openapi3.SchemaRef, error) {

	if from == nil && to != nil {
		return to, nil
	}

	if to == nil && from != nil {
		return from, nil
	}

	if to == nil && from == nil {
		return nil, nil
	}

	to_ref := gentypes.FromSchemaRef(to, ctx.settings.models)
	from_ref := gentypes.FromSchemaRef(from, ctx.settings.models)
	if to_ref.Equal(from_ref) {
		return to, nil
	}

	to_file := to_ref.Module
	if to_ref.Module != from_ref.Module {
		to_file = filepath.Join(RemoveExtensionOnFile(from_ref.Module), to_ref.Module)
	}

	to_component := to_ref.TypeName
	if to_component != from_ref.TypeName {
		to_component = strcase.ToCamel(to_ref.TypeName + to_component)
	}

	merge, err := MergeSchemaObjects(ctx, to.Value, from.Value)
	if err != nil {
		return nil, err
	}

	return &openapi3.SchemaRef{
		Ref:   fmt.Sprintf("%s#/%s", to_file, to_component),
		Value: merge,
	}, nil

}

func RemoveExtensionOnFile(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}
