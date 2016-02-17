package goscrape

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

//ResponseParseHandler is a defintion for the function to be called to parse data
type ResponseParseHandler func(response *Response) ParseResult

//ItemPipelineHandler is a definition for the function to be called to process data
type ItemPipelineHandler func(parsedItem ParseItem)

//URLRecord is a tool to help keep track of URL specific properties
type URLRecord struct {
	Retries uint8
	Wait    time.Duration
}

// Spider is an object that will perform crawling and parsing of content
type Spider struct {
	// User specific configs
	StartUrls    []string
	parseHandler ResponseParseHandler
	itemHandler  ItemPipelineHandler

	// Channels
	parseChannel chan *Response
	linksChannel chan string
	stopChannel  chan struct{}

	// URL tracking
	seenLinks map[string]*URLRecord
	linksMu   sync.Mutex

	// Workers
	workersWg sync.WaitGroup
	workersMu sync.RWMutex
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
		seenLinks:    make(map[string]*URLRecord),
		linksMu:      sync.Mutex{},
		workersWg:    sync.WaitGroup{},
		workersMu:    sync.RWMutex{},
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
		s.workersWg.Add(len(s.StartUrls))
		go func(url string) { s.linksChannel <- url }(url)
	}
	s.workersWg.Wait()

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

		s.linksMu.Lock()
		if _, ok := s.seenLinks[url]; ok {
			s.linksMu.Unlock()
			s.workersWg.Done()
			continue
		}

		s.seenLinks[url] = &URLRecord{
			Retries: 1,
			Wait:    1000 * time.Millisecond,
		}

		s.linksMu.Unlock()

		log.Println("Crawl - Got URL: " + url)

		content, err := s.fetchContent(url)

		if err != nil {
			s.seenLinks[url].Retries++
			go func() {
				// Sleep for 1 second and then retry the same URL
				time.Sleep(s.seenLinks[url].Wait)
				s.linksChannel <- url

				s.seenLinks[url].Wait *= 2
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
			s.workersMu.Lock()
			s.workersWg.Add(len(item.Urls))
			s.workersMu.Unlock()

			go func(item ParseResult) {
				for _, url := range item.Urls {
					s.linksChannel <- url
				}
			}(item)
		}

		if s.itemHandler != nil && item.Items != nil && len(item.Items) > 0 {
			for _, item := range item.Items {
				go s.itemHandler(item)
			}
		}

		s.workersWg.Done()
		response.Close()
	}
}

func (s *Spider) fetchContent(url string) (*Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error downloading content")
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading content")
		return nil, err
	}
	defer resp.Body.Close()

	return NewResponse(url, body)
}
