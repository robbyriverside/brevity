package brevity

type options struct {
	Verbose bool `short:"v" long:"verbose" description:"verbose output"`
	Debug   bool `short:"d" long:"debug" description:"Print debugging messages, like macro expansion"`
}

var Options = &options{}
