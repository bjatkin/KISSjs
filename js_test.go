package main

import (
	"testing"
)

func TestTokenizeJSScript(t *testing.T) {
	script := `({KISSimport:"t.js",nobundle:true});`
	ok := []kissJSToken{
		kissJSToken{tokenType: 0, value: "({"},
		kissJSToken{tokenType: 2, value: "KISSimport"},
		kissJSToken{tokenType: 8, value: ":"},
		kissJSToken{tokenType: 4, value: "\"t.js\""},
		kissJSToken{tokenType: 9, value: ","},
		kissJSToken{tokenType: 2, value: "nobundle"},
		kissJSToken{tokenType: 8, value: ":"},
		kissJSToken{tokenType: 4, value: "true"},
		kissJSToken{tokenType: 1, value: "})"},
		kissJSToken{tokenType: 10, value: ";"}}
	tokens := tokenizeJSScript(script)
	if len(ok) != len(tokens) {
		t.Errorf("Expecting %d tokens, but got %d", len(ok), len(tokens))
	}
	for i := 0; i < len(ok); i++ {
		if ok[i].tokenType != tokens[i].tokenType {
			t.Errorf("Expecting tokenType %d, but got %d", ok[i].tokenType, tokens[i].tokenType)
		}
		if ok[i].value != tokens[i].value {
			t.Errorf("Expecting value of %s, but got %s", ok[i].value, tokens[i].value)
		}
	}

	script = `console.log( "\"Test\"" );
let a = 10`
	tokens = tokenizeJSScript(script)
	ok = []kissJSToken{
		kissJSToken{tokenType: 16, value: "c"},
		kissJSToken{tokenType: 16, value: "o"},
		kissJSToken{tokenType: 16, value: "n"},
		kissJSToken{tokenType: 16, value: "s"},
		kissJSToken{tokenType: 16, value: "o"},
		kissJSToken{tokenType: 16, value: "l"},
		kissJSToken{tokenType: 16, value: "e"},
		kissJSToken{tokenType: 13, value: "."},
		kissJSToken{tokenType: 16, value: "l"},
		kissJSToken{tokenType: 16, value: "o"},
		kissJSToken{tokenType: 16, value: "g"},
		kissJSToken{tokenType: 12, value: "("},
		kissJSToken{tokenType: 4, value: "\"\\\"Test\\\"\""},
		kissJSToken{tokenType: 11, value: ")"},
		kissJSToken{tokenType: 10, value: ";"},
		kissJSToken{tokenType: 15, value: "\n"},
		kissJSToken{tokenType: 3, value: "let "},
		kissJSToken{tokenType: 16, value: "a"},
		kissJSToken{tokenType: 14, value: "="},
		kissJSToken{tokenType: 16, value: "1"},
		kissJSToken{tokenType: 16, value: "0"},
	}

	if len(ok) != len(tokens) {
		t.Errorf("Expecting %d tokens, but got %d", len(ok), len(tokens))
	}
	for i := 0; i < len(ok); i++ {
		if ok[i].tokenType != tokens[i].tokenType {
			t.Errorf("Expecting tokenType %d, but got %d", ok[i].tokenType, tokens[i].tokenType)
		}
		if ok[i].value != tokens[i].value {
			t.Errorf("Expecting value of %s but got %s", ok[i].value, tokens[i].value)
		}
	}
}

func TestParseJSTokens(t *testing.T) {
	script := `
	({KISSimport: "test.js", nobundle: true});
	({KISSimport: "again.js", nocompile: true});
	({KISSimport: "last.js", nobundle: false});

	let a = 10;
	let b = "50";
	console.log(a + b);
	var t = test({
		hi: 5,
	});
	`
	imports := []kissJSImport{
		kissJSImport{src: "test.js", nobundle: true},
		kissJSImport{src: "again.js", nocompile: true},
		kissJSImport{src: "last.js"},
	}
	lines := []string{
		"let a=10;",
		"let b=\"50\";",
		"console.log(a+b)",
		"var t=test({",
		"hi:5,",
		"});",
		"",
	}

	tokens := tokenizeJSScript(script)
	kissScript, err := parseJSTokens(tokens)
	if err != nil {
		t.Errorf("there was an error parsing the js script %s", err)
	}
	if len(kissScript.imports) != len(imports) {
		t.Errorf("Expected %d imports, got %d", len(kissScript.imports), len(imports))
	}
	if len(kissScript.lines) != len(lines) {
		t.Errorf("Expected %d lines, got %d", len(kissScript.lines), len(lines))
	}

	for i := 0; i < len(imports); i++ {
		if kissScript.imports[i].src != imports[i].src {
			t.Errorf("Expected src of %s, got %s", kissScript.imports[i].src, imports[i].src)
		}
		if kissScript.imports[i].nobundle != imports[i].nobundle {
			t.Errorf("Expected nobundle to be %v, got %v", kissScript.imports[i].nobundle, imports[i].nobundle)
		}
		if kissScript.imports[i].nocompile != imports[i].nocompile {
			t.Errorf("Expected nocompile to be %v, got %v", kissScript.imports[i].nocompile, imports[i].nocompile)
		}
	}

	for i := 0; i < len(lines); i++ {
		line := ""
		for _, v := range kissScript.lines[i].value {
			line += v.value
		}
		if line != lines[i] {
			t.Errorf("Expected line %s, got %s", lines[i], line)
		}
	}
}
