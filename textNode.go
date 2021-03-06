package main

import (
	"strings"
)

// TextNode contains text that does not appear in an xml tag
type TextNode struct {
	BaseNode
}

// Instance takes parameters from the node context and replaces template parameteres
func (node *TextNode) Instance(ctx InstNodeContext) error {
	if node.Visible() {
		data := strings.TrimSpace(node.Data())
		if len(data) == 0 || data[0] != '{' || data[len(data)-1] != '}' {
			return nil
		}

		paramNodes, ok := ctx.Parameters[data[1:len(data)-1]]
		if ok {
			for _, paramNode := range paramNodes {
				AppendChild(node, paramNode.Clone())
			}
		}

		node.SetVisible(false)
	}

	return node.BaseNode.Instance(ctx)
}

// Render returns the text on the data
func (node *TextNode) Render() string {
	var ret string
	if node.Visible() {
		ret += node.Data()
	}

	for _, child := range Children(node) {
		ret += child.Render()
	}

	return ret
}
