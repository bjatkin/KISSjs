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
	script := `
	({KISSimport: "test.js", remote: true});
	({KISSimport: "again.js", remote: false});
	({KISSimport: "last.js"});

	let a = 10;
	let b = "50";
	console.log(a + b);
	var t = test({
		hi: 5,
	});
	`
	imports := []Import{
		Import{Src: "test.js", Remote: true},
		Import{Src: "again.js", Remote: false},
		Import{Src: "last.js"},
	}
	lines := []string{
		"let a=10;",
		"let b=\"50\";",
		"console.log(a+b);",
		"var t=test({",
		"hi:5,",
		"});",
		"",
	}

	tokens := LexScript(script)
	kissScript, err := ParseTokens(tokens)
	if err != nil {
		t.Errorf("there was an error parsing the js script %s", err)
	}
	if len(kissScript.Imports) != len(imports) {
		t.Errorf("Expected %d imports, got %d", len(kissScript.Imports), len(imports))
	}
	if len(kissScript.Lines) != len(lines) {
		t.Errorf("Expected %d lines, got %d", len(kissScript.Lines), len(lines))
	}

	for i := 0; i < len(imports); i++ {
		if kissScript.Imports[i].Src != imports[i].Src {
			t.Errorf("Expected src of %s, got %s", kissScript.Imports[i].Src, imports[i].Src)
		}
		if kissScript.Imports[i].Remote != imports[i].Remote {
			t.Errorf("Expected Remote to be %v, got %v", kissScript.Imports[i].Remote, imports[i].Remote)
		}
	}

	for i := 0; i < len(lines); i++ {
		line := ""
		for _, v := range kissScript.Lines[i].Value {
			line += v.Value
		}
		if line != lines[i] {
			t.Errorf("Expected line %s, got %s", lines[i], line)
		}
	}
}

// TODO: we need better test coverage on this stuff
