package main

import (
	"KISS/js"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// JSNode is a node for any js script data
type JSNode struct {
	BaseNode
	Src    string
	Script js.Script
	Remote bool
	Depth  int
}

// Parse extracts the script information and arguments from the node and then calls parse on all it's children scripts
func (node *JSNode) Parse(ctx ParseNodeContext) error {
	hasSrc, srcAttr := GetAttr(node, "src")
	if hasSrc && node.firstChild != nil {
		return fmt.Errorf("error at node %s, can not have both a src value and a child text node", node)
	}
	if !hasSrc && node.firstChild == nil {
		return fmt.Errorf("error at node %s, node has neither a src element nore any child text, empty script nodes nod allowed", node)
	}

	hasRemote, _ := GetAttr(node, "remote")
	node.Remote = hasRemote
	if hasRemote && !hasSrc {
		return fmt.Errorf("error at node %s, can not specify remote without a src attribute", node)
	}
	if node.Remote {
		node.Src = srcAttr.Val
		return nil
	}

	script := ""
	if node.BaseNode.firstChild != nil {
		script = node.firstChild.Data()
		Detach(node.firstChild)
	}
	if hasSrc {
		node.Src = ctx.path + srcAttr.Val
		// TODO: OPTIM: it seems slow to re-read the same files over and over, perhaps we should have an abstraction that does some kind of caching
		scriptBytes, err := ioutil.ReadFile(node.Src)
		script = string(scriptBytes)
		if err != nil {
			return fmt.Errorf("error at node %s, %s", node, err)
		}
		ctx.path = getPath(node.Src)
	}

	tokens := js.LexScript(script)
	var err error
	node.Script, err = js.ParseTokens(tokens)
	if err != nil {
		return fmt.Errorf("error at node %s, %s", node, err)
	}

	// Add children
	for _, i := range node.Script.Imports {
		newNode := NewNode("script", JSType)
		attrs := []*html.Attribute{&html.Attribute{Key: "src", Val: i.Src}}
		if i.Remote {
			attrs = append(attrs, &html.Attribute{Key: "remote"})
		}
		newNode.SetAttrs(attrs)
		AppendChild(node, newNode)
	}

	node.Depth = ctx.depth
	ctx.depth++
	return node.BaseNode.Parse(ctx)
}

// Instance replaces props in a node with params
func (node *JSNode) Instance(ctx InstNodeContext) error {
	re := regexp.MustCompile(`\$[_a-zA-Z][_a-zA-Z0-9]*\$`)
	for i := 0; i < len(node.Script.Lines); i++ {
		line := &node.Script.Lines[i]
		for j := 0; j < len(line.Value); j++ {
			tok := &line.Value[j]
			// TODO: OPTIM: we only need to check template and value types
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
func (node *JSNode) Render() string {
	// fmt.Printf("node: %s %p\n", node, node)
	return "{" + node.Script.String() + "}"
}

// Clone creates a clone of the node
func (node *JSNode) Clone() Node {
	clone := &JSNode{
		BaseNode: BaseNode{data: node.Data(), attr: cloneAttrs(node.Attrs()), nType: node.Type(), visible: node.Visible()},
	}

	for _, child := range Children(node) {
		AppendChild(clone, child.Clone())
	}

	clone.Src = node.Src
	clone.Remote = node.Remote
	clone.Script = node.Script.Clone()

	return clone
}
