package generator_test

import (
	"encoding/json"
	"fmt"
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

	ctx := generator.GeneratorContext{}
	err := gen.Generate(&ctx)
	assert.Equal(t, nil, err)

	if data, err := json.Marshal(ctx.GetSpecification()); err != nil {
		t.Error(err)
	} else {
		fmt.Println(string(data))
	}
}
