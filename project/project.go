package project

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robbyriverside/brevity/internal/brevity"
	"github.com/robbyriverside/brief"
)

//go:embed templates
var templates embed.FS

// Dir of embedded templates
func Dir(dir string) (res []string, err error) {
	res = []string{}
	entries, err := templates.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		res = append(res, entry.Name())
	}
	return
}

// ActionFn project generation action
type ActionFn func(project string, option *brief.Node) error

// Project to be generated
type Project struct {
	err      error
	features *brief.Node
}

// Error found in the project
func (p *Project) Error() error {
	return p.err
}

// Stop the project by adding an error
func (p *Project) Stop(err error) *Project {
	p.err = err
	return p
}

func (p *Project) String() string {
	return fmt.Sprint(p.features)
}

// Read project from specfile
func Read(specfile string) *Project {
	var proj Project
	file, err := os.Open(specfile)
	if err != nil {
		proj.err = err
		return &proj
	}
	node, err := brief.Decode(file)
	if err != nil {
		proj.err = err
		return &proj
	}
	proj.features = node.FindNode("project")
	if proj.features == nil {
		proj.err = fmt.Errorf("brevity file %q does not contain project", specfile)
	}
	return &proj
}

// Generate a project into a destination folder
func (p *Project) Generate(dest string) error {
	if p.Error() != nil {
		return p.Error()
	}
	path, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("expanding destination %s failed: %s", dest, err)
	}
	if err := brevity.ValidateFolder(path); err != nil {
		return err
	}

	fmt.Println("generate:", path, p.features)

	return p.folders().makefile().docker().Error()
}
