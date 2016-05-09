package goscrape

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

//ResponseParseHandler is a defintion for the function to be called to parse data
type ResponseParseHandler func(response *Response) ParseResult

//ItemPipelineHandler is a definition for the function to be called to process data
type ItemPipelineHandler func(parsedItem ParseItem)

// FetchStatus is used to identify which links fetch operations resulted in what status
type FetchStatus uint8

// Declare multiple fetch statuses
const (
	Pending FetchStatus = iota
	Failed
	Succeeded
)

//URLRecord is a tool to help keep track of URL specific properties
type urlRecord struct {
	Retries uint8
	Wait    time.Duration
	Status  FetchStatus
}

// Response defines the attributes of parse response
type Response struct {
	URL     string
	Content []byte
}

// Spider is an object that will perform crawling and parsing of content
type Spider struct {
	// Configuration instance
	Configs *Configs

	// User specific configs
	parseHandler ResponseParseHandler
	itemHandler  ItemPipelineHandler

	// Channels
	parseChannel chan *Response
	linksChannel chan string
	stopChannel  chan struct{}

	// URL tracking
	seenLinks map[string]*urlRecord
	linksMu   sync.RWMutex

	// Workers
	workersWg sync.WaitGroup
	workersMu sync.RWMutex

	// Reporting
	LinksParsed []string
	LinksFailed []string
}

// NewSpider creates and returns an initialized instance of a spider
func NewSpider(configs *Configs, parseHandler ResponseParseHandler, itemHandler ItemPipelineHandler) *Spider {
	spider := Spider{
		Configs:      configs,
		parseHandler: parseHandler,
		itemHandler:  itemHandler,
		linksChannel: make(chan string),
		parseChannel: make(chan *Response),
		stopChannel:  make(chan struct{}),
		seenLinks:    make(map[string]*urlRecord),
		linksMu:      sync.RWMutex{},
		workersWg:    sync.WaitGroup{},
		workersMu:    sync.RWMutex{},
	}

	return &spider
}

// Start triggers the scraping pipeline process
func (s *Spider) Start() {
	workersCount := s.Configs.GetWorkersCount()
	for i := 1; i <= workersCount; i++ {
		go s.crawl()
		go s.parse()
	}

	for _, url := range s.Configs.StartUrls {
		log.Println("Starting with " + url)
		s.workersWg.Add(len(s.Configs.StartUrls))
		go func(url string) { s.linksChannel <- url }(url)
	}
	s.workersWg.Wait()

	for j := 1; j <= workersCount*2; j++ {
		s.stopChannel <- struct{}{}
	}

	// Cleanup
	time.Sleep(1000 * time.Millisecond)
	close(s.linksChannel)
	close(s.parseChannel)
	close(s.stopChannel)

	s.printReport()
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
		if urlRec, ok := s.seenLinks[url]; ok && urlRec.Status == Succeeded {
			s.linksMu.Unlock()
			s.workersWg.Done()
			continue
		}

		if _, exists := s.seenLinks[url]; !exists {
			s.seenLinks[url] = &urlRecord{
				Retries: 0,
				Wait:    s.Configs.GetRetryDuration(),
				Status:  Pending,
			}
		}

		s.linksMu.Unlock()

		log.Println("Crawl - Got URL: " + url)

		content, err := s.fetchContent(url)

		if err != nil {
			s.linksMu.Lock()
			urlRec := s.seenLinks[url]
			urlRec.Retries++
			urlRec.Status = Failed
			s.seenLinks[url] = urlRec
			s.linksMu.Unlock()

			s.linksMu.RLock()

			if s.seenLinks[url].Retries >= s.Configs.GetRetryMaxCount() {
				log.Println("Skip URL: ", url)
				s.LinksFailed = append(s.LinksFailed, url)
				s.linksMu.RUnlock()
				s.workersWg.Done()
				continue
			}
			s.linksMu.RUnlock()
			go func() {
				// Sleep for the amout of "wait" second and then retry the same URL
				s.linksMu.Lock()
				time.Sleep(s.seenLinks[url].Wait)

				s.seenLinks[url].Wait *= 2
				s.linksMu.Unlock()

				s.linksChannel <- url
			}()
			continue
		}

		s.linksMu.Lock()
		s.seenLinks[url].Status = Succeeded
		s.linksMu.Unlock()

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
				s.LinksParsed = append(s.LinksParsed, response.URL)
			}
		}

		s.workersWg.Done()
	}
}

func (s *Spider) fetchContent(urlStr string) (*Response, error) {
	request := s.Configs.GetRequest()
	client := http.Client{}
	urlInstance, err := url.Parse(urlStr)

	if err != nil {
		log.Fatalln("Malformed URL: %s", urlStr)
	}

	request.URL = urlInstance

	resp, err := client.Do(&request)

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		log.Println("Error downloading content")
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading content")
		return nil, err
	}

	response := Response{
		URL:     urlStr,
		Content: body,
	}

	return &response, nil
}

func (s *Spider) printReport() {
	fmt.Println("Links scraped:")
	fmt.Println("  - Count: ", len(s.LinksParsed))
	fmt.Println("Links failed:")
	fmt.Println("  - Count: ", len(s.LinksFailed))
}
