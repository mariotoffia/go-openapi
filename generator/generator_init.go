package generator

import (
	"os"
	"path"
	"strings"

	"github.com/mariotoffia/go-openapi/generator/gentypes"
	"gopkg.in/yaml.v3"
)

// prepareSpecificationFile will ensure that a spec file exist.
//
// If none has been provided by user, a temporary one will be created.
//
// Then all scanned modules (if any inclusion in settings) are added to
// the spec file to make sure that all models are included.
func prepareSpecificationFile(ctx *GeneratorContext) (string, error) {
	var tempSpecFile string

	// Add models to spec
	if len(ctx.settings.inclusion) > 0 {
		// Scan for modules
		modules, err := ScanForModules(ctx.settings.model_root, ctx.settings.inclusion)
		if err != nil {
			return "", err
		}

		var m map[string]any

		// No specification provided, create a temporary one
		if ctx.settings.spec == "" {
			tempSpecFile = path.Join(ctx.settings.model_root, "__go-openapi.yaml")
			ctx.settings.spec = tempSpecFile

			// Get default index.yaml
			index, err := ctx.settings.templates.GetFileAsString(string(TemplateIndex))
			if err != nil {
				return "", err
			}

			// Write index.yaml
			if err = os.WriteFile(tempSpecFile, []byte(index), 0644); err != nil {
				return "", err
			}

			if err := yaml.Unmarshal([]byte(index), &m); err != nil {
				return "", err
			}
		} else {
			// User set a existing spec -> use it
			data, err := os.ReadFile(ctx.settings.spec)
			if err != nil {
				return "", err
			}

			if err := yaml.Unmarshal(data, &m); err != nil {
				return "", err
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
							"$ref": gentypes.FromRefString(module.ToRef(object), ctx.settings.model_root).ToOpenAPI("yaml"),
							"type": "object",
						}
					}
				}
			}

			// Write spec
			data, err := yaml.Marshal(m)
			if err != nil {
				return "", err
			}

			if err = os.WriteFile(ctx.settings.spec, data, 0644); err != nil {
				return "", err
			}
		}
	}

	return tempSpecFile, nil
}
