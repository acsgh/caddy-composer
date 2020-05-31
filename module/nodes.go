package module

import (
	"bytes"
	"github.com/andybalholm/cascadia"
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

func appendContent(parent *html.Node, child *html.Node) {
	child.Parent = nil
	child.PrevSibling = nil
	child.NextSibling = nil

	parent.AppendChild(child)
}

func replaceContent(parent *html.Node, child *html.Node) {
	childToRemove := parent.FirstChild

	for childToRemove != nil {
		parent.RemoveChild(childToRemove)
		childToRemove = parent.FirstChild
	}

	parent.Attr = child.Attr
	parent.Data = child.Data
	parent.DataAtom = child.DataAtom

	childTmp := child.FirstChild

	for childTmp != nil {
		toAppend := childTmp
		childTmp = childTmp.NextSibling
		appendContent(parent, toAppend)
	}
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

func attachIfRequired(parent *html.Node, nodeName string, attrName string, nodes []*html.Node) {
	if parent != nil {
		for _, node := range nodes {
			src := attr(node, attrName, nil)

			if src != nil {
				if !containsNode(parent, nodeName, attrName, src) {
					appendContent(parent, node)
				}
			} else {
				appendContent(parent, node)
			}
		}
	}
}

func containsNode(parent *html.Node, nodeName string, attrName string, src *string) bool {
	nodes := cascadia.MustCompile(nodeName).MatchAll(parent)
	for _, node := range nodes {
		nodeSrc := attr(node, attrName, nil)

		if src != nil && nodeSrc != nil {
			if *src == *nodeSrc {
				return true
			}
		}
	}
	return false
}
