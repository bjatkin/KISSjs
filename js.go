package main

import (
	"io/ioutil"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type jsSnipit struct {
	imports   []*jsSnipit
	js        string
	src       string
	depth     int
	noCompile bool // When this is set the script is templated and searched for import statments
	noBundle  bool // When this is set the script is bundled into the bundle.js file (each snipit is added only once reguarless of how often it appers)
}

func extractScripts(root *html.Node, path string) ([]*jsSnipit, error) {
	ret := []*jsSnipit{}
	for _, node := range listNodes(root) {
		if node.Data == "script" {
			srcOK, src := getAttr(node, "src")
			noCompileOK, noCompile := getAttr(node, "nocompile")
			noBundleOK, noBundle := getAttr(node, "nobundle")
			add := jsSnipit{
				noCompile: noCompileOK || (noCompile != nil && noCompile.Val != "false"),
				noBundle:  noBundleOK || (noBundle != nil && noBundle.Val != "false"),
				depth:     1,
			}

			if srcOK {
				add.src = path + src.Val
			}

			if srcOK && !add.noCompile {
				script, err := ioutil.ReadFile(add.src)
				if err != nil {
					return ret, err
				}
				add.js = string(script)
			}

			if node.FirstChild != nil {
				add.js = node.FirstChild.Data
			}

			ret = append(ret, &add)
			snipits, err := compileJS(&add, getPath(add.src))
			if err != nil {
				return ret, err
			}

			ret = append(ret, snipits...)

			node.Parent.RemoveChild(node)
		}
	}

	ret = removeDuplicates(ret)
	ret = sortSnipits(ret)

	return ret, nil
}

// TODO: rewirte this to look at both the source code and the js source
func removeDuplicates(snipits []*jsSnipit) []*jsSnipit {
	ret := []*jsSnipit{}
	keys := make(map[string]int)
	for _, snipit := range snipits {
		value, ok := keys[snipit.src]
		if !ok || snipit.src == "" {
			keys[snipit.src] = len(ret)
			ret = append(ret, snipit)
		} else {
			if ret[value].depth < snipit.depth {
				ret[value].depth = snipit.depth
			}
		}
	}

	return ret
}

func sortSnipits(snipits []*jsSnipit) []*jsSnipit {
	ret := []*jsSnipit{}
	max := 0
	for _, snipit := range snipits {
		if snipit.depth > max {
			max = snipit.depth
		}
	}

	for i := max; i >= 0; i-- {
		for _, snipit := range snipits {
			if snipit.depth == i {
				ret = append(ret, snipit)
			}
		}
	}

	return ret
}

func compileJS(snipit *jsSnipit, path string) ([]*jsSnipit, error) {
	tokens := tokenizeJS(snipit.js)
	snipits, indexes := parseImports(tokens, []*jsSnipit{}, [][]int{})
	ret := []*jsSnipit{}

	for _, snipit := range snipits {
		ret = append(ret, snipit)
	}

	imports := []string{}
	for _, index := range indexes {
		imports = append(imports, snipit.js[index[0]:index[1]+2])
	}

	for _, i := range imports {
		snipit.js = strings.ReplaceAll(snipit.js, i, "")
	}

	for _, snipit := range snipits {
		if !snipit.noCompile {
			js, err := compileJS(snipit, getPath(path+snipit.src))
			if err != nil {
				return nil, err
			}
			ret = append(ret, js...)
		}
	}

	return ret, nil
}

type jsToken struct {
	tokenType int
	pattern   *regexp.Regexp
	value     string
	index     int
}

const (
	tokenTypeOpen = iota
	tokenTypeClose
	tokenTypeKeyword
	tokenTypeValue
	tokenTypeWhiteSpace
	tokenTypeColon
	tokenTypeComma
	tokenTypeAny
)

func tokenList() []jsToken {
	return []jsToken{
		jsToken{
			tokenType: tokenTypeOpen,
			pattern:   regexp.MustCompile(`^\({`),
		},
		jsToken{
			tokenType: tokenTypeClose,
			pattern:   regexp.MustCompile(`^}\)`),
		},
		jsToken{
			tokenType: tokenTypeKeyword,
			pattern:   regexp.MustCompile(`^KISSimport`),
		},
		jsToken{
			tokenType: tokenTypeKeyword,
			pattern:   regexp.MustCompile(`^nocompile`),
		},
		jsToken{
			tokenType: tokenTypeKeyword,
			pattern:   regexp.MustCompile(`^nobundle`),
		},
		jsToken{
			tokenType: tokenTypeValue,
			pattern:   regexp.MustCompile(`^(?:true|false)`),
		},
		jsToken{
			tokenType: tokenTypeValue,
			pattern:   regexp.MustCompile(`^['"][^'"]*['"]`),
		},
		jsToken{
			tokenType: tokenTypeWhiteSpace,
			pattern:   regexp.MustCompile(`^[ \t\r\n]+`),
		},
		jsToken{
			tokenType: tokenTypeColon,
			pattern:   regexp.MustCompile(`^:`),
		},
		jsToken{
			tokenType: tokenTypeComma,
			pattern:   regexp.MustCompile(`^,`),
		},
		jsToken{
			tokenType: tokenTypeAny,
			pattern:   regexp.MustCompile(`^.`),
		},
	}
}
func tokenizeJS(js string) []jsToken {
	ret := []jsToken{}
	tokens := tokenList()
	i := 0
	for i < len(js) {
		for _, token := range tokens {
			index := token.pattern.FindIndex([]byte(js[i:]))
			if index != nil {
				ret = append(ret,
					jsToken{
						tokenType: token.tokenType,
						value:     js[i : i+index[1]],
						index:     i,
					},
				)
				i += index[1]
				break
			}
		}
	}

	return ret
}

func parseImports(tokens []jsToken, snipits []*jsSnipit, indexes [][]int) ([]*jsSnipit, [][]int) {
	noSpace := []jsToken{}
	for _, token := range tokens {
		if token.tokenType != tokenTypeWhiteSpace {
			noSpace = append(noSpace, token)
		}
	}
	tokens = noSpace

	i := 0
	start := -1
	for i < len(tokens) {
		if tokens[i].tokenType == tokenTypeOpen {
			start = i
			break
		}
		i++
	}

	end := -1
	for i < len(tokens) {
		if tokens[i].tokenType == tokenTypeClose {
			end = i
			break
		}
		i++
	}

	if start > -1 && end > -1 {
		add := jsSnipit{}
		index := []int{tokens[start].index, tokens[end].index}
		i = start
		for i < end {
			if i+3 < len(tokens) &&
				tokens[i].tokenType == tokenTypeKeyword &&
				tokens[i+1].tokenType == tokenTypeColon &&
				tokens[i+2].tokenType == tokenTypeValue {
				if tokens[i].value == "KISSimport" {
					add.src = strings.Trim(tokens[i+2].value, "\"'")
				}
				if tokens[i].value == "nocompile" {
					add.noCompile = tokens[i+2].value == "true"
				}
				if tokens[i].value == "nobundle" {
					add.noBundle = tokens[i+2].value == "true"
				}
				if i+4 > len(tokens) ||
					tokens[i+3].tokenType != tokenTypeComma {
					break
				}
			}
			i++
		}
		snipits, indexes = parseImports(tokens[end:], append(snipits, &add), append(indexes, index))
	}

	return snipits, indexes
}
