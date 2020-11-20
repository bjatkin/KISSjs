package main

type kissNode interface {
	appendChild(*node)
	insertBefore(*node, *node)
	detach()
	children()
	listNodes()
	clone()
}

type node struct {
	parent, firstChild, prevSibling, nextSibling kissNode
	data                                         string
}

func newKissNode(data string) kissNode {
	return &node{data: data}
}

func (node *node) appendChild(add *node) {

}

func (node *node) insertBefore(add, child *node) {

}

func (node *node) children() {

}

func (node *node) listNodes() {

}

func (node *node) detach() {

}

func (node *node) clone() {

}

type jsNode struct {
	node
	src                 string
	tokens              []jsToken
	nobundle, nocompile bool
}

func newJSNode() kissNode {
	ret := &jsNode{}
	ret.data = "script"
	return ret
}

type cssNode struct {
	node
	selector []string
	styles   []cssStyle
	scope    string
}

func newCSSNode() kissNode {
	ret := &jsNode{}
	ret.data = "style"
	return ret
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
