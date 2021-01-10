package css

import (
	"fmt"
	"regexp"
	"strings"
)

type tokenType int

const (
	openBlock = tokenType(iota)
	closeBlock
	className
	idName
	elmName
	pseudoClass
	pseudoElm
	property
	whiteSpace
	newLine
	comma
	semiColon
	attrBlock
	openAttrBlock
	closeAttrBlock
	star
	child
	nextChild
	beginWith
	preceded
	equal
	openFn
	closeFn
	function
	keyframe
	percentage
	pixels
	color
	value
	template
	comment
	any
)

type tokenPattern struct {
	tType   tokenType
	pattern *regexp.Regexp
}

var tokenPatterns = []tokenPattern{
	tokenPattern{whiteSpace, regexp.MustCompile(`^[ \t]`)},
	tokenPattern{newLine, regexp.MustCompile(`^[\n\r]+`)},
	tokenPattern{closeFn, regexp.MustCompile(`^\)`)},
	tokenPattern{openFn, regexp.MustCompile(`^(attr|calc|cubic-bezier|hsl|hsla|linear-gradient|radial-gradient|repeating-linear-gradient|repeating-radial-gradient|rgb|rgba|var|url)\(`)},
	tokenPattern{elmName, regexp.MustCompile(`^(html|body|h[1-6]|div|hr|li|ol|p|ul|a|code|em|span|img|svg|canvas|table|tbody|td|tfoot|th|thead|tr|button|form|input|label|select|textarea)[ {>\.\[:~\+]`)},
	tokenPattern{color, regexp.MustCompile(`^#[0-9a-fA-F]{3,8}`)},
	tokenPattern{beginWith, regexp.MustCompile(`^[\^|]`)},
	tokenPattern{closeBlock, regexp.MustCompile(`^}`)},
	tokenPattern{openBlock, regexp.MustCompile(`^{`)},
	tokenPattern{equal, regexp.MustCompile(`^=`)},
	tokenPattern{template, regexp.MustCompile(`^"@[0-9a-zA-Z_-][0-9a-zA-Z_-]*@"`)},
	tokenPattern{closeAttrBlock, regexp.MustCompile(`^\]`)},
	tokenPattern{openAttrBlock, regexp.MustCompile(`^\[`)},
	tokenPattern{child, regexp.MustCompile(`^>`)},
	tokenPattern{nextChild, regexp.MustCompile(`^\+`)},
	tokenPattern{comma, regexp.MustCompile(`^,`)},
	tokenPattern{preceded, regexp.MustCompile(`^~`)},
	tokenPattern{semiColon, regexp.MustCompile(`^;`)},
	tokenPattern{star, regexp.MustCompile(`^\*`)},
	tokenPattern{keyframe, regexp.MustCompile(`^@keyframes [a-zA-Z0-9_-]+`)},
	tokenPattern{percentage, regexp.MustCompile(`^\d{0,3}%`)},
	tokenPattern{pixels, regexp.MustCompile(`^\d+px`)},
	tokenPattern{idName, regexp.MustCompile(`^#[a-zA-Z0-9_-]+`)},
	tokenPattern{className, regexp.MustCompile(`^\.[a-zA-Z0-9_-]+`)},
	tokenPattern{pseudoClass, regexp.MustCompile(`^:[a-zA-Z0-9_-]+`)},
	tokenPattern{pseudoElm, regexp.MustCompile(`^::[a-zA-Z0-9_-]+`)},
	tokenPattern{property, regexp.MustCompile(`^[a-zA-Z0-9_-]+:`)},
	tokenPattern{comment, regexp.MustCompile(`(?ms)^/\*.*\*/`)},
	// Catchall must be the last thing we match
	tokenPattern{any, regexp.MustCompile(`^.`)},
}

// Token is a css token type and value
type Token struct {
	Type  tokenType
	Value string
}

// Style is a css propery and value
type Style struct {
	Prop, Val string
}

// Rule is a css block with a selector and a set of styles
type Rule struct {
	Selectors []Selector
	Styles    []Style
}

// Selector is a css selector
type Selector struct {
	Sel     string
	PostSel string
}

// Anim is a css keyframe animation
type Anim struct {
	Name   string
	Frames []Frame
}

// Frame is a block inside a css keyframe animation
type Frame struct {
	Time   string
	Styles []Style
}

// Script is a parsed css script
type Script struct {
	Rules []Rule
	Anims []Anim
}

// Lex will produce tokens from a string of css rules
func Lex(css string) []Token {
	tokens := []Token{}
	i := 0
	for i < len(css) {
		for _, token := range tokenPatterns {
			index := token.pattern.FindIndex([]byte(css[i:]))
			if index != nil {
				add := Token{
					Type:  token.tType,
					Value: css[i : i+index[1]],
				}

				// if we overmatch the elmName so we need to give 1 char back
				if token.tType == elmName {
					old := len(add.Value)
					add.Value = strings.Trim(add.Value, " {>.[:~+")
					if len(add.Value) < old {
						i--
					}
				}

				tokens = append(tokens, add)
				i += index[1]
				break
			}
		}
	}

	// filter the results and lex attrBlocks
	ret := []Token{}
	i = 0
	for i < len(tokens) {
		if tokens[i].Type == comment || tokens[i].Type == newLine {
			i++
			continue
		}

		if tokens[i].Type == property {
			tokens[i].Value = strings.Trim(tokens[i].Value, ":")
			ret = append(ret, tokens[i])
			i++
			continue
		}

		if tokens[i].Type == openFn {
			count, fn := lexFn(tokens[i:])
			ret = append(ret, fn)
			i += count
			continue
		}

		if tokens[i].Type == openAttrBlock {
			count, block := lexAttrBlock(tokens[i:])
			ret = append(ret, block)
			i += count
			continue
		}

		ret = append(ret, tokens[i])
		i++
	}

	// lex property values
	tokens = ret
	ret = []Token{}
	i = 0
	for i < len(tokens) {
		if tokens[i].Type == whiteSpace {
			if i < len(tokens) && tokens[i+1].Type == openBlock {
				i++
				continue
			}
			// just set this to be not a selector type
			prev := any
			if i > 0 {
				prev = tokens[i-1].Type
			}
			if prev != elmName &&
				prev != className &&
				prev != idName &&
				prev != pseudoClass &&
				prev != pseudoElm &&
				prev != closeAttrBlock {
				i++
				continue
			}
		}

		if tokens[i].Type == property {
			ret = append(ret, tokens[i])
			i++
			count, val := lexValue(tokens[i:])
			ret = append(ret, val)
			i += count
			continue
		}

		ret = append(ret, tokens[i])
		i++
	}

	return ret
}

func lexFn(css []Token) (int, Token) {
	ret := Token{Type: function, Value: css[0].Value}
	i := 1
	for i < len(css) {
		tok := css[i].Type
		ret.Value += css[i].Value
		i++
		if tok == closeFn {
			return i, ret
		}
	}
	return 0, Token{}
}

func lexValue(css []Token) (int, Token) {
	ret := Token{Type: value}
	leading := true
	i := 0
	for i < len(css) {
		tok := css[i].Type
		if tok == whiteSpace && leading {
			i++
			continue
		}
		leading = false

		if tok == semiColon {
			return i + 1, ret
		}
		ret.Value += css[i].Value
		i++
	}
	return 0, Token{}
}

func lexAttrBlock(css []Token) (int, Token) {
	ret := Token{Type: attrBlock, Value: css[0].Value}
	i := 1
	for i < len(css) {
		tok := css[i].Type
		ret.Value += css[i].Value
		i++
		if tok == closeAttrBlock {
			return i, ret
		}
	}
	return 0, Token{}
}

// Parse will parse a serise of tokens generated by the lexer into a Script object
func Parse(css []Token) (Script, error) {
	ret := Script{}
	i := 0
	start := 0
	for i < len(css) {
		start = i

		count, rule := parseRule(css[i:])
		if count > 0 {
			i += count
			ret.Rules = append(ret.Rules, rule)
			continue
		}

		count, anim := parseAnim(css[i:])
		if count > 0 {
			i += count
			ret.Anims = append(ret.Anims, anim)
			continue
		}
		if i == start {
			return Script{}, fmt.Errorf("failed to parse css token, '%s'", css[i].Value)
		}
	}

	return ret, nil
}

func parseRule(css []Token) (int, Rule) {
	ret := Rule{}

	i, count := 0, 0
	count, ret.Selectors = parseSelector(css)
	i += count
	_, ret.Styles = parseBlock(css[i:])

	return 0, ret
}

func parseSelector(css []Token) (int, []Selector) {
	ret := []Selector{}
	selectorTokens := []tokenType{star, elmName, className, idName, attrBlock}
	postSelectorTokens := []tokenType{pseudoClass, pseudoElm}
	connectTokens := []tokenType{child, nextChild, preceded}

	i := 0
	add := Selector{}
	for i < len(css) {
		tok := css[i].Type

		found := false
		for _, check := range selectorTokens {
			if tok == check {
				found = true
				break
			}
		}
		if found {
			add.Sel += css[i].Value
			i++
			continue
		}

		found = false
		for _, check := range connectTokens {
			if tok == check {
				found = true
				break
			}
		}
		if found {
			add.Sel += css[i].Value
			i++
			continue
		}

		found = false
		for _, check := range postSelectorTokens {
			if tok == check {
				found = true
				break
			}
		}
		if found {
			add.PostSel += css[i].Value
			i++
			continue
		}

		if tok == whiteSpace {
			ret = append(ret, add)
			add = Selector{}
			i++
			continue
		}

		if tok == openBlock {
			ret = append(ret, add)
			return i, ret
		}

		i++
	}
	return 0, []Selector{}
}

func parseBlock(css []Token) (int, []Style) {
	ret := []Style{}

	// Expect the first token to be '{'
	i := 1
	add := Style{}
	for i < len(css) {
		tok := css[i].Type

		if tok == property {
			add.Prop = css[i].Value
			i++
			continue
		}

		if tok == value {
			add.Val = css[i].Value
			ret = append(ret, add)
			add = Style{}
			i++
			continue
		}

		if tok == closeBlock {
			return i, ret
		}

		i++
	}
	return 0, []Style{}
}

func parseAnim(css []Token) (int, Anim) {
	return 0, Anim{}
}
