package module

import (
	"bytes"
	"golang.org/x/net/html"
	"strings"
)

func attr(node *html.Node, name string, defaultValue *string) *string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return &attr.Val
		}
	}
	return defaultValue
}

func replaceContentWithString(parent *html.Node, data *string) {
	node := new(html.Node)
	node.Type = html.TextNode
	node.Data = *data
	replaceContent(parent, node)
}

func appendContent(parent *html.Node, child *html.Node) error {
	node, err := clone(child)

	if err != nil {
		return err
	}

	parent.AppendChild(node)
	return nil
}

func replaceContent(parent *html.Node, child *html.Node) error {
	childToRemove := parent.FirstChild

	for childToRemove != nil {
		parent.RemoveChild(childToRemove)
		childToRemove = parent.FirstChild
	}

	return appendContent(parent, child)
}

func renderToString(node *html.Node) (*string, error) {
	var b bytes.Buffer
	err := html.Render(&b, node)

	if err != nil {
		return nil, err
	}
	result := b.String()

	return &result, nil
}

func parseString(input *string) (*html.Node, error) {
	reader := strings.NewReader(*input)
	return html.Parse(reader)
}

func clone(node *html.Node) (*html.Node, error) {
	str, err := renderToString(node)

	if err != nil {
		return nil, err
	}

	println(*str)

	return parseString(str)
}
