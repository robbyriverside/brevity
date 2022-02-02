package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/robbyriverside/brevity/internal/brevity"

	"github.com/jessevdk/go-flags"
	"github.com/robbyriverside/brief"
)

/*
Generates code by reading a brevity file containing specs for 1 or more projects.
Each project spec references a project generator.  A project generator creates a project of a specific type.
For exmaple: A generator for...
	- golang backend project
	- flutter app project
	- golang tool project
*/

// Command for generate
type Command struct {
	Args struct {
		SpecFile    string `positional-arg-name:"specfile" description:"brevity specification file"`
		Destination string `positional-arg-name:"destination" description:"where to put the project root folder"`
	} `positional-args:"true" required:"true"`
	Library string `short:"l" long:"lib" description:"Brevity library location" env:"BREVITY_LIB"`
	Render  bool   `short:"r" long:"render" description:"Render files without actions"`
	specDir string
}

// Execute the project command
func (cmd *Command) Execute(args []string) error {
	gtor, err := cmd.CompileLibrary()
	specfile, err := filepath.Abs(cmd.Args.SpecFile)
	if err != nil {
		return err
	}
	cmd.specDir = filepath.Dir(specfile)
	node, err := cmd.ReadSpec()
	if err != nil {
		return err
	}
	return gtor.Generate(node)
}

// AddCommand to the parser
func AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("generate",
		"generate brevity projects",
		"creates files and folders for brevity projects",
		&Command{},
	)
	return err
}

// ReadNode reads a single node from a brief file
func ReadNode(specfile string) (*brief.Node, error) {
	file, err := os.Open(specfile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	dec := brief.NewDecoder(file, 4, filepath.Dir(specfile))
	dec.Debug = brevity.Options.Debug
	nodes, err := dec.Decode()
	if err != nil {
		return nil, err
	}
	if len(nodes) > 1 {
		return nil, fmt.Errorf("brief spec file %q has more than one top level form", specfile)
	}
	return nodes[0], nil
}

// ReadSpec brevity spec from file
func (cmd *Command) ReadSpec() (*brief.Node, error) {
	spec, err := ReadNode(cmd.Args.SpecFile)
	if err != nil {
		return nil, err
	}
	if spec.Type != "brevity" {
		return nil, fmt.Errorf("invalid brevity spec: top-level brevity")
	}
	for _, project := range spec.Body {
		if project.Type != "project" {
			return nil, fmt.Errorf("invalid brevity project")
		}
	}
	return spec, nil
}

// CompileLibrary for this command
func (cmd *Command) CompileLibrary() (*Generator, error) {
	libfile := filepath.Join(cmd.Library, "sections.brief")
	if _, err := os.Stat(libfile); os.IsNotExist(err) {
		return nil, fmt.Errorf("library file: %s error: %s", libfile, err)
	}
	gtor := cmd.New()

	libnode, err := ReadNode(libfile)
	if err != nil {
		return nil, fmt.Errorf("failed reading lib node: %s", err)
	}
	for _, node := range libnode.Body {
		section := NewSection(node.Name, node)
		gtor.AddSection(section)
	}
	return gtor, nil
}

// Generate brevity projects into a destination folder
func (gtor *Generator) Generate(brevity *brief.Node) error {
	path, err := filepath.Abs(gtor.Destination)
	if err != nil {
		return fmt.Errorf("expanding destination %s failed: %s", gtor.Destination, err)
	}
	gtor.Destination = path
	if err := ValidateFolder(path); err != nil {
		return err
	}
	// Generate code for each project
	for _, project := range brevity.Body {
		if len(project.Name) == 0 {
			return fmt.Errorf("invalid brevity spec: project must be named")
		}
		if err := gtor.Project(project); err != nil {
			return err
		}
	}
	return nil
}

// CompileSection found in brevity project
func (gtor *Generator) CompileSection(section *brief.Node) error {
	if len(gtor.Catalog) == 0 {
		return fmt.Errorf("empty generator catalog")
	}

	brevity.Debug("section catalog size", len(gtor.Catalog))
	if err := gtor.LoadSectionTemplates(section); err != nil {
		return err
	}

	if gtor.Template.DefinedTemplates() == "" {
		return fmt.Errorf("no templates found: section %s:%s", section.Type, section.Name)
	}
	brevity.Debug("section templates", gtor.Template.DefinedTemplates())
	return nil
}

// Project generates nth project in the spec
func (gtor *Generator) Project(project *brief.Node) error {
	if project.Name == "" {
		return fmt.Errorf("project name is required")
	}
	dir := filepath.Join(gtor.Destination, project.Name)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	if brevity.Options.Verbose {
		fmt.Println("--> project", project.Name, dir)
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	for _, section := range project.Body {
		err := gtor.CompileSection(section)
		if err != nil {
			return err
		}
		if err := gtor.ValidateSection(section); err != nil {
			return err
		}
	}
	if err := gtor.ExpandProjectMacros(project); err != nil {
		return err
	}
	for _, section := range project.Body {
		if err := gtor.ApplyTemplates(project, dir); err != nil {
			return err
		}
		if err := gtor.ApplyTemplates(section, dir); err != nil {
			return err
		}

		for _, subnode := range section.Body {
			if err := gtor.NextNode(subnode, dir); err != nil {
				return err
			}
		}

		// actions as we walk back up the tree
		// XXX: should be predictable, may also be actions on a second pass
		if err := gtor.ApplyActions(section, dir); err != nil {
			return err
		}
		if err := gtor.ApplyActions(project, dir); err != nil {
			return err
		}
	}
	return nil
}
