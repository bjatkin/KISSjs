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

	root, err := parseEntryFile(args.entry)
	if err != nil {
		fmt.Printf("Unable to parse entry file %s: %s\n", args.entry, err)
		return
	}

	ctx := ParseNodeContext{
		path:       getPath(args.entry),
		ImportTags: make(map[string]Node),
	}
	err = root.Parse(ctx)
	if err != nil {
		fmt.Printf("There was an error parsing the structure: %s\n", err)
		return
	}

	for _, node := range root.ListNodes() {
		fmt.Println(node)
	}
}

func parseEntryFile(file string) (Node, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	root, err := html.Parse(data)
	if err != nil {
		return nil, err
	}

	conv := convertNodeTree(nil, root)
	return hoistImports(conv), nil
}

func parseComponentFile(file string) ([]Node, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	htmlRoot, err := html.ParseFragment(data, nil)
	if err != nil {
		return nil, err
	}

	head := findOne(htmlRoot[0], "head")
	for _, node := range children(head) {
		escapeParent(node)
	}
	htmlRoot[0].RemoveChild(head)

	body := findOne(htmlRoot[0], "body")
	for _, node := range children(body) {
		escapeParent(node)
	}
	htmlRoot[0].RemoveChild(body)

	root := convertNodeTree(nil, htmlRoot[0])
	hoist := hoistImports(root)

	ret := []Node{}
	for _, node := range hoist.Children() {
		ret = append(ret, node)
	}

	return ret, nil
}

func convertNodeTree(parent Node, node *html.Node) Node {
	ret := ToKissNode(node)
	ret.SetParent(parent)

	if node.FirstChild != nil {
		ret.SetFirstChild(convertNodeTree(ret, node.FirstChild))
	}

	if node.NextSibling != nil {
		sibling := convertNodeTree(parent, node.NextSibling)
		sibling.SetPrevSibling(ret)
		ret.SetNextSibling(sibling)
	}

	return ret
}

func hoistImports(root Node) Node {
	imports := []Node{}
	for _, node := range root.ListNodes() {
		if strings.ToLower(node.Data()) == "component" {
			children := node.Children()
			if len(children) > 0 {
				root := NewNode("root")
				for _, child := range node.Children() {
					root.AppendChild(Detach(child))
				}
				((node).(*ImportNode)).ComponentRoot = root
			}
			Detach(node)
			imports = append(imports, node)
		}
	}

	for _, node := range imports {
		for _, child := range root.Children() {
			node.AppendChild(Detach(child))
		}
		root.AppendChild(node)
	}

	return root
}

// func writeCSS(file string, styles []*CSSRule) (*html.Node, error) {
// 	if len(styles) == 0 {
// 		return nil, nil
// 	}

// 	cssFile, err := os.Create(file)
// 	if err != nil {
// 		return nil, err
// 	}

// 	cssData := ""
// 	for _, style := range styles {
// 		cssData += style.String() + "\n"
// 	}
// 	_, err = cssFile.Write([]byte(cssData))

// 	styleNode := newNode("link",
// 		html.ElementNode,
// 		html.Attribute{Key: "rel", Val: "stylesheet"},
// 		html.Attribute{Key: "href", Val: removePath(file)},
// 	)

// 	return styleNode, err
// }

// func writeJS(file string, scripts []*jsSnipit) ([]*html.Node, error) {
// 	if len(scripts) == 0 {
// 		return nil, nil
// 	}
// 	jsFile, err := os.Create(file)
// 	if err != nil {
// 		return nil, err
// 	}

// 	nodes := []*html.Node{}
// 	jsScript := ""
// 	for _, script := range scripts {
// 		if script.noBundle && script.src != "" {
// 			nodes = append(nodes,
// 				newNode("script", html.ElementNode, html.Attribute{Key: "src", Val: script.src}),
// 			)
// 			continue
// 		}
// 		if script.js == "" {
// 			continue
// 		}
// 		jsScript += "{" + script.js + "}\n"
// 	}

// 	nodes = append(nodes,
// 		newNode("script", html.ElementNode, html.Attribute{Key: "src", Val: removePath(file)}),
// 	)

// 	_, err = jsFile.Write([]byte(jsScript))
// 	return nodes, err
// }

// func writeHTML(file string, root *html.Node) error {
// 	htmlFile, err := os.Create(file)
// 	if err != nil {
// 		return err
// 	}

// 	return html.Render(htmlFile, root)
// }

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

// The proper structure for this using XML semantics is
// <component tag="hello" src="world.html">
//      <hello></hello> <-- this would get instanced
// </component>
// <hello></hello> <-- this would not

// This is not how I'm doing things, this would be an obnoxious way to build things

// SOLUTION 1| Process all the imports and components first
//      Plus| This is a super simple method that could be easily implemented
//      Minus| seems like it would still be messy and bug prone, also makes the system more complex
// SOLUTION 2| make children method an itterator/ generator
//      Plus| makes it simpler to modify the tree and still have everything work
//      Minus| might be tricky to implement correctly/ bug free
//      Minus| requires a reset method in addition to children which complicates the API
// SOLUTION 3| Try to get imports and components set up in the conversion step
//      Plus| simple and could match the desired structure as a pre-processing step
//      Plus| you can pass imports as part of a NodeContext
//      Minus| does not solve the problem of dynamically changing decendents

// I don't like solution 1
// maybe SOLUTION 3
