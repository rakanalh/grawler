// Package processors provides a set of utility methods for processing
// CSS selectors / XPath nodes and converting them into texts
package processors

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/moovweb/gokogiri/xml"
)

// GetLinks is a helper function to extract links out of an
// xpath or css selector result
func GetLinks(nodes interface{}) ([]string, error) {
	var links []string
	if selection, ok := nodes.(goquery.Selection); ok {
		selection.Each(func(i int, child *goquery.Selection) {
			if link, exists := child.Attr("href"); exists {
				links = append(links, link)
			}
		})
	} else if xmlNodes, ok := nodes.([]xml.Node); ok {
		links = getXMLNodesAsString(xmlNodes)
	}
	return links, nil
}

// GetText returns the text of the provided nodes
func GetText(nodes interface{}) []string {
	var textNodes []string
	if selection, ok := nodes.(goquery.Selection); ok {
		selection.Each(func(i int, child *goquery.Selection) {
			textNodes = append(textNodes, child.Text())
		})
	} else if xmlNodes, ok := nodes.([]xml.Node); ok {
		textNodes = getXMLNodesAsString(xmlNodes)
	}
	return textNodes
}

func getXMLNodesAsString(xmlNodes []xml.Node) []string {
	var textNodes []string
	for _, node := range xmlNodes {
		textNodes = append(textNodes, node.String())
	}
	return textNodes
}
