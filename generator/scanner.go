package generator

import (
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/gobwas/glob"
	"gopkg.in/yaml.v3"
)

type Module struct {
	Path    string
	Objects []string
}

func (m Module) ToRef(object string) string {
	return fmt.Sprintf("%s#/%s", m.Path, object)
}

// ScanForModules will scan all the modules from the path and
// use the inclusion list to filter the modules.
//
// Open the file and parse it using the yaml parser. Include all
// top level object into a `Module` object.
func ScanForModules(path string, inclusion []Include) ([]Module, error) {
	modules := []Module{}

	f := os.DirFS(path)
	for _, inc := range inclusion {
		// Scan the directory for files matching the glob expression

		err := ScanDir(
			f,
			inc.Path,
			inc.Glob,
			func(f fs.FS, path string) error {
				// Create a new module
				module := Module{
					Path:    path,
					Objects: []string{},
				}

				data, err := FSysFileToBytes(f, path)
				if err != nil {
					return err
				}

				var m map[string]any
				if err := yaml.Unmarshal(data, &m); err != nil {
					return err
				}

				for k := range m {
					module.Objects = append(module.Objects, k)
				}

				modules = append(modules, module)
				return nil
			},
		)

		if err != nil {
			return nil, err
		}

	}

	return modules, nil
}

type ScanDirCallback func(f fs.FS, path string) error

func ScanDir(
	f fs.FS,
	directory string,
	globExpr string,
	cb ScanDirCallback) error {
	// Compile the glob expressions
	var glb glob.Glob
	var glbErr error

	if glb, glbErr = glob.Compile(globExpr, '/'); glbErr != nil {
		return fmt.Errorf(
			"failed to compile glob expression: %s error: %s",
			globExpr, glbErr.Error(),
		)
	}

	if directory == "" {
		directory = "."
	}

	err := fs.WalkDir(f, directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf(
				"failed to walk directory: %s, path: %s, err: %s ",
				directory, path, err.Error(),
			)
		}

		if d.IsDir() || !glb.Match(path) {
			return nil
		}

		if err := cb(f, path); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf(
			"failed to walk the fs, directory: %s error: %s",
			directory, err.Error(),
		)
	}

	return nil
}

// FSysFileToBytes reads a file from as _fqPath_ and return the
// contents of it.
//
// If it fails, it will return an error.
func FSysFileToBytes(f fs.FS, fqPath string) ([]byte, error) {
	file, err := f.Open(fqPath)

	if err != nil {
		return nil, err
	}

	var data []byte
	data, err = io.ReadAll(file)

	return data, err
}
