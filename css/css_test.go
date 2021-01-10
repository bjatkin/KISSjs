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
		test{
			css: `button[class|="p"]:hover .win:nth-child(2) {
				background-color: #ffffff33;
				width: 100%;
				border: 1px solid #a3a3a3;
				animation: zoom 1s;
			}`,
			check: Rule{
				Selectors: []Selector{
					Selector{Sel: "button[class|=\"p\"]", PostSel: ":hover"},
					Selector{Sel: ".win", PostSel: ":nth-child(2)"},
				},
				Styles: []Style{
					Style{Prop: "background-color", Val: "#ffffff33"},
					Style{Prop: "width", Val: "100%"},
					Style{Prop: "border", Val: "1px solid #a3a3a3"},
					Style{Prop: "animation", Val: "zoom 1s"},
				},
			},
		},
		test{
			css: `select::after table>tr+button {
				position: absolute;
				top: 10px;
				left: 10px;
				test: #fff url(xml+svg;stuffshere) after crap/100px;
			}`,
			check: Rule{
				Selectors: []Selector{
					Selector{Sel: "select", PostSel: "::after"},
					Selector{Sel: "table>tr+button"},
				},
				Styles: []Style{
					Style{Prop: "position", Val: "absolute"},
					Style{Prop: "top", Val: "10px"},
					Style{Prop: "left", Val: "10px"},
					Style{Prop: "test", Val: "#fff url(xml+svg;stuffshere) after crap/100px"},
				},
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

func TestParseAnim(t *testing.T) {
	type test struct {
		css   string
		check Anim
	}

	tests := []test{
		test{
			css: `@keyframes name {
				0% {
					background-color: #fff;
					left: 0px;
				}
				20% {
					background-color: #000;
					left: 50px;
				}
				100% {
					background-color: #777;
					left: 80px;
				}
			}`,
			check: Anim{
				Name: "name",
				Frames: []Frame{
					Frame{Time: "0%", Styles: []Style{
						Style{Prop: "background-color", Val: "#fff"},
						Style{Prop: "left", Val: "0px"},
					}},
					Frame{Time: "20%", Styles: []Style{
						Style{Prop: "background-color", Val: "#000"},
						Style{Prop: "left", Val: "50px"},
					}},
					Frame{Time: "100%", Styles: []Style{
						Style{Prop: "background-color", Val: "#777"},
						Style{Prop: "left", Val: "80px"},
					}},
				},
			},
		},
		test{
			css: `@keyframes long-nameThat_will-not_FAIL {
				0% {
					test: long value with l00ts of weird % stuff/10px;
				}
				50% {
					different: value;
				}
			}`,
			check: Anim{
				Name: "long-nameThat_will-not_FAIL",
				Frames: []Frame{
					Frame{Time: "0%", Styles: []Style{
						Style{Prop: "test", Val: "long value with l00ts of weird % stuff/10px"},
					}},
					Frame{Time: "50%", Styles: []Style{
						Style{Prop: "different", Val: "value"},
					}},
				},
			},
		},
	}

	for i, run := range tests {
		tokens := Lex(run.css)
		_, anim := parseAnim(tokens)

		if anim.Name != run.check.Name {
			t.Errorf("(%d) wrong animation name expecting %s but got %s", i, run.check.Name, anim.Name)
		}

		if len(anim.Frames) != len(run.check.Frames) {
			t.Errorf("(%d) incorect number of frames got %d but expected %d", i, len(anim.Frames), len(run.check.Frames))
		}

		for ii, frame := range anim.Frames {

			if frame.Time != run.check.Frames[ii].Time {
				t.Errorf("(%d|%d) wrong animation frame percentage expecting %s but got %s", i, ii, run.check.Frames[ii].Time, frame.Time)
			}

			if len(frame.Styles) != len(run.check.Frames[ii].Styles) {
				t.Errorf("(%d) incorect number of styles got %d but expected %d", i, len(frame.Styles), len(run.check.Frames[ii].Styles))
			}

			for iii, style := range frame.Styles {

				if style.Prop != run.check.Frames[ii].Styles[iii].Prop {
					t.Errorf("(%d|%d|%d) wrong style property expecting %s but got %s", i, ii, iii, run.check.Frames[ii].Styles[iii].Prop, style.Prop)
				}

				if style.Val != run.check.Frames[ii].Styles[iii].Val {
					t.Errorf("(%d|%d|%d) wrong style value expecting %s but got %s", i, ii, iii, run.check.Frames[ii].Styles[iii].Val, style.Val)
				}

			}
		}
	}
}

func TestParse(t *testing.T) {
	type test struct {
		css   string
		check Script
	}
	tests := []test{
		test{
			css: `div.move1 button.move2:nth-child(2) {
				width: 50px;
				height: 50px;
				background-color: #a0f;
				animation: zoom 2s;
				position: absolute;
				top: 100px;
				left: 100px;
			}
			
			@keyframes zoom {
				0% {
					top: 0px;
					left: 0px;
				}
				100% {
					top: 100px;
					left: 100px;
				}
			}`,
			check: Script{
				Rules: []Rule{
					Rule{
						Selectors: []Selector{
							Selector{Sel: "div.move1"},
							Selector{Sel: "button.move2", PostSel: ":nth-child(2)"},
						},
						Styles: []Style{
							Style{Prop: "width", Val: "50px"},
							Style{Prop: "height", Val: "50px"},
							Style{Prop: "background-color", Val: "#a0f"},
							Style{Prop: "animation", Val: "zoom 2s"},
							Style{Prop: "position", Val: "absolute"},
							Style{Prop: "top", Val: "100px"},
							Style{Prop: "left", Val: "100px"},
						},
					},
				},
				Anims: []Anim{
					Anim{
						Name: "zoom",
						Frames: []Frame{
							Frame{Time: "0%",
								Styles: []Style{
									Style{Prop: "top", Val: "0px"},
									Style{Prop: "left", Val: "0px"},
								},
							},
							Frame{Time: "100%",
								Styles: []Style{
									Style{Prop: "top", Val: "100px"},
									Style{Prop: "left", Val: "100px"},
								},
							},
						},
					},
				},
			},
		},
	}

	for i, run := range tests {
		tokens := Lex(run.css)
		script, err := Parse(tokens)

		if err != nil {
			t.Errorf("error parsing css script %s", err)
		}

		if len(script.Rules) != len(run.check.Rules) {
			t.Errorf("(%d) wrong number of rules got %d expected %d", i, len(script.Rules), len(run.check.Rules))
		}

		if len(script.Anims) != len(run.check.Anims) {
			t.Errorf("(%d) wrong number of rules got %d expected %d", i, len(script.Anims), len(run.check.Anims))
		}

		for ii, rule := range script.Rules {
			if len(rule.Selectors) != len(run.check.Rules[ii].Selectors) {
				t.Errorf(
					"(%d|%d) wrong number of selectors got %d expected %d",
					i, ii, len(rule.Selectors), len(run.check.Rules[ii].Selectors),
				)
			}

			if len(rule.Styles) != len(run.check.Rules[ii].Styles) {
				t.Errorf(
					"(%d|%d) wrong number of styles got %d expected %d",
					i, ii, len(rule.Styles), len(run.check.Rules[ii].Styles),
				)
			}

			for iii, sel := range rule.Selectors {
				if sel.Sel != run.check.Rules[ii].Selectors[iii].Sel {
					t.Errorf(
						"(%d|%d|%d) wrong selector got %s expected %s",
						i, ii, iii, sel.Sel, run.check.Rules[ii].Selectors[iii].Sel,
					)
				}

				if sel.PostSel != run.check.Rules[ii].Selectors[iii].PostSel {
					t.Errorf(
						"(%d|%d|%d) wrong post selector got %s expected %s",
						i, ii, iii, sel.PostSel, run.check.Rules[ii].Selectors[iii].PostSel,
					)
				}
			}

			for iii, style := range rule.Styles {
				if style.Prop != run.check.Rules[ii].Styles[iii].Prop {
					t.Errorf(
						"(%d|%d|%d) wrong style property got %s expected %s",
						i, ii, iii, style.Prop, run.check.Rules[ii].Styles[iii].Prop,
					)
				}

				if style.Val != run.check.Rules[ii].Styles[iii].Val {
					t.Errorf(
						"(%d|%d|%d) wrong style value got %s expected %s",
						i, ii, iii, style.Val, run.check.Rules[ii].Styles[iii].Val,
					)
				}
			}
		}

		for ii, anim := range script.Anims {
			if anim.Name != run.check.Anims[ii].Name {
				t.Errorf(
					"(%d|%d) wrong animation name got %s expected %s",
					i, ii, anim.Name, run.check.Anims[ii].Name,
				)
			}

			if len(anim.Frames) != len(run.check.Anims[ii].Frames) {
				t.Errorf(
					"(%d|%d) wrong number of animation frames got %d expected %d",
					i, ii, len(anim.Frames), len(run.check.Anims[ii].Frames),
				)
			}

			for iii, frame := range anim.Frames {
				if frame.Time != run.check.Anims[ii].Frames[iii].Time {
					t.Errorf(
						"(%d|%d|%d) wrong animation frame time got %s expected %s",
						i, ii, iii, frame.Time, run.check.Anims[ii].Frames[iii].Time,
					)
				}

				if len(frame.Styles) != len(run.check.Anims[ii].Frames[iii].Styles) {
					t.Errorf(
						"(%d|%d|%d) wrong number of animation frame styles got %d expected %d",
						i, ii, iii, len(frame.Styles), len(run.check.Anims[ii].Frames[iii].Styles),
					)
				}

				for iiii, style := range frame.Styles {
					if style.Prop != run.check.Anims[ii].Frames[iii].Styles[iiii].Prop {
						t.Errorf(
							"(%d|%d|%d|%d) wrong animation frame style prop got %s expected %s",
							i, ii, iii, iiii, style.Prop, run.check.Anims[ii].Frames[iii].Styles[iiii].Prop,
						)
					}

					if style.Val != run.check.Anims[ii].Frames[iii].Styles[iiii].Val {
						t.Errorf(
							"(%d|%d|%d|%d) wrong animation frame style prop got %s expected %s",
							i, ii, iii, iiii, style.Val, run.check.Anims[ii].Frames[iii].Styles[iiii].Val,
						)
					}
				}
			}

		}
	}
}
