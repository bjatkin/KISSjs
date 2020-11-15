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
