package generatortest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mariotoffia/go-openapi/generator"
	"github.com/mariotoffia/go-openapi/generator/gentypes"
	"github.com/stretchr/testify/assert"
)

func TestSimpleAllOf(t *testing.T) {

	// https://swagger.io/docs/specification/data-models/inheritance-and-polymorphism/
	// https://github.com/getkin/kin-openapi
	// https://github.com/deepmap/oapi-codegen
	cwd, _ := os.Getwd()
	gen := generator.NewSettings(generator.Templates{}).
		UseModelPath(
			filepath.Join(cwd, "./testdata/allof"),
			"github.com/mariotoffia/go-openapi/generator/testdata/allof",
		).
		UseOutputPath("_output").
		ToGenerator()

	ctx := generator.GeneratorContext{}
	err := gen.Generate(&ctx)
	assert.Equal(t, nil, err)

	spec := ctx.GetSpecification()

	m := map[string]gentypes.TypeDefinition{}

	for k, v := range spec.Components {
		if v.Definition != nil {
			m[k] = *v.Definition
		}

		if v.Reference != nil {
			component := ctx.ResolveTypeDefinition(v.Reference)
			if component != nil {
				m[k] = *component
			}
		}
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(string(data))
}
