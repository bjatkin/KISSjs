package main

import (
	"errors"
	"fmt"
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

func nodeIsWhiteSpace(node *html.Node) bool {
	if node.Type != html.TextNode {
		return false
	}
	if len(strings.TrimSpace(node.Data)) > 0 {
		return false
	}
	return true
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

	return n
}

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

func hydrate(node *html.Node, path string, props []prop) error {
	tree, err := newHTMLTree(node, path)
	if err != nil {
		return err
	}
	for _, prop := range props {
		key := "{" + prop.key + "}"
		for _, node := range tree.nodeList() {
			if prop.isSimple() {
				node.Data = strings.ReplaceAll(node.Data, key, prop.val[0].Data)
				for i := 0; i < len(node.Attr); i++ {
					attr := &node.Attr[i]
					attr.Val = strings.ReplaceAll(attr.Val, key, prop.val[0].Data)
				}
				continue
			}

			node.Data = strings.TrimSpace(node.Data)
			index := strings.Index(node.Data, key)
			for index >= 0 {
				new := []*html.Node{}
				if index > 0 {
					pre := clone(node)
					pre.Data = pre.Data[:index]
					new = append(new, pre)
				}

				for _, n := range prop.val {
					new = append(new, cloneDeep(n, nil, nil))
				}

				node.Data = node.Data[index+len(key):]
				err := tree.addSiblings(node, new...)
				if err != nil {
					return err
				}

				index = strings.Index(node.Data, key)
			}
		}
	}

	return nil
}

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

func generateScope(l int) string {
	ref := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ret := ""
	for i := 0; i < l; i++ {
		ret += string(ref[rand.Intn(len(ref))])
	}
	return ret
}

type importNode struct {
	node     *html.Node
	tag, src string
	scope    string
}

func newImportNode(node *html.Node) (importNode, error) {
	ret := importNode{node: node}
	if node.Data != "component" {
		return ret, fmt.Errorf("%s node is not a vaild import node", node.Data)
	}
	for _, attr := range node.Attr {
		if attr.Key == "src" {
			ret.src = attr.Val
		}
		if attr.Key == "tag" {
			ret.tag = attr.Val
		}
	}
	if ret.src == "" {
		return ret, errors.New("Missing src attribute")
	}
	if ret.tag == "" {
		return ret, errors.New("Missing tag attribute")
	}

	ret.scope = generateScope(6)

	return ret, nil
}

type componentNode struct {
	node  *html.Node
	props []prop
	class importNode
	scope string
	depth int
}

func newComponentNode(node *html.Node, class importNode) (componentNode, error) {
	ret := componentNode{node: node}

	n := node.Parent
	for n != nil {
		n = n.Parent
		ret.depth++
	}

	for _, attr := range node.Attr {
		ret.props = append(ret.props,
			prop{key: attr.Key, val: []*html.Node{newNode(attr.Val, html.TextNode)}},
		)
	}

	n = node.FirstChild
	for n != nil {
		if nodeIsWhiteSpace(n) {
			n = n.NextSibling
			continue
		}
		ret.props = append(ret.props,
			prop{
				key: n.Data,
				val: children(n),
			},
		)
		n = n.NextSibling
	}

	ret.scope = generateScope(6)
	ret.class = class

	return ret, nil
}
