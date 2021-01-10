package css

import (
	"testing"
)

func TestLex(t *testing.T) {
	type test struct {
		css   string
		check []Token
	}

	tests := []test{
		test{ // TEST 0
			css: `div {
				color: #fff;
				border: 1px solid black;
			}`,
			check: []Token{
				Token{elmName, "div"},
				Token{openBlock, "{"},
				Token{property, "color"},
				Token{value, "#fff"},
				Token{property, "border"},
				Token{value, "1px solid black"},
				Token{closeBlock, "}"},
			},
		},
		test{ // TEST 1
			css: `.class {
				color: white;
				margin: 0px 10px;
			}
			
			button.primary:focus {
				border: none;
				color: #3f12dd88;
			}`,
			check: []Token{
				Token{className, ".class"},
				Token{openBlock, "{"},
				Token{property, "color"},
				Token{value, "white"},
				Token{property, "margin"},
				Token{value, "0px 10px"},
				Token{closeBlock, "}"},
				Token{elmName, "button"},
				Token{className, ".primary"},
				Token{pseudoClass, ":focus"},
				Token{openBlock, "{"},
				Token{property, "border"},
				Token{value, "none"},
				Token{property, "color"},
				Token{value, "#3f12dd88"},
				Token{closeBlock, "}"},
			},
		},
		test{ // TEST 2
			css: `div[h^="g"] {
				text-decoration: none;
			}`,
			check: []Token{
				Token{elmName, "div"},
				Token{attrBlock, `[h^="g"]`},
				Token{openBlock, "{"},
				Token{property, "text-decoration"},
				Token{value, "none"},
				Token{closeBlock, "}"},
			},
		},
		test{ // TEST 3
			css: `table>#test+tr {
				animation: test 5s;
				background: #fff url(data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyIiBoZWlnaHQ9IjMiPjxwYXRoIGQ9Im0gMCwxIDEsMiAxLC0yIHoiLz48L3N2Zz4=) no-repeat scroll 95% center/10px 15px;
			}`,
			check: []Token{
				Token{elmName, "table"},
				Token{child, ">"},
				Token{idName, "#test"},
				Token{nextChild, "+"},
				Token{elmName, "tr"},
				Token{openBlock, "{"},
				Token{property, "animation"},
				Token{value, "test 5s"},
				Token{property, "background"},
				Token{value, "#fff url(data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyIiBoZWlnaHQ9IjMiPjxwYXRoIGQ9Im0gMCwxIDEsMiAxLC0yIHoiLz48L3N2Zz4=) no-repeat scroll 95% center/10px 15px"},
				Token{closeBlock, "}"},
			},
		},
		test{ // TEST 4
			css: `div {
				width: 50px;
				height: 50px;
				background-color: gray;
				animation: zoom 2s infinite;
				position: absolute;
				top: 10px;
				left: 10px;
			}
			
			@keyframes zoom {
				0% {
					left: 10px;
					background-color: gray;
				}
				50% {
					left: 100px;
					background-color: white;
				}
				100% {
					left: 10px;
					background-color: gray;
				}
			}`,
			check: []Token{
				Token{elmName, "div"},
				Token{openBlock, "{"},
				Token{property, "width"},
				Token{value, "50px"},
				Token{property, "height"},
				Token{value, "50px"},
				Token{property, "background-color"},
				Token{value, "gray"},
				Token{property, "animation"},
				Token{value, "zoom 2s infinite"},
				Token{property, "position"},
				Token{value, "absolute"},
				Token{property, "top"},
				Token{value, "10px"},
				Token{property, "left"},
				Token{value, "10px"},
				Token{closeBlock, "}"},
				Token{keyframe, "@keyframes zoom"},
				Token{openBlock, "{"},
				Token{percentage, "0%"},
				Token{openBlock, "{"},
				Token{property, "left"},
				Token{value, "10px"},
				Token{property, "background-color"},
				Token{value, "gray"},
				Token{closeBlock, "}"},
				Token{percentage, "50%"},
				Token{openBlock, "{"},
				Token{property, "left"},
				Token{value, "100px"},
				Token{property, "background-color"},
				Token{value, "white"},
				Token{closeBlock, "}"},
				Token{percentage, "100%"},
				Token{openBlock, "{"},
				Token{property, "left"},
				Token{value, "10px"},
				Token{property, "background-color"},
				Token{value, "gray"},
				Token{closeBlock, "}"},
				Token{closeBlock, "}"},
			},
		},
		test{ // TEST 5
			css: `div a.test #again {
				color: blue;
			}`,
			check: []Token{
				Token{elmName, "div"},
				Token{whiteSpace, " "},
				Token{elmName, "a"},
				Token{className, ".test"},
				Token{whiteSpace, " "},
				Token{idName, "#again"},
				Token{openBlock, "{"},
				Token{property, "color"},
				Token{value, "blue"},
				Token{closeBlock, "}"},
			},
		},
	}

	for i, run := range tests {
		tokens := Lex(run.css)

		if len(tokens) != len(run.check) {
			t.Errorf("(%d): Incorect token count expected %d tokens but got %d", i, len(run.check), len(tokens))
		}

		for ii, tok := range run.check {
			if ii >= len(tokens) {
				// prevent a crash when tokens have different counts
				return
			}
			if tokens[ii].Type != tok.Type {
				t.Errorf("(%d) Incorrect token type expected %d but got %d", i, tok.Type, tokens[ii].Type)
			}
			if tokens[ii].Value != tok.Value {
				t.Errorf("(%d) Incorrect token value expected %s but got %s", i, tok.Value, tokens[ii].Value)
			}
		}
	}
}

func TestParseSelector(t *testing.T) {
	type test struct {
		css   string
		check []Selector
	}

	tests := []test{
		test{
			css: `div a table {
				color: black;
			}`,
			check: []Selector{
				Selector{Sel: "div"},
				Selector{Sel: "a"},
				Selector{Sel: "table"},
			},
		},
		test{
			css: `.class[href|="google"] {}`,
			check: []Selector{
				Selector{Sel: `.class[href|="google"]`},
			},
		},
		test{
			css: `#test div .class a:visited button>p::before {}`,
			check: []Selector{
				Selector{Sel: "#test"},
				Selector{Sel: "div"},
				Selector{Sel: ".class"},
				Selector{Sel: "a", PostSel: ":visited"},
				Selector{Sel: "button>p", PostSel: "::before"},
			},
		},
	}

	for i, run := range tests {
		tokens := Lex(run.css)
		_, selectors := parseSelector(tokens)

		if len(selectors) != len(run.check) {
			t.Errorf("(%d): Incorrect token count expected %d tokens but got %d", i, len(run.check), len(selectors))
		}

		for ii, sel := range selectors {
			if ii >= len(run.check) {
				// prevent a crash when tokens have different counts
				return
			}
			if sel.Sel != run.check[ii].Sel {
				t.Errorf("(%d), selector %d has a value of %s, was expecting %s", i, ii, sel.Sel, run.check[ii].Sel)
			}

			if sel.PostSel != run.check[ii].PostSel {
				t.Errorf("(%d), post selector %d has a value of %s, was expecting %s", i, ii, sel.PostSel, run.check[ii].PostSel)
			}
		}
	}
}

func TestParseBlock(t *testing.T) {
	type test struct {
		css   string
		check []Style
	}

	tests := []test{
		test{
			css: `{
				color: black;
				background-color: blue;
				font-color: red;
			}`,
			check: []Style{
				Style{Prop: "color", Val: "black"},
				Style{Prop: "background-color", Val: "blue"},
				Style{Prop: "font-color", Val: "red"},
			},
		},
		test{
			css: `{
				test: #fff url(xml+svg;stuffshere);
				animation: ani 1s loop;
				width: 50px;
				height: 75px;
			}`,
			check: []Style{
				Style{Prop: "test", Val: "#fff url(xml+svg;stuffshere)"},
				Style{Prop: "animation", Val: "ani 1s loop"},
				Style{Prop: "width", Val: "50px"},
				Style{Prop: "height", Val: "75px"},
			},
		},
	}

	for i, run := range tests {
		tokens := Lex(run.css)
		_, block := parseBlock(tokens)

		if len(block) != len(run.check) {
			t.Errorf("(%d): Incorrect token count expected %d tokens but got %d", i, len(run.check), len(block))
		}

		for ii, prop := range block {
			if ii >= len(run.check) {
				// prevent a crash when tokens have different counts
				return
			}
			if prop.Prop != run.check[ii].Prop {
				t.Errorf("(%d), propery %d has a value of %s, was expecting %s", i, ii, prop.Prop, run.check[ii].Prop)
			}

			if prop.Val != run.check[ii].Val {
				t.Errorf("(%d), property value %d has a value of %s, was expecting %s", i, ii, prop.Val, run.check[ii].Val)
			}
		}
	}
}

func TestParseRule(t *testing.T) {
	type test struct {
		css   string
		check Rule
	}

	tests := []test{
		test{
			css: `.test {
				color: black;
			}`,
			check: Rule{
				Selectors: []Selector{Selector{Sel: ".test"}},
				Styles:    []Style{Style{Prop: "color", Val: "black"}},
			},
		},
	}

	for i, run := range tests {
		tokens := Lex(run.css)
		_, rule := parseRule(tokens)

		if len(rule.Selectors) != len(run.check.Selectors) {
			t.Errorf("(%d): Incorrect selector count expected %d tokens but got %d", i, len(run.check.Selectors), len(rule.Selectors))
		}
		if len(rule.Styles) != len(run.check.Styles) {
			t.Errorf("(%d): Incorrect style count expected %d tokens but got %d", i, len(run.check.Styles), len(rule.Styles))
		}

		for ii, sel := range rule.Selectors {
			if ii >= len(run.check.Selectors) {
				// prevent a crash when tokens have different counts
				return
			}
			if sel.Sel != run.check.Selectors[ii].Sel {
				t.Errorf("(%d), selector %d has a value of %s, was expecting %s", i, ii, sel.Sel, run.check.Selectors[ii].Sel)
			}

			if sel.PostSel != run.check.Selectors[ii].PostSel {
				t.Errorf("(%d), post selector %d has a value of %s, was expecting %s", i, ii, sel.PostSel, run.check.Selectors[ii].PostSel)
			}
		}

		for ii, sty := range rule.Styles {
			if ii >= len(run.check.Styles) {
				// prevent a crash when tokens have different counts
				return
			}
			if sty.Prop != run.check.Styles[ii].Prop {
				t.Errorf("(%d), styles prop %d has a value of %s, was expecting %s", i, ii, sty.Prop, run.check.Styles[ii].Prop)
			}

			if sty.Val != run.check.Styles[ii].Val {
				t.Errorf("(%d), style value %d has a value of %s, was expecting %s", i, ii, sty.Val, run.check.Styles[ii].Val)
			}
		}
	}
}
