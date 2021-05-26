package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/robbyriverside/brevity/internal/brevity"
	"github.com/robbyriverside/brief"

	"github.com/google/shlex"
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

// Add the agenda for an element
func (cat Catalog) Add(elem string) *Agenda {
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
	templates := gen.GetNode("templates")
	if templates != nil {
		for _, tmpl := range templates.Body {
			elem, ok := tmpl.Keys["element"]
			if !ok {
				return fmt.Errorf("missing template:%q element keyword", tmpl.Name)
			}
			agenda := gtor.Catalog.Add(elem)
			agenda.AddTemplate(tmpl)
		}
	}
	actions := gen.GetNode("actions")
	if actions != nil {
		for _, action := range actions.Body {
			elem, ok := action.Keys["element"]
			if !ok {
				return fmt.Errorf("missing action:%q element keyword", action.Name)
			}
			agenda := gtor.Catalog.Add(elem)
			agenda.AddAction(action)
		}
	}
	return nil
}

// LoadGlobTemplates loads templates from the fileglob into generator
func (gtor *Generator) LoadGlobTemplates(fileglob string) error {
	filenames, err := filepath.Glob(fileglob)
	if err != nil {
		return err
	}
	if len(filenames) == 0 {
		return nil
	}
	_, err = gtor.Template.ParseFiles(filenames...)
	return err
}

func (gtor *Generator) loadLocalTemplates(node *brief.Node) error {
	local, ok := node.Keys["templates"]
	if !ok {
		return nil
	}
	return gtor.LoadGlobTemplates(local)
}

// LoadSectionTemplates load templates for a section
func (gtor *Generator) LoadSectionTemplates(section *brief.Node, lib string) error {
	if err := gtor.LoadGlobTemplates(filepath.Join(lib, section.Type, "templates", "*.tmpl")); err != nil {
		return err
	}
	if section.Name != "" {
		if err := gtor.LoadGlobTemplates(filepath.Join(lib, section.Type, "templates", section.Name, "*.tmpl")); err != nil {
			return err
		}
	}
	// local brevity templates
	if err := gtor.loadLocalTemplates(section.Parent.Parent); err != nil {
		return err
	}
	// local project templates
	if err := gtor.loadLocalTemplates(section.Parent); err != nil {
		return err
	}
	// local section templates
	return gtor.loadLocalTemplates(section)
}

// ApplyTemplates executes templates for this spec node
func (gtor *Generator) ApplyTemplates(spec *brief.Node, dir string) error {
	agenda, ok := gtor.Catalog[spec.Type]
	if !ok {
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
	agenda, ok := gtor.Catalog[spec.Type]
	if !ok {
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
	if brevity.Options.Verbose {
		fmt.Printf("template %s on %s:%s -> %s\n", action.Name, spec.Type, spec.Name, filename)
	}
	path := filepath.Dir(filename)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := gtor.Template.ExecuteTemplate(file, action.Name, spec); err != nil {
		return err
	}
	return file.Sync()
}

// ExecAction executes an action
func (gtor *Generator) ExecAction(action, spec *brief.Node, dir string) error {
	// MAYBE: call a generator
	exectmpl, ok := action.Keys["exec"]
	if !ok {
		return fmt.Errorf("template %s has no file", action.Name)
	}

	execute, err := ExecValueTemplate(exectmpl, spec)
	if err != nil {
		return err
	}
	args, err := shlex.Split(execute)
	if err != nil {
		return err
	}
	if brevity.Options.Verbose {
		fmt.Printf("action %s on %s:%s exec: %s\n", action.Name, spec.Type, spec.Name, strings.Join(args, " "))
	}
	cmd := exec.Command(args[0], args[1:]...)
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
