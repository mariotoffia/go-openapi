package generator_test

import (
	"testing"

	"github.com/mariotoffia/go-openapi/generator"
	"github.com/stretchr/testify/assert"
)

func TestSimpleAllOf(t *testing.T) {

	// https://swagger.io/docs/specification/data-models/inheritance-and-polymorphism/
	// https://github.com/getkin/kin-openapi
	// https://github.com/deepmap/oapi-codegen
	gen := generator.NewSettings(generator.Templates{}).
		UseModelPath(
			"./testdata/allof",
			"github.com/mariotoffia/go-openapi/generator/testdata/allof",
		).
		UseOutputPath("_output").
		ToGenerator()

	err := gen.Generate(generator.GeneratorContext{})
	assert.Equal(t, nil, err)
}
