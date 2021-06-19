package generator

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/robbyriverside/brevity/internal/brevity"
	"github.com/robbyriverside/brief"

	"github.com/Masterminds/sprig"
	"github.com/google/shlex"
	"github.com/sirupsen/logrus"
)

// Dictionary is used to index elements within a section
type Dictionary struct {
	Map  map[string]*brief.Node
	List []*brief.Node
}

// NewDictionary construct empty dictionary
func NewDictionary() *Dictionary {
	return &Dictionary{
		Map:  make(map[string]*brief.Node),
		List: make([]*brief.Node, 0),
	}
}

// Add node to a dictionary
func (dict *Dictionary) Add(node *brief.Node) {
	if len(node.Name) > 0 {
		dict.Map[node.Name] = node
	}
	dict.List = append(dict.List, node)
}

// Agenda contains steps to perform
type Agenda struct {
	Templates *Dictionary
	Actions   *Dictionary
	Found     bool
}

// NewAgenda constructor
func NewAgenda() *Agenda {
	return &Agenda{
		Templates: NewDictionary(),
		Actions:   NewDictionary(),
	}
}

// AddTemplate add template to an agenda
func (a *Agenda) AddTemplate(node *brief.Node) {
	a.Templates.Add(node)
}

// AddAction add action to an agenda
func (a *Agenda) AddAction(node *brief.Node) {
	a.Actions.Add(node)
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
	Catalog         Catalog
	Template        *template.Template
	Render          bool
	LibDir, SpecDir string
}

// New Generator ctor
func (cmd *Command) New() *Generator {
	return &Generator{
		Catalog:  Catalog{},
		Template: template.New("top").Funcs(sprig.GenericFuncMap()),
		Render:   cmd.Render,
		LibDir:   cmd.Library,
		SpecDir:  cmd.specDir,
	}
}

// ValidateTemplate ensure correct template node
func ValidateTemplate(tmpl *brief.Node, pos int) error {
	if len(tmpl.Name) == 0 {
		return fmt.Errorf("template %d has no name", pos)
	}
	_, ok := tmpl.Keys["element"]
	if !ok {
		return fmt.Errorf("missing template:%q element keyword", tmpl.Name)
	}
	_, ok = tmpl.Keys["file"]
	if !ok {
		return fmt.Errorf("missing template:%q file keyword", tmpl.Name)
	}
	return nil
}

// ValidateAction ensure correct action node
func ValidateAction(act *brief.Node, pos int) error {
	if len(act.Name) == 0 {
		return fmt.Errorf("action %d has no name", pos)
	}
	_, ok := act.Keys["element"]
	if !ok {
		return fmt.Errorf("missing action:%q element keyword", act.Name)
	}
	_, ok = act.Keys["exec"]
	if !ok {
		return fmt.Errorf("missing action:%q exec keyword", act.Name)
	}
	return nil
}

func (gtor *Generator) compile(gen *brief.Node) error {
	templates := gen.Child("templates")
	if templates == nil {
		return fmt.Errorf("generator.brief missing templates node")
	}
	if templates != nil {
		for i, tmpl := range templates.Body {
			if err := ValidateTemplate(tmpl, i); err != nil {
				return err
			}
			elem := tmpl.Keys["element"]
			agenda := gtor.Catalog.Add(elem)
			agenda.AddTemplate(tmpl)
		}
	}
	actions := gen.Child("actions")
	if actions == nil {
		return fmt.Errorf("generator.brief missing actions node")
	}
	if actions != nil {
		for i, action := range actions.Body {
			if err := ValidateAction(action, i); err != nil {
				return err
			}
			elem := action.Keys["element"]
			agenda := gtor.Catalog.Add(elem)
			agenda.AddAction(action)
		}
	}
	return nil
}

// ValidateSection returns and error if any template elements are missing
func (gtor *Generator) ValidateSection(section *brief.Node) error {
	gtor.Catalog.validateNode(section)
	missing := []string{}
	for key, agenda := range gtor.Catalog {
		if key == "project" {
			continue
		}
		if !agenda.Found {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("invalid %s spec: missing elements %s", section.Type, missing)
	}
	return nil
}

func (cat Catalog) validateNode(node *brief.Node) {
	agenda, found := cat[node.Type]
	if found {
		agenda.Found = found
	}
	for _, n := range node.Body {
		cat.validateNode(n)
	}
}

// LoadGenerator if the genfile exists
func (gtor *Generator) LoadGenerator(genfile string) error {
	node, err := ReadNode(genfile)
	if err != nil {
		return err
	}
	return gtor.compile(node)
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
	filename := filepath.Join(gtor.SpecDir, local)
	return gtor.LoadGlobTemplates(filename)
}

// SectionNames subdirs of the templates directory
func (gtor *Generator) SectionNames(section *brief.Node) (map[string]bool, error) {
	tdir := filepath.Join(gtor.LibDir, section.Type, "templates")
	files, err := ioutil.ReadDir(tdir)
	if err != nil {
		return nil, err
	}
	result := map[string]bool{}
	for _, info := range files {
		if info.IsDir() {
			result[info.Name()] = true
		}
	}
	return result, nil
}

// LoadSectionTemplates load templates for a section
func (gtor *Generator) LoadSectionTemplates(section *brief.Node) error {
	if err := gtor.LoadGlobTemplates(filepath.Join(gtor.LibDir, section.Type, "templates", "*.tmpl")); err != nil {
		return err
	}
	if section.Name != "" {
		if err := gtor.LoadGlobTemplates(filepath.Join(gtor.LibDir, section.Type, "templates", section.Name, "*.tmpl")); err != nil {
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
	for _, action := range agenda.Templates.List {
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
	if gtor.Render {
		if brevity.Options.Debug {
			for _, action := range agenda.Actions.List {
				fmt.Printf("*** action:%q element:%q exec:%q\n", action.Name, action.Keys["element"], action.Keys["exec"])
			}
		}
		return nil
	}
	for _, action := range agenda.Actions.List {
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
	filetmpl, err := template.New("value").Funcs(sprig.GenericFuncMap()).Parse(value)
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
