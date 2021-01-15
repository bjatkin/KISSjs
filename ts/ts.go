package ts

import (
	"fmt"
	"regexp"
)

// These are the typscript token types
const (
	Keyword = iota
	Export
	OpenObject
	CloseObject
	OpenCloseString
	EscapedToken
	Star
	Comma
	NewLine
	Value
	WhiteSpace
	Comment
	CommentStart
	BlockCommentStart
	BlockCommentEnd
	Any
)

var tokenPatterns = []tokenPattern{
	tokenPattern{Keyword, regexp.MustCompile(`^(import|from)`)},
	tokenPattern{Export, regexp.MustCompile(`^export`)},
	tokenPattern{OpenObject, regexp.MustCompile(`^{`)},
	tokenPattern{CloseObject, regexp.MustCompile(`^}`)},
	tokenPattern{OpenCloseString, regexp.MustCompile(`^[\x60'"]`)},
	tokenPattern{EscapedToken, regexp.MustCompile(`^\\.`)},
	tokenPattern{Star, regexp.MustCompile(`^\*`)},
	tokenPattern{Comma, regexp.MustCompile(`^,`)},
	tokenPattern{WhiteSpace, regexp.MustCompile(`^[ \t]+`)},
	tokenPattern{NewLine, regexp.MustCompile(`^[\n\r]+`)},
	tokenPattern{Value, regexp.MustCompile(`^\$[_a-zA-Z][_a-zA-Z0-9]*\$`)},
}

// tokenPattern matches a regex with a tokenType
type tokenPattern struct {
	tType   int
	pattern *regexp.Regexp
}

// Token is a token type and value
type Token struct {
	Type    int
	Value   string
	LineNum int
}

// Script is a parsed ts file
type Script struct {
	Imports []string
	Tokens  []Token
}

func (script Script) String() string {
	var ret string
	for _, token := range script.Tokens {
		ret += token.Value
	}
	return ret
}

// Clone preforms a deep clone of the script object
func (script Script) Clone() Script {
	clone := Script{}
	for _, imp := range script.Imports {
		clone.Imports = append(clone.Imports, imp)
	}

	for _, tok := range script.Tokens {
		clone.Tokens = append(clone.Tokens, Token{tok.Type, tok.Value, tok.LineNum})
	}

	return clone
}

// Lex lexes a ts script and returns a series of tokens
func Lex(script string) []Token {
	tokens := []Token{}
	var i, start, line int
	for i < len(script) {
		start = i
		for _, token := range tokenPatterns {
			index := token.pattern.FindIndex([]byte(script[i:]))
			if index != nil {
				tokens = append(tokens,
					Token{
						Type:    token.tType,
						Value:   script[i : i+index[1]],
						LineNum: line,
					},
				)
				i += index[1]
				if token.tType == NewLine {
					line++
				}
				break
			}
		}
		if i == start {
			tokens = append(tokens,
				Token{
					Type:    Any,
					Value:   script[i : i+1],
					LineNum: line,
				},
			)
			i++
		}
	}

	// lex higher level tokens together
	ret := []Token{}
	i = 0
	for i < len(tokens) {
		if tokens[i].Type == Export {
			i++
			continue
		}
		if tokens[i].Type == OpenCloseString {
			count, str := lexString(tokens[i:])
			ret = append(ret, str)
			i += count
			continue
		}
		if tokens[i].Type == CommentStart ||
			tokens[i].Type == BlockCommentStart {
			count := lexComment(tokens[i:])
			i += count
			continue
		}
		ret = append(ret, tokens[i])
		i++
	}

	return ret
}

func lexString(script []Token) (int, Token) {
	open := script[0].Value
	ret := Token{Type: Value, Value: open, LineNum: script[0].LineNum}
	i := 1
	for i < len(script) {
		tok := script[i]
		ret.Value += tok.Value
		i++
		if tok.Type == OpenCloseString && tok.Value == open {
			return i, ret
		}
	}
	return 0, Token{}
}

func lexComment(script []Token) int {
	end := NewLine
	if script[0].Type == BlockCommentStart {
		end = BlockCommentEnd
	}

	i := 1
	for i < len(script) {
		if script[i].Type == end {
			return i
		}
	}
	return 0
}

// Parse will parse a series of tokens passed from the lexer into the Script struct
func Parse(script []Token) (Script, error) {
	ret := Script{}
	var i int
	for i < len(script) {
		tok := script[i]
		if tok.Value == "import" {
			count, tsImport := parseImport(script[i:])
			if count == 0 {
				return ret, fmt.Errorf("could not parse import statment on line %d", script[i].LineNum)
			}
			i += count
			ret.Imports = append(ret.Imports, tsImport)
			continue
		}
		ret.Tokens = append(ret.Tokens, tok)
		i++
	}

	return ret, nil
}

// parseImport parses an import from script tokens
func parseImport(script []Token) (int, string) {
	var i int
	var ret string
	next := Keyword //import
	for i < len(script) {
		tok := script[i]
		if tok.Type == WhiteSpace || tok.Type == Any || tok.Type == Comma {
			i++
			continue
		}
		if tok.Type != next {
			return 0, ""
		}

		if tok.Type == NewLine {
			if ret == "" {
				return 0, ""
			}
			return i + 1, ret
		}

		if tok.Type == Keyword {
			if tok.Value == "import" {
				next = OpenObject
			}
			if tok.Value == "from" {
				next = Value
			}
		}

		if tok.Type == OpenObject {
			next = CloseObject
		}
		if tok.Type == CloseObject {
			next = Keyword
		}

		if tok.Type == Value {
			next = NewLine
			ret = tok.Value[1:len(tok.Value)-1] + ".ts"
		}
		i++
	}
	return 0, ""
}
