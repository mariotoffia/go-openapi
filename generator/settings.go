package generator

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type Include struct {
	Path string
	Glob string
}
type Settings struct {
	loader    *openapi3.Loader
	models    string
	output    string
	spec      string
	inclusion []Include
	templates Templates
}

func NewSettings(templates Templates) *Settings {
	return &Settings{
		templates: templates,
	}
}

// ToGenerator creates a generator from the settings
func (sett *Settings) ToGenerator() *Generator {
	return NewGenerator(*sett)
}

// UseModelPath sets the base path where all models are resolved
// relative to.
func (sett *Settings) UseModelPath(models string) *Settings {
	sett.models = models
	return sett
}

// UseOutputPath sets the base path where all generated files will be written.
func (sett *Settings) UseOutputPath(output string) *Settings {
	sett.output = output
	return sett
}

// UseSpec sets the path where a open api spec file is located.
//
// If none is provided, a default one will be created.
//
// NOTE: If model scanning is performed, those will be added to
// the spec _components.schema_ section.
func (sett *Settings) UseSpec(spec string) *Settings {
	sett.spec = spec
	return sett
}

// Include adds a list of paths to include in the generation.
//
// NOTE: Those must be relative to the module path.
//
// It is possible to use glob expressions by adding : and the glob expression
// For example: 'mypkg:{a,b}-model.yaml' will add path 'mypkg' and all files
// matching the glob expression.
func (sett *Settings) Include(inclusion ...string) *Settings {
	// Add all the paths to the inclusion list
	for _, inc := range inclusion {
		idx := strings.Index(inc, ":")
		if idx == -1 {
			sett.inclusion = append(sett.inclusion, Include{Path: inc})
		} else {
			sett.inclusion = append(sett.inclusion, Include{
				Path: inc[:idx],
				Glob: inc[idx+1:],
			})
		}
	}

	return sett
}

// UseLoader will override the default loader used to load the OpenAPI
func (sett *Settings) UseLoader(loader *openapi3.Loader) *Settings {
	sett.loader = loader
	return sett
}
