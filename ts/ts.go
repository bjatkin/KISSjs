package ts

import (
	"fmt"
	"regexp"
)

const (
	keyword = iota
	openObject
	closeObject
	openCloseString
	escapedToken
	star
	comma
	newLine
	value
	whiteSpace
	comment
	commentStart
	blockCommentStart
	blockCommentEnd
	any
)

var tokenPatterns = []tokenPattern{
	tokenPattern{keyword, regexp.MustCompile(`^(import|from)`)},
	tokenPattern{openObject, regexp.MustCompile(`^{`)},
	tokenPattern{closeObject, regexp.MustCompile(`^}`)},
	tokenPattern{openCloseString, regexp.MustCompile(`^[\x60'"]`)},
	tokenPattern{escapedToken, regexp.MustCompile(`^\\.`)},
	tokenPattern{star, regexp.MustCompile(`^\*`)},
	tokenPattern{comma, regexp.MustCompile(`^,`)},
	tokenPattern{whiteSpace, regexp.MustCompile(`^[ \t]+`)},
	tokenPattern{newLine, regexp.MustCompile(`^[\n\r]+`)},
	tokenPattern{value, regexp.MustCompile(`^\$[_a-zA-Z][_a-zA-Z0-9]*\$`)},
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
				if token.tType == newLine {
					line++
				}
				break
			}
		}
		if i == start {
			tokens = append(tokens,
				Token{
					Type:    any,
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
		if tokens[i].Type == openCloseString {
			count, str := lexString(tokens[i:])
			ret = append(ret, str)
			i += count
			continue
		}
		if tokens[i].Type == commentStart ||
			tokens[i].Type == blockCommentStart {
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
	ret := Token{Type: value, Value: open, LineNum: script[0].LineNum}
	i := 1
	for i < len(script) {
		tok := script[i]
		ret.Value += tok.Value
		i++
		if tok.Type == openCloseString && tok.Value == open {
			return i, ret
		}
	}
	return 0, Token{}
}

func lexComment(script []Token) int {
	end := newLine
	if script[0].Type == blockCommentStart {
		end = blockCommentEnd
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
	next := keyword //import
	for i < len(script) {
		tok := script[i]
		if tok.Type == whiteSpace || tok.Type == any || tok.Type == comma {
			i++
			continue
		}
		if tok.Type != next {
			return 0, ""
		}

		if tok.Type == newLine {
			if ret == "" {
				return 0, ""
			}
			return i + 1, ret
		}

		if tok.Type == keyword {
			if tok.Value == "import" {
				next = openObject
			}
			if tok.Value == "from" {
				next = value
			}
		}

		if tok.Type == openObject {
			next = closeObject
		}
		if tok.Type == closeObject {
			next = keyword
		}

		if tok.Type == value {
			next = newLine
			ret = tok.Value
		}
		i++
	}
	return 0, ""
}
