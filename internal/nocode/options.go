package nocode

type options struct {
	Verbose bool `short:"v" long:"verbose" description:"verbose output"`
}

var Options = &options{}
