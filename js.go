package main

import (
	"io/ioutil"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Used
type jsSnipit struct {
	imports []*jsSnipit
	js      string
	src     string
	depth   int
	compile bool // When this is set the script is templated and searched for import statments
	bundle  bool // When this is set the script is bundled into the bundle.js file (each snipit is added only once reguarless of how often it appers)
}

// Used
func extractScriptsNEW(root *html.Node, path string) ([]*jsSnipit, error) {
	ret := []*jsSnipit{}
	for _, node := range listNodes(root) {
		if node.Data == "script" {
			ok, src := getAttr(node, "src")
			add := jsSnipit{
				compile: true,
				bundle:  true,
				depth:   1,
			}

			if ok {
				script, err := ioutil.ReadFile(path + src.Val)
				if err != nil {
					return ret, err
				}
				add.js = string(script)
				add.src = src.Val
			} else {
				add.js = node.FirstChild.Data
			}

			ret = append(ret, &add)
			snipits, err := compileJS(&add, getPath(path+src.Val))
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

// Used
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

// Used
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

// Used
func compileJS(snipit *jsSnipit, path string) ([]*jsSnipit, error) {
	importRX := regexp.MustCompile(`{ *KISSimport: *["']([^"']*)["'] *[^}]*}`)
	importFiles := importRX.FindAllStringSubmatch(snipit.js, -1)
	for _, KissImport := range importFiles {
		snipit.js = strings.ReplaceAll(snipit.js, KissImport[0], "")
	}

	ret := []*jsSnipit{}
	for _, KISSimport := range importFiles {
		file := path + KISSimport[1]
		script, err := ioutil.ReadFile(file)
		if err != nil {
			return ret, err
		}

		add := jsSnipit{
			js:      string(script),
			compile: true,
			bundle:  true,
			depth:   snipit.depth + 1,
			src:     file,
		}
		ret = append(ret, &add)
		snipits, err := compileJS(&add, getPath(file))
		if err != nil {
			return ret, err
		}

		ret = append(ret, snipits...)
	}
	return ret, nil
}
