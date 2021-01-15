package main

import (
	"KISS/ts"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// TSNode is a node for any ts script data
type TSNode struct {
	BaseNode
	Src    string
	Script ts.Script
	Depth  int
}

// Parse extracts the script informaiton and arguments from the node and then calls parse on all it's children scripts
func (node *TSNode) Parse(ctx ParseNodeContext) error {
	hasSrc, srcAttr := GetAttr(node, "src")
	if hasSrc && node.firstChild != nil {
		return fmt.Errorf("error at node %s, can not have both a src value and a child text node", node)
	}
	if !hasSrc && node.firstChild == nil {
		return fmt.Errorf("error at node %s, node has neither a src element nore any child text, empty script nodes not allowed", node)
	}

	script := ""
	if node.BaseNode.firstChild != nil {
		script = node.firstChild.Data()
		node.Src = ctx.path
		Detach(node.firstChild)
	}
	if hasSrc {
		node.Src = ctx.path + srcAttr.Val
		scriptBytes, err := ioutil.ReadFile(node.Src)
		script = string(scriptBytes)
		if err != nil {
			return fmt.Errorf("error at node %s, %s", node, err)
		}
		ctx.path = getPath(node.Src)
	}

	tokens := ts.Lex(script)
	var err error
	node.Script, err = ts.Parse(tokens)
	if err != nil {
		return fmt.Errorf("error at node %s, %s", node, err)
	}

	// Add children
	for _, i := range node.Script.Imports {
		newNode := NewNode("script", TSType)
		newNode.(*TSNode).Src = ctx.path + i
		attrs := []*html.Attribute{
			&html.Attribute{Key: "src", Val: i},
			&html.Attribute{Key: "type", Val: "text/typescript"},
		}
		newNode.SetAttrs(attrs)
		AppendChild(node, newNode)
	}

	node.Depth = ctx.depth
	ctx.depth++
	return node.BaseNode.Parse(ctx)
}

// Instance replaces props in a node with params
func (node *TSNode) Instance(ctx InstNodeContext) error {
	re := regexp.MustCompile(`\$[_a-zA-Z][_a-zA-Z0-9]*\$`)
	for i := 0; i < len(node.Script.Tokens); i++ {
		tok := node.Script.Tokens[i]
		if tok.Type == ts.Value {
			matches := re.FindAll([]byte(tok.Value), -1)
			for _, match := range matches {
				val := ""
				pnode, ok := ctx.Parameters[string(match[1:len(match)-1])]
				if ok {
					if len(pnode) == 1 {
						val = pnode[0].Data()
					}
					if len(pnode) > 1 {
						return fmt.Errorf("error at node %s, tried to replace %s with multiple param nodes", node, match)
					}
					if len(pnode) == 1 && pnode[0].Type() != TextType {
						return fmt.Errorf("error at node %s, tried to replace %s with a non-text parameter", node, match)
					}
				}
				tok.Value = strings.ReplaceAll(tok.Value, string(match), val)
			}
		}
	}

	return nil
}

// Render converts a node into a textual representation
func (node *TSNode) Render() string {
	return "{\n" + node.Script.String() + "\n}\n"
}

// Clone creats a clone of the node
func (node *TSNode) Clone() Node {
	clone := &TSNode{
		BaseNode: BaseNode{data: node.Data(), attr: cloneAttrs(node.Attrs()), nType: node.Type(), visible: node.Visible()},
	}

	for _, child := range Children(node) {
		AppendChild(clone, child.Clone())
	}

	clone.Src = node.Src
	clone.Script = node.Script.Clone()

	return clone
}
