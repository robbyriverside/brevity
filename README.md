# Brevity

Brevity is an App Meta-Generator.

Brevity generates project code files and executes init commands to define a Go application.  The project specification is metadata written in the [brief](https://github.com/robbyriverside/brief) specification language which is syntax simplified XML.

The contents of files are defined using Go text-templates.  The generated files are allowed to be any format, even script files that can be later executed as an init command.  Further, Brevity is a meta-generator because it can generate template files or even __brief__ files which can then be used for code generation.

# Generators

These are the code generators configured using the __brief__ language.  

In future, there will be a meta-generator that uses a meta-level (also brief) specification that generates the config for the code generators.

## Project

Generate code for a Go project.

Options:
 - cli:  which command line flags interpreter to install
   - Future:  urfave, go-flags, cobra
 - api: which web api package to install
   - Future:  gin, protobuf
 - mocks: which mock generator to install
   - Future:  go-mocks
  
