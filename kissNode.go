package main

// TEST THIS PACKAGE

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/net/html"
)

type kissNodeType int

const (
	baseNode = kissNodeType(iota)
	kissJsNode
	kissCSSNode
	kissComponentNode
)

type kissNode interface {
	AppendChild(kissNode)
	InsertBefore(kissNode, kissNode) error
	Children() []kissNode
	ListNodes() []kissNode
	String() string
	Parse(kissNodeContext) error
	Type() kissNodeType
	Data() string
	SetData(string)
	Attrs() []*html.Attribute
	SetAttrs([]*html.Attribute)
	Parent() kissNode
	SetParent(kissNode)
	FirstChild() kissNode
	SetFirstChild(kissNode)
	PrevSibling() kissNode
	SetPrevSibling(kissNode)
	NextSibling() kissNode
	SetNextSibling(kissNode)
}

type kissNodeContext struct {
	path           string
	componentScope string
}

type node struct {
	parent, firstChild, prevSibling, nextSibling kissNode
	data                                         string
	attr                                         []*html.Attribute
}

func newKissNode(data string) kissNode {
	switch data {
	case "script":
		return &jsNode{node: node{data: data}}
	case "style":
		return &cssNode{node: node{data: data}}
	case "component":
		return &importNode{node: node{data: data}}
	default:
		return &node{data: data}
	}
}

func htmlNodeToKissNode(node *html.Node) kissNode {
	ret := newKissNode(node.Data)
	attrs := []*html.Attribute{}
	for _, attr := range node.Attr {
		attrs = append(attrs,
			&html.Attribute{
				Namespace: attr.Namespace,
				Key:       attr.Key,
				Val:       attr.Val},
		)
	}
	ret.SetAttrs(attrs)
	return ret
}

func (node *node) AppendChild(add kissNode) {
	add.SetParent(node)

	lastChild := node.firstChild
	for lastChild.NextSibling() != nil {
		lastChild = lastChild.NextSibling()
	}

	add.SetPrevSibling(lastChild)
	add.SetNextSibling(nil)
}

func (node *node) InsertBefore(add, child kissNode) error {
	check := node.FirstChild()
	for check != nil {
		if check == child {
			add.SetParent(node)
			add.SetPrevSibling(check.PrevSibling())
			add.SetNextSibling(check)
			return nil
		}
	}
	return fmt.Errorf("node %s is not a child of %s", child, node)
}

func (node *node) Children() []kissNode {
	child := node.FirstChild()
	ret := []kissNode{}
	for child != nil {
		ret = append(ret, child)
		child = child.NextSibling()
	}
	return ret
}

func (node *node) ListNodes() []kissNode {
	return itterateNodes(node, true)
}

func itterateNodes(node kissNode, root bool) []kissNode {
	ret := []kissNode{}

	if node.FirstChild() != nil {
		ret = append(ret, itterateNodes(node.FirstChild(), false)...)
	}

	if node.NextSibling() != nil && !root {
		ret = append(ret, itterateNodes(node.NextSibling(), false)...)
	}

	return ret
}

func kissDetach(node kissNode) {
	if node.Parent() != nil {
		if node.PrevSibling() == nil && node.NextSibling() != nil {
			node.Parent().SetFirstChild(node.NextSibling())
		}
		node.SetParent(nil)
	}
	if node.PrevSibling() != nil {
		node.PrevSibling().SetNextSibling(node.NextSibling())
	}
	if node.NextSibling() != nil {
		node.NextSibling().SetPrevSibling(node.PrevSibling())
	}
	node.SetNextSibling(nil)
	node.SetPrevSibling(nil)
}

func kissClone(node *node) kissNode {
	return kissCloneDeep(node, true)
}

func kissCloneDeep(node kissNode, root bool) kissNode {
	if node == nil {
		return nil
	}

	ret := newKissNode(node.Data())
	attrs := []*html.Attribute{}
	for _, attr := range node.Attrs() {
		attrs = append(attrs,
			&html.Attribute{
				Namespace: attr.Namespace,
				Key:       attr.Key,
				Val:       attr.Val},
		)
	}

	ret.SetAttrs(attrs)
	ret.SetFirstChild(kissCloneDeep(node.FirstChild(), false))
	if !root {
		ret.SetNextSibling(kissCloneDeep(node.NextSibling(), false))
	}

	return ret
}

func (node *node) Parse(ctx kissNodeContext) error {
	for _, child := range node.Children() {
		err := child.Parse(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (node *node) Data() string {
	return node.data
}

func (node *node) Type() kissNodeType {
	return baseNode
}

func (node *node) SetData(data string) {
	node.data = data
}

func (node *node) Attrs() []*html.Attribute {
	return node.attr
}

func (node *node) SetAttrs(attrs []*html.Attribute) {
	node.attr = attrs
}

func (node *node) Parent() kissNode {
	return node.parent
}

func (node *node) SetParent(parent kissNode) {
	node.parent = parent
}

func (node *node) FirstChild() kissNode {
	return node.firstChild
}

func (node *node) SetFirstChild(child kissNode) {
	node.firstChild = child
}

func (node *node) PrevSibling() kissNode {
	return node.prevSibling
}

func (node *node) SetPrevSibling(sibling kissNode) {
	node.prevSibling = sibling
}

func (node *node) NextSibling() kissNode {
	return node.nextSibling
}

func (node *node) SetNextSibling(sibling kissNode) {
	node.nextSibling = sibling
}

func (node *node) String() string {
	ret := "<" + node.data
	for _, attr := range node.attr {
		ret += " " + attr.Namespace + ":" + attr.Key + "=" + attr.Val
	}
	ret += ">"
	if node.firstChild != nil {
		ret += "..."
	}
	ret += "</" + node.data + ">"
	return ret
}

type jsNode struct {
	node
	src                 string
	script              kissJSScript
	nobundle, nocompile bool
}

func (node *jsNode) parse(ctx kissNodeContext) error {
	hasSrc, srcAttr := getKissAttr(&node.node, "src")
	if hasSrc && node.firstChild != nil {
		return fmt.Errorf("error at node %s, can not have both a src value and a child text node", node)
	}

	hasNoCompile, _ := getKissAttr(&node.node, "nocompile")
	node.nocompile = hasNoCompile

	hasNoBundle, _ := getKissAttr(&node.node, "nobundle")
	node.nobundle = hasNoBundle
	if hasNoBundle && !hasSrc {
		return fmt.Errorf("error at node %s, can not specify nobundle without a src attribute", node)
	}

	if hasNoBundle && hasNoCompile {
		return nil
	}

	script := ""
	if node.node.firstChild != nil {
		script = node.firstChild.Data()
		kissDetach(node.firstChild)
	}
	if hasSrc {
		node.src = srcAttr.Val
		scriptBytes, err := ioutil.ReadFile(ctx.path + node.src)
		script = string(scriptBytes)
		if err != nil {
			return fmt.Errorf("error at node %s, %s", node.String(), err)
		}
	}

	tokens := tokenizeJSScript(script)
	var err error
	node.script, err = parseJSTokens(tokens)
	if err != nil {
		return fmt.Errorf("error at node %s, %s", node.String(), err)
	}

	// Add children
	for _, i := range node.script.imports {
		newNode := newKissNode("script")
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

	node.node.Parse(ctx)
	return nil
}

type cssNode struct {
	node
	selector []string
	styles   []cssStyle
	scope    string
}

// Question, we need a scope here? Should we add it to the parse function call?
// Should we reach up into parents in the parse command? Does that make sense?
func (node *cssNode) Parse(ctx kissNodeContext) error {
	return nil
}

type importNode struct {
	node
	src string
}

type parameter struct {
	key, val string
}

type componentNode struct {
	node
	parameters          []parameter
	nobundleID          string
	nobundle, nocompile bool
}

type parameterNode struct {
	node
}

type rootComponentNode struct {
	node
}

func getKissAttr(node *node, key string) (bool, *html.Attribute) {
	for _, attr := range node.attr {
		if attr.Key == key {
			return true, attr
		}
	}
	return false, nil
}
