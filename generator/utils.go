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

// ContainsString will check if a string is in a slice of strings.
func ContainsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// RemoveSchemaRef will remove a _ref_ from the _references_ slice if found.
func RemoveSchemaRef(references openapi3.SchemaRefs, ref *openapi3.SchemaRef) openapi3.SchemaRefs {
	for j := range references {
		if references[j] == ref {
			return append(references[:j], references[j+1:]...)
		}
	}

	return references
}
