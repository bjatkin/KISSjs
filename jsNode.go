package main

import (
	"KISS/js"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// JSNode is a node for any script data
type JSNode struct {
	BaseNode
	Src    string
	Script js.Script
	Remote bool
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
		node.AppendChild(newNode)
	}

	return node.BaseNode.Parse(ctx)
}

// Instance replaces props in a node with params
func (node *JSNode) Instance(ctx InstNodeContext) error {
	re := regexp.MustCompile(`\$[_a-zA-Z][_a-zA-Z0-9]*\$`)
	for i := 0; i < len(node.Script.Lines); i++ {
		line := &node.Script.Lines[i]
		for j := 0; j < len(line.Value); j++ {
			tok := &line.Value[j]
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

// FindEntry locates all the entry points for the HTML, JS and CSS code in the tree
func (node *JSNode) FindEntry(ctx RenderNodeContext) RenderNodeContext {
	if !node.Visible() {
		Detach(node)
		return ctx
	}

	if node.Remote {
		ctx.files = ctx.files.Merge(&File{
			Name:    node.Src,
			Type:    JSFileType,
			Entries: []Node{node},
			Remote:  true,
		})
		Detach(node)
		return ctx
	}

	if ctx.callerType != JSType {
		ctx.files = ctx.files.Merge(&File{
			Name:    "bundle",
			Type:    JSFileType,
			Entries: []Node{node},
		})
		Detach(node)
	}

	ctx.callerType = JSType
	for _, node := range node.Children() {
		ctx = node.FindEntry(ctx)
	}

	return ctx
}

// Render converts a node into a textual representation
func (node *JSNode) Render() string {
	ret := "{" + node.Script.String() + "}"

	for _, child := range node.Children() {
		script := child.Render()
		ret = script + ret
	}

	return ret
}

// Clone clones a parse node context
func (node *JSNode) Clone() Node {
	clone := JSNode{
		BaseNode: BaseNode{data: node.Data(), attr: node.Attrs(), nType: node.Type(), visible: node.Visible()},
	}

	for _, child := range node.Children() {
		clone.AppendChild(child.Clone())
	}

	clone.Src = node.Src
	clone.Remote = node.Remote
	clone.Script = node.Script.Clone()

	return &clone
}
