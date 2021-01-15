package main

import (
	"math/rand"
	"strings"

	"golang.org/x/net/html"
)

func newNode(data string, nodeType html.NodeType, attr ...html.Attribute) *html.Node {
	return &html.Node{
		Data: data,
		Type: nodeType,
		Attr: attr,
	}
}

func escapeParent(node *html.Node) *html.Node {
	if node.Parent == nil || node.Parent.Parent == nil {
		return node
	}
	node.Parent.Parent.InsertBefore(detach(node), node.Parent)
	return node
}

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

func findOne(root *html.Node, query string) *html.Node {
	for _, node := range listNodes(root) {
		if strings.ToLower(node.Data) == strings.ToLower(query) {
			return node
		}
	}
	return nil
}

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

func getAttr(node *html.Node, key string) *html.Attribute {
	for i := 0; i < len(node.Attr); i++ {
		if node.Attr[i].Key == key {
			return &node.Attr[i]
		}
	}
	return nil
}

func cloneAttrs(attrs []*html.Attribute) []*html.Attribute {
	var ret []*html.Attribute
	for _, attr := range attrs {
		ret = append(ret,
			&html.Attribute{
				Namespace: attr.Namespace,
				Key:       attr.Key,
				Val:       attr.Val,
			},
		)
	}
	return ret
}

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

func randomID(l int) string {
	ref := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ret := ""
	for i := 0; i < l; i++ {
		ret += string(ref[rand.Intn(len(ref))])
	}
	return ret
}
