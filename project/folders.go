package project

import (
	"errors"
	"os"
	"path/filepath"
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

var folderActions = map[string]ActionFn{
	"cli": func(project, feature, option string) error {
		if err := makeFolder(project, "cmd", project); err != nil {
			return err
		}
		// TODO: use template to generate main.go file
		// ??? use mustache ???  or text/template
		return nil
	},
}
