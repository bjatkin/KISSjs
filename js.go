package main

import (
	"fmt"
	"regexp"
)

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
	tokenTypeDot
	tokenTypeEqual
	tokenTypeNewLine
	tokenTypeAny
)

var tokenPatterns = []JSTokenPattern{
	JSTokenPattern{
		tokenType: tokenTypeOpenImport,
		pattern:   regexp.MustCompile(`^\({`),
	},
	JSTokenPattern{
		tokenType: tokenTypeCloseImport,
		pattern:   regexp.MustCompile(`^}\)`),
	},
	JSTokenPattern{
		tokenType: tokenTypeKissKeyword,
		pattern:   regexp.MustCompile(`^KISSimport`),
	},
	JSTokenPattern{
		tokenType: tokenTypeKissKeyword,
		pattern:   regexp.MustCompile(`^nocompile`),
	},
	JSTokenPattern{
		tokenType: tokenTypeKissKeyword,
		pattern:   regexp.MustCompile(`^nobundle`),
	},
	JSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^function {0,1}`),
	},
	JSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^var `),
	},
	JSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^let `),
	},
	JSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^yield `),
	},
	JSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^new `),
	},
	JSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^async {0,1}`),
	},
	JSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^await {0,1}`),
	},
	JSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^import {0,1}`),
	},
	JSTokenPattern{
		tokenType: tokenTypeEqual,
		pattern:   regexp.MustCompile(`^=`),
	},
	JSTokenPattern{
		tokenType: tokenTypeValue,
		pattern:   regexp.MustCompile(`^(?:true|false)`),
	},
	JSTokenPattern{
		tokenType: tokenTypeEscapedOpenCloseString,
		pattern:   regexp.MustCompile(`^\\['"]`),
	},
	JSTokenPattern{
		tokenType: tokenTypeOpenCloseString,
		pattern:   regexp.MustCompile(`^[\x60'"]`),
	},
	JSTokenPattern{
		tokenType: tokenTypeWhiteSpace,
		pattern:   regexp.MustCompile(`^[ \t]+`),
	},
	JSTokenPattern{
		tokenType: tokenTypeNewLine,
		pattern:   regexp.MustCompile(`^[\n\r]+`),
	},
	JSTokenPattern{
		tokenType: tokenTypeSemiColon,
		pattern:   regexp.MustCompile(`^;`),
	},
	JSTokenPattern{
		tokenType: tokenTypeColon,
		pattern:   regexp.MustCompile(`^:`),
	},
	JSTokenPattern{
		tokenType: tokenTypeComma,
		pattern:   regexp.MustCompile(`^,`),
	},
	JSTokenPattern{
		tokenType: tokenTypeOpenExpression,
		pattern:   regexp.MustCompile(`^[\(\[{]`),
	},
	JSTokenPattern{
		tokenType: tokenTypeCloseExpression,
		pattern:   regexp.MustCompile(`^[\)\]}]`),
	},
	JSTokenPattern{
		tokenType: tokenTypeDot,
		pattern:   regexp.MustCompile(`^\.`),
	},
	JSTokenPattern{
		tokenType: tokenTypeAny,
		pattern:   regexp.MustCompile(`^.`),
	},
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
	src                 string
	nobundle, nocompile bool
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
	tok := ret.value[len(ret.value)-1].tokenType
	switch tok {
	case tokenTypeSemiColon:
		return i, ret
	case tokenTypeOpenExpression:
		return i, ret
	case tokenTypeCloseExpression:
		return i, ret
	case tokenTypeEqual:
		return i, ret
	case tokenTypeColon:
		return i, ret
	case tokenTypeComma:
		return i, ret
	case tokenTypeOpenImport:
		return i, ret
	default:
		ret.value = append(ret.value, JSToken{tokenTypeSemiColon, ";"})
		return i, ret
	}
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
			if keyword == "nobundle" {
				ret.nobundle = (val == "true")
			}
			if keyword == "nocompile" {
				ret.nocompile = (val == "true")
			}
			expectedToken = tokenTypeComma
		case tokenTypeComma:
			expectedToken = tokenTypeKissKeyword
		}
		i++
	}
	return 0, JSImport{}
}
