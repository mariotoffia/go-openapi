package gentypes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ComponentReference is a reference to a `ComponentDefinition`.
type ComponentReference struct {
	// TypeName returns the name of the component type.
	TypeName string
	// NameSpace is the nesting of a type e.g. _/components/schemas/MyType_
	// will have a namespace of _components/schemas_ and the
	// Type will be _MyType_.
	NameSpace string
	// Module is the filename of the file excluding the extension.
	Module string
	// Path returns the relative path from the _rootPath_
	Path string
	// The root path where the _Path_ is relative to. RootPath
	// is *always* a rooted absolute path and never a relative path!
	RootPath string
}

// NewComponentReference creates a new reference and cleans up the parameters
// to allow for comparison even if they differ slightly such in whitespace
// or e.g. './'.
//
// Note that the root path must always be absolute path and never relative.
//
// If the _path_ is absolute or the _rootPath_ is not absolute it will panic.
// It will also panic if the resulting path is "above" _rootPath_.
//
// NOTE: It will automatically make sure that the _path_ is relative to the
// _rootPath_ and do not contain any navigational elements such as `..`.
//
// _typeName_ may also include slashes to indicate a nesting in a namespace.
func NewComponentReference(typeName, module, path, rootPath string) *ComponentReference {

	if strings.Trim(rootPath, " ") == "" {
		panic("rootPath must not be empty")
	}

	if !filepath.IsAbs(rootPath) {
		panic(fmt.Sprintf("rootPath must be an absolute path: %s", rootPath))
	}

	if filepath.IsAbs(path) {
		panic(fmt.Sprintf("path must be a relative path: %s", path))
	}

	fq := filepath.Join(rootPath, path)
	fq = filepath.Clean(fq)

	if !strings.HasPrefix(fq, rootPath) {
		panic(fmt.Sprintf("path '%s' is above root path '%s'", fq, rootPath))
	}

	var namespace string
	if strings.Contains(typeName, "/") {
		namespace = filepath.Dir(typeName)
		typeName = filepath.Base(typeName)
	}

	return &ComponentReference{
		TypeName:  strings.Trim(typeName, " "),
		NameSpace: TrimPath(namespace),
		Module:    TrimPath(RemoveExtensionOnFile(module)),
		Path:      TrimPath(strings.TrimPrefix(fq, rootPath)),
		RootPath:  strings.TrimSuffix(rootPath, "/ "),
	}
}

// HasType is set to `true` when the Type is set.
// IsEmpty returns `true` if the reference is empty.
func (tr *ComponentReference) IsEmpty() bool {
	return tr.TypeName == "" && tr.Module == "" &&
		tr.Path == "" && tr.RootPath == "" &&
		tr.NameSpace == ""
}

// String creates a full representation of the reference.
func (tr *ComponentReference) String() string {
	return fmt.Sprintf(
		"%s#/%s",
		filepath.Join(tr.RootPath, tr.Path, tr.Module),
		filepath.Join(tr.NameSpace, tr.TypeName),
	)
}

// ToOpenAPI creates a open api specification compatible reference.
func (tr *ComponentReference) ToOpenAPI(ext string) string {

	return fmt.Sprintf(
		"%s.%s#/%s", filepath.Join(tr.Path, tr.Module),
		strings.TrimPrefix(ext, "."),
		filepath.Join(tr.NameSpace, tr.TypeName),
	)
}

// RelativeModulePath returns the relative module path from the root path.
func (tr *ComponentReference) RelativeModulePath() string {
	return filepath.Join(tr.Path, tr.Module)
}

// ToGoPackage will produce the modules go package name based on the
// _basePackage_ and the _Path_.
func (tr *ComponentReference) ToGoPackage(basePackage string) string {
	return strings.ToLower(filepath.Join(basePackage, tr.Path))
}

// ToFqFilePath will return the fully qualified path to the file
// that contains the type.
//
// NOTE: The _ext_ will create the complete filename since `TypeReference`
// only holds the module name and not the physical file name.
func (tr *ComponentReference) ToFqFilePath(ext string) string {
	return filepath.Join(
		tr.RootPath,
		tr.Path,
		fmt.Sprintf("%s.%s", tr.Module, strings.TrimPrefix(ext, ".")))
}

// ToRelativeFilePath will return the fully relative path to the file
// that contains the type.
func (tr *ComponentReference) ToRelativeFilePath(ext string) string {
	return filepath.Join(
		tr.Path,
		fmt.Sprintf("%s.%s", tr.Module, strings.TrimPrefix(ext, ".")),
	)
}

func (tr *ComponentReference) IsOnRootPath(rootPath string) bool {
	return tr.RootPath == rootPath
}

func (tr *ComponentReference) Equal(other *ComponentReference) bool {
	if tr == nil && other == nil {
		return true
	}

	if tr != nil || other == nil {
		return false
	}

	if tr == nil || other != nil {
		return false
	}

	return tr.TypeName == other.TypeName &&
		tr.Module == other.Module &&
		tr.Path == other.Path &&
		tr.RootPath == other.RootPath &&
		tr.NameSpace == other.NameSpace
}

func FromRefString(ref, rootPath string) *ComponentReference {
	if !strings.Contains(ref, "#/") {
		panic(fmt.Sprintf("'%s' is not a valid openapi reference", ref))
	}

	parts := strings.Split(ref, "#/")
	if len(parts) == 2 {
		return NewComponentReference(
			strings.Trim(parts[1], " "),
			filepath.Base(parts[0]),
			filepath.Dir(parts[0]),
			rootPath,
		)
	}

	panic(fmt.Sprintf("'%s' is not a valid openapi reference", ref))
}

// ToReference will create a `TypeReference` from a `openapi3.SchemaRef`.
func FromSchemaRef(ref *openapi3.SchemaRef, rootPath string) *ComponentReference {
	if ref == nil || ref.Ref == "" {
		panic("ref is nil or empty")
	}

	return FromRefString(ref.Ref, rootPath)
}

// FromTypeDefinition will render a reference to _td_ type
// that is uniform and always is the same.
//
// The reference path is relative to the _td_, use `ToRootRelativePath`
// to get a path relative to the root.
//
// If _relPath_ is set it will resolve the _Path_ relative to the _td_
// and *then* make it relative to the _rootPath_. If omitted, the _Path_
// will be the _TypeDefinition.Path_.
//
// NOTE: When _relPath_ do contain a file it will be used instead of the _TypeDefinition_ file.
func FromTypeDefinition(td *TypeDefinition, relPath ...string) *ComponentReference {
	if len(relPath) == 0 {
		return NewComponentReference(td.ID.TypeName, td.ID.Module, td.ID.Path, td.ID.RootPath)
	}

	path := filepath.Join(relPath...)
	fq := filepath.Join(td.ID.RootPath, td.ID.Path, path)
	fq = filepath.Clean(fq)

	if !strings.HasPrefix(fq, td.ID.RootPath) {
		panic(fmt.Sprintf("path '%s' is above root path '%s'", fq, td.ID.RootPath))
	}

	path = TrimPath(strings.TrimPrefix(fq, td.ID.RootPath))

	if filepath.Ext(path) == "" {
		return NewComponentReference(td.ID.TypeName, td.ID.Module, path, td.ID.RootPath)
	}

	return NewComponentReference(
		td.ID.TypeName,
		RemoveExtensionOnFile(filepath.Base(path)),
		filepath.Dir(path),
		td.ID.RootPath,
	)
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

// GoPackage renders a full package name from a module package
// and a path.
func GoPackage(modulePackage, path string) string {
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

func RemoveExtensionOnFile(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}
