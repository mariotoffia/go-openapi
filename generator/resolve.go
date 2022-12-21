package generator

import (
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mariotoffia/go-openapi/generator/gentypes"
)

// IsDefinition checks if the _ref_ is a definition or a reference.
func IsDefinition(ref *openapi3.SchemaRef) bool {
	return ref.Ref == ""
}

// IsReference check if the _ref_ is a reference or a definition.
func IsReference(ref *openapi3.SchemaRef) bool {
	return ref.Ref != ""
}

// ResolveReferenceAndSwitchIfNeeded will create a `ComponentReference`. If the _ref_ is under the specification root path
// then it will be used. If it is under model root path it will use that instead to create the `ComponentReference`
// as root path. If specification and module is on the same root path, the the longest path will be used as root path.
func ResolveReferenceAndSwitchIfNeeded(
	ctx *GeneratorContext,
	componentId *gentypes.ComponentReference,
	ref *openapi3.SchemaRef,
) *gentypes.ComponentReference {
	// Create the fully qualified path to the reference
	ref_path := filepath.Join(componentId.RootPath, componentId.Path, ref.Ref)
	ref_path = filepath.Clean(ref_path)

	var type_ref *gentypes.ComponentReference

	if IsSpecificationRooted(ctx, ref_path) {
		if Is_A_SpecificationRef(ctx, componentId) {
			// Specification rooted component id -> no switch
			type_ref = gentypes.FromSchemaRef(ref, ctx.settings.spec_root)
		} else {
			// Switch root to module root
			type_ref = gentypes.FromSchemaRef(ref, ctx.settings.model_root)
		}
	} else {
		if !Is_A_SpecificationRef(ctx, componentId) {
			// Module rooted component id -> no switch
			type_ref = gentypes.FromSchemaRef(ref, ctx.settings.model_root)
		} else {
			// Switch root to spec root
			type_ref = gentypes.FromSchemaRef(ref, ctx.settings.spec_root)
		}
	}

	return type_ref
}

func ResolveGoPackage(ctx *GeneratorContext, ref *gentypes.ComponentReference) string {

	if Is_A_SpecificationRef(ctx, ref) {
		return ref.ToGoPackage(ctx.settings.spec_package)
	}

	return ref.ToGoPackage(ctx.settings.model_package)
}

func Is_A_SpecificationRef(ctx *GeneratorContext, ref *gentypes.ComponentReference) bool {
	return IsSpecificationRooted(ctx, ref.RootPath)
}

func IsSpecificationRooted(ctx *GeneratorContext, path string) bool {

	if strings.HasPrefix(path, ctx.settings.model_root) {
		if strings.HasPrefix(path, ctx.settings.spec_root) {
			// Both are rooted - if longest is spec then it is spec rooted
			return len(ctx.settings.spec_root) > len(ctx.settings.model_root)
		}

		return false
	}

	return strings.HasPrefix(path, ctx.settings.spec_root)
}
