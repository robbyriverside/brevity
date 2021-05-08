package generate

import (
	"text/template"

	"github.com/robbyriverside/brevity/project"
	"github.com/robbyriverside/brief"
)

type Agenda struct {
	Templates []*brief.Node
	Actions   []*brief.Node
}

func NewAgenda() *Agenda {
	return &Agenda{
		Templates: make([]*brief.Node, 0),
		Actions:   make([]*brief.Node, 0),
	}
}

func (a *Agenda) AddTemplate(node *brief.Node) {
	a.Templates = append(a.Templates, node)
}

func (a *Agenda) AddAction(node *brief.Node) {
	a.Actions = append(a.Actions, node)
}

type Generator struct {
	Catalog  Catalog
	Template *template.Template
}

type Catalog map[string]*Agenda

func (cat Catalog) Get(elem string) *Agenda {
	agenda, ok := cat[elem]
	if !ok {
		agenda = NewAgenda()
		cat[elem] = agenda
	}
	return agenda
}

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
