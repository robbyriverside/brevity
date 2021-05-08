package main

import (
	"log"

	"github.com/robbyriverside/brevity/internal/brevity"
	"github.com/robbyriverside/brevity/internal/generate"

	"github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewParser(brevity.Options, flags.Default)
	parser.Name = "brevity"

	if err := generate.AddCommand(parser); err != nil {
		log.Fatal(err)
	}

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			return
		}
		log.Fatal(err)
	}
}
