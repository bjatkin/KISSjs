package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/net/html"
)

// NodeType is an identifyer for different node types
type NodeType int

// The different types of nodes
const (
	TypeBaseNode = NodeType(iota)
	TypeJSNode
	TypeCSSNode
	TypeImportNode
	TypeParameterNode
	TypeComponentNode
	TypeRootComponentNode
)

// Node is a interface for all objects that can behave like html nodes
type Node interface {
	AppendChild(Node)
	InsertBefore(Node, Node) error
	Children() []Node
	ListNodes() []Node
	String() string
	Parse(ParseNodeContext) error
	Data() string
	SetData(string)
	Attrs() []*html.Attribute
	SetAttrs([]*html.Attribute)
	Parent() Node
	SetParent(Node)
	FirstChild() Node
	SetFirstChild(Node)
	PrevSibling() Node
	SetPrevSibling(Node)
	NextSibling() Node
	SetNextSibling(Node)
}

// ParseNodeContext passes contextual infromation from parent to child nodes
type ParseNodeContext struct {
	path               string
	componentScope     string
	ImportTags         map[string]Node
	SkipComponentCheck bool
}

// BaseNode is the most basic node
type BaseNode struct {
	parent, firstChild, prevSibling, nextSibling Node
	data                                         string
	attr                                         []*html.Attribute
}

// NewNode creates a new node
func NewNode(data string) Node {
	switch data {
	case "script":
		return &JSNode{BaseNode: BaseNode{data: data}}
	case "style":
		return &CSSNode{BaseNode: BaseNode{data: data}}
	case "component":
		return &ImportNode{BaseNode: BaseNode{data: data}}
	default:
		return &BaseNode{data: data}
	}
}

// ToKissNode converts an html node into a normal node
func ToKissNode(node *html.Node) Node {
	ret := NewNode(node.Data)
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

// AppendChild adds a the new node as the last child of the node
func (node *BaseNode) AppendChild(new Node) {
	new.SetParent(node)

	lastChild := node.firstChild
	if lastChild != nil {
		for lastChild.NextSibling() != nil {
			lastChild = lastChild.NextSibling()
		}
	}

	new.SetPrevSibling(lastChild)
	new.SetNextSibling(nil)

	if lastChild != nil {
		lastChild.SetNextSibling(new)
		return
	}
	node.SetFirstChild(new)
}

// InsertBefore adds the new node as a child directly before the specified child node
// and error is thrown if child is not a direct child of the base node
func (node *BaseNode) InsertBefore(new, child Node) error {
	check := node.FirstChild()
	for check != nil {
		if check == child {
			new.SetParent(node)
			new.SetPrevSibling(check.PrevSibling())
			new.SetNextSibling(check)
			return nil
		}
	}
	return fmt.Errorf("node %s is not a child of %s", child, node)
}

// Children returns an array of all the base nodes direct children
func (node *BaseNode) Children() []Node {
	child := node.FirstChild()
	ret := []Node{}
	for child != nil {
		ret = append(ret, child)
		child = child.NextSibling()
	}
	return ret
}

// ListNodes returns an array of all the base nodes decendents
func (node *BaseNode) ListNodes() []Node {
	return itterateNodes(node, true)
}

func itterateNodes(node Node, root bool) []Node {
	ret := []Node{node}

	if node.FirstChild() != nil {
		ret = append(ret, itterateNodes(node.FirstChild(), false)...)
	}

	if node.NextSibling() != nil && !root {
		ret = append(ret, itterateNodes(node.NextSibling(), false)...)
	}

	return ret
}

// Detach removes the node from it's parents and siblings and then patches the hole left behind
func Detach(node Node) Node {
	if node.Parent() != nil {
		if node.PrevSibling() == nil {
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

	return node
}

// Clone creates a deep copy of a node, but does not copy over the connections to the original parent and siblings
func Clone(node Node) Node {
	return cloneDeep(node, true)
}

func cloneDeep(node Node, root bool) Node {
	if node == nil {
		return nil
	}

	ret := NewNode(node.Data())
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
	ret.SetFirstChild(cloneDeep(node.FirstChild(), false))
	if !root {
		ret.SetNextSibling(cloneDeep(node.NextSibling(), false))
	}

	return ret
}

// Parse builds the nodes structure and then calls parse on all it's child nodes
func (node *BaseNode) Parse(ctx ParseNodeContext) error {
	if !ctx.SkipComponentCheck {
		for tag, root := range ctx.ImportTags {
			if strings.ToLower(node.Data()) == tag {
				compNode := ToComponentNode(node)
				compNode.AppendChild(Clone(root))
				return compNode.Parse(ctx)
			}
		}
	}
	ctx.SkipComponentCheck = false

	for _, child := range node.Children() {
		err := child.Parse(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Data returns the nodes data field
func (node *BaseNode) Data() string {
	return node.data
}

// SetData is used to set the nodes data field
func (node *BaseNode) SetData(data string) {
	node.data = data
}

// Attrs returns the attributes of the node
func (node *BaseNode) Attrs() []*html.Attribute {
	return node.attr
}

// SetAttrs sets the attributes of the node
func (node *BaseNode) SetAttrs(attrs []*html.Attribute) {
	node.attr = attrs
}

// Parent returns the parent of the node
func (node *BaseNode) Parent() Node {
	return node.parent
}

// SetParent sets the new parent of the node
func (node *BaseNode) SetParent(parent Node) {
	node.parent = parent
}

// FirstChild returns the first child of the node
func (node *BaseNode) FirstChild() Node {
	return node.firstChild
}

// SetFirstChild sets the new first child of the node
func (node *BaseNode) SetFirstChild(child Node) {
	node.firstChild = child
}

// PrevSibling returns the previous sibling of the node
func (node *BaseNode) PrevSibling() Node {
	return node.prevSibling
}

// SetPrevSibling sets the new previous sibling of the node
func (node *BaseNode) SetPrevSibling(sibling Node) {
	node.prevSibling = sibling
}

// NextSibling returns the next sibling of the node
func (node *BaseNode) NextSibling() Node {
	return node.nextSibling
}

// SetNextSibling sets the new next sibling of the node
func (node *BaseNode) SetNextSibling(sibling Node) {
	node.nextSibling = sibling
}

// String returns the nodes string representation
func (node *BaseNode) String() string {
	ret := "<" + node.data
	for _, attr := range node.attr {
		ret += " "
		if len(attr.Namespace) > 0 {
			ret += attr.Namespace + ":"
		}
		ret += attr.Key + "=\"" + attr.Val + "\""
	}
	ret += ">"
	if node.firstChild != nil {
		ret += "..."
	}
	ret += "</>"
	return ret
}

// JSNode is a node for any script data
type JSNode struct {
	BaseNode
	src                 string
	script              JSScript
	nobundle, nocompile bool
}

// Parse extracts the script information and arguments from the node and then calls parse on all it's children scripts
func (node *JSNode) Parse(ctx ParseNodeContext) error {
	hasSrc, srcAttr := GetAttr(&node.BaseNode, "src")
	if hasSrc && node.firstChild != nil {
		return fmt.Errorf("error at node %s, can not have both a src value and a child text node", node)
	}

	hasNoCompile, _ := GetAttr(&node.BaseNode, "nocompile")
	node.nocompile = hasNoCompile

	hasNoBundle, _ := GetAttr(&node.BaseNode, "nobundle")
	node.nobundle = hasNoBundle
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
		node.src = ctx.path + srcAttr.Val
		scriptBytes, err := ioutil.ReadFile(node.src)
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
		newNode := NewNode("script")
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

	ctx.path = getPath(node.src)
	ctx.SkipComponentCheck = true
	node.BaseNode.Parse(ctx)
	return nil
}

// CSSNode is a node for all style data
type CSSNode struct {
	BaseNode
	Rules []*CSSRule
	Scope string
}

// Parse extracts all css rules and applies the correct scope to them
func (node *CSSNode) Parse(ctx ParseNodeContext) error {
	node.Scope = ctx.componentScope
	// extract the css rules
	css := ""
	if node.FirstChild() != nil {
		css = node.FirstChild().Data()
		Detach(node.FirstChild())
	}

	rules, err := ParseCSS(css)
	if err != nil {
		return err
	}

	node.Rules = rules

	// apply the correct scope
	if ctx.componentScope != "" {
		for _, rule := range node.Rules {
			rule.AddClass(ctx.componentScope)
		}
	}

	return nil
}

// ImportNode is a node for all component definition
type ImportNode struct {
	BaseNode
	Tag           string
	Src           string
	ComponentRoot Node
}

// Parse validates the import node and builds all the related context nodes
func (node *ImportNode) Parse(ctx ParseNodeContext) error {
	hasTag, tagAttr := GetAttr(&node.BaseNode, "tag")
	if !hasTag {
		return fmt.Errorf("error at ndoe %s, import node must have a tag attribute", node)
	}
	hasSrc, srcAttr := GetAttr(&node.BaseNode, "src")
	if hasSrc && node.ComponentRoot != nil {
		return fmt.Errorf("error at node %s, can not have both a src value and a child node", node)
	}

	if hasSrc {
		node.Src = ctx.path + srcAttr.Val
		children, err := parseComponentFile(node.Src)
		if err != nil {
			return fmt.Errorf("error at node %s, there was an error parsing coponent src", err)
		}

		root := NewNode("root")
		for _, child := range children {
			root.AppendChild(child)
		}
		node.ComponentRoot = root
	}

	err := node.ComponentRoot.Parse(ctx)
	if err != nil {
		return err
	}

	ctx.path = getPath(node.Src)
	ctx.ImportTags[strings.ToLower(tagAttr.Val)] = node.ComponentRoot
	ctx.SkipComponentCheck = true

	return node.BaseNode.Parse(ctx)
}

// Parameter is a simple key value struct
type Parameter struct {
	Key, Val string
}

// ComponentNode is a node for all components that match imports nodes
type ComponentNode struct {
	BaseNode
	NobundleID          string
	NoBundle, NoCompile bool
}

// ToComponentNode converts any kiss node into a component type node
// Warning! this function should never be used on sibling or parent nodes,
// only use this function on child nodes or siblings child nodes
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
	return ret
}

// Parse uses the it's class to add a root component and then calls parse on all it's children
func (node *ComponentNode) Parse(ctx ParseNodeContext) error {
	fmt.Printf("THIS IS A COMPONENT %s\n", node)
	ctx.SkipComponentCheck = true
	node.BaseNode.Parse(ctx)
	return nil
}

// ParameterNode is a node for component parameteres (e.g. all child nodes except the rootComponentNode)
type ParameterNode struct {
	BaseNode
}

func (node *ParameterNode) Parse(ctx ParseNodeContext) error {
	return nil
}

// RootComponentNode is the container for the instance nodes of the parent component node
type RootComponentNode struct {
	BaseNode
}

func (node *RootComponentNode) Parse(ctx ParseNodeContext) error {
	return nil
}

// GetAttr returns the html attribute with the matching key if it exsists, it also returns true if the attribute was found and false otherwise
func GetAttr(node *BaseNode, key string) (bool, *html.Attribute) {
	for _, attr := range node.attr {
		if attr.Key == key {
			return true, attr
		}
	}
	return false, nil
}
