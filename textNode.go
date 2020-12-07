package main

import (
	"fmt"
	"strings"
)

// TextNode contains text that does not appear in an xml tag
type TextNode struct {
	BaseNode
}

// Instance replaces props in a node with params
func (node *TextNode) Instance(ctx NodeContext) error {
	if node.Visible() {
		fmt.Println("HERE ", node)
		data := strings.TrimSpace(node.Data())
		if len(data) == 0 || data[0] != '{' || data[len(data)-1] != '}' {
			return nil
		}

		paramNodes, ok := ctx.Parameters[data[1:len(data)-1]]
		if ok {
			for _, paramNode := range paramNodes {
				node.AppendChild(paramNode.Clone())
			}

			node.SetVisible(false)
		}
	}

	return node.BaseNode.Instance(ctx)
}

// Render returns the text on the data
func (node *TextNode) Render() string {
	ret := ""
	if node.Visible() {
		ret += node.Data()
	}

	for _, child := range node.Children() {
		ret += child.Render()
	}

	return ret
}
