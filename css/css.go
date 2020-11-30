package css

import "regexp"

type tokenType int

const (
	colonToken = tokenType(iota)
	doubleColonToken
	dotToken
	hashToken
	starToken
	semiColonToken
	tildaToken
	commaToken
	greaterThanToken
	openAttrToken
	closeAttrToken
	templateToken
	pipeToken
	equalToken
	openBlockToken
	closeBlockToken
	hatToken
	openFnToken
	closeFnToken
	spaceToken
	anyToken
)

type tokenPattern struct {
	ttype   tokenType
	pattern *regexp.Regexp
}

var tokenPatterns = []tokenPattern{
	tokenPattern{spaceToken, regexp.MustCompile(`^ `)},
	tokenPattern{closeFnToken, regexp.MustCompile(`^)`)},
	tokenPattern{openFnToken, regexp.MustCompile(`^(`)},
	tokenPattern{hatToken, regexp.MustCompile(`^\^`)},
	tokenPattern{closeBlockToken, regexp.MustCompile(`^}`)},
	tokenPattern{openBlockToken, regexp.MustCompile(`^{`)},
	tokenPattern{equalToken, regexp.MustCompile(`^=`)},
	tokenPattern{pipeToken, regexp.MustCompile(`^|`)},
	tokenPattern{templateToken, regexp.MustCompile(`^"@[0-9a-zA-Z_-][0-9a-zA-Z_-]*@"`)},
	tokenPattern{closeAttrToken, regexp.MustCompile(`^\]`)},
	tokenPattern{openAttrToken, regexp.MustCompile(`^\[`)},
	tokenPattern{greaterThanToken, regexp.MustCompile(`^>`)},
	tokenPattern{commaToken, regexp.MustCompile(`^,`)},
	tokenPattern{tildaToken, regexp.MustCompile(`^~`)},
	tokenPattern{semiColonToken, regexp.MustCompile(`^;`)},
	tokenPattern{starToken, regexp.MustCompile(`^\*`)},
	tokenPattern{hashToken, regexp.MustCompile(`^#`)},
	tokenPattern{dotToken, regexp.MustCompile(`^\.`)},
	tokenPattern{doubleColonToken, regexp.MustCompile(`^::`)},
	tokenPattern{colonToken, regexp.MustCompile(`^:`)},
	tokenPattern{anyToken, regexp.MustCompile(`^.`)},
}

type Token struct {
	Type  tokenType
	Value string
}

func tokenize(css string) []Token {
	return nil
}
