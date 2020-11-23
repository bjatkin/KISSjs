package main

import (
	"testing"
)

func TestParseCSS(t *testing.T) {
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
	matches := []CSSRule{
		CSSRule{
			Selector: []string{"div"},
			Styles: []CSSStyle{
				CSSStyle{Style: "color", Value: "#fff"},
				CSSStyle{Style: "border", Value: "1px solid black"},
			},
		},
		CSSRule{
			Selector: []string{".newclass"},
			Styles: []CSSStyle{
				CSSStyle{Style: "test", Value: "\"@this_is_a_test@\""},
				CSSStyle{Style: "again", Value: "100px"},
			},
		},
	}
	rules, err := ParseCSS(css)
	if err != nil {
		t.Errorf("There was an error while parsing the css %s", err)
	}
	if len(matches) != len(rules) {
		t.Errorf("Expected %d rules but got %d rules", len(matches), len(rules))
	}
	for i := 0; i < len(matches); i++ {
		if len(matches[i].Selector) != len(rules[i].Selector) {
			t.Errorf("Expected %d selectors but got %d", len(matches[i].Selector), len(rules[i].Selector))
		}
		for j := 0; j < len(matches[i].Selector); j++ {
			if rules[i].Selector[j] != matches[i].Selector[j] {
				t.Errorf("Expected Selector %s, but got %s", rules[i].Selector[j], matches[i].Selector[j])
			}
		}

		if len(rules[i].Styles) != len(matches[i].Styles) {
			t.Errorf("Expected %d styles but got %d", len(matches[i].Selector), len(rules[i].Selector))
		}
		for k := 0; k < len(matches[i].Styles); k++ {
			if rules[i].Styles[k] != matches[i].Styles[k] {
				t.Errorf("Expected Styles %s, but got %s", rules[i].Styles[k], matches[i].Styles[k])
			}
		}
	}
}
