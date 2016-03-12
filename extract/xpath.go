package extract

import (
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/html"
	"github.com/moovweb/gokogiri/xml"
)

// XPath is a structure that
type XPath struct {
	doc *html.HtmlDocument
}

// NewXPath is a constructor method
func NewXPath(content []byte) (*XPath, error) {
	doc, err := gokogiri.ParseHtml(content)

	if err != nil {
		return nil, err
	}

	return &XPath{doc: doc}, nil
}

// Extract is asda
func (extractor *XPath) Extract(expr string) ([]xml.Node, error) {
	nodes, err := extractor.doc.Search(expr)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// Close the document once we're done using it
func (extractor *XPath) Close() {
	if extractor.doc != nil {
		extractor.doc.Free()
	}
}
