package brevity

import "fmt"

type options struct {
	Verbose bool `short:"v" long:"verbose" description:"verbose output"`
	Debug   bool `short:"d" long:"debug" description:"Print debugging messages, like macro expansion"`
}

var Options = &options{}

func Debug(values ...interface{}) {
	if Options.Debug {
		fmt.Print("*** ")
		fmt.Println(values...)
	}
}
