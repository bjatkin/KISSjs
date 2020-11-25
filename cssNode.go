package main

// CSSNode is a node for all style data
type CSSNode struct {
	BaseNode
	Rules []*CSSRule
	Scope string
}

// Type returns CSSType
func (node *CSSNode) Type() NodeType {
	return CSSType
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

func (node *CSSNode) Render() string {
	return ""
}
