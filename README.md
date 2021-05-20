# Brevity

Brevity is an Code-Generator that can be used for any target software language.  Brevity is written in Go and uses the Go text/template engine to generate source code.  Libraries of code templates are used to compose a sophisticated application using only a specification.  The generated code is defined in layers to create the starting point for custom code development at many levels of abstraction: cli-only, cli plus web api, etc.

Brevity generates project code files and executes init commands to define an application.  The project specification is metadata written in the [brief](https://github.com/robbyriverside/brief) specification language which is syntax simplified XML.

The contents of files are defined using Go text-templates.  The generated files are allowed to be any format, even script files that can be later executed as an init command.  Further, Brevity is a meta-generator because it can generate template files or even [brief](https://github.com/robbyriverside/brief) files which can then be used for code generation.

__This page under construction__
