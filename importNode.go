package main

import (
	"fmt"
	"strings"
)

// ImportNode is a node for all component definition
type ImportNode struct {
	BaseNode
	Tag           string
	Src           string
	ComponentRoot Node
}

// Parse validates the import node and builds all the related context nodes
func (node *ImportNode) Parse(ctx ParseNodeContext) error {
	hasTag, tagAttr := GetAttr(node, "tag")
	if !hasTag {
		return fmt.Errorf("error at node %s, import node must have a tag attribute", node)
	}

	hasSrc, srcAttr := GetAttr(node, "src")
	if hasSrc && node.ComponentRoot != nil {
		return fmt.Errorf("error at node %s, can not have both a src value and a child node", node)
	}

	if hasSrc {
		node.Src = ctx.path + srcAttr.Val
		children, err := parseComponentFile(node.Src)
		if err != nil {
			return fmt.Errorf("error at node %s, %s there was an error parsing component src", node, err)
		}

		root := NewNode("root", BaseType)
		root.SetVisible(false)
		for _, child := range children {
			AppendChild(root, child)
		}
		node.ComponentRoot = root
	}

	compCtx := ctx.Clone()
	compCtx.path = getPath(node.Src)
	err := node.ComponentRoot.Parse(compCtx)
	if err != nil {
		return err
	}

	ctx.ImportTags = append(ctx.ImportTags,
		ImportTag{
			tag:  strings.ToLower(tagAttr.Val),
			root: node.ComponentRoot,
			path: compCtx.path,
		},
	)

	return node.BaseNode.Parse(ctx)
}

// Clone creates a deep copy of a node, but does not copy over the connections to the original parent and siblings
func (node *ImportNode) Clone() Node {
	clone := &ImportNode{
		BaseNode: BaseNode{data: node.Data(), attr: cloneAttrs(node.Attrs()), nType: node.Type(), visible: node.Visible()},
	}

	for _, child := range Children(node) {
		AppendChild(clone, child.Clone())
	}

	clone.Tag = node.Tag
	clone.Src = node.Src
	clone.ComponentRoot = node.ComponentRoot

	return clone
}
