package goscrape

import (
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/html"
	"github.com/moovweb/gokogiri/xml"
)

// XPathSearcher Defines an interface that implements searching through XPath
type XPathSearcher interface {
	XPath(content string) ([]string, error)
}

// Response defines the attributes of parse response
type Response struct {
	URL     string
	Content []byte
	doc     *html.HtmlDocument
}

// NewResponse creates a Response instance
func NewResponse(url string, content []byte) (*Response, error) {
	doc, err := gokogiri.ParseHtml(content)

	if err != nil {
		return nil, err
	}

	response := Response{
		URL:     url,
		Content: content,
		doc:     doc,
	}

	return &response, nil
}

// XPath searches for all nodes that match provided xpath
func (r *Response) XPath(xpath string) ([]string, error) {
	if r.doc == nil {

	}
	results, err := r.doc.Search(xpath)
	if err != nil {
		return nil, err
	}
	return r.normalizeXMLNodes(results), nil
}

// Close the document once we're done using it
func (r *Response) Close() {
	if r.doc != nil {
		r.doc.Free()
	}
}

// normalizeXMLNodes is used to convert the xml.Node array to a string one
func (r *Response) normalizeXMLNodes(nodes []xml.Node) []string {
	result := []string{}
	for _, node := range nodes {
		result = append(result, node.String())
	}
	return result
}
