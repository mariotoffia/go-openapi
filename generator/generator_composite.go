package generator

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mariotoffia/go-openapi/generator/gentypes"
)

func HandleComposite(
	config gentypes.CreateConfig,
	td *gentypes.TypeDefinition,
	refs openapi3.SchemaRefs) (composites []gentypes.Composition, schema *openapi3.Schema, err error) {

	var has_inline []gentypes.Composition
	var has_inline_refs []*openapi3.SchemaRef

	// Copy the schema
	scm := *td.Schema
	schema = &scm

	for i := range refs {
		composite := refs[i]
		cfg := config

		if composite.Ref != "" {
			cfg.Name = gentypes.FromSchemaRef(composite, config.RootPath).TypeName
		}

		cd, err := gentypes.CreateComposite(composite, cfg)

		if err != nil {
			panic(err)
		}

		if cd.Inline {
			has_inline = append(has_inline, *cd)
			has_inline_refs = append(has_inline_refs, composite)
		} else {
			composites = append(composites, *cd)
		}
	}

	if len(has_inline) > 0 {

		// Remove the inline schema refs
		for ref := range has_inline_refs {
			for i := range refs {
				if refs[i] == has_inline_refs[ref] {
					refs = removeSchemaRef(refs, i)
				}
			}
		}

		for i := range has_inline {
			fmt.Println(i)
			/*
				if has_inline[i].Schema.Type == "object" {

					s, merge_err := MergeSchemaObjects(schema, has_inline[i].Schema)

					if merge_err != nil {
						err = merge_err
						return
					}

					schema = s

				} else {

					if add_err := AddSchemaToObject(schema, has_inline_refs[i] , false ); err != nil {
						err = add_err
						return
					}

				}*/
		}
	}

	return
}

func removeSchemaRef(slice []*openapi3.SchemaRef, index int) []*openapi3.SchemaRef {
	return append(slice[:index], slice[index+1:]...)
}
