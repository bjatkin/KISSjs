package main

import (
	"fmt"
	"regexp"
	"strings"
)

// ComponentNode is a node for all components that match imports nodes
type ComponentNode struct {
	BaseNode
	TempNodes []Node
	TempAttr  []Node
}

// Parse uses the it's class to add a root component and then calls parse on all it's children
func (node *ComponentNode) Parse(ctx ParseNodeContext) error {
	err := node.BaseNode.Parse(ctx)
	if err != nil {
		return err
	}

	var root Node
	for _, tag := range ctx.ImportTags {
		if strings.ToLower(node.Data()) == tag.tag {
			root = tag.root.Clone()
			root.SetVisible(false)
			AppendChild(node, root)

			ctx.path = tag.path
			break
		}
	}

	// Collect all the attributes here
	node = collectNodes(node, root)

	return nil
}

// Instance takes parameters from the node context and replaces template parameteres
func (node *ComponentNode) Instance(ctx InstNodeContext) error {
	re := regexp.MustCompile(`{[_a-zA-Z][_a-zA-Z0-9]*}`)
	for _, attr := range node.Attrs() {
		matches := re.FindAll([]byte(attr.Val), -1)
		for _, match := range matches {
			node, ok := ctx.Parameters[string(match[1:len(match)-1])]
			if ok {
				if len(node) != 1 || node[0].Type() != TextType {
					return fmt.Errorf("error at node %s, tried to replace %s with a non-text parameter", node, match)
				}
				attr.Val = strings.ReplaceAll(attr.Val, string(match), node[0].Data())
			}
		}
	}

	ctx.componentScope = "k-" + randomID(6)
	ctx.Parameters = make(map[string][]Node)

	for _, attr := range node.Attrs() {
		ctx.Parameters[strings.ToLower(attr.Key)] = []Node{NewNode(attr.Val, TextType)}
	}

	for _, child := range Children(node) {
		if child.Data() == "root" {
			err := child.Instance(ctx)
			if err != nil {
				return err
			}
			break
		}
		ctx.Parameters[strings.ToLower(child.Data())] = Children(child)
	}

	for _, child := range Children(node) {
		err := child.Instance(ctx)
		if err != nil {
			return err
		}
	}

	// Hide all the parameters
	for _, child := range Children(node) {
		if child.Data() == "root" {
			continue
		}
		child.SetVisible(false)
		for _, desc := range Descendants(child) {
			desc.SetVisible(false)
		}
	}

	return nil
}

func collectNodes(node *ComponentNode, root Node) *ComponentNode {
	// Collect all the attributes here
	re := regexp.MustCompile(`{[_a-zA-Z][_a-zA-Z0-9]*}`)
	descs := Descendants(root)
	for i := 0; i < len(descs); i++ {
		if re.Match([]byte(descs[i].Data())) {
			node.TempNodes = append(node.TempNodes, descs[i])
		}
		for _, attrib := range descs[i].Attrs() {
			if re.Match([]byte(attrib.Val)) {
				node.TempAttr = append(node.TempAttr, descs[i])
			}
		}
	}

	return node
}

// Clone creates a deep copy of a node, but does not copy over the connections to the original parent and siblings
func (node *ComponentNode) Clone() Node {
	clone := &ComponentNode{
		BaseNode: BaseNode{data: node.Data(), attr: cloneAttrs(node.Attrs()), nType: node.Type(), visible: node.Visible()},
	}

	var root Node
	for _, child := range Children(node) {
		AppendChild(clone, child.Clone())
		if child.Data() == "root" {
			root = child
		}
	}

	clone = collectNodes(clone, root)

	return clone
}
