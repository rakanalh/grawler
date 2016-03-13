# goscrape
A Go scraper inspired by Python's Scrapy (http://scrapy.org).

This repository is still a work in progress

Roadmap:

1. Better logging
2. Exporting scraped content to JSON, XML or other formats
3. Web Interface for viewing jobs / creating dynamic scrapers
4. Sitemap spider
5. **Tests**

# Example Code

```
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/rakanalh/goscrape"
	"github.com/rakanalh/goscrape/extract"
	"github.com/rakanalh/goscrape/processors"
)

func parseHandler(response *goscrape.Response) goscrape.ParseResult {
	log.Println(response.URL)
	if !strings.Contains(response.URL, "category") {
		return goscrape.ParseResult{}
	}

	xpath, err := extract.NewXPath(response.Content)
	nodes, err := xpath.Extract("//ul[@class=\"list-thumb-right\"]/li/a[position()=1]/@href")
	urls, err := processors.GetLinks(nodes)

	// css, err := extract.NewCss(response.Content)
	// selection, err := css.Extract("ul.list-thumb-right > li > a")
	// urls, err := processors.GetLinks(selection)
	fmt.Println(urls)

	if err != nil {
		panic("Could not parse Xpath")
	}

	log.Println("Processing URL: " + response.URL)

	for i, url := range urls {
		urls[i] = "http://www.testdomain.net" + url
	}

	result := goscrape.ParseResult{
		Urls: urls,
	}
	return result
}

func itemHandler(item goscrape.ParseItem) {
	fmt.Println(item)
}

func main() {
	startUrls := []string{
		"http://www.testdomain.net/category/test_uri",
	}

	configs := goscrape.Configs{
		StartUrls:    startUrls,
		WorkersCount: 3,
	}

	spider := goscrape.NewSpider(&configs, parseHandler, itemHandler)
	spider.Start()
}
```
