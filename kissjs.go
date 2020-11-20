package main

import "regexp"

const (
	tokenTypeOpenImport = iota
	tokenTypeCloseImport
	tokenTypeKeyword
	tokenTypeStringStartEnd
	tokenTypeEscapedStringStartEnd
	tokenTypeBoolValue
	tokenTypeNumberValue
	tokenTypeWhiteSpace
	tokenTypeColon
	tokenTypeComma
	tokenTypeSemiColon
	tokenTypeAny
	tokenTypeCloseExpression
	tokenTypeOpenExpression
	tokenTypeDot
)

var tokenPatternList = []kissJSTokenPattern{
	kissJSTokenPattern{
		tokenType: tokenTypeOpenImport,
		pattern:   regexp.MustCompile(`^\({`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeCloseImport,
		pattern:   regexp.MustCompile(`^})`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^KISSimport`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^nocompile`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeKeyword,
		pattern:   regexp.MustCompile(`^nobundle`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeBoolValue,
		pattern:   regexp.MustCompile(`^(?:true|false)`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeEscapedStringStartEnd,
		pattern:   regexp.MustCompile(`^\['"]`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeStringStartEnd,
		pattern:   regexp.MustCompile(`^['"]`),
	},
	kissJSTokenPattern{
		tokenType: tokenTypeWhiteSpace,
		pattern:   regexp.MustCompile(`^[ \t\r\n]+`),
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
		tokenType: tokenTypeAny,
		pattern:   regexp.MustCompile(`^.`),
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
	value      string
	terminated bool
}
