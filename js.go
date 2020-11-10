package main

import (
	"errors"
	"io/ioutil"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type jsSnipit struct {
	imports []*jsSnipit
	js      string
	src     string
	compile bool
	depth   int
}

func jsFromImport(file string, depth int) (jsSnipit, error) {
	ret := jsSnipit{}
	script, err := ioutil.ReadFile(file)
	if err != nil {
		return ret, err
	}
	ret.js = string(script)
	ret.src = file
	ret.compile = true

	err = ret.collectImports(0, getPath(file))
	return ret, err
}

func jsFromNode(node *html.Node, depth int, path string) (jsSnipit, error) {
	ret := jsSnipit{}
	if node.DataAtom != atom.Script || node.Type != html.ElementNode {
		return ret, errors.New("only script nodes can be converted to a jsSnipit object")
	}

	for _, attr := range node.Attr {
		if attr.Key == "src" {
			ret.src = path + attr.Val
		}
		if attr.Key == "compile" && attr.Val == "true" {
			ret.compile = true
		}
	}

	if ret.src != "" && ret.compile {
		err := ret.readSrc()
		if err != nil {
			return ret, err
		}
	}

	if node.FirstChild != nil && ret.js == "" {
		ret.js = node.FirstChild.Data
	}

	err := ret.collectImports(0, getPath(ret.src))
	return ret, err
}

func (js *jsSnipit) collectImports(depth int, path string) error {
	importRX := regexp.MustCompile(`{ *KISSimport: *['"]([^'"}\n]*)['"] *}`)
	importFiles := importRX.FindAllStringSubmatch(js.js, -1)

	for _, file := range importFiles {
		alreadyImported := false
		for _, roi := range js.imports {
			if path+file[1] == roi.src {
				alreadyImported = true
			}
		}
		if alreadyImported {
			continue
		}

		i, err := jsFromImport(path+file[1], depth)
		if err != nil {
			return err
		}
		js.imports = append(js.imports, &i)
		i.depth = depth
		depth++
		js.imports = append(js.imports, i.imports...)
	}

	return nil
}

func (js *jsSnipit) sortImports() {
	max := 0
	for _, i := range js.imports {
		if max < i.depth {
			max = i.depth
		}
	}

	sortedImports := []*jsSnipit{}
	for i := max; i >= 0; i-- {
		for _, js := range js.imports {
			if js.depth == i {
				sortedImports = append(sortedImports, js)
			}
		}
	}

	js.imports = sortedImports
}

func (js *jsSnipit) readSrc() error {
	script, err := ioutil.ReadFile(js.src)
	if err != nil {
		return err
	}
	js.js = string(script)
	return nil
}

func (js *jsSnipit) hydrate(props []prop) {
	for _, prop := range props {
		if !prop.isSimple() {
			continue
		}
		js.js = strings.ReplaceAll(js.js, "$"+prop.key+"$", prop.val[0].Data)
	}
}
