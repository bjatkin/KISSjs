package main

import "strings"

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

func (node *CSSNode) Instance(ctx NodeContext) error {
	for i := 0; i < len(node.Rules); i++ {
		rule := node.Rules[i]
		rule.AddClass(ctx.componentScope)
		for j := 0; j < len(rule.Styles); j++ {
			style := &rule.Styles[j]
			for k := 0; k < len(style.Value); k++ {
				val := &style.Value
				for name, param := range ctx.Parameters {
					*val = strings.ReplaceAll(
						*val,
						"\"@"+name+"@\"",
						param[0].Data(),
					)
				}
			}
		}
	}
	return nil
}

func (node *CSSNode) FindEntry(ctx RenderNodeContext) RenderNodeContext {
	ctx.files = ctx.files.Merge(&File{
		Name:    "bundle",
		Type:    CSSFileType,
		Entries: []Node{node},
	})
	Detach(node)

	return ctx
}

func (node *CSSNode) Render() string {
	ret := ""

	for _, rule := range node.Rules {
		ret += rule.String()
	}

	return ret
}

func (node *CSSNode) Clone() Node {
	clone := CSSNode{
		BaseNode: BaseNode{data: node.Data(), attr: node.Attrs(), nType: node.Type(), visible: node.Visible()},
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
