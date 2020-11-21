package main

import (
	"fmt"

	"golang.org/x/net/html"
)

func newMain(osArgs []string) {
	args, err := parseArgs(osArgs)

	if err != nil {
		fmt.Printf("Parsing Args was unsuccessful: %s\n", err)
		return
	}

	root, err := parseEntryFile(args.entry)
	if err != nil {
		fmt.Printf("Unable to parse entry file %s: %s\n", args.entry, err)
		return
	}

	kissRoot := convertNodeTree(root)
	ctx := kissNodeContext{
		path: getPath(args.entry),
	}
	kissRoot.Parse(ctx)
}

func convertNodeTree(node *html.Node) kissNode {
	ret := htmlNodeToKissNode(node)
	fmt.Println(ret)

	if node.FirstChild != nil {
		ret.SetFirstChild(convertNodeTree(node.FirstChild))
	}

	if node.NextSibling != nil {
		ret.SetNextSibling(convertNodeTree(node.NextSibling))
	}

	return ret
}
