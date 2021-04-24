package main

import (
	"log"

	"github.com/robbyriverside/nocode"
	"github.com/robbyriverside/nocode/project"

	"github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewParser(nocode.Options, flags.Default)
	parser.Name = "nocode"

	if err := project.AddCommand(parser); err != nil {
		log.Fatal(err)
	}

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			return
		}
		log.Fatal(err)
	}
}
