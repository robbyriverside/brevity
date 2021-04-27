package project

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/robbyriverside/nocode/internal/nocode"
)

const (
	NameFeature = "name"
	CLIFeature  = "cli"
	APIFeature  = "api"
	MockFeature = "mocks"
)

//go:embed templates
var templates embed.FS

// ActionFn project generation action
type ActionFn func(project, feature, option string) error

// Features for the project
type Features map[string]string

// Project to be generated
type Project struct {
	err      error
	features Features
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
	err := nocode.Read(specfile, &proj.features)
	if err != nil {
		proj.err = err
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
	if err := nocode.ValidateFolder(path); err != nil {
		return err
	}

	fmt.Println("generate:", path, p.features)

	return p.folders().makefile().docker().Error()
}
