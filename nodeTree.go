package main

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
)

type htmlTree struct {
	root       *html.Node
	imports    []importNode
	components []componentNode
	styles     []*cssRule
	scripts    []*jsSnipit
	path       string
}

func newHTMLTree(root *html.Node, path string) (htmlTree, error) {
	ret := htmlTree{
		root: root,
		path: path,
	}
	err := ret.collectImportNodes()
	if err != nil {
		return ret, err
	}
	err = ret.collectComponents()
	if err != nil {
		return ret, err
	}

	err = ret.collectScripts()
	if err != nil {
		return ret, err
	}

	err = ret.collectStyle()
	return ret, err
}

func (tree *htmlTree) nodeList() []*html.Node {
	return listNodes(tree.root)
}

func (tree *htmlTree) find(query string) []*html.Node {
	ret := []*html.Node{}
	for _, n := range tree.nodeList() {
		if n.Data == strings.ToLower(query) {
			ret = append(ret, n)
		}
	}

	return ret
}

func (tree *htmlTree) findOne(query string) *html.Node {
	ret := tree.find(query)
	if len(ret) > 0 {
		return ret[0]
	}
	return nil
}

// func (tree *htmlTree) itter(node *html.Node) []*html.Node {
// 	ret := []*html.Node{node}

// 	if node.FirstChild != nil {
// 		ret = append(ret, tree.itter(node.FirstChild)...)
// 	}

// 	if node.NextSibling != nil {
// 		ret = append(ret, tree.itter(node.NextSibling)...)
// 	}

// 	return ret
// }

func (tree *htmlTree) addChild(parent, child *html.Node) error {
	if !tree.nodeInTree(parent) {
		return errors.New("Parent node not in html tree")
	}
	parent.AppendChild(detach(child))
	return nil
}

func (tree *htmlTree) addChildren(parent *html.Node, children ...*html.Node) error {
	for _, child := range children {
		err := tree.addChild(parent, child)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tree *htmlTree) addSibling(oldNode, newNode *html.Node) error {
	if !tree.nodeInTree(oldNode) {
		return errors.New("OldNode not in the html tree")
	}

	newNode = detach(newNode)
	oldNode.Parent.InsertBefore(newNode, oldNode)
	return nil
}

func (tree *htmlTree) addSiblings(oldNode *html.Node, newNodes ...*html.Node) error {
	for _, node := range newNodes {
		err := tree.addSibling(oldNode, node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tree *htmlTree) delete(nodes ...*html.Node) error {
	for _, node := range nodes {
		if !tree.nodeInTree(node) {
			return errors.New("Child node does not belong to this htmlTree")
		}
	}
	for _, node := range nodes {
		node.Parent.RemoveChild(node)
	}
	return nil
}

func (tree *htmlTree) nodeInTree(node *html.Node) bool {
	for _, n := range tree.nodeList() {
		if n == node {
			return true
		}
	}
	return false
}

func (tree *htmlTree) collectImportNodes() error {
	imports := tree.find("component")
	for _, i := range imports {
		iNode, err := newImportNode(i)
		if err != nil {
			return err
		}
		tree.imports = append(tree.imports, iNode)
	}
	return nil
}

func (tree *htmlTree) collectComponents() error {
	if len(tree.imports) == 0 {
		tree.collectImportNodes()
	}

	components := []componentNode{}
	for _, i := range tree.imports {
		comps := tree.find(i.tag)
		for _, c := range comps {
			cNode, err := newComponentNode(c, i)
			if err != nil {
				return err
			}
			components = append(components, cNode)
			childNodes, err := cNode.components(tree.path)
			if err != nil {
				return err
			}
			components = append(components, childNodes...)
		}
	}
	tree.components = components

	return nil
}

func (tree *htmlTree) sortComponents() {
	max := 0
	for _, c := range tree.components {
		if c.depth > max {
			max = c.depth
		}
	}
	sortedComponents := []componentNode{}
	for i := max; i >= 0; i-- {
		for _, comp := range tree.components {
			if comp.depth == i {
				sortedComponents = append(
					sortedComponents,
					comp,
				)
			}
		}
	}

	tree.components = sortedComponents
}

func (tree *htmlTree) collectStyle() error {
	styleNode := tree.findOne("style")
	if styleNode == nil {
		return nil
	}

	if styleNode.FirstChild != nil {
		rules, err := cssFromNode(styleNode.FirstChild)
		if err != nil {
			return err
		}

		tree.styles = append(tree.styles, rules...)
	}
	return nil
}

func (tree *htmlTree) collectScripts() error {
	scriptNodes := tree.find("script")
	for _, node := range scriptNodes {
		snipit, err := jsFromNode(node, 0, tree.path)
		if err != nil {
			return err
		}
		snipit.sortImports()

		tree.scripts = append(tree.scripts, &snipit)
		tree.scripts = append(tree.scripts, snipit.imports...)
	}
	return nil
}
