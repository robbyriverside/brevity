package generator

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/robbyriverside/brevity/internal/brevity"
	"github.com/robbyriverside/brief"
)

// ExecuteMacro template which defines a brief spec then decode into nodes
func ExecuteMacro(tmpl *template.Template, section *brief.Node) ([]*brief.Node, error) {
	var out strings.Builder
	tmpl.Execute(&out, section)
	if brevity.Options.Debug {
		fmt.Println("*** macro:\n", string(section.Encode()))
		fmt.Println("*** expansion:\n", out.String())
	}
	in := strings.NewReader(out.String())
	dec := brief.NewDecoder(in, 4)
	dec.Padding = section.Parent.Indent
	return dec.Decode()
}

// MergeKeys of two nodes, the left node takes precedence
func MergeKeys(node, other *brief.Node) {
	if node.Name == "" {
		node.Name = other.Name
	}
	for key, value := range other.Keys {
		_, ok := node.Keys[key]
		if !ok {
			node.Keys[key] = value
		}
	}
}

// MergeBody of two nodes, the left node takes precendence
func MergeBody(node, other *brief.Node) {
	body := append(node.Body, other.Body...)
	nodes := MergeNodes(body, true)
	node.Body = nodes
}

// MergeNode combines two nodes recursively
func MergeNode(node, other *brief.Node) {
	MergeKeys(node, other)
	MergeBody(node, other)
	if node.Content == "" {
		node.Content = other.Content
	}
}

// MergeNodes combine a set of nodes of the same Type
// useNames means they are only merged if they have the same name
func MergeNodes(body []*brief.Node, useNames bool) []*brief.Node {
	current := body
	remain := make([]*brief.Node, 0)
	result := make([]*brief.Node, 0)
	for {
		if len(current) == 0 {
			break
		}
		first := current[0]
		result = append(result, first)
		rest := current[1:]
		for _, next := range rest {
			if first.Type == next.Type {
				if !useNames || first.Name == next.Name {
					MergeNode(first, next)
					continue
				}
			}
			remain = append(remain, next)
		}
		current = remain
		remain = make([]*brief.Node, 0)
	}
	return result
}

// ExpandProjectMacros loops over project sections until all macros are expanded
func (gtor *Generator) ExpandProjectMacros(project *brief.Node) error {
	current := project.Body
	body := make([]*brief.Node, 0)
	expanded := make([]*brief.Node, 0)
	for {
		for _, section := range current {
			macroName := fmt.Sprintf("@macro.%s", section.Type)
			tmpl := gtor.Template.Lookup(macroName)
			if tmpl != nil {
				nodes, err := ExecuteMacro(tmpl, section)
				if err != nil {
					return err
				}
				for _, n := range nodes {
					n.Parent = section.Parent
				}
				expanded = append(expanded, nodes...)
				continue
			}

			body = append(body, section)
		}

		if len(expanded) == 0 {
			break
		}
		current = expanded
		expanded = make([]*brief.Node, 0)
	}

	project.Body = MergeNodes(body, false)
	for _, node := range project.Body {
		node.Parent = project
	}
	return nil
}
