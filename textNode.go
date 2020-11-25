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
	if !node.Visible() {
		return nil
	}

	data := strings.TrimSpace(node.Data())
	if len(data) == 0 || data[0] != '{' || data[len(data)-1] != '}' {
		return nil
	}

	paramNodes, ok := ctx.Parameters[data]
	if ok {
		for _, paramNode := range paramNodes {
			// node.AppendChild(Clone(paramNode))
			node.AppendChild(paramNode.Clone())
		}

		// Set these all to visible as the component node may have hidden the originals
		for _, desc := range node.Descendants() {
			desc.SetVisible(true)
		}

		node.SetVisible(false)
	}
	return nil
}

// Render returns the text on the data
func (node *TextNode) Render(file *File) (*File, []*File) {
	fmt.Println("TEXT RENDER:", node)
	if node.Visible() {
		file.Content += node.Data()
	}

	var ret []*File
	for _, children := range node.Children() {
		mainFile, files := children.Render(file)
		file.Content += mainFile.Content
		ret = append(ret, files...)
	}

	return file, ret
}
