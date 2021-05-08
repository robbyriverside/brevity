package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/robbyriverside/brief"
)

/*
Generates code by reading a brevity file containing specs for 1 or more projects.
Each project spec references a project generator.  A project generator creates a projec of a specific type.
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
}

// Execute the project command
func (cmd *Command) Execute(args []string) error {
	node, err := Read(cmd.Args.SpecFile)
	if err != nil {
		return err
	}
	return Generate(cmd.Args.Destination, node)
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

// Read brevity spec from file
func Read(specfile string) (*brief.Node, error) {
	file, err := os.Open(specfile)
	if err != nil {
		return nil, err
	}
	node, err := brief.Decode(file)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// Generate brevity projects into a destination folder
func Generate(dest string, brevity *brief.Node) error {
	path, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("expanding destination %s failed: %s", dest, err)
	}
	if err := ValidateFolder(path); err != nil {
		return err
	}
	for i, project := range brevity.Body {
		dict := Dictionary{}
		if err = GenProject(i, project, dict); err != nil {
			return err
		}
	}
	return nil
}