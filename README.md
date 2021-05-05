# Brevity

App Meta-Generator.

Write layered specifications and compose a project in phases.

Each new phase integrates with the existing code.

The generated source is readable and flexible for custom edits.

# Generators

These are the code generators configured using brief language.  

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
  
