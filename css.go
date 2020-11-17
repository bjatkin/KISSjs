package main

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
)

// Used
type cssStyle struct {
	key, val string
}

// Used
type cssRule struct {
	selector []string
	styles   []cssStyle
}

// Used
func scopeCSS(component *html.Node) ([]*cssRule, error) {
	scope := "KISS-" + generateScope(6)
	style := findOne(component, "style")
	if style == nil {
		return []*cssRule{}, nil
	}

	rules, err := cssFromNode(style.FirstChild)
	if err != nil {
		return []*cssRule{}, nil
	}

	for _, rule := range rules {
		rule.addClass(scope)
	}
	for _, node := range listNodes(component.FirstChild) {
		addClass(node, scope)
	}

	return rules, nil
}

// Used
func extractCSS(root *html.Node) ([]*cssRule, error) {

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

	// scope all the component nodes
	allStyles := []*cssRule{}
	for _, tag := range importTags {
		for _, node := range listNodes(root) {
			if strings.ToLower(node.Data) == strings.ToLower(tag) {
				styles, err := scopeCSS(node.FirstChild)
				if err != nil {
					return allStyles, err
				}
				allStyles = append(allStyles, styles...)
				style := findOne(node, "style")
				if style != nil {
					style.Parent.RemoveChild(style)
				}
			}
		}
	}

	// extract the head style
	style := findOne(root, "style")
	if style != nil {
		styles, err := cssFromNode(style.FirstChild)
		if err != nil {
			return allStyles, err
		}
		allStyles = append(allStyles, styles...)
		style.Parent.RemoveChild(style)
	}

	return allStyles, nil
}

// Used
func cssFromNode(node *html.Node) ([]*cssRule, error) {
	ret := []*cssRule{}
	if node.Type != html.TextNode {
		return ret, errors.New("only text nodes can be processed as css rules")
	}
	for _, rule := range strings.Split(node.Data, "}") {
		rule = strings.TrimSpace(rule)
		if len(rule) == 0 {
			continue
		}
		css := &cssRule{}
		half := strings.Index(rule, "{")
		if half < 0 {
			return ret, errors.New("could not find style section in css rule, missing '{' or '}' character")
		}

		css.selector = strings.Split(strings.TrimSpace(rule[:half]), " ")
		styles := strings.Split(strings.TrimSpace(rule[half+1:]), ";")

		for _, style := range styles {
			if len(style) == 0 {
				continue
			}
			split := strings.Split(style, ":")
			if len(split) != 2 {
				return ret, errors.New("css style does not contain both a key and value. Expecting key and value split by ':'")
			}
			css.styles = append(css.styles, cssStyle{key: strings.TrimSpace(split[0]), val: strings.TrimSpace(split[1])})
		}

		ret = append(ret, css)
	}

	return ret, nil
}

// Used
func (css *cssRule) addClass(class string) {
	for i := 0; i < len(css.selector); i++ {
		if strings.Index(css.selector[i], class) < 0 {
			css.selector[i] += "." + class
		}
	}
}

// Used
func (css *cssRule) String() string {
	ret := strings.Join(css.selector, " ")

	ret += "{\n"
	for _, style := range css.styles {
		ret += "\t" + style.key + ": " + style.val + ";\n"
	}

	return ret + "}"
}
