package ts

import (
	"fmt"
	"testing"
)

func TestLex(t *testing.T) {
	type test struct {
		ts    string
		check []Token
	}

	tests := []test{
		test{
			ts: `import {t, d} from '../ts/utils.ts'
			let b: int = $b$;`,
			check: []Token{
				Token{keyword, "import", 0},
				Token{whiteSpace, " ", 0},
				Token{openObject, "{", 0},
				Token{any, "t", 0},
				Token{comma, ",", 0},
				Token{whiteSpace, " ", 0},
				Token{any, "d", 0},
				Token{closeObject, "}", 0},
				Token{whiteSpace, " ", 0},
				Token{keyword, "from", 0},
				Token{whiteSpace, " ", 0},
				Token{value, "'../ts/utils.ts'", 0},
				Token{newLine, "\n", 0},
				Token{whiteSpace, "			", 1},
				Token{any, "l", 1},
				Token{any, "e", 1},
				Token{any, "t", 1},
				Token{whiteSpace, " ", 1},
				Token{any, "b", 1},
				Token{any, ":", 1},
				Token{whiteSpace, " ", 1},
				Token{any, "i", 1},
				Token{any, "n", 1},
				Token{any, "t", 1},
				Token{whiteSpace, " ", 1},
				Token{any, "=", 1},
				Token{whiteSpace, " ", 1},
				Token{value, "$b$", 1},
				Token{any, ";", 1},
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
			ts: `import {a, d} from "../test/script.ts"
import {} from "../hello/world.ts"
let a: bool = true;`,
			check: Script{
				Imports: []string{
					"../test/script.ts",
					"hello/world.ts",
				},
				Tokens: []Token{
					Token{any, "l", 2},
					Token{any, "e", 2},
					Token{any, "t", 2},
					Token{whiteSpace, " ", 2},
					Token{any, "a", 2},
					Token{any, ":", 2},
					Token{whiteSpace, " ", 2},
					Token{any, "b", 2},
					Token{any, "o", 2},
					Token{any, "o", 2},
					Token{any, "l", 2},
					Token{whiteSpace, " ", 2},
					Token{any, "=", 2},
					Token{whiteSpace, " ", 2},
					Token{any, "t", 2},
					Token{any, "r", 2},
					Token{any, "u", 2},
					Token{any, "e", 2},
					Token{any, ";", 2},
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
		if len(script.Tokens) != len(run.check.Tokens) {
			fmt.Println("\n\n\n", script.Tokens)
			t.Errorf("(%d) wrong number of tokens got %d, but expected %d", i, len(script.Tokens), len(run.check.Tokens))

		}

		// TODO: finish this
	}
}
