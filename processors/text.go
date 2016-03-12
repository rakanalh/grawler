package processors

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/moovweb/gokogiri/xml"
)

func GetLinks(nodes interface{}) ([]string, error) {
	var links []string
	if selection, ok := nodes.(goquery.Selection); ok {
		selection.Each(func(i int, child *goquery.Selection) {
			if link, exists := child.Attr("href"); exists {
				links = append(links, link)
			}
		})
	} else if xmlNodes, ok := nodes.([]xml.Node); ok {
		links = getXmlNodesAsString(xmlNodes)
	}
	return links, nil
}

func GetText(nodes interface{}) []string {
	var textNodes []string
	if selection, ok := nodes.(goquery.Selection); ok {
		selection.Each(func(i int, child *goquery.Selection) {
			textNodes = append(textNodes, child.Text())
		})
	} else if xmlNodes, ok := nodes.([]xml.Node); ok {
		textNodes = getXmlNodesAsString(xmlNodes)
	}
	return textNodes
}

func getXmlNodesAsString(xmlNodes []xml.Node) []string {
	var textNodes []string
	for _, node := range xmlNodes {
		textNodes = append(textNodes, node.String())
	}
	return textNodes
}
