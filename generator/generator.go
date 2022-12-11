package generator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

type Generator struct {
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

func (gen *Generator) Generate() error {
	var tempSpecFile string

	// Add models to spec
	if len(gen.settings.inclusion) > 0 {
		// Scan for modules
		modules, err := ScanForModules(gen.settings.models, gen.settings.inclusion)
		if err != nil {
			return err
		}

		var m map[string]any

		// No specification provided, create a temporary one
		if gen.settings.spec == "" {
			tempSpecFile = path.Join(gen.settings.models, "__go-openapi.yaml")
			gen.settings.spec = tempSpecFile

			// Get default index.yaml
			index, err := gen.settings.templates.GetFileAsString(string(TemplateIndex))
			if err != nil {
				return err
			}

			// Write index.yaml
			if err = os.WriteFile(tempSpecFile, []byte(index), 0644); err != nil {
				return err
			}

			if err := yaml.Unmarshal([]byte(index), &m); err != nil {
				return err
			}
		} else {
			// User set a existing spec -> use it
			data, err := os.ReadFile(gen.settings.spec)
			if err != nil {
				return err
			}

			if err := yaml.Unmarshal(data, &m); err != nil {
				return err
			}
		}

		if len(modules) > 0 {
			// Add modules to spec
			schemas := m["components"].(map[string]interface{})["schemas"].(map[string]interface{})
			for _, module := range modules {

				if strings.Contains(module.Path, "__go-openapi.yaml") {
					continue
				}

				for _, object := range module.Objects {
					if _, ok := schemas[object]; !ok {
						schemas[object] = map[string]interface{}{
							"$ref": module.ToRef(object),
							"type": "object",
						}
					}
				}
			}

			// Write spec
			data, err := yaml.Marshal(m)
			if err != nil {
				return err
			}

			if err = os.WriteFile(gen.settings.spec, data, 0644); err != nil {
				return err
			}
		}
	}

	// Remove temp spec file
	defer func() {
		if tempSpecFile != "" {
			os.Remove(tempSpecFile)
		}
	}()

	// Load the specification
	doc, err := gen.settings.loader.LoadFromFile(gen.settings.spec)

	if err != nil {
		return err
	}

	err = doc.Validate(gen.settings.loader.Context)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	DumpSchemas(&buffer, doc.Components.Schemas)

	fmt.Println(buffer.String())

	return nil
}

func DumpSchemas(buffer *bytes.Buffer, schemas map[string]*openapi3.SchemaRef) {
	for k, v := range schemas {
		fmt.Fprintln(buffer, "schema: ", k)
		DumpSchemaRef(buffer, v)
	}
}
func DumpValue(buffer *bytes.Buffer, v *openapi3.Schema) {
	if v == nil {
		fmt.Println("nil")
		return
	}

	data, _ := v.MarshalJSON()
	fmt.Fprintln(buffer, string(data))
}

func DumpSchemaRef(buffer *bytes.Buffer, v *openapi3.SchemaRef) {
	if v.Ref != "" {
		fmt.Fprintln(buffer, "ref: ", v.Ref)
	}
	DumpSchemaRefs(buffer, v.Value.AllOf)
	DumpSchemaRefs(buffer, v.Value.AnyOf)
	DumpSchemaRefs(buffer, v.Value.OneOf)
	DumpValue(buffer, v.Value)
}

func DumpSchemaRefs(buffer *bytes.Buffer, refs openapi3.SchemaRefs) {
	if len(refs) == 0 {
		return
	}

	for _, v := range refs {
		if v.Value.Discriminator != nil && len(v.Value.Discriminator.Mapping) > 0 {
			fmt.Fprintln(buffer, "discriminator: ", v.Value.Discriminator.PropertyName)
			for k, v := range v.Value.Discriminator.Mapping {
				fmt.Fprintln(buffer, "mapping: ", k, " -> ", v)
			}
		}

		DumpSchemaRef(buffer, v)
	}
}
