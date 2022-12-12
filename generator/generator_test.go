package generator_test

import (
	"fmt"
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
	gen := generator.NewSettings(generator.Templates{}).
		UseModelPath(
			"./testdata/allof",
			"github.com/mariotoffia/go-openapi/generator/testdata/allof",
		).
		UseOutputPath("_output").
		ToGenerator()

	err := gen.Generate()
	assert.Equal(t, nil, err)
}

func TestResolvePath(t *testing.T) {
	root := gentypes.GoModFqPath()
	path := "../anyof"

	fmt.Println(filepath.Join(root, path))
}
