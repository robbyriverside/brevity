package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/robbyriverside/brief"
)

func (p *Project) folders() *Project {
	if p.Error() != nil {
		return p
	}

	var found bool
	name := p.features.Name
	if len(name) == 0 {
		return p.Stop(errors.New("required feature: name, not found"))
	}
	for _, option := range p.features.Body {
		if action, ok := folderActions[option.Type]; ok {
			if err := action(name, option); err != nil {
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
	"cli": func(project string, option *brief.Node) (err error) {
		if err := makeFolder(project, "cmd", project); err != nil {
			return err
		}

		tmpl, err := FolderTemplates(option.Name)
		if err != nil {
			return
		}
		main, err := os.Create(filepath.Join(project, "cmd", project, "main.go"))
		if err != nil {
			return fmt.Errorf("failed creating main.go: %s", err)
		}
		defer main.Close()
		data := struct {
			Name     string
			Packages []string
		}{
			Name:     project,
			Packages: []string{"one", "two", "three"},
		}
		if err = tmpl.ExecuteTemplate(main, option.Name, &data); err != nil {
			return fmt.Errorf("cli template %s failed: %s", option, err)
		}
		return nil
	},
}
