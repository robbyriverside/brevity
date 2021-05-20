package generator

import (
	"fmt"
	"os"
	"text/template"

	"github.com/robbyriverside/brief"
)

// Dictionary is used to lookup elements within a hierarchy
type Dictionary map[string]*brief.Node

// Agenda contains steps to perform
type Agenda struct {
	Templates map[string]*brief.Node
	Actions   map[string]*brief.Node
}

// NewAgenda ctor
func NewAgenda() *Agenda {
	return &Agenda{
		Templates: make(map[string]*brief.Node),
		Actions:   make(map[string]*brief.Node),
	}
}

// AddTemplate add template to an agenda
func (a *Agenda) AddTemplate(node *brief.Node) {
	a.Templates[node.Name] = node
}

// AddAction add action to an agenda
func (a *Agenda) AddAction(node *brief.Node) {
	a.Actions[node.Name] = node
}

// Catalog of agenda for elements in the spac
type Catalog map[string]*Agenda

// Get the agenda for an element
func (cat Catalog) Get(elem string) *Agenda {
	agenda, ok := cat[elem]
	if !ok {
		agenda = NewAgenda()
		cat[elem] = agenda
	}
	return agenda
}

// Generator for code
type Generator struct {
	Catalog  Catalog
	Template *template.Template
}

// New Generator ctor
func New() *Generator {
	return &Generator{
		Catalog:  Catalog{},
		Template: template.New(""),
	}
}

func (gtor *Generator) compile(gen *brief.Node, files *FileSet) error {
	if files.Err != nil {
		return files.Err // stops recursive generator files
	}
	var cat Catalog
	basefile, ok := gen.Get("extend")
	if ok {
		node, err := Read(basefile)
		if err != nil {
			return err
		}
		if err := gtor.compile(node, files.Add(basefile)); err != nil {
			return err
		}
	}
	templates := gen.GetNode("templates")
	if templates != nil {
		for _, tmpl := range templates.Body {
			elem, ok := tmpl.Get("element")
			if !ok {
				return fmt.Errorf("missing template:%q element keyword", tmpl.Name)
			}
			agenda := cat.Get(elem)
			agenda.AddTemplate(tmpl)
		}
	}
	actions := gen.GetNode("actions")
	if actions != nil {
		for _, action := range actions.Body {
			elem, ok := action.Get("element")
			if !ok {
				return fmt.Errorf("missing action:%q element keyword", action.Name)
			}
			agenda := cat.Get(elem)
			agenda.AddAction(action)
		}
	}
	dir, ok := templates.Get("dir")
	if !ok {
		return fmt.Errorf("templates dir parameter is required")
	}
	if stat, err := os.Stat(dir); os.IsNotExist(err) || !stat.IsDir() {
		return fmt.Errorf("templates dir:%q must be an existing directory", dir)
	}
	// glob := filepath.Join(dir, "*.tmpl")
	// if _, err := gtor.Template.ParseGlob(glob); err != nil {
	// 	return fmt.Errorf("failed parsing templates dir:%q - %s", dir, err)
	// }
	return nil
}

// GenFile generates a file from a template
func (gtor *Generator) GenFile(node *brief.Node, dict Dictionary) error {
	// TODO: generate a file by executing a template
	return nil
}

// ExecAction executes an action
func (gtor *Generator) ExecAction(node *brief.Node, dict Dictionary) error {
	// TODO: exec an os script
	//       OR call a generator
	return nil
}

// NextNode recursively generates node hierarchy
func (gtor *Generator) NextNode(node *brief.Node, dict Dictionary) error {
	dict[node.Type] = node

	agenda := gtor.Catalog.Get(node.Type)
	for _, tmpl := range agenda.Templates {
		gtor.GenFile(tmpl, dict)
	}

	for _, subnode := range node.Body {
		gtor.NextNode(subnode, dict)
	}

	// actions as we walk back up the tree
	// no reason why, may also be actions on a second pass
	for _, act := range agenda.Actions {
		gtor.ExecAction(act, dict)
	}
	delete(dict, node.Type)
	return nil
}
