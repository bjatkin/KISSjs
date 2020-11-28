package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/net/html"
)

// JSNode is a node for any script data
type JSNode struct {
	BaseNode
	Src                 string
	Script              JSScript
	NoBundle, NoCompile bool
}

// Parse extracts the script information and arguments from the node and then calls parse on all it's children scripts
func (node *JSNode) Parse(ctx NodeContext) error {
	hasSrc, srcAttr := GetAttr(node, "src")
	if hasSrc && node.firstChild != nil {
		return fmt.Errorf("error at node %s, can not have both a src value and a child text node", node)
	}
	if !hasSrc && node.firstChild == nil {
		return fmt.Errorf("error at node %s, node has neither a src element nore any child text, empty script nodes nod allowed", node)
	}

	hasNoCompile, _ := GetAttr(node, "nocompile")
	node.NoCompile = hasNoCompile

	hasNoBundle, _ := GetAttr(node, "nobundle")
	node.NoBundle = hasNoBundle
	if hasNoBundle && !hasSrc {
		return fmt.Errorf("error at node %s, can not specify nobundle without a src attribute", node)
	}

	if hasNoBundle && hasNoCompile {
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

	tokens := tokenizeJSScript(script)
	var err error
	node.Script, err = parseJSTokens(tokens)
	if err != nil {
		return fmt.Errorf("error at node %s, %s", node, err)
	}
	node.Instance(ctx)
	if strings.Index(node.Script.String(), "Biggest") > 0 {
		fmt.Println(node.Script.String())

	}

	// Add children
	for _, i := range node.Script.imports {
		newNode := NewNode("script", JSType)
		attrs := []*html.Attribute{&html.Attribute{Key: "src", Val: i.src}}
		if i.nobundle {
			attrs = append(attrs, &html.Attribute{Key: "nobundle"})
		}
		if i.nocompile {
			attrs = append(attrs, &html.Attribute{Key: "nocompile"})
		}
		newNode.SetAttrs(attrs)
		node.AppendChild(newNode)
	}

	return node.BaseNode.Parse(ctx)
}

func (node *JSNode) Instance(ctx NodeContext) error {
	for i := 0; i < len(node.Script.lines); i++ {
		line := &node.Script.lines[i]
		for j := 0; j < len(line.value); j++ {
			tok := &line.value[j]
			for name, param := range ctx.Parameters {
				tok.value = strings.ReplaceAll(tok.value, "$"+name+"$", param[0].Data())
			}
		}
	}
	return nil
}

func (node *JSNode) FindEntry(ctx RenderNodeContext) RenderNodeContext {
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

func (node *JSNode) Render() string {
	ret := "{" + node.Script.String() + "}"

	for _, child := range node.Children() {
		script := child.Render()
		ret = script + ret
	}

	return ret
}

func (node *JSNode) Clone() Node {
	clone := JSNode{
		BaseNode: BaseNode{data: node.Data(), attr: node.Attrs(), nType: node.Type(), visible: node.Visible()},
	}

	for _, child := range node.Children() {
		clone.AppendChild(child.Clone())
	}

	clone.Src = node.Src
	clone.NoBundle = node.NoBundle
	clone.NoCompile = node.NoCompile
	clone.Script = node.Script.clone()

	return &clone
}
