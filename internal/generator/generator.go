package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/robbyriverside/brief"
	"github.com/sirupsen/logrus"
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
	// TODO: handle extension
	// basefile, ok := gen.Get("extend")
	// if ok {
	// 	node, err := Read(basefile)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if err := gtor.compile(node, files.Add(basefile)); err != nil {
	// 		return err
	// 	}
	// }
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
	return nil
}

func (gtor *Generator) loadTemplates(section *brief.Node, lib string) error {
	fileglob := filepath.Join(lib, section.Type, "templates", "*.tmpl")
	if _, err := gtor.Template.ParseGlob(fileglob); err != nil {
		return err
	}
	if section.Name != "" {
		fileglob := filepath.Join(lib, section.Type, "templates", section.Name, "*.tmpl")
		if _, err := gtor.Template.ParseGlob(fileglob); err != nil {
			return err
		}
	}
	return nil
}

// ApplyTemplates executes templates for this spec node
func (gtor *Generator) ApplyTemplates(spec *brief.Node, dir string) error {
	agenda := gtor.Catalog.Get(spec.Type)
	if agenda == nil {
		return nil
	}
	for _, action := range agenda.Templates {
		if err := gtor.GenFile(action, spec, dir); err != nil {
			return err
		}
	}
	return nil
}

// ApplyActions executes actions for this spec node
func (gtor *Generator) ApplyActions(spec *brief.Node, dir string) error {
	agenda := gtor.Catalog.Get(spec.Type)
	if agenda == nil {
		return nil
	}
	for _, action := range agenda.Actions {
		if err := gtor.ExecAction(action, spec, dir); err != nil {
			return err
		}
	}
	return nil
}

// NextNode recursively generates node hierarchy
func (gtor *Generator) NextNode(node *brief.Node, dir string) error {
	if err := gtor.ApplyTemplates(node, dir); err != nil {
		return err
	}

	for _, subnode := range node.Body {
		if err := gtor.NextNode(subnode, dir); err != nil {
			return err
		}
	}

	// actions as we walk back up the tree
	// no reason why, may also be actions on a second pass
	return gtor.ApplyActions(node, dir)
}

// ExecValueTemplate for templates inside action key values
func ExecValueTemplate(value string, node *brief.Node) (string, error) {
	if !strings.Contains(value, "{{") {
		return value, nil
	}
	filetmpl, err := template.New("value").Parse(value)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	if err := filetmpl.Execute(&out, node); err != nil {
		return "", err
	}
	return out.String(), nil
}

// GenFile generates a file from a template
func (gtor *Generator) GenFile(action, spec *brief.Node, dir string) error {
	tmpl := gtor.Template.Lookup(action.Name)
	if tmpl == nil {
		return fmt.Errorf("no template found for %s", action.Name)
	}
	filetmpl, ok := action.Keys["file"]
	if !ok {
		return fmt.Errorf("template %s has no file", action.Name)
	}

	filename, err := ExecValueTemplate(filetmpl, spec)
	if err != nil {
		return err
	}
	filename = filepath.Join(dir, filename)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return gtor.Template.ExecuteTemplate(file, action.Name, spec)
}

// ExecAction executes an action
func (gtor *Generator) ExecAction(action, spec *brief.Node, dir string) error {
	// FIXME: call a generator
	exectmpl, ok := action.Keys["exec"]
	if !ok {
		return fmt.Errorf("template %s has no file", action.Name)
	}

	execute, err := ExecValueTemplate(exectmpl, spec)
	if err != nil {
		return err
	}
	cmd := exec.Command(execute)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"output": string(out),
			"action": action.Name,
			"dir":    dir,
		}).Error("failed action")
		return err
	}
	return nil
}
