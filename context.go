package goscrape

import "net/http"

// Define selector types
const (
	XPathSelector = iota
	CSSSelector
)

type SelectorType int

// ScrapeContext is an object that holds property of
// the scraping process
type ScrapeContext struct {
	Request      http.Request
	SelectorType SelectorType
}
