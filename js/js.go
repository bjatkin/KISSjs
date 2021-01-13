package js

import (
	"fmt"
	"regexp"
)

const (
	openImport = iota
	closeImport
	kissKeyword
	keyword
	value
	openCloseString
	escapedOpenCloseString
	whiteSpace
	colon
	comma
	semiColon
	closeExpression
	openExpression
	openObject
	closeObject
	dot
	equal
	newLine
	template
	comment
	commentStart
	blockCommentStart
	blockCommentEnd
	logicalOp
	any
)

var tokenPatterns = []tokenPattern{
	tokenPattern{openImport, regexp.MustCompile(`^\({`)},
	tokenPattern{closeImport, regexp.MustCompile(`^}\)`)},
	tokenPattern{kissKeyword, regexp.MustCompile(`^KISSimport`)},
	tokenPattern{kissKeyword, regexp.MustCompile(`^remote`)},
	tokenPattern{keyword, regexp.MustCompile(`^function {0,1}`)},
	tokenPattern{keyword, regexp.MustCompile(`^var `)},
	tokenPattern{keyword, regexp.MustCompile(`^let `)},
	tokenPattern{keyword, regexp.MustCompile(`^yield `)},
	tokenPattern{keyword, regexp.MustCompile(`^new `)},
	tokenPattern{keyword, regexp.MustCompile(`^return `)},
	tokenPattern{keyword, regexp.MustCompile(`^const `)},
	tokenPattern{keyword, regexp.MustCompile(`^document`)},
	tokenPattern{keyword, regexp.MustCompile(`^async {0,1}`)},
	tokenPattern{keyword, regexp.MustCompile(`^await {0,1}`)},
	tokenPattern{keyword, regexp.MustCompile(`^import {0,1}`)},
	tokenPattern{equal, regexp.MustCompile(`^=`)},
	tokenPattern{value, regexp.MustCompile(`^(?:true|false)`)},
	tokenPattern{escapedOpenCloseString, regexp.MustCompile(`^\\['"]`)},
	tokenPattern{openCloseString, regexp.MustCompile(`^[\x60'"]`)},
	tokenPattern{whiteSpace, regexp.MustCompile(`^[ \t]+`)},
	tokenPattern{newLine, regexp.MustCompile(`^[\n\r]+`)},
	tokenPattern{semiColon, regexp.MustCompile(`^;`)},
	tokenPattern{colon, regexp.MustCompile(`^:`)},
	tokenPattern{comma, regexp.MustCompile(`^,`)},
	tokenPattern{openObject, regexp.MustCompile(`^{`)},
	tokenPattern{closeObject, regexp.MustCompile(`^}`)},
	tokenPattern{openExpression, regexp.MustCompile(`^[\(\[]`)},
	tokenPattern{closeExpression, regexp.MustCompile(`^[\)\]]`)},
	tokenPattern{dot, regexp.MustCompile(`^\.`)},
	tokenPattern{template, regexp.MustCompile(`^\$[_a-zA-Z][_a-zA-Z0-9]*\$`)},
	tokenPattern{commentStart, regexp.MustCompile(`^\/\/`)},
	tokenPattern{blockCommentStart, regexp.MustCompile(`^\/\*`)},
	tokenPattern{blockCommentEnd, regexp.MustCompile(`^\*\/`)},
	tokenPattern{logicalOp, regexp.MustCompile(`^(&&|\|\|)`)},
	tokenPattern{any, regexp.MustCompile(`^.`)},
}

// tokenPattern matches a regex with a tokenType
type tokenPattern struct {
	tType   int
	pattern *regexp.Regexp
}

// Token is a token type and value
type Token struct {
	Type  int
	Value string
}

// Import is a js import statment
type Import struct {
	Src    string
	Remote bool
}

// Line is a line of js
type Line struct {
	Value []Token
}

// Script is a parsed js file
type Script struct {
	Imports []Import
	Lines   []Line
}

func (script Script) String() string {
	ret := ""
	for _, line := range script.Lines {
		for _, token := range line.Value {
			ret += token.Value
		}
	}
	return ret
}

// Clone preforms a deep clone of the script object
func (script Script) Clone() Script {
	clone := Script{}
	for _, imp := range script.Imports {
		clone.Imports = append(clone.Imports,
			Import{
				Src:    imp.Src,
				Remote: imp.Remote,
			},
		)
	}

	for _, line := range script.Lines {
		newLine := Line{}
		for _, tok := range line.Value {
			newLine.Value = append(newLine.Value, Token{Type: tok.Type, Value: tok.Value})
		}
		clone.Lines = append(clone.Lines, newLine)
	}

	return clone
}

// LexScript lexes a js script and returns a series of tokens
func LexScript(script string) []Token {
	tokens := []Token{}
	i := 0
	for i < len(script) {
		for _, token := range tokenPatterns {
			index := token.pattern.FindIndex([]byte(script[i:]))
			if index != nil {
				tokens = append(tokens,
					Token{
						Type:  token.tType,
						Value: script[i : i+index[1]],
					},
				)
				i += index[1]
				break
			}
		}
	}

	// filter results
	ret := []Token{}
	i = 0
	for i < len(tokens) {
		if tokens[i].Type == whiteSpace {
			i++
			continue
		}
		if tokens[i].Type == openCloseString {
			count, str := lexString(tokens[i:])
			ret = append(ret, str)
			i += count
			continue
		}
		if tokens[i].Type == commentStart ||
			tokens[i].Type == blockCommentStart {
			count, _ := lexComment(tokens[i:])
			i += count
			continue
		}
		ret = append(ret, tokens[i])
		i++
	}

	return ret
}

func lexString(script []Token) (int, Token) {
	if script[0].Type != openCloseString {
		return 0, Token{}
	}
	ret := Token{Type: value, Value: script[0].Value}
	i := 1
	for i < len(script) {
		tok := script[i].Type
		ret.Value += script[i].Value
		i++
		if tok == openCloseString {
			return i, ret
		}
	}
	return 0, Token{}
}

func lexComment(script []Token) (int, Token) {
	if script[0].Type != commentStart &&
		script[0].Type != blockCommentStart {
		return 0, Token{}
	}
	ret := Token{Type: comment, Value: script[0].Value}
	i := 1
	endType := newLine
	if script[0].Type == blockCommentStart {
		endType = blockCommentEnd
	}
	for i < len(script) {
		tok := script[i].Type
		ret.Value += script[i].Value
		i++
		if tok == endType {
			return i, ret
		}
	}
	return 0, Token{}
}

// ParseTokens will parse a series of tokens passed from the lexer into a Script object
func ParseTokens(script []Token) (Script, error) {
	ret := Script{}
	i := 0
	start := 0
	for i < len(script) {
		start = i
		if script[i].Type == whiteSpace {
			i++
			continue
		}
		count, jsImport := parseImportStatment(script[i:])
		if count > 0 {
			i += count
			ret.Imports = append(ret.Imports, jsImport)
			continue
		}
		count, line := parseLine(script[i:])
		if count > 0 {
			i += count
			ret.Lines = append(ret.Lines, line)
		}
		if i == start {
			return Script{}, fmt.Errorf("failed to parse js token, '%s'", script[i].Value)
		}
	}

	ret.Lines = addSemiColons(ret.Lines)

	return ret, nil
}

func parseLine(script []Token) (int, Line) {
	ret := Line{}
	i := 0
	for script[i].Type == newLine {
		i++
		if i >= len(script) {
			return i, Line{}
		}
	}
	for i < len(script) {
		tok := script[i].Type
		if tok == newLine {
			i++
			break
		}
		if tok == semiColon {
			i++
			break
		}
		ret.Value = append(ret.Value, script[i])
		i++
	}
	if len(ret.Value) == 0 {
		return 0, ret
	}
	return i, ret
}

func addSemiColons(lines []Line) []Line {
	for i := 0; i < len(lines); i++ {
		line := &lines[i]
		if len(line.Value) == 0 {
			continue
		}
		preTok := line.Value[len(line.Value)-1].Type
		postTok := any
		if i+1 < len(lines) && len(lines[i+1].Value) > 0 {
			postTok = lines[i+1].Value[0].Type
		}

		add := true
		switch preTok {
		case semiColon:
			add = false
		case openExpression:
			add = false
		case openObject:
			add = false
		case closeObject:
			add = false
		case dot:
			add = false
		case equal:
			add = false
		case colon:
			add = false
		case comma:
			add = false
		case openImport:
			add = false
		case logicalOp:
			add = false
		}

		switch postTok {
		case closeExpression:
			add = false
		case closeObject:
			add = false
		case closeImport:
			add = false
		}

		if preTok == closeObject && postTok == keyword {
			add = true
		}
		if preTok == closeObject && postTok == any {
			add = true
		}
		if preTok == closeExpression && postTok == closeExpression {
			add = false
		}

		if add {
			line.Value = append(line.Value, Token{semiColon, ";"})
		}
	}

	return lines
}

func parseImportStatment(script []Token) (int, Import) {
	i := 0
	for script[i].Type == newLine {
		i++
		if i >= len(script) {
			return 0, Import{}
		}
	}
	if script[i].Type != openImport {
		return 0, Import{}
	}
	ret := Import{}
	expectedToken := kissKeyword
	keyword := ""
	i++
	for i < len(script) {
		tok := script[i].Type
		if tok == whiteSpace {
			i++
			continue
		}
		if tok == closeImport {
			i++
			if script[i].Type != semiColon {
				break
			}
			i++
			return i, ret
		}

		if tok != expectedToken {
			break
		}
		switch tok {
		case kissKeyword:
			keyword = script[i].Value
			expectedToken = colon
		case colon:
			expectedToken = value
		case value:
			val := script[i].Value
			if keyword == "KISSimport" {
				ret.Src = val[1 : len(val)-1]
			}
			if keyword == "remote" {
				ret.Remote = (val == "true")
			}
			expectedToken = comma
		case comma:
			expectedToken = kissKeyword
		}
		i++
	}
	return 0, Import{}
}
