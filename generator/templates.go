package generator

import (
	"embed"
	"html/template"
	"io/fs"
	"strings"
)

type WellKnownTemplates string

const (
	// TemplateIndex is the name of the main open api 3 spec file
	// to use when generating the open api 3 spec or models.
	//
	// This is when a spec do not exist and models are generated,
	// otherwise a user specified spec is needed.
	TemplateIndex WellKnownTemplates = "index.yaml"
)

//go:embed templates
var templates embed.FS

type Templates struct {
	templates fs.FS
}

func (tpl *Templates) UseTemplates(templates fs.FS) *Templates {
	tpl.templates = templates
	return tpl
}

// GetTemplate will return a template from the template folder.
//
// NOTE: It will scan the user provided first if such exist before checking
// the embedded template folder.
func (tpl *Templates) GetTemplate(fqName string) (*template.Template, error) {
	if !strings.HasPrefix(fqName, "templates/") {
		fqName = "templates/" + fqName
	}

	t, err := template.ParseFS(tpl.templates, fqName)
	if err == nil {
		return t, nil
	}

	return template.ParseFS(templates, fqName)
}

// GetFile will check if it exists in user provided templates folder
// or in the embedded templates folder.
func (tpl *Templates) GetFileAsString(fqPath string) (string, error) {
	data, err := templates.ReadFile("templates/index.yaml")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
