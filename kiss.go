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

	// compiled, err := compileEntryFile(entryFile)
	// err = renderHTMLTree(compiled, output)

	root, err := compileFileNEW(entryFile)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	scripts, err := extractScriptsNEW(root, getPath(entryFile))
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	styles, err := extractCSSNEW(root)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	cleanTree(root)

	body := findOne(root, "body")
	if body == nil {
		fmt.Printf("Error: body node missing from compiled html document")
	}
	scriptNode := newNode("script", html.ElementNode, html.Attribute{Key: "src", Val: removePath(output) + ".js"})
	body.AppendChild(scriptNode)

	jsFile, err := os.Create(output + ".js")
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	jsScript := ""
	for _, script := range scripts {
		jsScript += "{" + script.js + "}\n"
	}
	jsFile.Write([]byte(jsScript))

	head := findOne(root, "head")
	if head == nil {
		fmt.Println("Error: head node missing from compiled html document")
		return
	}
	styleNode := newNode("link",
		html.ElementNode,
		html.Attribute{Key: "rel", Val: "stylesheet"},
		html.Attribute{Key: "href", Val: removePath(output) + ".css"},
	)
	head.AppendChild(styleNode)

	cssFile, err := os.Create(output + ".css")
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	cssData := ""
	for _, style := range styles {
		cssData += style.String() + "\n"
	}
	cssFile.Write([]byte(cssData))

	file, err := os.Create(output + ".html")
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	err = html.Render(file, root)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
}

func printUsageMSG() {
	fmt.Printf("Usage: \n\n\tkiss entry [-o output] [-g globals]")
}

func cleanTree(root *html.Node) {
	// Get all the import tags so we can find components
	importTags := []string{}
	for _, node := range listNodes(root) {
		if node.Data == "component" {
			ok, root := getAttr(node, "root")
			if !ok || root.Val == "false" {
				new := true
				_, newTag := getAttr(node, "tag")
				for _, tag := range importTags {
					if newTag.Val == tag {
						new = false
						break
					}
				}
				if new {
					importTags = append(importTags, newTag.Val)
				}
			}
		}
	}

	// Get all the component nodes
	componentNodes := []*html.Node{}
	for _, tag := range importTags {
		for _, node := range listNodes(root) {
			if strings.ToLower(node.Data) == strings.ToLower(tag) {
				componentNodes = append(componentNodes, node)
			}
		}
	}

	// Remove Component Nodes
	for _, component := range componentNodes {
		root := component.FirstChild
		for _, node := range children(root) {
			component.InsertBefore(detach(node), root)
		}
		component.RemoveChild(root)
		for _, node := range children(component) {
			component.Parent.InsertBefore(detach(node), component)
		}
		component.Parent.RemoveChild(component)
	}

	// Remove Import statements
	for _, node := range listNodes(root) {
		ok, root := getAttr(node, "root")
		if node.Data == "component" && (!ok || root.Val != "true") {
			node.Parent.RemoveChild(node)
		}
	}
}

func compileFileNEW(file string) (*html.Node, error) {
	root, err := inlineComponents(file, true)
	if err != nil {
		return root, err
	}

	// Get all the import tags so we can find components
	importTags := []string{}
	for _, node := range listNodes(root) {
		if node.Data == "component" {
			ok, root := getAttr(node, "root")
			if !ok || root.Val == "false" {
				new := true
				_, newTag := getAttr(node, "tag")
				for _, tag := range importTags {
					if newTag.Val == tag {
						new = false
						break
					}
				}
				if new {
					importTags = append(importTags, newTag.Val)
				}
			}
		}
	}

	// Get all the component nodes
	componentNodes := []*html.Node{}
	for _, tag := range importTags {
		for _, node := range listNodes(root) {
			if strings.ToLower(node.Data) == strings.ToLower(tag) {
				componentNodes = append(componentNodes, node)
			}
		}
	}

	// Instantiate the inlined components
	for _, node := range componentNodes {
		processComponent(node)
	}

	return root, err
}

func inlineComponents(file string, entry bool) (*html.Node, error) {
	root, err := parseEntryFile(file)
	if !entry {
		root = newNode("component", html.ElementNode, html.Attribute{Key: "root", Val: "true"})
		cNode, err := parseComponentFile(file)
		if err != nil {
			return root, err
		}
		for _, node := range cNode {
			root.AppendChild(detach(node))
		}
	}

	importNodes := []*html.Node{}
	for _, node := range listNodes(root) {
		if node.Data == "component" {
			ok, root := getAttr(node, "root")
			if !ok || root.Val == "false" {
				importNodes = append(importNodes, node)
			}
		}
	}

	for _, iNode := range importNodes {
		for _, node := range listNodes(root) {
			_, tag := getAttr(iNode, "tag")
			if strings.ToLower(node.Data) == strings.ToLower(tag.Val) {
				_, src := getAttr(iNode, "src")
				child, err := inlineComponents(getPath(file)+src.Val, false)
				if err != nil {
					return root, err
				}

				node.AppendChild(child)
			}
		}
	}

	return root, err
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

	root, err := html.ParseFragment(data, nil)
	if err != nil {
		return nil, err
	}

	head := findOne(root[0], "head")
	for _, node := range children(head) {
		escapeParent(node)
	}
	root[0].RemoveChild(head)

	body := findOne(root[0], "body")
	for _, node := range children(body) {
		escapeParent(node)
	}
	root[0].RemoveChild(body)

	return children(root[0]), nil
}

func processComponent(component *html.Node) {
	simple, complex := getParameters(component)

	for _, param := range simple {
		for _, node := range children(findOne(component, "style")) {
			hydrateNode(node, param, "\"@", "@\"")
		}
		for _, script := range children(findOne(component, "component")) {
			if script.Data != "script" {
				continue
			}
			for _, node := range children(script) {
				hydrateNode(node, param, "$", "$")
			}
		}
		for _, node := range listNodes(component.FirstChild) {
			if node.Parent != nil &&
				(node.Parent.Data == "script" || node.Parent.Data == "style") {
				continue
			}
			hydrateNode(node, param, "{", "}")
		}
	}

	for _, param := range complex {
		for _, node := range listNodes(component.FirstChild) {
			if node.Data == "script" || node.Data == "style" {
				continue
			}
			if node.Parent != nil &&
				(node.Parent.Data == "script" || node.Parent.Data == "style") {
				continue
			}
			hydrateNodeComplex(node, param, "{", "}")
		}
	}

	for _, node := range children(component) {
		if node.Data == "component" {
			continue
		}
		component.RemoveChild(node)
	}
}

func hydrateNode(node *html.Node, param simpleParameter, ss, ee string) {
	key := ss + param.name + ee
	node.Data = strings.ReplaceAll(node.Data, key, param.value)
	for i := 0; i < len(node.Attr); i++ {
		node.Attr[i].Key = strings.ReplaceAll(node.Attr[i].Key, key, param.value)
		node.Attr[i].Val = strings.ReplaceAll(node.Attr[i].Val, key, param.value)
	}
}

func hydrateNodeComplex(node *html.Node, param complexParameter, ss, ee string) {
	key := ss + param.name + ee
	index := strings.Index(node.Data, key)
	if index >= 0 {
		newNodes := []*html.Node{}
		if index > 0 {
			newNodes = append(newNodes, newNode(node.Data[:index], node.Type, node.Attr...))
		}
		for _, val := range param.value {
			newNodes = append(newNodes, cloneDeep(val, nil, nil))
		}

		for _, newNode := range newNodes {
			node.Parent.InsertBefore(detach(newNode), node)
		}
		node.Data = node.Data[index+len(key):]
		index = strings.Index(node.Data, key)
	}
}

type simpleParameter struct {
	name, value string
}

type complexParameter struct {
	name   string
	parent *html.Node
	value  []*html.Node
}

func getParameters(component *html.Node) ([]simpleParameter, []complexParameter) {
	simple := []simpleParameter{}
	complex := []complexParameter{}
	for _, attr := range component.Attr {
		simple = append(simple,
			simpleParameter{
				name:  attr.Key,
				value: attr.Val,
			},
		)
	}

	for _, node := range children(component) {
		if node.Data == "component" || nodeIsWhiteSpace(node) {
			continue
		}
		if len(children(node)) == 1 &&
			node.FirstChild.Type == html.TextNode {
			simple = append(simple,
				simpleParameter{
					name:  node.Data,
					value: node.FirstChild.Data,
				},
			)
			continue
		}
		complex = append(complex,
			complexParameter{
				name:   node.Data,
				parent: node,
				value:  children(node),
			},
		)
	}

	return simple, complex
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

	tree.sortComponents()
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
		for _, c := range subTree.components {
			c.depth++
		}
		tree.components = append(tree.components, subTree.components...)
	}

	// Remove Compiled Scripts
	scripts := tree.find("script")
	for _, script := range scripts {
		found, compile := getAttr(script, "compile")
		if found && compile.Val == "true" {
			tree.delete(script)
		}
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

	tree.sortComponents()

	for _, comp := range tree.components {
		changed := true
		for changed {
			updated, err := comp.hydrate()
			if err != nil {
				return tree, err
			}
			changed = changed || updated
		}
	}

	// Scope the CSS
	// for _, node := range tree.nodeList() {
	// 	addClass(node, comp.scope)
	// }
	// for _, style := range tree.styles {
	// 	style.addClass(comp.scope)
	// }

	// Compile sub components
	// tree.sortComponents()
	// for _, subComp := range tree.components {
	// 	subTree, err := compileComponent(subComp, getPath(path+comp.class.src))
	// 	if err != nil {
	// 		return tree, err
	// 	}

	// 	err = tree.addSiblings(subComp.node, children(subTree.findOne("body"))...)
	// 	if err != nil {
	// 		return tree, err
	// 	}
	// 	tree.delete(comp.node)

	// 	tree.scripts = append(tree.scripts, subTree.scripts...)
	// 	tree.styles = append(tree.styles, subTree.styles...)
	// 	for _, c := range subTree.components {
	// 		c.depth++
	// 	}
	// 	tree.components = append(tree.components, subTree.components...)
	// }

	// Hydrate Nodes
	// tree.sortComponents()
	// for _, subComp := range tree.components {
	// 	fmt.Printf("HERE")
	// 	for _, node := range tree.nodeList() {
	// 		err = hydrate(node, getPath(subComp.class.src), subComp.props)
	// 		if err != nil {
	// 			return tree, err
	// 		}
	// 	}

	// 	for _, style := range tree.styles {
	// 		style.hydrate(subComp.props)
	// 	}

	// 	for _, script := range tree.scripts {
	// 		script.hydrate(subComp.props)
	// 	}
	// }

	// tree.delete(tree.find("script")...)

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

func removePath(fileName string) string {
	last := strings.LastIndex(fileName, "/")
	return fileName[last+1:]
}
