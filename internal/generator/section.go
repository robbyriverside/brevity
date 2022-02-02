package generator

import (
	"fmt"
	"text/template"

	"github.com/robbyriverside/brief"
)

// Section parser
type Section struct {
	Name string
	Node *brief.Node
}

// NewSection parser
func NewSection(name string, node *brief.Node) *Section {
	return &Section{
		Name: name,
		Node: node,
	}
}

// Compile section into template
func (sec *Section) Compile(subname string, tmpl *template.Template) error {
	top := sec.Node.Child("definitions")
	defn := make(map[string]*brief.Node)
	for _, node := range top.Body {
		switch node.Type {
		case "define":
			tmpl.Parse(define(node))
		case "definitions":
			if node.Name == "" {
				return fmt.Errorf("sub definitions must have names! See section %s", sec.Name)
			}
			defn[node.Name] = node
		}
	}
	if sec.Node.Name != "" {
		sub := defn[subname]
		for _, node := range sub.Body {
			if node.Type == "define" {
				tmpl.Parse(define(node))
			}
		}
	}
	return nil
}

func define(node *brief.Node) string {
	return fmt.Sprintf("{{define %q -}}%s{{- end}}", node.Name, node.Content)
}
