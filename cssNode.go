package main

import (
	"KISS/css"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// CSSNode is a node for all style data
type CSSNode struct {
	BaseNode
	Href   string
	Script css.Script
	Remote bool
}

// Parse extracts all css rules and applies the correct scope to them
func (node *CSSNode) Parse(ctx ParseNodeContext) error {
	hasHref, hrefAttr := GetAttr(node, "href")
	hasRemote, _ := GetAttr(node, "remote")
	node.Remote = hasRemote
	if hasHref && node.firstChild != nil {
		return fmt.Errorf("error at node %s, can not both href value and a child text node", node)
	}
	if !hasHref && node.firstChild == nil {
		return fmt.Errorf("error at node %s, node has neither a href attribute nor any child text, empty style nodes not allowed", node)
	}
	if hasRemote && !hasHref {
		return fmt.Errorf("error at node %s, can not specify remote without an href attribute", node)
	}
	if node.Remote {
		node.Href = hrefAttr.Val
		return nil
	}

	// extract the css rules
	cssString := ""
	if node.FirstChild() != nil {
		cssString = node.FirstChild().Data()
		Detach(node.FirstChild())
	}
	if hasHref {
		node.Href = ctx.path + hrefAttr.Val
		styleBytes, err := ioutil.ReadFile(node.Href)
		cssString = string(styleBytes)
		if err != nil {
			return fmt.Errorf("error at node %s, %s", node, err)
		}
	}

	tokens := css.Lex(cssString)
	script, err := css.Parse(tokens)
	if err != nil {
		return err
	}

	node.Script = script

	return nil
}

// Instance takes parameters from the node context and replaces template parameteres
func (node *CSSNode) Instance(ctx InstNodeContext) error {
	re := regexp.MustCompile(`"@[_a-zA-Z][_a-zA-Z0-9]*@"`)
	for i := 0; i < len(node.Rules); i++ {
		rule := node.Rules[i]
		rule.AddClass(ctx.componentScope)
		for j := 0; j < len(rule.Styles); j++ {
			style := &rule.Styles[j]
			for k := 0; k < len(style.Value); k++ {
				matches := re.FindAll([]byte(style.Value), -1)
				for _, match := range matches {
					p := ""
					pnode, ok := ctx.Parameters[string(match[2:len(match)-2])]
					if ok {
						if len(pnode) == 1 {
							p = pnode[0].Data()
						}
						if len(pnode) > 1 {
							return fmt.Errorf("error at node %s, tried to replace %s with multiple param nodes", node, match)
						}
						if len(pnode) == 1 && pnode[0].Type() != TextType {
							return fmt.Errorf("error at node %s, tried to replace %s with a non-text parameter", node, match)
						}
					}
					style.Value = strings.ReplaceAll(style.Value, string(match), p)
				}
			}
		}
	}
	return nil
}

// FindEntry locates all the entry points for the HTML, JS and CSS code in the tree
func (node *CSSNode) FindEntry(ctx RenderNodeContext) RenderNodeContext {
	if node.Remote {
		ctx.files = ctx.files.Merge(&File{
			Name:    node.Href,
			Type:    CSSFileType,
			Entries: []Node{node},
			Remote:  true,
		})
		Detach(node)
		return ctx
	}

	ctx.files = ctx.files.Merge(&File{
		Name:    "bundle",
		Type:    CSSFileType,
		Entries: []Node{node},
	})
	Detach(node)

	return ctx
}

// Render converts a node into a textual representation
func (node *CSSNode) Render() string {
	ret := ""

	for _, rule := range node.Rules {
		ret += rule.String()
	}

	return ret
}

// Clone creates a deep copy of a node, but does not copy over the connections to the original parent and siblings
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

	return &clone
}
