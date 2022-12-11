package generator_test

import (
	"testing"

	"github.com/mariotoffia/go-openapi/generator"
	"github.com/stretchr/testify/assert"
)

func TestSimpleAllOf(t *testing.T) {
	gen := generator.NewSettings(generator.Templates{}).
		UseModelPath("./testdata/allof").
		UseOutputPath("_output").
		ToGenerator()

	err := gen.Generate()
	assert.Equal(t, nil, err)
}
