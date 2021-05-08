package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/robbyriverside/brief"
)

// MakeFolder from list of dir names
func MakeFolder(names ...string) error {
	return os.MkdirAll(filepath.Join(names...), os.ModePerm)
}

// Dictionary is used to lookup element values within a hierarchy
type Dictionary map[string]*brief.Node

// GenProject generates nth project
func GenProject(nth int, project *brief.Node, dict Dictionary) error {
	genfile, ok := project.Get("generate")
	if !ok {
		return fmt.Errorf("generate key is required and must contain filename in project %s", project.Name)
	}
	gen, err := Read(genfile)
	if err != nil {
		return err
	}
	gtor, err := NewGenerator(gen)
	if err != nil {
		return err
	}

	return gtor.GenNode(nth, project, dict)
}

// GenFile generates a file from a template
func (gtor *Generator) GenFile(nth int, node *brief.Node, dict Dictionary) error {
	// TODO: generate a file by executing a template
	return nil
}

// ExecAction executes an action
func (gtor *Generator) ExecAction(nth int, node *brief.Node, dict Dictionary) error {
	// TODO: exec an os script
	//       OR call a generator
	return nil
}

// GenNode recursively generates node hierarchy
func (gtor *Generator) GenNode(nth int, node *brief.Node, dict Dictionary) error {
	dict[node.Type] = node

	agenda := gtor.Catalog.Get(node.Type)
	for i, tmpl := range agenda.Templates {
		gtor.GenFile(i, tmpl, dict)
	}

	for i, subnode := range node.Body {
		gtor.GenNode(i, subnode, dict)
	}

	// actions as we walk back up the tree
	// no reason why, may also be actions on a second pass
	for i, act := range agenda.Actions {
		gtor.ExecAction(i, act, dict)
	}
	return nil
}
