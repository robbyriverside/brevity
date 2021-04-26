package project

import (
	"github.com/jessevdk/go-flags"
)

// Command for the project
type Command struct {
	Args struct {
		Destination string `positional-arg-name:"destination" description:"where to put the project root folder"`
		SpecFile    string `positional-arg-name:"specfile" description:"project specification file"`
	} `positional-args:"true" required:"true"`
}

// Execute the project command
func (cmd *Command) Execute(args []string) error {
	return Read(cmd.Args.SpecFile).Generate(cmd.Args.Destination)
}

// AddCommand to the parser
func AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("project",
		"generate a Go project",
		"creates files and folders for a basic Go project using the name of the project/folder",
		&Command{},
	)
	return err
}
