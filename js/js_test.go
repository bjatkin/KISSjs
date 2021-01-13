package js

import (
	"testing"
)

func TestLexScript(t *testing.T) {
	script := `({KISSimport:"t.js", remote: true});`
	ok := []Token{
		Token{Type: openImport, Value: "({"},
		Token{Type: kissKeyword, Value: "KISSimport"},
		Token{Type: colon, Value: ":"},
		Token{Type: value, Value: `"t.js"`},
		Token{Type: comma, Value: ","},
		Token{Type: kissKeyword, Value: "remote"},
		Token{Type: colon, Value: ":"},
		Token{Type: value, Value: "true"},
		Token{Type: closeImport, Value: "})"},
		Token{Type: semiColon, Value: ";"}}
	tokens := LexScript(script)
	if len(ok) != len(tokens) {
		t.Errorf("Expecting %d tokens, but got %d", len(ok), len(tokens))
	}
	for i := 0; i < len(ok); i++ {
		if ok[i].Type != tokens[i].Type {
			t.Errorf("Expecting tokenType %d, but got %d", ok[i].Type, tokens[i].Type)
		}
		if ok[i].Value != tokens[i].Value {
			t.Errorf("Expecting value of %s, but got %s", ok[i].Value, tokens[i].Value)
		}
	}

	script = `console.log( "\"Test\"" );
let a = 10`
	tokens = LexScript(script)
	ok = []Token{
		Token{Type: any, Value: "c"},
		Token{Type: any, Value: "o"},
		Token{Type: any, Value: "n"},
		Token{Type: any, Value: "s"},
		Token{Type: any, Value: "o"},
		Token{Type: any, Value: "l"},
		Token{Type: any, Value: "e"},
		Token{Type: dot, Value: "."},
		Token{Type: any, Value: "l"},
		Token{Type: any, Value: "o"},
		Token{Type: any, Value: "g"},
		Token{Type: openExpression, Value: "("},
		Token{Type: value, Value: `"\"Test\""`},
		Token{Type: closeExpression, Value: ")"},
		Token{Type: semiColon, Value: ";"},
		Token{Type: newLine, Value: "\n"},
		Token{Type: keyword, Value: "let "},
		Token{Type: any, Value: "a"},
		Token{Type: equal, Value: "="},
		Token{Type: any, Value: "1"},
		Token{Type: any, Value: "0"},
	}

	if len(ok) != len(tokens) {
		t.Errorf("Expecting %d tokens, but got %d", len(ok), len(tokens))
	}
	for i := 0; i < len(ok); i++ {
		if ok[i].Type != tokens[i].Type {
			t.Errorf("Expecting tokenType %d, but got %d", ok[i].Type, tokens[i].Type)
		}
		if ok[i].Value != tokens[i].Value {
			t.Errorf("Expecting value of %s but got %s", ok[i].Value, tokens[i].Value)
		}
	}
}

func TestParseTokens(t *testing.T) {
	type test struct {
		script     string
		check      Script
		checkLines []string
	}

	tests := []test{
		test{
			script: `
			({KISSimport: "test.js", remote: true});
			({KISSimport: "again.js", remote: false});
			({KISSimport: "last.js"});

			let a = 10;
			let b = "50"
			console.log(a + b);
			var t = test({
				hi: 5,
			});
			`,
			check: Script{
				Imports: []Import{
					Import{Src: "test.js", Remote: true},
					Import{Src: "again.js", Remote: false},
					Import{Src: "last.js"},
				},
			},
			checkLines: []string{
				"let a=10;",
				"let b=\"50\";",
				"console.log(a+b);",
				"var t=test({",
				"hi:5,",
				"});",
				"",
			},
		},
		test{
			script: `
			if (object.good() &&
				object.notBad()) {
					let doing = "a thing"
				}
			`,
			check: Script{},
			checkLines: []string{
				"if(object.good()&&",
				"object.notBad()){",
				"let doing=\"a thing\"",
				"};",
			},
		},
		test{
			script: `fetch(test, {
				method: "POST",
				body: JSON.stringify({
					data: data,
				})
			}).`,
			check: Script{},
			checkLines: []string{
				"fetch(test,{",
				`method:"POST",`,
				`body:JSON.stringify({`,
				`data:data,`,
				`})`,
				`}).`,
			},
		},
	}

	for i, run := range tests {
		tokens := LexScript(run.script)
		kissScript, err := ParseTokens(tokens)
		if err != nil {
			t.Errorf("(%d) there was an error parsing the js script %s", i, err)
		}
		if len(kissScript.Imports) != len(run.check.Imports) {
			t.Errorf("(%d) Expected %d imports, got %d", i, len(run.check.Imports), len(kissScript.Imports))
		}
		if len(kissScript.Lines) != len(run.checkLines) {
			t.Errorf("(%d) Expected %d lines, got %d", i, len(run.checkLines), len(kissScript.Lines))
		}

		for i := 0; i < len(kissScript.Imports); i++ {
			if kissScript.Imports[i].Src != run.check.Imports[i].Src {
				t.Errorf("(%d) Expected src of %s, got %s", i, kissScript.Imports[i].Src, run.check.Imports[i].Src)
			}
			if kissScript.Imports[i].Remote != run.check.Imports[i].Remote {
				t.Errorf("(%d) Expected Remote to be %v, got %v", i, kissScript.Imports[i].Remote, run.check.Imports[i].Remote)
			}
		}

		for ii := 0; ii < len(kissScript.Lines); ii++ {
			line := ""
			for _, v := range kissScript.Lines[ii].Value {
				line += v.Value
			}
			if line != run.checkLines[ii] {
				t.Errorf("(%d|%d) Expected line %s, got %s", i, ii, run.checkLines[ii], line)
			}
		}
	}
}

// TODO: we need better test coverage on this stuff
