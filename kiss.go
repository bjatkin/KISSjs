package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	args, err := parseArgs(os.Args)

	if err != nil {
		fmt.Printf("Usage:\n\tkiss entry [-o output] [-g globals]\n")
		return
	}

	root, outlined, err := compileFile(args.entry, args.globals)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	body := findOne(root, "body")
	if body == nil {
		fmt.Printf("Error: body node missing from compiled html document")
		return
	}
	head := findOne(root, "head")
	if head == nil {
		fmt.Printf("Error: head node missing from compiled html document")
		return
	}

	scripts, err := extractScripts(root, getPath(args.entry))
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	jsNodes, err := writeJS(args.output+".js", scripts)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	for _, node := range jsNodes {
		body.AppendChild(node)
	}

	styles, err := extractCSS(root)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	cssNode, err := writeCSS(args.output+".css", styles)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	head.AppendChild(cssNode)

	componentNodes, err := getComponentNodes(root)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	for _, outline := range outlined {
		scope := getAttr(outline, "scope")

		var div *html.Node
		for _, node := range componentNodes {
			rootScope := getAttr(node.FirstChild, "scope")
			if rootScope != nil && rootScope.Val == scope.Val {
				div = node.FirstChild.FirstChild
			}
		}

		scripts, err := extractScripts(outline, getPath(args.entry))
		if err != nil {
			fmt.Printf("Error: %s", err)
			return
		}
		outlineJSNodes, err := writeJS(getPath(args.output)+"/lazyComponents/"+scope.Val+".js", scripts)
		if err != nil {
			fmt.Printf("Error: %s", err)
			return
		}

		for _, node := range outlineJSNodes {
			// Add to the div
			src := getAttr(node, "src").Val
			div.Attr = append(div.Attr, html.Attribute{Key: "js", Val: src})
		}

		styles, err := extractCSS(outline)
		if err != nil {
			fmt.Printf("Error: %s", err)
			return
		}

		outlineCSSNode, err := writeCSS(getPath(args.output)+"/lazyComponents/"+scope.Val+".css", styles)
		if err != nil {
			fmt.Printf("Error: %s", err)
			return
		}
		if outlineCSSNode != nil {
			href := getAttr(outlineCSSNode, "href").Val
			div.Attr = append(div.Attr, html.Attribute{Key: "css", Val: href})
		}

		err = writeHTML(getPath(args.output)+"/lazyComponents/"+scope.Val+".html", outline)
		if err != nil {
			fmt.Printf("Error: %s", err)
			return
		}
		div.Attr = append(div.Attr, html.Attribute{Key: "src", Val: "lazyComponents/" + scope.Val + ".html"})
	}

	err = cleanTree(root)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	err = writeHTML(args.output+".html", root)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
}

func cleanTree(root *html.Node) error {
	componentNodes, err := getComponentNodes(root)
	if err != nil {
		return err
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
		root := getAttr(node, "root")
		if node.Data == "component" && (root != nil && root.Val != "false") {
			node.Parent.RemoveChild(node)
		}
	}

	return nil
}

func compileFile(entryFile, globalFile string) (*html.Node, []*html.Node, error) {
	root, err := parseEntryFile(entryFile)
	if err != nil {
		return nil, nil, err
	}
	globals, err := parseComponentFile(globalFile)
	if err != nil {
		return nil, nil, err
	}

	root = processGlobals(root, globals)

	root, err = inlineComponents(root, getPath(entryFile))
	if err != nil {
		return nil, nil, err
	}

	componentNodes, err := getComponentNodes(root)
	if err != nil {
		return nil, nil, err
	}

	// Instantiate the inlined components
	for _, node := range componentNodes {
		processComponent(node)
	}

	root, children, err := outlineComponents(root, getPath(entryFile))
	if err != nil {
		return nil, nil, err
	}

	return root, children, err
}

func processGlobals(root *html.Node, globals []*html.Node) *html.Node {
	globalRoot := newNode("root", html.ElementNode)
	for _, node := range globals {
		globalRoot.AppendChild(detach(node))
	}
	globalParams, globalComplexParams := getParameters(globalRoot)

	for _, param := range globalParams {
		for _, node := range children(findOne(root, "style")) {
			hydrateNode(node, param, "\"@", "@\"")
		}
		for _, script := range children(root) {
			if script.Data != "script" {
				continue
			}
			for _, node := range children(script) {
				hydrateNode(node, param, "$", "$")
			}
		}
		for _, node := range listNodes(root) {
			if node.Parent != nil &&
				(node.Parent.Data == "script" || node.Parent.Data == "style") {
				continue
			}
			hydrateNode(node, param, "{", "}")
		}
	}

	for _, param := range globalComplexParams {
		for _, node := range listNodes(root) {
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

	return root
}

func inlineComponents(root *html.Node, path string) (*html.Node, error) {
	importNodes, err := getImportNodes(root)
	if err != nil {
		return nil, err
	}

	for _, iNode := range importNodes {
		for _, node := range listNodes(root) {
			tag := getAttr(iNode, "tag")
			if strings.ToLower(node.Data) == strings.ToLower(tag.Val) {
				child := newNode("component", html.ElementNode, html.Attribute{Key: "root", Val: "true"})
				cNode := children(iNode)
				newPath := path
				src := getAttr(iNode, "src")
				if src != nil {
					var err error
					cNode, err = parseComponentFile(path + src.Val)
					if err != nil {
						return nil, err
					}
					newPath = getPath(path + src.Val)
				}
				for _, node := range cNode {
					child.AppendChild(detach(cloneDeep(node, nil, nil)))
				}
				child, err := inlineComponents(child, newPath)
				if err != nil {
					return root, err
				}

				node.AppendChild(child)
			}
		}
		for _, node := range children(iNode) {
			detach(node)
		}
	}

	return root, nil
}

func outlineComponents(root *html.Node, path string) (*html.Node, []*html.Node, error) {
	componentNodes, err := getComponentNodes(root)
	if err != nil {
		return nil, nil, err
	}

	outlined := []*html.Node{}
	scope := "KISS-" + generateScope(6)
	for _, node := range componentNodes {
		nobundle := getAttr(node, "nobundle")
		id := getAttr(node, "nobundleid")
		if nobundle != nil && nobundle.Val != "false" {
			cRoot := findOne(node, "component")
			addScope := true
			for _, attr := range cRoot.Attr {
				if attr.Key == "scope" {
					addScope = false
				}
			}
			if addScope {
				cRoot.Attr = append(cRoot.Attr, html.Attribute{Key: "scope", Val: scope})
			}

			new := detach(clone(cRoot))
			for _, child := range children(cRoot) {
				new.AppendChild(detach(child))
			}
			outlined = append(outlined, new)

			div := newNode("div", html.ElementNode, html.Attribute{Key: "id", Val: id.Val})
			cRoot.AppendChild(div)
		}
	}

	return root, outlined, nil
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
	nocompile := getAttr(component, "nocompile")
	if nocompile != nil && nocompile.Val != "false" {
		return
	}

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

func writeCSS(file string, styles []*cssRule) (*html.Node, error) {
	if len(styles) == 0 {
		return nil, nil
	}

	cssFile, err := os.Create(file)
	if err != nil {
		return nil, err
	}

	cssData := ""
	for _, style := range styles {
		cssData += style.String() + "\n"
	}
	_, err = cssFile.Write([]byte(cssData))

	styleNode := newNode("link",
		html.ElementNode,
		html.Attribute{Key: "rel", Val: "stylesheet"},
		html.Attribute{Key: "href", Val: removePath(file)},
	)

	return styleNode, err
}

func writeJS(file string, scripts []*jsSnipit) ([]*html.Node, error) {
	if len(scripts) == 0 {
		return nil, nil
	}
	jsFile, err := os.Create(file)
	if err != nil {
		return nil, err
	}

	nodes := []*html.Node{}
	jsScript := ""
	for _, script := range scripts {
		if script.noBundle && script.src != "" {
			nodes = append(nodes,
				newNode("script", html.ElementNode, html.Attribute{Key: "src", Val: script.src}),
			)
			continue
		}
		if script.js == "" {
			continue
		}
		jsScript += "{" + script.js + "}\n"
	}

	nodes = append(nodes,
		newNode("script", html.ElementNode, html.Attribute{Key: "src", Val: removePath(file)}),
	)

	_, err = jsFile.Write([]byte(jsScript))
	return nodes, err
}

func writeHTML(file string, root *html.Node) error {
	htmlFile, err := os.Create(file)
	if err != nil {
		return err
	}

	return html.Render(htmlFile, root)
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

func validImportNodes(nodes []*html.Node) error {
	for _, node := range nodes {
		tag := getAttr(node, "tag")
		src := getAttr(node, "src")
		child := node.FirstChild
		if tag == nil || tag.Val == "" {
			return errors.New("missing component tag or missing tag value")
		}
		if src != nil && child != nil {
			return errors.New("both a src and inner nodes provide on component, only one is supported at a time")
		}
	}
	return nil
}

func getImportNodes(root *html.Node) ([]*html.Node, error) {
	importNodes := []*html.Node{}
	for _, node := range listNodes(root) {
		if node.Data == "component" {
			root := getAttr(node, "root")
			if root == nil || root.Val != "true" {
				add := true
				for _, iNode := range importNodes {
					iTag := getAttr(iNode, "tag")
					tag := getAttr(node, "tag")
					if tag != nil && iTag.Val == tag.Val {
						add = false
					}
				}
				if add {
					importNodes = append(importNodes, node)
				}
			}
		}
	}

	if err := validImportNodes(importNodes); err != nil {
		return nil, err
	}

	return importNodes, nil
}

func getImportTags(root *html.Node) ([]string, error) {
	importNodes, err := getImportNodes(root)
	if err != nil {
		return []string{}, err
	}
	importTags := []string{}
	for _, node := range importNodes {
		tag := getAttr(node, "tag")
		importTags = append(importTags, tag.Val)
	}
	return importTags, nil
}

func getComponentNodes(root *html.Node) ([]*html.Node, error) {
	importNodes, err := getImportNodes(root)
	if err != nil {
		return nil, err
	}

	componentNodes := []*html.Node{}
	for _, tag := range importNodes {
		for _, node := range listNodes(root) {
			if strings.ToLower(node.Data) == strings.ToLower(getAttr(tag, "tag").Val) {
				componentNodes = append(componentNodes, node)
			}
		}
	}
	return componentNodes, nil
}
