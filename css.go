package main

import (
	"errors"
	"fmt"
	"strings"
)

// CSSStyle is a style propery and style value
type CSSStyle struct {
	Style, Value string
}

// CSSRule is a css selector and set of styles
type CSSRule struct {
	Selector []string
	Styles   []CSSStyle
}

func (css *CSSRule) clone() *CSSRule {
	clone := &CSSRule{}
	for _, sel := range css.Selector {
		clone.Selector = append(clone.Selector, sel)
	}
	for _, style := range css.Styles {
		clone.Styles = append(clone.Styles,
			CSSStyle{
				Style: style.Style,
				Value: style.Value,
			},
		)
	}

	return clone
}

// AddClass adds a class to the css rule selector
func (css *CSSRule) AddClass(class string) {
	for i := 0; i < len(css.Selector); i++ {
		if strings.Index(css.Selector[i], class) < 0 {
			css.Selector[i] += "." + class
		}
	}
}

// String returns the string representation of the css rule
func (css *CSSRule) String() string {
	ret := strings.Join(css.Selector, " ")

	ret += "{"
	for _, style := range css.Styles {
		ret += style.Style + ":" + style.Value + ";"
	}

	return ret + "}"
}

func test() {
	css := `
	div {
		color: #fff;
		border: 1px solid black;
	}	

	.newclass {
		test: "@this_is_a_test@";	
		again: 100px;
	}
	`

	rules, _ := ParseCSS(css)
	for _, rule := range rules {
		fmt.Printf("rule: %#v\n", rule)
	}
}

// ParseCSS converts a string of css into css rules
func ParseCSS(css string) ([]*CSSRule, error) {
	ret := []*CSSRule{}
	for _, rule := range strings.Split(css, "}") {
		rule = strings.TrimSpace(rule)
		if len(rule) == 0 {
			continue
		}
		add := &CSSRule{}
		half := strings.Index(rule, "{")
		if half < 0 {
			return nil, errors.New("could not find style section in css rule, missing '{' or '}' chaaracter")
		}

		add.Selector = strings.Split(strings.TrimSpace(rule[:half]), " ")
		styles := strings.Split(strings.TrimSpace(rule[half+1:]), ";")

		for _, style := range styles {
			if len(style) == 0 {
				continue
			}
			split := strings.Split(style, ":")
			if len(split) != 2 {
				return nil, errors.New("css style does not contain both a key and value. Expecting key and value split by ':'")
			}
			add.Styles = append(add.Styles, CSSStyle{Style: strings.TrimSpace(split[0]), Value: strings.TrimSpace(split[1])})
		}

		ret = append(ret, add)
	}

	return ret, nil
}
