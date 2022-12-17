package generator

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func ContainsInterface(slice []any, s any) bool {
	for i := range slice {
		if s == slice[i] {
			return true
		}
	}

	return false
}

func ContainsSchemaRef(slice openapi3.SchemaRefs, ref *openapi3.SchemaRef) bool {
	for i := range slice {
		if ref.Ref == slice[i].Ref {
			return true
		}
	}

	return false
}

func MergeStrings(to, from string) string {

	if to == "" {
		return from
	}

	if from == "" {
		return to
	}

	return fmt.Sprintf("%s\n\n%s", to, from)

}
