package main

import "golang.org/x/net/html"

type componentNode struct {
	node     *html.Node
	children []*html.Node
}
