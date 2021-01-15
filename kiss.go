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
		fmt.Printf("Usage:\n\tkiss entry [-o output] [-g globals] [-v view_location]\n")
		return
	}

	globals := make(map[string][]Node)
	if args.globals != "" {
		comps, err := parseComponentFile(args.globals)
		if err != nil {
			fmt.Printf("Unable to parse the global args file %s: %s\n", args.globals, err)
			return
		}
		for _, comp := range comps {
			globals[strings.ToLower(comp.Data())] = Children(comp)
		}
	}

	root, err := parseEntryFile(args.entry)
	if err != nil {
		fmt.Printf("Unable to parse entry file %s: %s\n", args.entry, err)
		return
	}

	pctx := ParseNodeContext{
		path: getPath(args.entry),
	}
	err = root.Parse(pctx)
	if err != nil {
		fmt.Printf("There was an error parsing the structure: %s\n", err)
		return
	}
	ictx := InstNodeContext{
		Parameters: globals,
	}
	err = root.Instance(ictx)
	if err != nil {
		fmt.Printf("There was an error instancing the structure: %s\n", err)
		return
	}

	err = Render(args.output, args.viewLocation, root)
	if err != nil {
		fmt.Printf("There was an error writing the output files, %s", err)
		return
	}
}

const comp = "comp"

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
	root, err = convertInstanceComponents(root)
	if err != nil {
		return nil, err
	}
	root = hoistImports(root)
	root, err = convertComponents(root)
	if err != nil {
		return nil, err
	}

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
	root, err = convertInstanceComponents(root)
	if err != nil {
		return nil, err
	}
	root = hoistImports(root)
	root, err = convertComponents(root)
	if err != nil {
		return nil, err
	}

	ret := []Node{}
	for _, node := range Children(root) {
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
	for _, node := range Descendants(root) {
		if strings.ToLower(node.Data()) == comp {
			children := Children(node)
			if len(children) > 0 {
				root := NewNode("root", BaseType)
				root.SetVisible(false)
				for _, child := range Children(node) {
					AppendChild(root, Detach(child))
				}
				((node).(*ImportNode)).ComponentRoot = root
			}
			Detach(node)
			imports = append(imports, node)
		}
	}

	for _, node := range imports {
		for _, child := range Children(root) {
			AppendChild(node, Detach(child))
		}
		AppendChild(root, node)
	}

	return root
}

func convertInstanceComponents(root Node) (Node, error) {
	desc := Descendants(root)
	for i := 0; i < len(desc); i++ {
		node := desc[i]
		if strings.ToLower(node.Data()) == comp {
			hasTag, _ := GetAttr(node, "tag")
			if hasTag {
				continue
			}

			tagName := "tag-" + randomID(6)
			attrs := node.Attrs()
			add := NewNode(tagName, BaseType, attrs...)
			attrs = append(attrs, &html.Attribute{Key: "tag", Val: tagName})
			node.SetAttrs(attrs)

			// Steal the childrent from the component node
			add.SetFirstChild(node.FirstChild())
			for _, child := range Children(node) {
				child.SetParent(add)
			}
			node.SetFirstChild(nil)

			err := InsertBefore(node.Parent(), node, add)
			if err != nil {
				return nil, err
			}
		}
	}

	return root, nil
}

func convertComponents(root Node) (Node, error) {
	tags := []string{}
	for _, node := range Descendants(root) {
		if strings.ToLower(node.Data()) == comp {
			hasTag, tagAttr := GetAttr(node, "tag")
			if !hasTag {
				return nil, fmt.Errorf("component node missing tag value %s", node)
			}
			tags = append(tags, tagAttr.Val)
		}
		for _, tag := range tags {
			if strings.ToLower(node.Data()) == strings.ToLower(tag) {
				comp := NewNode(node.Data(), ComponentType, node.Attrs()...)
				comp.SetParent(node.Parent())
				if node.Parent() != nil && node.PrevSibling() == nil {
					node.Parent().SetFirstChild(comp)
				}
				comp.SetFirstChild(node.FirstChild())
				for _, child := range Children(node) {
					child.SetParent(comp)
				}
				comp.SetPrevSibling(node.PrevSibling())
				if node.PrevSibling() != nil {
					node.PrevSibling().SetNextSibling(comp)
				}
				comp.SetNextSibling(node.NextSibling())
				if node.NextSibling() != nil {
					node.NextSibling().SetPrevSibling(comp)
				}

				comp.SetVisible(false)
			}
		}
	}

	return root, nil
}

func removeWhiteSpace(root Node) Node {
	for _, node := range Descendants(root) {
		if len(strings.TrimSpace(node.Data())) == 0 {
			Detach(node)
		}
	}
	return root
}

func fragmentNodes(root Node) Node {
	re := regexp.MustCompile(`{[_a-zA-Z][_a-zA-Z0-9]*}`)
	for _, node := range Descendants(root) {
		data := strings.TrimSpace(node.Data())
		matches := re.FindAllIndex([]byte(data), -1)
		if len(matches) == 0 {
			continue
		}
		if len(matches) == 1 && (matches[0][1]-matches[0][0] == len(data)) {
			continue
		}
		newData := []string{}
		prevIndex := 0
		for _, match := range matches {
			newData = append(newData, data[prevIndex:match[0]])
			newData = append(newData, data[match[0]:match[1]])
			prevIndex = match[1]
		}
		newData = append(newData, data[prevIndex:])
		node.SetVisible(false)
		for _, ndata := range newData {
			if len(strings.TrimSpace(ndata)) == 0 {
				continue
			}
			new := NewNode(ndata, TextType)
			AppendChild(node, new)
		}
	}

	return root
}

// The various file types supported by the system
const (
	JSFileType = iota
	HTMLFileType
	CSSFileType
	TSFileType
)

// File represents a simple file
type File struct {
	Name    string
	Entries []Node
	Path    string
	Type    int
	Remote  bool
}

// FileList type so I can add methods
type FileList []*File

// Merge merges a new file into the file list
func (files FileList) Merge(add *File) FileList {
	for _, file := range files {
		if file.Name == add.Name &&
			file.Type == add.Type {
			file.Entries = append(add.Entries, file.Entries...)
			return files
		}
	}

	return append([]*File{add}, files...)
}

// WriteFile writes a file to a specified dir
func WriteFile(dest, data string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(data))
	return err
}

// Render takes a node and renders the full tree into an array of files
func Render(outputDir, viewLocation string, root Node) error {
	var head, body Node
	for _, desc := range Descendants(root) {
		if desc.Data() == "head" {
			head = desc
		}
		if desc.Data() == "body" {
			body = desc
		}
	}

	if head == nil || body == nil {
		return fmt.Errorf("Missing head or body node")
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0700)
	}

	cssNodes := FindNodes(root, CSSType)
	var cssBundle string
	for _, node := range cssNodes {
		cssNode := node.(*CSSNode)
		if !cssNode.Remote {
			cssBundle += node.Render()
		} else {
			AppendChild(head,
				NewNode("link", BaseType, &html.Attribute{Key: "rel", Val: "stylesheet"}, &html.Attribute{Key: "href", Val: cssNode.Href}))
		}
		Detach(node)
	}
	if len(cssNodes) > 0 {
		err := WriteFile(outputDir+"/bundle.css", cssBundle)
		if err != nil {
			return err
		}
		AppendChild(head,
			NewNode("link", BaseType, &html.Attribute{Key: "rel", Val: "stylesheet"}, &html.Attribute{Key: "href", Val: viewLocation + "/bundle.css"}),
		)
	}

	var jsBundle string
	jsNodes := FindNodes(root, JSType)
	maxDepth := 0
	for _, node := range jsNodes {
		depth := node.(*JSNode).Depth
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	sorted := []Node{}
	for i := maxDepth; i >= 0; i-- {
		for ii := 0; ii < len(jsNodes); ii++ {
			if jsNodes[ii].(*JSNode).Depth == i {
				sorted = append(sorted, jsNodes[ii])
			}
		}
	}
	jsNodes = sorted
	for _, node := range sorted {
		fmt.Println(node, node.(*JSNode).Depth)
	}

	done := []string{}
	for _, node := range jsNodes {
		jsNode := node.(*JSNode)
		if jsNode.Remote {
			AppendChild(body,
				NewNode("script", BaseType, &html.Attribute{Key: "src", Val: jsNode.Src}))
			Detach(node)
			continue
		}

		src := node.Render()
		new := true
		for _, check := range done {
			if check == src {
				new = false
			}
		}
		if new {
			jsBundle = jsBundle + src
			done = append(done, src)
		}

		Detach(node)
	}
	if len(jsNodes) > 0 {
		err := WriteFile(outputDir+"/bundle.js", jsBundle)
		if err != nil {
			return err
		}

	}

	var tsBundle string
	tsNodes := FindNodes(root, TSType)
	maxDepth = 0
	for _, node := range tsNodes {
		depth := node.(*TSNode).Depth
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	sorted = []Node{}
	for i := maxDepth; i >= 0; i-- {
		for ii := 0; ii < len(tsNodes); ii++ {
			if tsNodes[ii].(*TSNode).Depth == i {
				sorted = append(sorted, tsNodes[ii])
			}
		}
	}
	tsNodes = sorted

	done = []string{}
	for _, node := range tsNodes {
		src := node.Render()
		new := true
		for _, check := range done {
			if check == src {
				new = false
			}
		}
		if new {
			tsBundle = tsBundle + src
			done = append(done, src)
		}
		Detach(node)
	}
	if len(tsNodes) > 0 {
		err := WriteFile(outputDir+"/bundle.ts", tsBundle)
		if err != nil {
			return err
		}
	}
	if len(jsNodes) > 0 || len(tsNodes) > 0 {
		AppendChild(body,
			NewNode("script", BaseType, &html.Attribute{Key: "src", Val: viewLocation + "/bundle.js"}),
		)
	}

	err := WriteFile(outputDir+"/index.html", root.Render())
	if err != nil {
		return err
	}

	return nil
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
		if node.Data == comp {
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
