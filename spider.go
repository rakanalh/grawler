package goscrape

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type ResponseParseHandler func(response Response) ParseResult
type ItemPipelineHandler func()

// Spider is an object that will perform crawling and parsing of content
type Spider struct {
	Context      ScrapeContext
	StartUrls    []string
	parseHandler ResponseParseHandler
	itemHandler  ItemPipelineHandler
	parseChannel chan Response
	linksChannel chan string
	seenLinks    map[string]bool
	mu           sync.Mutex
}

// NewSpider creates and returns an initialized instance of a spider
func NewSpider(startUrls []string, parseHandler ResponseParseHandler, itemHandler ItemPipelineHandler) *Spider {
	spider := Spider{
		StartUrls:    startUrls,
		parseHandler: parseHandler,
		itemHandler:  itemHandler,
		linksChannel: make(chan string),
		parseChannel: make(chan Response),
		seenLinks:    make(map[string]bool),
		mu:           sync.Mutex{},
	}

	return &spider
}

func (s *Spider) Start() {
	for i := 1; i <= 4; i++ {
		go s.crawl()
		go s.parse()
	}

	for _, url := range s.StartUrls {
		log.Println("Starting with " + url)
		s.linksChannel <- url
	}
}

// Crawl all pages defined as start pages
func (s *Spider) crawl() {
	for {
		url := <-s.linksChannel

		s.mu.Lock()
		if s.seenLinks[url] {
			s.mu.Unlock()
			continue
		}
		s.seenLinks[url] = true
		s.mu.Unlock()

		log.Println("Crawl - Got URL: " + url)

		content := s.fetchContent(url)
		go func(response Response) {
			s.parseChannel <- response
		}(content)
	}
}

func (s *Spider) parse() {
	for {
		response := <-s.parseChannel
		item := s.parseHandler(response)

		if item.Urls != nil && len(item.Urls) > 0 {
			go func(item ParseResult) {
				for _, url := range item.Urls {
					s.linksChannel <- url
				}
			}(item)
		}
	}
}

func (s *Spider) fetchContent(url string) Response {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(fmt.Errorf("Error downloading content"))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading content")
	}
	defer resp.Body.Close()

	return Response{Content: body}
}
