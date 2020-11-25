package main

import "fmt"

// CSSNode is a node for all style data
type CSSNode struct {
	BaseNode
	Rules []*CSSRule
	Scope string
}

// Parse extracts all css rules and applies the correct scope to them
func (node *CSSNode) Parse(ctx NodeContext) error {
	node.Scope = ctx.componentScope
	// extract the css rules
	css := ""
	if node.FirstChild() != nil {
		css = node.FirstChild().Data()
		Detach(node.FirstChild())
	}

	rules, err := ParseCSS(css)
	if err != nil {
		return err
	}

	node.Rules = rules

	// apply the correct scope
	if ctx.componentScope != "" {
		for _, rule := range node.Rules {
			rule.AddClass(ctx.componentScope)
		}
	}

	return nil
}

func (node *CSSNode) Render(file *File) (*File, []*File) {
	fmt.Println("CSS RENDER:", node)
	content := ""
	for _, rule := range node.Rules {
		content += rule.String()
	}

	ret := []*File{}
	if file.Type == CSSFileType {
		file.Content += content
	} else {
		ret = append(ret, &File{
			Type:    CSSFileType,
			Name:    "bundle",
			Content: content,
		})
	}

	return file, ret
}

func (node *CSSNode) Clone() Node {
	clone := CSSNode{
		BaseNode: BaseNode{data: node.Data(), attr: node.Attrs(), visible: node.Visible()},
	}

	for _, child := range node.Children() {
		clone.AppendChild(child.Clone())
	}

	for _, rule := range node.Rules {
		clone.Rules = append(clone.Rules, rule.clone())
	}
	clone.Scope = node.Scope

	return &clone
}
