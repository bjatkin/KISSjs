package main

import (
	"math/rand"
	"strings"

	"golang.org/x/net/html"
)

// Used
func newNode(data string, nodeType html.NodeType, attr ...html.Attribute) *html.Node {
	return &html.Node{
		Data: data,
		Type: nodeType,
		Attr: attr,
	}
}

func nodeIsWhiteSpace(node *html.Node) bool {
	if node.Type != html.TextNode {
		return false
	}
	if len(strings.TrimSpace(node.Data)) > 0 {
		return false
	}
	return true
}

// Used
// Does some weird ordering stuff...
func escapeParent(node *html.Node) *html.Node {
	if node.Parent == nil || node.Parent.Parent == nil {
		return node
	}
	node.Parent.Parent.InsertBefore(detach(node), node.Parent)
	return node
}

// Used
func detach(node *html.Node) *html.Node {
	if node.PrevSibling == nil {
		if node.Parent != nil {
			node.Parent.FirstChild = node.NextSibling
		}
	} else {
		node.PrevSibling.NextSibling = node.NextSibling
	}
	if node.NextSibling == nil {
		if node.Parent != nil {
			node.Parent.LastChild = node.PrevSibling
		}
	} else {
		node.NextSibling.PrevSibling = node.PrevSibling
	}

	node.NextSibling = nil
	node.PrevSibling = nil
	node.Parent = nil
	return node
}

// Used
func cloneDeep(n *html.Node, parent *html.Node, prev *html.Node) *html.Node {
	if n == nil {
		return nil
	}

	ret := clone(n)
	ret.FirstChild = cloneDeep(n.FirstChild, ret, nil)
	ret.Parent = parent
	ret.PrevSibling = prev
	ret.NextSibling = cloneDeep(n.NextSibling, parent, ret)
	if ret.NextSibling == nil && parent != nil {
		parent.LastChild = ret
	}

	return ret
}

// Used
func clone(node *html.Node) *html.Node {
	ret := newNode(node.Data, node.Type)
	ret.DataAtom = node.DataAtom
	ret.Namespace = node.Namespace

	retAttr := []html.Attribute{}
	for _, attr := range node.Attr {
		retAttr = append(retAttr,
			html.Attribute{
				Namespace: attr.Namespace,
				Key:       attr.Key,
				Val:       attr.Val,
			},
		)
	}

	ret.Attr = retAttr
	return ret
}

// Used
func find(root *html.Node, query string) []*html.Node {
	ret := []*html.Node{}
	for _, node := range listNodes(root) {
		if strings.ToLower(node.Data) == strings.ToLower(query) {
			ret = append(ret, node)
		}
	}
	return ret
}

// Used
func findOne(root *html.Node, query string) *html.Node {
	for _, node := range listNodes(root) {
		if strings.ToLower(node.Data) == strings.ToLower(query) {
			return node
		}
	}
	return nil
}

// Used
func children(node *html.Node) []*html.Node {
	ret := []*html.Node{}
	if node == nil {
		return ret
	}
	n := node.FirstChild
	for n != nil {
		ret = append(ret, n)
		n = n.NextSibling
	}
	return ret
}

// Used
func getAttr(node *html.Node, key string) (bool, *html.Attribute) {
	for i := 0; i < len(node.Attr); i++ {
		if node.Attr[i].Key == key {
			return true, &node.Attr[i]
		}
	}
	return false, &html.Attribute{}
}

// func hydrate(node *html.Node, path string, props []prop) (bool, error) {
// 	tree, err := newHTMLTree(node, path)
// 	if err != nil {
// 		return false, err
// 	}
// 	changed := false
// 	for _, prop := range props {
// 		key := "{" + prop.key + "}"
// 		for _, node := range tree.nodeList() {
// 			if prop.isSimple() {
// 				old := node.Data
// 				node.Data = strings.ReplaceAll(node.Data, key, prop.val[0].Data)
// 				changed = changed || (old == node.Data)

// 				for i := 0; i < len(node.Attr); i++ {
// 					attr := &node.Attr[i]
// 					old := attr.Val
// 					attr.Val = strings.ReplaceAll(attr.Val, key, prop.val[0].Data)
// 					changed = changed || (old == attr.Val)
// 				}
// 				continue
// 			}

// 			node.Data = strings.TrimSpace(node.Data)
// 			index := strings.Index(node.Data, key)
// 			for index >= 0 {
// 				changed = true
// 				new := []*html.Node{}
// 				if index > 0 {
// 					pre := clone(node)
// 					pre.Data = pre.Data[:index]
// 					new = append(new, pre)
// 				}

// 				for _, n := range prop.val {
// 					new = append(new, cloneDeep(n, nil, nil))
// 				}

// 				node.Data = node.Data[index+len(key):]
// 				err := tree.addSiblings(node, new...)
// 				if err != nil {
// 					return false, err
// 				}

// 				index = strings.Index(node.Data, key)
// 			}
// 		}
// 	}

// 	return changed, nil
// }

// Used
func addClass(node *html.Node, class string) {
	added := false
	for _, attr := range node.Attr {
		if attr.Key == "class" {
			if strings.Index(attr.Val, class) < 0 {
				attr.Val += " " + class
			}
			added = true
		}
	}
	if !added {
		node.Attr = append(
			node.Attr,
			html.Attribute{Key: "class", Val: class},
		)
	}
}

// Used
// Keep in mind that this will list all the siblings of root as well
func listNodes(root *html.Node) []*html.Node {
	ret := []*html.Node{root}

	if root.FirstChild != nil {
		ret = append(ret, listNodes(root.FirstChild)...)
	}

	if root.NextSibling != nil {
		ret = append(ret, listNodes(root.NextSibling)...)
	}

	return ret
}

// Used
func generateScope(l int) string {
	ref := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ret := ""
	for i := 0; i < l; i++ {
		ret += string(ref[rand.Intn(len(ref))])
	}
	return ret
}

// type importNode struct {
// 	node     *html.Node
// 	tag, src string
// 	scope    string
// }

// func newImportNode(node *html.Node) (importNode, error) {
// 	ret := importNode{node: node}
// 	if node.Data != "component" {
// 		return ret, fmt.Errorf("%s node is not a vaild import node", node.Data)
// 	}
// 	for _, attr := range node.Attr {
// 		if attr.Key == "src" {
// 			ret.src = attr.Val
// 		}
// 		if attr.Key == "tag" {
// 			ret.tag = attr.Val
// 		}
// 	}
// 	if ret.src == "" {
// 		return ret, errors.New("Missing src attribute")
// 	}
// 	if ret.tag == "" {
// 		return ret, errors.New("Missing tag attribute")
// 	}

// 	ret.scope = generateScope(6)

// 	return ret, nil
// }

// type componentNode struct {
// 	node  *html.Node
// 	tree  *htmlTree
// 	props []prop
// 	class importNode
// 	scope string
// 	depth int
// }

// func newComponentNode(node *html.Node, class importNode) (componentNode, error) {
// 	ret := componentNode{node: node}

// 	n := node.Parent
// 	for n != nil {
// 		n = n.Parent
// 		ret.depth++
// 	}

// 	for _, attr := range node.Attr {
// 		ret.props = append(ret.props,
// 			prop{key: attr.Key, val: []*html.Node{newNode(attr.Val, html.TextNode)}},
// 		)
// 	}

// 	n = node.FirstChild
// 	for n != nil {
// 		if nodeIsWhiteSpace(n) {
// 			n = n.NextSibling
// 			continue
// 		}
// 		ret.props = append(ret.props,
// 			prop{
// 				key: n.Data,
// 				val: children(n),
// 			},
// 		)
// 		n = n.NextSibling
// 	}

// 	ret.scope = generateScope(6)
// 	ret.class = class

// 	return ret, nil
// }

// func (comp *componentNode) components(path string) ([]componentNode, error) {
// 	ret := []componentNode{}
// 	file := path + comp.class.src
// 	cNodes, err := parseComponentFile(file)
// 	root := newNode("ComponentRoot", html.ElementNode)
// 	if err != nil {
// 		return ret, err
// 	}

// 	tree, err := newHTMLTree(root, getPath(file))
// 	if err != nil {
// 		return ret, err
// 	}
// 	tree.addChildren(root, cNodes...)
// 	comp.tree = &tree

// 	for _, c := range tree.components {
// 		components, err := c.components(getPath(file))
// 		if err != nil {
// 			return ret, err
// 		}
// 		for _, add := range components {
// 			add.depth++
// 			ret = append(ret, add)
// 		}
// 	}

// 	return ret, err
// }

// func (comp *componentNode) hydrate() (bool, error) {
// 	changed := false
// 	for _, node := range listNodes(comp.tree.root) {
// 		c, err := hydrate(node, getPath(comp.tree.path), comp.props)
// 		changed = changed || c
// 		if err != nil {
// 			return changed, err
// 		}
// 	}
// 	return false, nil
// }
