package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mariotoffia/go-openapi/generator"
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

	err := gen.Generate(generator.GeneratorContext{})
	assert.Equal(t, nil, err)
}
