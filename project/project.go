package project

import (
	"embed"
	"fmt"
	"os"

	"github.com/robbyriverside/brief"
)

//go:embed golang/templates
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
	nodes, err := brief.Decode(file)
	if err != nil {
		proj.err = err
		return &proj
	}
	node := nodes[0]
	proj.features = node.FindNode("project")
	if proj.features == nil {
		proj.err = fmt.Errorf("brevity file %q does not contain project", specfile)
	}
	return &proj
}
