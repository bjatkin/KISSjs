package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	if len(os.Args) != 2 && len(os.Args) != 4 && len(os.Args) != 6 {
		printUsageMSG()
		return
	}

	entryFile := os.Args[1]
	output := strings.Split(entryFile, ".")[0] + "_compiled"
	if len(os.Args) == 4 {
		if os.Args[2] != "-o" {
			printUsageMSG()
			return
		}

		output = os.Args[3]
	}

	compiled, err := compileEntryFile(entryFile)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	err = renderHTMLTree(compiled, output)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

}

func printUsageMSG() {
	fmt.Printf("Usage: \n\n\tkiss entry [-o output] [-g globals]")
}

func parseEntryFile(file string) (*html.Node, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return html.Parse(data)
}

func parseComponentFile(file string) ([]*html.Node, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return html.ParseFragment(data, nil)
}

func compileEntryFile(file string) (htmlTree, error) {
	node, err := parseEntryFile(file)
	if err != nil {
		return htmlTree{}, err
	}

	tree, err := newHTMLTree(node, getPath(file))
	if err != nil {
		return tree, err
	}

	for _, comp := range tree.components {
		subTree, err := compileComponent(comp, getPath(file))
		if err != nil {
			return tree, err
		}
		err = tree.addSiblings(comp.node, children(subTree.findOne("body"))...)
		if err != nil {
			return tree, err
		}
		tree.delete(comp.node)

		tree.scripts = append(tree.scripts, subTree.scripts...)
		tree.styles = append(tree.styles, subTree.styles...)
	}

	return tree, nil
}

func compileComponent(comp componentNode, path string) (htmlTree, error) {
	nodes, err := parseComponentFile(path + comp.class.src)
	if err != nil {
		return htmlTree{}, err
	}

	root := newNode("componentRoot", html.ElementNode)
	for _, n := range nodes {
		root.AppendChild(n)
	}

	tree, err := newHTMLTree(root, getPath(path+comp.class.src))
	if err != nil {
		return tree, err
	}

	// Scope the CSS
	for _, node := range tree.nodeList() {
		addClass(node, comp.scope)
	}
	for _, style := range tree.styles {
		style.addClass(comp.scope)
	}

	// Compile sub components
	for _, subComp := range tree.components {
		subTree, err := compileComponent(subComp, getPath(path+comp.class.src))
		if err != nil {
			return tree, err
		}

		err = tree.addSiblings(subComp.node, children(subTree.findOne("body"))...)
		if err != nil {
			return tree, err
		}
		tree.delete(comp.node)

		tree.scripts = append(tree.scripts, subTree.scripts...)
		tree.styles = append(tree.styles, subTree.styles...)
	}

	for _, node := range tree.nodeList() {
		err = hydrate(node, getPath(comp.class.src), comp.props)
		if err != nil {
			return tree, err
		}
	}

	for _, style := range tree.styles {
		style.hydrate(comp.props)
	}

	for _, script := range tree.scripts {
		script.hydrate(comp.props)
	}

	tree.delete(tree.find("script")...)

	return tree, nil
}

func renderHTMLTree(tree htmlTree, output string) error {
	// Render the output.css file
	cssFile, err := os.Create(output + ".css")
	if err != nil {
		return err
	}
	cssRules := ""
	for _, rule := range tree.styles {
		cssRules += rule.String() + "\n"
	}
	cssFile.Write([]byte(cssRules))

	// Render the output.js file
	// TODO: this will bundle everything which is not really what we want
	//       we need to bundle only the component and main js but on the imports
	jsFile, err := os.Create(output + ".js")
	if err != nil {
		return err
	}
	jsScript := ""
	for _, script := range tree.scripts {
		jsScript += script.js + "\n"
	}
	jsFile.Write([]byte(jsScript))

	head := tree.findOne("head")
	if head == nil {
		return errors.New("html tree is missing a head node")
	}
	// Link the output.css file to the tree
	cssLink := newNode(
		"link",
		html.ElementNode,
		html.Attribute{Key: "rel", Val: "stylesheet"},
		html.Attribute{Key: "href", Val: output + ".css"},
	)
	err = tree.addChild(head, cssLink)
	if err != nil {
		return err
	}

	// Link all the output.js file to the tree
	jsLink := newNode(
		"script",
		html.ElementNode,
		html.Attribute{Key: "href", Val: output + ".js"},
	)
	err = tree.addChild(head, jsLink)
	if err != nil {
		return err
	}

	// Render the output.html file
	htmlFile, err := os.Create(output + ".html")
	return html.Render(htmlFile, tree.root)
}

func getPath(fileName string) string {
	last := strings.LastIndex(fileName, "/")
	if last < -1 {
		return ""
	}
	return fileName[:last+1]
}
