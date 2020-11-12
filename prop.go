package main

import (
	"golang.org/x/net/html"
)

type prop struct {
	key string
	val []*html.Node
}

func (prop *prop) isSimple() bool {
	if len(prop.val) != 1 {
		return false
	}
	if prop.val[0].Type != html.TextNode {
		return false
	}
	return true
}
