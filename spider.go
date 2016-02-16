package goscrape

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

//ResponseParseHandler is a defintion for the function to be called to parse data
type ResponseParseHandler func(response *Response) ParseResult

//ItemPipelineHandler is a definition for the function to be called to process data
type ItemPipelineHandler func()

// Spider is an object that will perform crawling and parsing of content
type Spider struct {
	Context      ScrapeContext
	StartUrls    []string
	parseHandler ResponseParseHandler
	itemHandler  ItemPipelineHandler
	parseChannel chan *Response
	linksChannel chan string
	stopChannel  chan struct{}
	seenLinks    map[string]bool
	mu           sync.Mutex
	wg           sync.WaitGroup
	wgMu         sync.RWMutex
}

// NewSpider creates and returns an initialized instance of a spider
func NewSpider(startUrls []string, parseHandler ResponseParseHandler, itemHandler ItemPipelineHandler) *Spider {
	spider := Spider{
		StartUrls:    startUrls,
		parseHandler: parseHandler,
		itemHandler:  itemHandler,
		linksChannel: make(chan string),
		parseChannel: make(chan *Response),
		stopChannel:  make(chan struct{}),
		seenLinks:    make(map[string]bool),
		mu:           sync.Mutex{},
		wg:           sync.WaitGroup{},
		wgMu:         sync.RWMutex{},
	}

	return &spider
}

// Start triggers the scraping pipeline process
func (s *Spider) Start() {
	count := 1
	for i := 1; i <= count; i++ {
		go s.crawl()
		go s.parse()
	}

	for _, url := range s.StartUrls {
		log.Println("Starting with " + url)
		s.wg.Add(len(s.StartUrls))
		go func(url string) { s.linksChannel <- url }(url)
	}
	s.wg.Wait()

	for j := 1; j <= count*2; j++ {
		s.stopChannel <- struct{}{}
	}

	// Cleanup
	time.Sleep(1000 * time.Millisecond)
	close(s.linksChannel)
	close(s.parseChannel)
	close(s.stopChannel)
}

// Crawl all pages defined as start pages
func (s *Spider) crawl() {
LOOP:
	for {
		var url string
		select {
		case <-s.stopChannel:
			break LOOP
		case url = <-s.linksChannel:
			// pass
		}

		s.mu.Lock()
		if s.seenLinks[url] {
			s.mu.Unlock()
			s.wg.Done()
			continue
		}
		s.seenLinks[url] = true
		s.mu.Unlock()

		log.Println("Crawl - Got URL: " + url)

		content, err := s.fetchContent(url)

		if err != nil {
			go func() {
				// Sleep for 1 second and then retry the same URL
				time.Sleep(1000 * time.Millisecond)
				s.linksChannel <- url
			}()
			continue
		}

		go func(response *Response) {
			s.parseChannel <- response
		}(content)
	}
}

func (s *Spider) parse() {
LOOP:
	for {
		var response *Response
		select {
		case <-s.stopChannel:
			break LOOP
		case response = <-s.parseChannel:
			//pass
		}

		item := s.parseHandler(response)

		if item.Urls != nil && len(item.Urls) > 0 {
			s.wgMu.Lock()
			s.wg.Add(len(item.Urls))
			s.wgMu.Unlock()

			go func(item ParseResult) {
				for _, url := range item.Urls {
					s.linksChannel <- url
				}
			}(item)
		}

		s.wg.Done()
	}
}

func (s *Spider) fetchContent(url string) (*Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(fmt.Errorf("Error downloading content"))
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading content")
		return nil, err
	}
	defer resp.Body.Close()

	return &Response{Url: url, Content: body}, nil
}
