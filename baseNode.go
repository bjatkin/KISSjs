package main

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// NodeType is the type of the Node
type NodeType int

// The list of node types
const (
	BaseType = iota
	ImportType
	ComponentType
	TextType
	JSType
	CSSType
)

// ToNodeType Extracts a NodeType from an html.Node
func ToNodeType(node *html.Node) NodeType {
	if node.Type == html.TextNode {
		return TextType
	}
	data := strings.ToLower(node.Data)
	if data == "component" {
		return ImportType
	}
	if data == "style" {
		return CSSType
	}
	if data == "script" {
		return JSType
	}
	return BaseType
}

func (nType NodeType) String() string {
	switch nType {
	case ImportType:
		return "Import Node"
	case ComponentType:
		return "Component Node"
	case TextType:
		return "Text Node"
	case JSType:
		return "Java Script Node"
	case CSSType:
		return "CSS Style Node"
	default:
		return "Base Node"
	}
}

// Node is a interface for all objects that can behave like html nodes
type Node interface {
	AppendChild(Node)
	InsertBefore(Node, Node) error
	Children() []Node
	Descendants() []Node
	String() string
	Parse(NodeContext) error
	Instance(NodeContext) error
	Render() string
	FindEntry(RenderNodeContext) RenderNodeContext
	Type() NodeType
	Clone() Node
	Visible() bool
	SetVisible(bool)
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

// TODO: Consider removing Type() from the interface?
// TODO: Consider making the interface more narrow by removing antying not required. e.g. Children/ Descendants/ AppendChild etc.
// TODO: Could you remove the Parent() and parent from the node interface? This would strongly enforce the flow of the program

// NodeContext passes contextual infromation from parent to child nodes durring parsing
type NodeContext struct {
	path           string
	componentScope string
	ImportTags     []ImportTag
	Parameters     map[string][]Node
}

type RenderNodeContext struct {
	callerType NodeType
	files      FileList
}

// ImportTag represents an import tag
type ImportTag struct {
	tag  string
	root Node
	path string
}

// Clone clones a parse node context
func (ctx NodeContext) Clone() NodeContext {
	ret := NodeContext{
		path:           ctx.path,
		componentScope: ctx.componentScope,
		Parameters:     make(map[string][]Node),
	}

	for _, tag := range ctx.ImportTags {
		ret.ImportTags = append(ret.ImportTags,
			ImportTag{
				tag:  tag.tag,
				root: tag.root,
				path: tag.path,
			},
		)
	}

	for param, node := range ctx.Parameters {
		ret.Parameters[param] = node
	}
	return ret
}

// BaseNode is the most basic node
type BaseNode struct {
	parent, firstChild, prevSibling, nextSibling Node
	nType                                        NodeType
	visible                                      bool
	data                                         string
	attr                                         []*html.Attribute
}

// NewNode creates a new node
func NewNode(data string, nType NodeType, attr ...*html.Attribute) Node {
	newAttr := []*html.Attribute{}
	for _, old := range attr {
		newAttr = append(newAttr,
			&html.Attribute{
				Namespace: old.Namespace,
				Key:       old.Key,
				Val:       old.Val,
			},
		)
	}

	base := BaseNode{data: data, attr: newAttr, visible: true, nType: nType}

	switch nType {
	case JSType:
		return &JSNode{BaseNode: base}
	case CSSType:
		return &CSSNode{BaseNode: base}
	case ImportType:
		base.visible = false
		return &ImportNode{BaseNode: base}
	case ComponentType:
		base.visible = false
		return &ComponentNode{BaseNode: base}
	case TextType:
		return &TextNode{BaseNode: base}
	default:
		return &base
	}
}

// ToKissNode converts an html node into a normal node
func ToKissNode(node *html.Node) Node {
	ret := NewNode(node.Data, ToNodeType(node))
	if node.Type == html.CommentNode {
		ret.SetVisible(false)
	}

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

// Descendants returns an array of all the base nodes decendents
func (node *BaseNode) Descendants() []Node {
	return listDescendants(node, true)
}

// listDescendants
func listDescendants(node Node, root bool) []Node {
	ret := []Node{node}

	if node.FirstChild() != nil {
		ret = append(ret, listDescendants(node.FirstChild(), false)...)
	}

	if node.NextSibling() != nil && !root {
		ret = append(ret, listDescendants(node.NextSibling(), false)...)
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
func (node *BaseNode) Clone() Node {
	clone := NewNode(node.data, node.nType, node.attr...)
	clone.SetVisible(node.Visible())

	for _, child := range node.Children() {
		clone.AppendChild(child.Clone())
	}

	return clone
}

// Parse builds the nodes structure and then calls parse on all it's child nodes
func (node *BaseNode) Parse(ctx NodeContext) error {
	for _, child := range node.Children() {
		err := child.Instance(ctx)
		if err != nil {
			return err
		}
		err = child.Parse(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Instance takes parameters from the node context and replaces template parameteres
func (node *BaseNode) Instance(ctx NodeContext) error {
	added := false
	for _, attr := range node.Attrs() {
		if attr.Key == "class" {
			attr.Val += " " + ctx.componentScope
			added = true
		}
	}

	if !added {
		node.SetAttrs(append(node.Attrs(), &html.Attribute{Key: "class", Val: ctx.componentScope}))
	}

	re := regexp.MustCompile(`{[_a-zA-Z][_a-zA-Z0-9]*}`)
	for _, attr := range node.Attrs() {
		matches := re.FindAll([]byte(attr.Val), -1)
		for _, match := range matches {
			node, ok := ctx.Parameters[string(match[1:len(match)-1])]
			if ok {
				if len(node) != 1 || node[0].Type() != TextType {
					return fmt.Errorf("error at node %s, tried to replace %s with a non-text parameter", node, match)
				}
				attr.Val = strings.ReplaceAll(attr.Val, string(match), node[0].Data())
			}
		}
	}

	return nil
}

// Render converts a node into a textual representation
func (node *BaseNode) Render() string {
	ret := ""
	if node.visible {
		ret += "<" + node.Data()
		for _, attr := range node.attr {
			if len(attr.Val) == 0 {
				continue
			}
			ret += " "
			if len(attr.Namespace) > 0 {
				ret += attr.Namespace + ":"
			}
			ret += attr.Key + "=\"" + attr.Val + "\""
		}
		ret += ">"
	}

	for _, child := range node.Children() {
		ret += child.Render()
	}

	if node.visible {
		ret += "</" + node.Data() + ">"
	}
	return ret
}

func (node *BaseNode) FindEntry(ctx RenderNodeContext) RenderNodeContext {
	for _, child := range node.Children() {
		ctx.callerType = BaseType
		ctx = child.FindEntry(ctx)
	}
	return ctx
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

// Visible indicates if a node will be rendered or not
func (node *BaseNode) Visible() bool {
	return node.visible
}

// SetVisible sets the visibility of the node
func (node *BaseNode) SetVisible(set bool) {
	node.visible = set
}

// Type returns the type of the node
func (node *BaseNode) Type() NodeType {
	return node.nType
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

// GetAttr returns the html attribute with the matching key if it exsists, it also returns true if the attribute was found and false otherwise
func GetAttr(node Node, key string) (bool, *html.Attribute) {
	for _, attr := range node.Attrs() {
		if attr.Key == key {
			return true, attr
		}
	}
	return false, nil
}
