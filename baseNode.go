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
	BaseType = NodeType(iota)
	ImportType
	ComponentType
	TextType
	JSType
	TSType
	CSSType
)

// ToNodeType Extracts a NodeType from an html.Node
func ToNodeType(node *html.Node) NodeType {
	if node.Type == html.TextNode {
		return TextType
	}
	data := strings.ToLower(node.Data)
	if data == "comp" {
		return ImportType
	}
	if data == "style" {
		return CSSType
	}
	if data == "script" {
		sType := getAttr(node, "type")
		if sType != nil && sType.Val == "text/typescript" {
			return TSType
		}
		return JSType
	}

	rel := getAttr(node, "rel")
	if data == "link" && rel != nil && rel.Val == "stylesheet" {
		return CSSType
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
		return "Javascript Node"
	case TSType:
		return "Typescript Node"
	case CSSType:
		return "CSS Style Node"
	default:
		return "Base Node"
	}
}

// Node is a interface for all objects that can behave like html nodes
type Node interface {
	String() string
	Parse(ParseNodeContext) error
	Instance(InstNodeContext) error
	Render() string
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

// TODO: why are path and depth lowercase but ImportTags is not?

// ParseNodeContext passes contextual infromation from parent to child nodes durring parsing
type ParseNodeContext struct {
	path       string
	ImportTags []ImportTag
	depth      int
}

// TODO: why is compScope lower case but parameters is not?

// InstNodeContext passes contextual infromation from parent to child nodes durring instancing
type InstNodeContext struct {
	componentScope string
	Parameters     map[string][]Node
}

// RenderNodeContext passes contextual informaiton from parent to child nodes durring rendering
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
func (ctx ParseNodeContext) Clone() ParseNodeContext {
	ret := ParseNodeContext{
		path: ctx.path,
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
	case TSType:
		return &TSNode{BaseNode: base}
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

// AppendChild adds a the new node as the last child of the parent node
func AppendChild(parent, child Node) {
	child.SetParent(parent)

	lastChild := parent.FirstChild()
	if lastChild != nil {
		for lastChild.NextSibling() != nil {
			lastChild = lastChild.NextSibling()
		}
	}

	child.SetPrevSibling(lastChild)
	child.SetNextSibling(nil)

	if lastChild != nil {
		lastChild.SetNextSibling(child)
		return
	}
	parent.SetFirstChild(child)
}

// InsertBefore adds the new node as a child directly before the specified child node
// and error is thrown if child is not a direct child of the base node
func InsertBefore(parent, child, new Node) error {
	check := parent.FirstChild()
	for check != nil {
		if check == child {
			prev := child.PrevSibling()
			if prev != nil {
				prev.SetNextSibling(new)
			} else {
				parent.SetFirstChild(new)
			}
			new.SetPrevSibling(prev)
			child.SetPrevSibling(new)

			new.SetParent(parent)
			new.SetNextSibling(child)

			return nil
		}
		check = check.NextSibling()
	}
	return fmt.Errorf("node %s is not a child of %s", child, parent)
}

// Children returns an array of all the parent nodes direct children
func Children(parent Node) []Node {
	child := parent.FirstChild()
	ret := []Node{}
	for child != nil {
		ret = append(ret, child)
		child = child.NextSibling()
	}
	return ret
}

// Descendants returns an array of all the base nodes decendents
func Descendants(root Node) []Node {
	return listDescendants(root, true)
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
	clone := NewNode(node.data, node.nType, cloneAttrs(node.attr)...)
	clone.SetVisible(node.Visible())

	for _, child := range Children(node) {
		AppendChild(clone, child.Clone())
	}

	return clone
}

// Parse builds the nodes structure and then calls parse on all it's child nodes
func (node *BaseNode) Parse(ctx ParseNodeContext) error {
	for _, child := range Children(node) {
		err := child.Parse(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Instance takes parameters from the node context and replaces template parameteres
func (node *BaseNode) Instance(ctx InstNodeContext) error {
	AddClass(node, ctx.componentScope)
	re := regexp.MustCompile(`{[_a-zA-Z][_a-zA-Z0-9]*}`)
	for _, attr := range node.Attrs() {
		matches := re.FindAll([]byte(attr.Val), -1)
		for _, match := range matches {
			param := ""
			pnode, ok := ctx.Parameters[string(match[1:len(match)-1])]
			if ok {
				if len(pnode) == 1 {
					param = pnode[0].Data()
				}
				if len(pnode) > 1 {
					return fmt.Errorf("error at node %s, tried to replace %s with multiple param nodes", node, match)
				}
				if len(pnode) == 1 && pnode[0].Type() != TextType {
					return fmt.Errorf("error at node %s, tried to replace %s with a non-text parameter", node, match)
				}
			}
			attr.Val = strings.ReplaceAll(attr.Val, string(match), param)
		}
	}

	for _, child := range Children(node) {
		err := child.Instance(ctx)
		if err != nil {
			return err
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

	for _, child := range Children(node) {
		ret += child.Render()
	}

	if node.Data() == "hr" ||
		node.Data() == "br" ||
		node.Data() == "link" {
		return ret
	}

	if node.visible {
		ret += "</" + node.Data() + ">"
	}
	return ret
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

// AddClass adds the class string to the nodes class attribute
func AddClass(node Node, class string) {
	classes := []string{}
	for _, attr := range node.Attrs() {
		if attr.Key == "class" {
			classes = strings.Split(attr.Val, " ")
		}
	}

	if len(classes) == 0 {
		node.SetAttrs(append(node.Attrs(), &html.Attribute{Key: "class"}))
	}

	for i := 0; i < len(classes); i++ {
		if class == classes[i] {
			return
		}
	}

	classes = append(classes, class)
	for _, attr := range node.Attrs() {
		if attr.Key == "class" {
			attr.Val = strings.Join(classes, " ")
			return
		}
	}
}

// FindNodes finds all child nodes of the root of a given NodeType
func FindNodes(root Node, nType NodeType) []Node {
	ret := []Node{}
	desc := Descendants(root)
	for i := 0; i < len(desc); i++ {
		if desc[i].Type() == nType {
			ret = append(ret, desc[i])
		}
	}
	return ret
}
