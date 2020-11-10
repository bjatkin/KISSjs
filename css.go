package main

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
)

type cssStyle struct {
	key, val string
}

type cssRule struct {
	selector []string
	styles   []cssStyle
}

func cssFromNode(node *html.Node) (cssRule, error) {
	ret := cssRule{}
	if node.Type != html.TextNode {
		return ret, errors.New("only text nodes can be processed as css rules")
	}
	half := strings.Index(node.Data, "{")
	end := strings.Index(node.Data, "}")
	if half < 0 || end <= half {
		return ret, errors.New("could not find style section in css rule, missing '{' or '}' character")
	}

	ret.selector = strings.Split(strings.TrimSpace(node.Data[:half]), " ")
	styles := strings.Split(strings.TrimSpace(node.Data[half+1:end]), ";")

	for _, style := range styles {
		if len(style) == 0 {
			continue
		}
		split := strings.Split(style, ":")
		if len(split) != 2 {
			return ret, errors.New("css style does not contain both a key and value. Expecting key and value split by ':'")
		}
		ret.styles = append(ret.styles, cssStyle{key: strings.TrimSpace(split[0]), val: strings.TrimSpace(split[1])})
	}

	return ret, nil
}

func (css *cssRule) addClass(class string) {
	for _, sel := range css.selector {
		if strings.Index(sel, class) < 0 {
			sel += "." + class
		}
	}
}

func (css *cssRule) String() string {
	ret := strings.Join(css.selector, " ")

	ret += "{\n"
	for _, style := range css.styles {
		ret += "\t" + style.key + ": " + style.val + ";\n"
	}

	return ret + "}"
}

func (css *cssRule) hydrate(props []prop) {
	for _, prop := range props {
		if !prop.isSimple() {
			continue
		}
		for _, sel := range css.selector {
			sel = strings.ReplaceAll(sel, "\"{"+prop.key+"}\"", prop.val[0].Data)
		}
		for _, style := range css.styles {
			style.key = strings.ReplaceAll(style.key, "\"{"+prop.key+"}\"", prop.val[0].Data)
			style.val = strings.ReplaceAll(style.val, "\"{"+prop.key+"}\"", prop.val[0].Data)
		}
	}
}
