package generate

import (
	"text/template"

	"github.com/robbyriverside/brevity/project"
	"github.com/robbyriverside/brief"
)

// Dictionary is used to lookup elements within a hierarchy
type Dictionary map[string]*brief.Node

// Agenda contains steps to perform
type Agenda struct {
	Templates []*brief.Node
	Actions   []*brief.Node
}

// NewAgenda ctor
func NewAgenda() *Agenda {
	return &Agenda{
		Templates: make([]*brief.Node, 0),
		Actions:   make([]*brief.Node, 0),
	}
}

// AddTemplate add template to an agenda
func (a *Agenda) AddTemplate(node *brief.Node) {
	a.Templates = append(a.Templates, node)
}

// AddAction add action to an agenda
func (a *Agenda) AddAction(node *brief.Node) {
	a.Actions = append(a.Actions, node)
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

// NewGenerator ctor
func NewGenerator(gen *brief.Node) (*Generator, error) {
	var cat Catalog
	templates := gen.GetNode("templates")
	if templates != nil {
		for _, tmpl := range templates.Body {
			elem, ok := tmpl.Get("element")
			if !ok {
				continue // TODO: might be an error required arg:element
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
				continue // TODO: might be an error required arg:element
			}
			agenda := cat.Get(elem)
			agenda.AddAction(action)
		}
	}
	dir, unknown := templates.Get("dir")
	tmpl, err := project.LoadTemplates(dir, unknown)
	if err != nil {
		return nil, err
	}
	return &Generator{
		Catalog:  cat,
		Template: tmpl,
	}, nil
}

// FileTemplate generates a file from a template
func (gtor *Generator) FileTemplate(nth int, node *brief.Node, dict Dictionary) error {
	// TODO: generate a file by executing a template
	return nil
}

// ExecAction executes an action
func (gtor *Generator) ExecAction(nth int, node *brief.Node, dict Dictionary) error {
	// TODO: exec an os script
	//       OR call a generator
	return nil
}

// NextNode recursively generates node hierarchy
func (gtor *Generator) NextNode(nth int, node *brief.Node, dict Dictionary) error {
	dict[node.Type] = node

	agenda := gtor.Catalog.Get(node.Type)
	for i, tmpl := range agenda.Templates {
		gtor.FileTemplate(i, tmpl, dict)
	}

	for i, subnode := range node.Body {
		gtor.NextNode(i, subnode, dict)
	}

	// actions as we walk back up the tree
	// no reason why, may also be actions on a second pass
	for i, act := range agenda.Actions {
		gtor.ExecAction(i, act, dict)
	}
	delete(dict, node.Type)
	return nil
}
