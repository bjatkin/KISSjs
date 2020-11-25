package main

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/net/html"
)

// JSNode is a node for any script data
type JSNode struct {
	BaseNode
	Src                 string
	Script              JSScript
	NoBundle, NoCompile bool
}

// Type returns JSType
func (node *JSNode) Type() NodeType {
	return JSType
}

// Parse extracts the script information and arguments from the node and then calls parse on all it's children scripts
func (node *JSNode) Parse(ctx NodeContext) error {
	hasSrc, srcAttr := GetAttr(node, "src")
	if hasSrc && node.firstChild != nil {
		return fmt.Errorf("error at node %s, can not have both a src value and a child text node", node)
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

func (node *JSNode) Render() string {
	if node.Src == "" {
		return ""
	}
	return "<script src=\"" + node.Src + "\"></script>"
}
