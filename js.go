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

// func jsFromImport(file string, depth int) (jsSnipit, error) {
// 	ret := jsSnipit{}
// 	script, err := ioutil.ReadFile(file)
// 	if err != nil {
// 		return ret, err
// 	}
// 	ret.js = string(script)
// 	ret.src = file
// 	ret.compile = true

// 	err = ret.collectImports(0, getPath(file))
// 	return ret, err
// }

// func jsFromNode(node *html.Node, depth int, path string) (jsSnipit, error) {
// 	ret := jsSnipit{}
// 	if node.DataAtom != atom.Script || node.Type != html.ElementNode {
// 		return ret, errors.New("only script nodes can be converted to a jsSnipit object")
// 	}

// 	for _, attr := range node.Attr {
// 		if attr.Key == "src" {
// 			ret.src = path + attr.Val
// 		}
// 		if attr.Key == "compile" && attr.Val == "true" {
// 			ret.compile = true
// 		}
// 	}

// 	if ret.src != "" && ret.compile {
// 		err := ret.readSrc()
// 		if err != nil {
// 			return ret, err
// 		}
// 	}

// 	if node.FirstChild != nil && ret.js == "" {
// 		ret.js = node.FirstChild.Data
// 	}

// 	err := ret.collectImports(0, getPath(ret.src))
// 	return ret, err
// }

// func (js *jsSnipit) collectImports(depth int, path string) error {
// 	importRX := regexp.MustCompile(`{ *KISSimport: *['"]([^'"}\n]*)['"] *}`)
// 	importFiles := importRX.FindAllStringSubmatch(js.js, -1)

// 	for _, file := range importFiles {
// 		alreadyImported := false
// 		for _, roi := range js.imports {
// 			if path+file[1] == roi.src {
// 				alreadyImported = true
// 			}
// 		}
// 		if alreadyImported {
// 			continue
// 		}

// 		i, err := jsFromImport(path+file[1], depth)
// 		if err != nil {
// 			return err
// 		}
// 		js.imports = append(js.imports, &i)
// 		i.depth = depth
// 		depth++
// 		js.imports = append(js.imports, i.imports...)
// 	}

// 	return nil
// }

// func (js *jsSnipit) sortImports() {
// 	max := 0
// 	for _, i := range js.imports {
// 		if max < i.depth {
// 			max = i.depth
// 		}
// 	}

// 	sortedImports := []*jsSnipit{}
// 	for i := max; i >= 0; i-- {
// 		for _, js := range js.imports {
// 			if js.depth == i {
// 				sortedImports = append(sortedImports, js)
// 			}
// 		}
// 	}

// 	js.imports = sortedImports
// }

// func (js *jsSnipit) readSrc() error {
// 	script, err := ioutil.ReadFile(js.src)
// 	if err != nil {
// 		return err
// 	}
// 	js.js = string(script)
// 	return nil
// }

// func (js *jsSnipit) hydrate(props []prop) bool {
// 	changed := false
// 	for _, prop := range props {
// 		if !prop.isSimple() {
// 			continue
// 		}
// 		old := js.js
// 		js.js = strings.ReplaceAll(js.js, "$"+prop.key+"$", prop.val[0].Data)
// 		changed = changed || (old == js.js)
// 	}
// 	return changed
// }
