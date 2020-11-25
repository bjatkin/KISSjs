package main

import (
	"fmt"
	"strings"
)

// ComponentNode is a node for all components that match imports nodes
type ComponentNode struct {
	BaseNode
	NoBundleID          string
	NoBundle, NoCompile bool
}

// ToComponentNode converts any kiss node into a component type node
// Warning! this function should not be used in a Parse, Inline, or Render function
// if this function is used in a Parse or Inline function it
// should never be used on sibling or parent nodes,
// only on child nodes or a sibling's child nodes
func ToComponentNode(node Node) *ComponentNode {
	ret := &ComponentNode{}
	ret.SetParent(node.Parent())
	if node.Parent() != nil && node.PrevSibling() == nil {
		node.Parent().SetFirstChild(ret)
	}
	ret.SetFirstChild(node.FirstChild())
	for _, child := range node.Children() {
		child.SetParent(ret)
	}
	ret.SetPrevSibling(node.PrevSibling())
	if node.PrevSibling() != nil {
		node.PrevSibling().SetNextSibling(ret)
	}
	ret.SetNextSibling(node.NextSibling())
	if node.NextSibling() != nil {
		node.NextSibling().SetPrevSibling(ret)
	}

	ret.SetData(node.Data())
	ret.SetAttrs(node.Attrs())
	ret.SetVisible(false)
	return ret
}

// Parse uses the it's class to add a root component and then calls parse on all it's children
func (node *ComponentNode) Parse(ctx NodeContext) error {
	err := node.BaseNode.Parse(ctx)
	if err != nil {
		return err
	}
	ctx.Parameters = make(map[string][]Node)

	for _, desc := range node.Descendants() {
		desc.SetVisible(false)
	}

	for _, attr := range node.Attrs() {
		ctx.Parameters["{"+strings.ToLower(attr.Key)+"}"] = []Node{NewNode(attr.Val, TextType)}
	}

	for _, child := range node.Children() {
		ctx.Parameters["{"+strings.ToLower(child.Data())+"}"] = child.Children()
	}

	var root Node
	for _, tag := range ctx.ImportTags {
		if strings.ToLower(node.Data()) == tag.tag {
			root = tag.root.Clone()
			root.SetVisible(false)
			for _, desc := range root.Descendants() {
				err := desc.Instance(ctx)
				if err != nil {
					return fmt.Errorf("error at node %s, could not copy data from class into component %s", node, err)
				}
			}
			node.AppendChild(root)

			ctx.path = tag.path
			break
		}
	}

	return nil
}

func (node *ComponentNode) Clone() Node {
	clone := ComponentNode{
		BaseNode: BaseNode{data: node.Data(), attr: node.Attrs(), visible: node.Visible()},
	}

	for _, child := range node.Children() {
		clone.AppendChild(child.Clone())
	}

	clone.NoBundleID = node.NoBundleID
	clone.NoBundle = node.NoBundle
	clone.NoCompile = node.NoCompile

	return &clone
}
