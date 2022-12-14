package gentypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveRelativeReferenceFromType(t *testing.T) {
	id := NewComponentReference("Report", "report", "generator/testdata/allof", GoModFqPath())
	q := TypeDefinition{
		ID:        *id,
		GoPackage: GoPackage("", "./generator/testdata/allof"),
	}

	assert.Equal(t,
		"generator/testdata/anyof/anyof",
		FromTypeDefinition(&q, "../anyof/anyof.yaml").RelativeModulePath(),
	)

	assert.Equal(t,
		"generator/testdata/allof/gurka/nisse",
		FromTypeDefinition(&q, "./gurka/nisse.yaml").RelativeModulePath(),
	)

	assert.Equal(t,
		"generator/testdata/allof/gurka/nisse",
		FromTypeDefinition(&q, "gurka/nisse.yaml").RelativeModulePath(),
	)

	assert.Equal(t,
		"generator/testdata/allof/nisse",
		FromTypeDefinition(&q, "nisse.yaml").RelativeModulePath(),
	)
}

func TestResolveRelativeReferenceFromTypeAboveRootPathShallFail(t *testing.T) {
	id := NewComponentReference("Report", "report", "generator/testdata/allof", GoModFqPath())
	q := TypeDefinition{
		ID:        *id,
		GoPackage: GoPackage("", "./generator/testdata/allof"),
	}

	assert.Panics(t, func() {
		FromTypeDefinition(&q, "../../../../anyof/anyof.yaml")
	})

}
