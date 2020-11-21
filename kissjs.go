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

var tokenPatterns = []kissJSTokenPattern{
	kissJSTokenPattern{
		tokenType: tokenTypeOpenImport,
		pattern:   regexp.MustCompile(`^\({`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeCloseImport,
		pattern:   regexp.MustCompile(`^}\)`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKissKeyword,
		pattern:   regexp.MustCompile(`^KISSimport`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKissKeyword,
		pattern:   regexp.MustCompile(`^nocompile`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKissKeyword,
		pattern:   regexp.MustCompile(`^nobundle`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^function {0,1}`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^var `),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^let `),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^yield `),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^new `),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^async {0,1}`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^await {0,1}`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^import {0,1}`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeEqual,
		pattern:   regexp.MustCompile(`^=`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeValue,
		pattern:   regexp.MustCompile(`^(?:true|false)`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeEscapedOpenCloseString,
		pattern:   regexp.MustCompile(`^\\['"]`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeOpenCloseString,
		pattern:   regexp.MustCompile(`^[\x60'"]`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeWhiteSpace,
		pattern:   regexp.MustCompile(`^[ \t]+`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeNewLine,
		pattern:   regexp.MustCompile(`^[\n\r]+`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeSemiColon,
		pattern:   regexp.MustCompile(`^;`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeColon,
		pattern:   regexp.MustCompile(`^:`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeComma,
		pattern:   regexp.MustCompile(`^,`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeOpenExpression,
		pattern:   regexp.MustCompile(`^[\(\[{]`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeCloseExpression,
		pattern:   regexp.MustCompile(`^[\)\]}]`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeDot,
		pattern:   regexp.MustCompile(`^\.`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeAny,
		pattern:   regexp.MustCompile(`^.`),
	},
}

type kissJSTokenPattern struct {
	tokenType int
	pattern   *regexp.Regexp
}

type kissJSToken struct {
	tokenType int
	value     string
}

type kissJSImport struct {
	src                 string
	nobundle, nocompile bool
}

type kissJSLine struct {
	value []kissJSToken
}

type kissJSScript struct {
	imports []kissJSImport
	lines   []kissJSLine
}

func (script kissJSScript) String() string {
	ret := ""
	for _, line := range script.lines {
		for _, token := range line.value {
			ret += token.value
		}
	}
	return ret
}

//TEST
func testJSStuff() error {
	script := `
({KISSimport: "test/this.js", nobundle: true});
({KISSimport: "test/again.js", nobundle: true, nocompile: true});
({KISSimport: "test/this.js", nobundle: false});


console.log("this is a \"TEST\" of my system");
let a = 10
if (true) {
	a += 10;
}



console.log('test \'LOWER\' down');
`
	tokens := tokenizeJSScript(script)
	parsed, err := parseJSTokens(tokens)
	if err != nil {
		return err
	}
	for _, i := range parsed.imports {
		fmt.Printf("%+v\n", i)
	}
	fmt.Println("======")
	fmt.Println(parsed)
	return nil
}

func tokenizeJSScript(script string) []kissJSToken {
	ret := []kissJSToken{}
	i := 0
	for i < len(script) {
		for _, token := range tokenPatterns {
			index := token.pattern.FindIndex([]byte(script[i:]))
			if index != nil {
				ret = append(ret,
					kissJSToken{
						tokenType: token.tokenType,
						value:     script[i : i+index[1]],
					},
				)
				i += index[1]
				break
			}
		}
	}

	return ret
}

func parseJSTokens(script []kissJSToken) (kissJSScript, error) {
	ret := kissJSScript{}
	tokens := []kissJSToken{kissJSToken{tokenType: tokenTypeNewLine}}
	i := 0
	for i < len(script) {
		tok := script[i].tokenType
		if tok == tokenTypeWhiteSpace {
			i++
			continue
		}
		if tok == tokenTypeOpenCloseString {
			add, token, err := parseJSString(script[i:])
			if err != nil {
				return kissJSScript{}, err
			}
			i += add
			tokens = append(tokens, token)
		}
		tokens = append(tokens, script[i])
		i++
	}

	i = 0
	for i < len(tokens) {
		if tokens[i].tokenType == tokenTypeOpenImport {
			add, jsImport, err := parseJSImportStatment(tokens[i:])
			if err != nil {
				return kissJSScript{}, err
			}
			ret.imports = append(ret.imports, jsImport)
			i += add
			continue
		}
		if tokens[i].tokenType == tokenTypeNewLine || i == 0 {
			add, jsLine := parseJSLine(tokens[i:])
			i += add
			if len(jsLine.value) > 0 {
				ret.lines = append(ret.lines, jsLine)
			}
			continue
		}
		i++
	}

	return ret, nil
}

func parseJSLine(script []kissJSToken) (int, kissJSLine) {
	ret := kissJSLine{}
	i := 1
	for i < len(script) {
		tok := script[i].tokenType
		if tok == tokenTypeNewLine {
			break
		}
		if tok == tokenTypeOpenImport {
			break
		}
		ret.value = append(ret.value, script[i])
		i++
	}
	if len(ret.value) == 0 {
		return i, ret
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
	default:
		ret.value = append(ret.value, kissJSToken{tokenTypeSemiColon, ";"})
		return i, ret
	}
}

func parseJSImportStatment(script []kissJSToken) (int, kissJSImport, error) {
	ret := kissJSImport{}
	i := 1
	expectedToken := tokenTypeKissKeyword
	keyword := ""
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
			return i, ret, nil
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
				ret.src = val
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
	return 0, kissJSImport{}, fmt.Errorf("failed to parse kiss import unexpected token in stream '%s'", script[i].value)
}

func parseJSString(script []kissJSToken) (int, kissJSToken, error) {
	ret := kissJSToken{tokenType: tokenTypeValue, value: script[0].value}
	i := 1
	for i < len(script) {
		tok := script[i].tokenType
		ret.value += script[i].value
		i++
		if tok == tokenTypeOpenCloseString {
			return i, ret, nil
		}
	}
	return 0, kissJSToken{}, fmt.Errorf("unexpected token in stream %s", script[i].value)
}
