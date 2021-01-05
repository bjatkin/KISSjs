package main

import (
	"fmt"
	"regexp"
)

//
//	TODO: this should probably live in a sub package rather than the main package. this would probably clear up the namespace a lot to move it
//  -bja 01.04.2021

const (
	tokenTypeOpenImport = iota
	tokenTypeCloseImport
	tokenTypeKissKeyword
	tokenTypeKeyword
	tokenTypeValue
	tokenTypeOpenCloseString
	tokenTypeEscapedOpenCloseString
	tokenTypeWhiteSpace
	tokenTypeColon
	tokenTypeComma
	tokenTypeSemiColon
	tokenTypeCloseExpression
	tokenTypeOpenExpression
	tokenTypeOpenObject
	tokenTypeCloseObject
	tokenTypeDot
	tokenTypeEqual
	tokenTypeNewLine
	tokenTypeTemplate
	tokenTypeComment
	tokenTypeCommentStart
	tokenTypeBlockCommentStart
	tokenTypeBlockCommentEnd
	tokenTypeAny
)

var tokenPatterns = []JSTokenPattern{
	JSTokenPattern{tokenTypeOpenImport, regexp.MustCompile(`^\({`)},
	JSTokenPattern{tokenTypeCloseImport, regexp.MustCompile(`^}\)`)},
	JSTokenPattern{tokenTypeKissKeyword, regexp.MustCompile(`^KISSimport`)},
	JSTokenPattern{tokenTypeKissKeyword, regexp.MustCompile(`^remote`)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^function {0,1}`)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^var `)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^let `)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^yield `)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^new `)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^return `)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^const `)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^document`)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^async {0,1}`)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^await {0,1}`)},
	JSTokenPattern{tokenTypeKeyword, regexp.MustCompile(`^import {0,1}`)},
	JSTokenPattern{tokenTypeEqual, regexp.MustCompile(`^=`)},
	JSTokenPattern{tokenTypeValue, regexp.MustCompile(`^(?:true|false)`)},
	JSTokenPattern{tokenTypeEscapedOpenCloseString, regexp.MustCompile(`^\\['"]`)},
	JSTokenPattern{tokenTypeOpenCloseString, regexp.MustCompile(`^[\x60'"]`)},
	JSTokenPattern{tokenTypeWhiteSpace, regexp.MustCompile(`^[ \t]+`)},
	JSTokenPattern{tokenTypeNewLine, regexp.MustCompile(`^[\n\r]+`)},
	JSTokenPattern{tokenTypeSemiColon, regexp.MustCompile(`^;`)},
	JSTokenPattern{tokenTypeColon, regexp.MustCompile(`^:`)},
	JSTokenPattern{tokenTypeComma, regexp.MustCompile(`^,`)},
	JSTokenPattern{tokenTypeOpenObject, regexp.MustCompile(`^{`)},
	JSTokenPattern{tokenTypeCloseObject, regexp.MustCompile(`^}`)},
	JSTokenPattern{tokenTypeOpenExpression, regexp.MustCompile(`^[\(\[]`)},
	JSTokenPattern{tokenTypeCloseExpression, regexp.MustCompile(`^[\)\]]`)},
	JSTokenPattern{tokenTypeDot, regexp.MustCompile(`^\.`)},
	JSTokenPattern{tokenTypeTemplate, regexp.MustCompile(`^\$[_a-zA-Z][_a-zA-Z0-9]*\$`)},
	JSTokenPattern{tokenTypeCommentStart, regexp.MustCompile(`^\/\/`)},
	JSTokenPattern{tokenTypeBlockCommentStart, regexp.MustCompile(`^\/\*`)},
	JSTokenPattern{tokenTypeBlockCommentEnd, regexp.MustCompile(`^\*\/`)},
	JSTokenPattern{tokenTypeAny, regexp.MustCompile(`^.`)},
}

// JSTokenPattern matches a regex with a tokenType
type JSTokenPattern struct {
	tokenType int
	pattern   *regexp.Regexp
}

// JSToken is a token type and value
type JSToken struct {
	tokenType int
	value     string
}

// JSImport is a js import statment
type JSImport struct {
	src    string
	remote bool
}

// JSLine is a line of js
type JSLine struct {
	value []JSToken
}

// JSScript is a parsed js file
type JSScript struct {
	imports []JSImport
	lines   []JSLine
}

func (script JSScript) String() string {
	ret := ""
	for _, line := range script.lines {
		for _, token := range line.value {
			ret += token.value
		}
	}
	return ret
}

func (script JSScript) clone() JSScript {
	clone := JSScript{}
	for _, imp := range script.imports {
		clone.imports = append(clone.imports,
			JSImport{
				src:    imp.src,
				remote: imp.remote,
			},
		)
	}

	for _, line := range script.lines {
		newLine := JSLine{}
		for _, tok := range line.value {
			newLine.value = append(newLine.value, JSToken{tokenType: tok.tokenType, value: tok.value})
		}
		clone.lines = append(clone.lines, newLine)
	}

	return clone
}

func tokenizeJSScript(script string) []JSToken {
	tokens := []JSToken{}
	i := 0
	for i < len(script) {
		for _, token := range tokenPatterns {
			index := token.pattern.FindIndex([]byte(script[i:]))
			if index != nil {
				tokens = append(tokens,
					JSToken{
						tokenType: token.tokenType,
						value:     script[i : i+index[1]],
					},
				)
				i += index[1]
				break
			}
		}
	}

	// filter results
	ret := []JSToken{}
	i = 0
	for i < len(tokens) {
		if tokens[i].tokenType == tokenTypeWhiteSpace {
			i++
			continue
		}
		if tokens[i].tokenType == tokenTypeOpenCloseString {
			count, str := tokenizeJSString(tokens[i:])
			ret = append(ret, str)
			i += count
			continue
		}
		if tokens[i].tokenType == tokenTypeCommentStart ||
			tokens[i].tokenType == tokenTypeBlockCommentStart {
			count, _ := tokenizeJSComment(tokens[i:])
			i += count
			continue
		}
		ret = append(ret, tokens[i])
		i++
	}

	return ret
}

func tokenizeJSString(script []JSToken) (int, JSToken) {
	if script[0].tokenType != tokenTypeOpenCloseString {
		return 0, JSToken{}
	}
	ret := JSToken{tokenType: tokenTypeValue, value: script[0].value}
	i := 1
	for i < len(script) {
		tok := script[i].tokenType
		ret.value += script[i].value
		i++
		if tok == tokenTypeOpenCloseString {
			return i, ret
		}
	}
	return 0, JSToken{}
}

func tokenizeJSComment(script []JSToken) (int, JSToken) {
	if script[0].tokenType != tokenTypeCommentStart &&
		script[0].tokenType != tokenTypeBlockCommentStart {
		return 0, JSToken{}
	}
	ret := JSToken{tokenType: tokenTypeComment, value: script[0].value}
	i := 1
	endType := tokenTypeNewLine
	if script[0].tokenType == tokenTypeBlockCommentStart {
		endType = tokenTypeBlockCommentEnd
	}
	for i < len(script) {
		tok := script[i].tokenType
		ret.value += script[i].value
		i++
		if tok == endType {
			return i, ret
		}
	}
	return 0, JSToken{}
}

func parseJSTokens(script []JSToken) (JSScript, error) {
	ret := JSScript{}
	i := 0
	start := 0
	for i < len(script) {
		start = i
		tok := script[i].tokenType
		if tok == tokenTypeWhiteSpace {
			i++
			continue
		}
		count, jsImport := parseJSImportStatment(script[i:])
		if count > 0 {
			i += count
			ret.imports = append(ret.imports, jsImport)
			continue
		}
		count, line := parseJSLine(script[i:])
		if count > 0 {
			i += count
			ret.lines = append(ret.lines, line)
		}
		if i == start {
			return JSScript{}, fmt.Errorf("failed to parse js token, '%s'", script[i].value)
		}
	}

	ret.lines = addSemiColons(ret.lines)

	return ret, nil
}

func parseJSLine(script []JSToken) (int, JSLine) {
	ret := JSLine{}
	i := 0
	for script[i].tokenType == tokenTypeNewLine {
		i++
		if i >= len(script) {
			return i, JSLine{}
		}
	}
	for i < len(script) {
		tok := script[i].tokenType
		if tok == tokenTypeNewLine {
			i++
			break
		}
		if tok == tokenTypeSemiColon {
			i++
			break
		}
		ret.value = append(ret.value, script[i])
		i++
	}
	if len(ret.value) == 0 {
		return 0, ret
	}
	return i, ret
}

func addSemiColons(lines []JSLine) []JSLine {
	for i := 0; i < len(lines); i++ {
		line := &lines[i]
		if len(line.value) == 0 {
			continue
		}
		preTok := line.value[len(line.value)-1].tokenType
		postTok := tokenTypeAny
		if i+1 < len(lines) && len(lines[i+1].value) > 0 {
			postTok = lines[i+1].value[0].tokenType
		}

		add := true
		switch preTok {
		case tokenTypeSemiColon:
			add = false
		case tokenTypeOpenExpression:
			add = false
		case tokenTypeOpenObject:
			add = false
		case tokenTypeCloseObject:
			add = false
		case tokenTypeDot:
			add = false
		case tokenTypeEqual:
			add = false
		case tokenTypeColon:
			add = false
		case tokenTypeComma:
			add = false
		case tokenTypeOpenImport:
			add = false
		}

		switch postTok {
		case tokenTypeCloseExpression:
			add = false
		}

		if preTok == tokenTypeCloseObject && postTok == tokenTypeKeyword {
			add = true
		}
		if preTok == tokenTypeCloseObject && postTok == tokenTypeAny {
			add = true
		}
		if preTok == tokenTypeCloseExpression && postTok == tokenTypeCloseExpression {
			add = false
		}

		if add {
			line.value = append(line.value, JSToken{tokenTypeSemiColon, ";"})
		}
	}

	return lines
}

func parseJSImportStatment(script []JSToken) (int, JSImport) {
	i := 0
	for script[i].tokenType == tokenTypeNewLine {
		i++
		if i >= len(script) {
			return 0, JSImport{}
		}
	}
	if script[i].tokenType != tokenTypeOpenImport {
		return 0, JSImport{}
	}
	ret := JSImport{}
	expectedToken := tokenTypeKissKeyword
	keyword := ""
	i++
	for i < len(script) {
		tok := script[i].tokenType
		if tok == tokenTypeWhiteSpace {
			i++
			continue
		}
		if tok == tokenTypeCloseImport {
			i++
			if script[i].tokenType != tokenTypeSemiColon {
				break
			}
			i++
			return i, ret
		}

		if tok != expectedToken {
			break
		}
		switch tok {
		case tokenTypeKissKeyword:
			keyword = script[i].value
			expectedToken = tokenTypeColon
		case tokenTypeColon:
			expectedToken = tokenTypeValue
		case tokenTypeValue:
			val := script[i].value
			if keyword == "KISSimport" {
				ret.src = val[1 : len(val)-1]
			}
			if keyword == "remote" {
				ret.remote = (val == "true")
			}
			expectedToken = tokenTypeComma
		case tokenTypeComma:
			expectedToken = tokenTypeKissKeyword
		}
		i++
	}
	return 0, JSImport{}
}
