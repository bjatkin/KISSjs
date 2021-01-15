package ts

import (
	"testing"
)

func TestLex(t *testing.T) {
	type test struct {
		ts    string
		check []Token
	}

	tests := []test{
		test{
			ts: `import {t, d} from '../ts/utils'
			let b: int = $b$;`,
			check: []Token{
				Token{Keyword, "import", 0},
				Token{WhiteSpace, " ", 0},
				Token{OpenObject, "{", 0},
				Token{Any, "t", 0},
				Token{Comma, ",", 0},
				Token{WhiteSpace, " ", 0},
				Token{Any, "d", 0},
				Token{CloseObject, "}", 0},
				Token{WhiteSpace, " ", 0},
				Token{Keyword, "from", 0},
				Token{WhiteSpace, " ", 0},
				Token{Value, "'../ts/utils'", 0},
				Token{NewLine, "\n", 0},
				Token{WhiteSpace, "			", 1},
				Token{Any, "l", 1},
				Token{Any, "e", 1},
				Token{Any, "t", 1},
				Token{WhiteSpace, " ", 1},
				Token{Any, "b", 1},
				Token{Any, ":", 1},
				Token{WhiteSpace, " ", 1},
				Token{Any, "i", 1},
				Token{Any, "n", 1},
				Token{Any, "t", 1},
				Token{WhiteSpace, " ", 1},
				Token{Any, "=", 1},
				Token{WhiteSpace, " ", 1},
				Token{Value, "$b$", 1},
				Token{Any, ";", 1},
			},
		},
	}

	for i, run := range tests {
		tokens := Lex(run.ts)

		if len(run.check) != len(tokens) {
			t.Errorf("(%d) wrong number of tokens expected %d but got %d", i, len(run.check), len(tokens))
		}

		for ii, tok := range run.check {
			if tok.Value != tokens[ii].Value {
				t.Errorf("(%d|%d) wrong token value, expected %s but got %s", i, ii, tok.Value, tokens[ii].Value)
			}

			if tok.Type != tokens[ii].Type {
				t.Errorf("(%d|%d) wrong token type, expected %d but got %d", i, ii, tok.Type, tokens[ii].Type)
			}

			if tok.LineNum != tokens[ii].LineNum {
				t.Errorf("(%d|%d) wrong token line number, expected %d but got %d", i, ii, tok.LineNum, tokens[ii].LineNum)
			}
		}
	}
}

func TestParse(t *testing.T) {
	type test struct {
		ts    string
		check Script
	}

	tests := []test{
		test{
			ts: `import {a, d} from "../test/script"
import {} from "hello/world"
let a: bool = true;`,
			check: Script{
				Imports: []string{
					"../test/script.ts",
					"hello/world.ts",
				},
				Tokens: []Token{
					Token{Any, "l", 2},
					Token{Any, "e", 2},
					Token{Any, "t", 2},
					Token{WhiteSpace, " ", 2},
					Token{Any, "a", 2},
					Token{Any, ":", 2},
					Token{WhiteSpace, " ", 2},
					Token{Any, "b", 2},
					Token{Any, "o", 2},
					Token{Any, "o", 2},
					Token{Any, "l", 2},
					Token{WhiteSpace, " ", 2},
					Token{Any, "=", 2},
					Token{WhiteSpace, " ", 2},
					Token{Any, "t", 2},
					Token{Any, "r", 2},
					Token{Any, "u", 2},
					Token{Any, "e", 2},
					Token{Any, ";", 2},
				},
			},
		},
	}

	for i, run := range tests {
		tokens := Lex(run.ts)
		script, err := Parse(tokens)
		if err != nil {
			t.Errorf("(%d) there was an error parsing the script %s", i, err)
		}

		if len(script.Imports) != len(run.check.Imports) {
			t.Errorf("(%d) wrong number of imports got %d, but expected %d", i, len(script.Imports), len(run.check.Imports))
		}

		for ii, imp := range run.check.Imports {
			if imp != script.Imports[ii] {
				t.Errorf("(%d|%d) expected import %s, but got %s", i, ii, imp, script.Imports[ii])
			}
		}

		if len(script.Tokens) != len(run.check.Tokens) {
			t.Errorf("(%d) wrong number of tokens got %d, but expected %d", i, len(script.Tokens), len(run.check.Tokens))

		}

		for ii, tok := range run.check.Tokens {
			if tok.Value != script.Tokens[ii].Value {
				t.Errorf("(%d|%d) wrong token value got %s, but expected %s", i, ii, tok.Value, script.Tokens[ii].Value)
			}

			if tok.Type != script.Tokens[ii].Type {
				t.Errorf("(%d|%d) wrong token type got %d, expected %d", i, ii, tok.Type, script.Tokens[ii].Type)
			}

			if tok.LineNum != script.Tokens[ii].LineNum {
				t.Errorf("(%d|%d) wrong token line number got %d, expected %d", i, ii, tok.LineNum, script.Tokens[ii].Type)
			}
		}
	}
}
