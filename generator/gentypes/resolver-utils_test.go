package gentypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveRelativeReferenceFromType(t *testing.T) {
	q := TypeDefinitionImpl{
		RootPath:    GoModFqPath(),
		Path:        "generator/testdata/allof",
		OpenAPIFile: "report.yaml",
		GoFile:      "report.go",
		GoPackage:   RenderPackage("", "./generator/testdata/allof"),
		Name:        "Report",
	}

	assert.Equal(t,
		"generator/testdata/anyof/anyof.yaml",
		ResolvePathFromTypeDefinition(&q, "../anyof/anyof.yaml"),
	)

	assert.Equal(t,
		"generator/testdata/allof/gurka/nisse.yaml",
		ResolvePathFromTypeDefinition(&q, "./gurka/nisse.yaml"),
	)

	assert.Equal(t,
		"generator/testdata/allof/gurka/nisse.yaml",
		ResolvePathFromTypeDefinition(&q, "gurka/nisse.yaml"),
	)

	assert.Equal(t,
		"generator/testdata/allof/nisse.yaml",
		ResolvePathFromTypeDefinition(&q, "nisse.yaml"),
	)
}

func TestResolveRelativeReferenceFromTypeAboveRootPathShallFail(t *testing.T) {
	q := TypeDefinitionImpl{
		RootPath:    GoModFqPath(),
		Path:        "generator/testdata/allof",
		OpenAPIFile: "report.yaml",
		GoFile:      "report.go",
		GoPackage:   RenderPackage("", "./generator/testdata/allof"),
		Name:        "Report",
	}

	assert.Panics(t, func() {
		ResolvePathFromTypeDefinition(&q, "../../../../anyof/anyof.yaml")
	})

}
