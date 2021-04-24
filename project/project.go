package project

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/robbyriverside/nocode"
)

// Project to be generated
type Project struct {
	Name     string   `json:"name"`
	CLI      string   `json:"cli"`
	Makefile []string `json:"makefile"`
}

// MustRead project from specfile and exit on error
func MustRead(specfile string) *Project {
	var proj Project
	err := nocode.Read(specfile, &proj)
	if err != nil {
		log.Fatal(err)
	}
	return &proj
}

// Generate a project into a destination folder
func (p *Project) Generate(dest string) error {
	path, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("expanding destination %s failed: %s", dest, err)
	}
	if err := nocode.ValidateFolder(path); err != nil {
		return err
	}

	fmt.Println("generate", p.Name, "with", p.CLI, "to", path)
	fmt.Println(*p)

	return nil
}
