package goscrape

import (
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/xml"
)

// XPathSearcher Defines an interface that implements searching through XPath
type XPathSearcher interface {
	XPath(content string) ([]string, error)
}

type CssSearcher interface {
	Css(content string) ([]string, error)
}

type Response struct {
	Url     string
	Content []byte
}

func (r *Response) XPath(xpath string) ([]string, error) {
	doc, _ := gokogiri.ParseHtml(r.Content)
	defer doc.Free()

	results, err := doc.Search(xpath)
	if err != nil {
		return nil, err
	}
	return r.normalizeXmlNodes(results), nil
}

func (r *Response) Css(content string) ([]string, error) {
	return []string{}, nil
}

func (r *Response) normalizeXmlNodes(nodes []xml.Node) []string {
	result := []string{}
	for _, node := range nodes {
		result = append(result, node.String())
	}
	return result
}
