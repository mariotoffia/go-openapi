package gentypes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ToOpenAPIReference will render a reference to _td_ type
// that is uniform and always is the same.
//
// The reference is normalized against the `RootPath()`
func ToOpenAPIReference(td *TypeDefinitionImpl) string {
	path := ToRootRelativePath(td.GetRootPath(), td.GetPath())
	file := strings.Trim(td.GetOpenAPIFile(), " ")

	return fmt.Sprintf("%s/%s#/%s", path, file, td.GetName())
}

// ToGoReference creates a reference to the _td_ type
// by constructing the go package name / type name.
func ToGoReference(td *TypeDefinitionImpl) string {
	return fmt.Sprintf("%s/%s", td.GetGoPackage(), td.GetName())
}

// ResolvePathFromTypeDefinition will resolve a path
// relative a _td_. The path may contain .. and/or sub paths.
//
// This is useful when a _$ref_ is defined in the _td_ and we need
// to resolve where it is pointing to.
//
// The returned _path_ is relative to the `RootPath()`.
//
// NOTE: The pay may or may not contain a file name.
//
// If the resulting path is above the `RootPath()` it will
// panic.
func ResolvePathFromTypeDefinition(td *TypeDefinitionImpl, path string) string {
	fq := filepath.Join(td.GetRootPath(), td.GetPath(), path)
	fq = filepath.Clean(fq)

	if !strings.HasPrefix(fq, td.GetRootPath()) {
		panic(fmt.Sprintf("path '%s' is above root path '%s'", fq, td.GetRootPath()))
	}

	path = strings.TrimPrefix(fq, td.GetRootPath())
	return TrimPath(path)
}

// TrimPath will remove any leading dot slash, space
// and trailing spaces and slash. It will also
// remove any leading slash.
func TrimPath(path string) string {
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	return strings.Trim(path, " ")
}

// TrimOpenAPIReference will trim the open api reference
// so it is possible to compare them even if they are
// expressed in different ways.
//
// For example: "./path/file.yaml#/name" is equivalent
// to "path/file.yaml#/name".
//
// This function will strip all leading dot slash, space
// and trailing spaces.
func TrimOpenAPIReference(ref string) string {
	ref = strings.TrimPrefix(ref, "./")
	ref = strings.TrimPrefix(ref, "/")
	return strings.Trim(ref, " ")
}

// ToRootRelativePath will create a path that is relative to _rootPath_
// without any leading dot slash, space and trailing spaces and
// trailing slash.
//
// It also cleans the path from any double slashes and ..
func ToRootRelativePath(rootPath, path string) string {
	path = filepath.Join(rootPath, path)
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, rootPath)
	path = TrimPath(path)
	return path
}

// Checks if the reference is an explicit reference.
//
// An explicit reference is when the reference is naming
// the component e.g. _#/components/schemas/MyType_ or
// _file.yaml#/MyType_.
func IsOpenAPIExplicitComponentReference(ref string) bool {
	return strings.Contains((ref), "#/")
}

// IsOpenAPIFileReference will check if the _ref_
// do contain a reference that contains either .yaml
// or .json.
func IsOpenAPIFileReference(ref string) bool {
	return strings.Contains(ref, ".yaml") || strings.Contains(ref, ".json")
}

// OpenAPIFileAndComponent extracts the file and component part of
// the _ref_. If internal reference the file part will be empty.
//
// NOTE: If no #/ is found it both _file_ and _component_ will be empty.
func OpenAPIFileAndComponent(ref string) (file string, component string) {
	if !IsOpenAPIExplicitComponentReference(ref) {
		return
	}

	parts := strings.Split(ref, "#/")
	if len(parts) == 1 {
		component = strings.Trim(parts[0], " ")
	} else if len(parts) == 2 {
		file = TrimPath(parts[0])
		component = strings.Trim(parts[1], " ")
	} else {
		panic("incorrect reference: " + ref)
	}

	return
}

// GoModFqPath will return the root path of the project (go.mod).
//
// If not found an empty string is returned. If _cwd_ is
// empty it will use current working directory.
func GoModFqPath(cwd ...string) string {
	// get current path
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if len(cwd) > 0 {
		path = filepath.Join(cwd...)
	}

	// iterate backwards and look for go.mod
	for {
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			return path
		}

		if path == "/" {
			return ""
		}

		path = filepath.Join(path, "..")
	}
}

// RenderPackage renders a full package name from a module package
// and a path.
func RenderPackage(modulePackage, path string) string {
	if modulePackage == "" {
		modulePackage = GoModPackage("")
	}

	return fmt.Sprintf("%s/%s", modulePackage, TrimPath(path))
}

// GoModPackage will load the _go.mod_ file and extract the
// module name.
//
// If module is not found an empty string is returned.
func GoModPackage(mod string) string {
	if mod == "" {
		mod = GoModFqPath()
	}

	data, err := os.ReadFile(filepath.Join(mod, "go.mod"))
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module") {
			return strings.Trim(strings.TrimPrefix(line, "module"), " \t")
		}
	}
	return ""
}
