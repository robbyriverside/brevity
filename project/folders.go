package project

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

func (p *Project) folders() *Project {
	if p.Error() != nil {
		return p
	}

	var found bool
	name, ok := p.features[NameFeature]
	if !ok {
		return p.Stop(errors.New("required feature: name, not found"))
	}
	for key, option := range p.features {
		if key == NameFeature {
			continue
		}
		if action, ok := folderActions[key]; ok {
			if err := action(name, key, option); err != nil {
				return p.Stop(err)
			}
			found = true
		}
	}
	if !found {
		if err := makeFolder(name); err != nil {
			return p.Stop(err)
		}
	}

	return p
}

func makeFolder(names ...string) error {
	return os.MkdirAll(filepath.Join(names...), os.ModePerm)
}

const (
	rootTemplate = "templates"
)

var (
	folderTemplates = map[string]*template.Template{}
)

// FolderTemplates created lazy by reading embed filesystem
func FolderTemplates(name string) (tmpl *template.Template, err error) {
	tmpl, ok := folderTemplates[name]
	if !ok {
		tmpl, err = template.ParseFS(templates, filepath.Join(rootTemplate, name, "*.tmpl"))
		if err == nil {
			folderTemplates[name] = tmpl
		}
	}
	return
}

var folderActions = map[string]ActionFn{
	"cli": func(project, feature, option string) (err error) {
		if err := makeFolder(project, "cmd", project); err != nil {
			return err
		}

		tmpl, err := FolderTemplates(feature)
		if err != nil {
			return
		}
		main, err := os.Create(filepath.Join(project, "cmd", project, "main.go"))
		if err != nil {
			return fmt.Errorf("failed creating main.go: %s", err)
		}
		defer main.Close()
		data := struct {
			Name string
		}{
			Name: project,
		}
		if err = tmpl.ExecuteTemplate(main, "go-flags", &data); err != nil {
			return fmt.Errorf("cli template failed: %s", err)
		}
		return nil
	},
}
