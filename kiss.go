package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
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

	ctx := NodeContext{
		path:       getPath(args.entry),
		Parameters: make(map[string][]Node),
	}
	err = root.Parse(ctx)
	if err != nil {
		fmt.Printf("There was an error parsing the structure: %s\n", err)
		return
	}

	err = Render(args.output, root)
	if err != nil {
		fmt.Printf("There was an error writing the output files, %s", err)
		return
	}
}

func parseEntryFile(file string) (Node, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	htmlRoot, err := html.Parse(data)
	if err != nil {
		return nil, err
	}
	if htmlRoot.FirstChild.Type == html.DoctypeNode {
		htmlRoot.FirstChild = htmlRoot.FirstChild.NextSibling
		htmlRoot.FirstChild.PrevSibling = nil
	}

	root := convertNodeTree(nil, htmlRoot)
	root.SetVisible(false)
	root = fragmentNodes(root)
	root = removeWhiteSpace(root)
	root = hoistImports(root)
	root = convertComponents(root)

	return root, nil
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
	root = removeWhiteSpace(root)
	root = fragmentNodes(root)
	root = hoistImports(root)
	root = convertComponents(root)

	ret := []Node{}
	for _, node := range root.Children() {
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
	for _, node := range root.Descendants() {
		if strings.ToLower(node.Data()) == "component" {
			children := node.Children()
			if len(children) > 0 {
				root := NewNode("root", BaseType)
				root.SetVisible(false)
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

func convertComponents(root Node) Node {
	tags := []string{}
	for _, node := range root.Descendants() {
		if strings.ToLower(node.Data()) == "component" {
			_, tagAttr := GetAttr(node, "tag")
			tags = append(tags, tagAttr.Val)
		}
		for _, tag := range tags {
			if strings.ToLower(node.Data()) == strings.ToLower(tag) {
				ToComponentNode(node)
			}
		}
	}

	return root
}

func removeWhiteSpace(root Node) Node {
	for _, node := range root.Descendants() {
		if len(strings.TrimSpace(node.Data())) == 0 {
			Detach(node)
		}
	}
	return root
}

func fragmentNodes(root Node) Node {
	re := regexp.MustCompile(`{[_a-zA-Z][_a-zA-Z0-9]*}`)
	for _, node := range root.Descendants() {
		matches := re.FindAllIndex([]byte(node.Data()), -1)
		if len(matches) == 0 {
			continue
		}
		newData := []string{}
		prevIndex := 0
		for _, match := range matches {
			newData = append(newData, node.Data()[prevIndex:match[0]])
			newData = append(newData, node.Data()[match[0]:match[1]])
			prevIndex = match[1]
		}
		newData = append(newData, node.Data()[prevIndex:])
		node.SetVisible(false)
		for _, data := range newData {
			new := NewNode(data, TextType)
			node.AppendChild(new)
		}
	}

	return root
}

const (
	JSFileType = iota
	HTMLFileType
	CSSFileType
)

// File represents a simple file
type File struct {
	Name    string
	Content string
	Type    int
}

func (file *File) WriteFile(dir string) error {
	ext := ".html"
	if file.Type == JSFileType {
		ext = ".js"
	}
	if file.Type == CSSFileType {
		ext = ".css"
	}

	f, err := os.Create(dir + "/" + file.Name + ext)
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(file.Content))

	return err
}

// Render takes a node and renders the full tree into an array of files
func Render(outputDir string, root Node) error {
	var head, body Node
	for _, desc := range root.Descendants() {
		if desc.Data() == "head" {
			head = desc
		}
		if desc.Data() == "body" {
			body = desc
		}
	}
	head.AppendChild(NewNode("link", BaseType, &html.Attribute{Key: "rel", Val: "stylesheet"}, &html.Attribute{Key: "href", Val: "bundle.css"}))
	body.AppendChild(NewNode("script", BaseType, &html.Attribute{Key: "src", Val: "bundle.js"}))

	html, otherFiles := root.Render(&File{
		Type: HTMLFileType,
		Name: "index",
	})

	i := 0
	for i < len(otherFiles) {
		base := otherFiles[i]
		j := i + 1
		for j < len(otherFiles) {
			other := otherFiles[j]
			if base.Name == other.Name &&
				base.Type == other.Type {
				base.Content += other.Content
				l := len(otherFiles) - 1
				otherFiles[j] = otherFiles[l]
				otherFiles = otherFiles[:l]
				continue
			}
			j++
		}
		i++
	}

	for _, file := range otherFiles {
		err := file.WriteFile(outputDir)
		if err != nil {
			return err
		}
	}
	return html.WriteFile(outputDir)
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
