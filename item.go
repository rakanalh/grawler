package goscrape

// ParseItem defines the single item of scraped data
type ParseItem map[string]string

// ParseResult is a container to all user-defined items
type ParseResult struct {
	Items []ParseItem
	Urls  []string
}
