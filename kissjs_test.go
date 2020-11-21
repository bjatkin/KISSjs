package main

import (
	"testing"
)

func TestTokenizeJSScript(t *testing.T) {
	script := `({KISSimport:"t.js",nobundle:true})`
	ok := []kissJSToken{
		kissJSToken{tokenType: 0, value: "({"},
		kissJSToken{tokenType: 2, value: "KISSimport"},
		kissJSToken{tokenType: 8, value: ":"},
		kissJSToken{tokenType: 5, value: "\""},
		kissJSToken{tokenType: 16, value: "t"},
		kissJSToken{tokenType: 13, value: "."},
		kissJSToken{tokenType: 16, value: "j"},
		kissJSToken{tokenType: 16, value: "s"},
		kissJSToken{tokenType: 5, value: "\""},
		kissJSToken{tokenType: 9, value: ","},
		kissJSToken{tokenType: 2, value: "nobundle"},
		kissJSToken{tokenType: 8, value: ":"},
		kissJSToken{tokenType: 4, value: "true"},
		kissJSToken{tokenType: 1, value: "})"}}
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

	script = `console.log("\"Test\"");
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
		kissJSToken{tokenType: 5, value: "\""},
		kissJSToken{tokenType: 6, value: "\\\""},
		kissJSToken{tokenType: 16, value: "T"},
		kissJSToken{tokenType: 16, value: "e"},
		kissJSToken{tokenType: 16, value: "s"},
		kissJSToken{tokenType: 16, value: "t"},
		kissJSToken{tokenType: 6, value: "\\\""},
		kissJSToken{tokenType: 5, value: "\""},
		kissJSToken{tokenType: 11, value: ")"},
		kissJSToken{tokenType: 10, value: ";"},
		kissJSToken{tokenType: 15, value: "\n"},
		kissJSToken{tokenType: 3, value: "let "},
		kissJSToken{tokenType: 16, value: "a"},
		kissJSToken{tokenType: 7, value: " "},
		kissJSToken{tokenType: 14, value: "="},
		kissJSToken{tokenType: 7, value: " "},
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
