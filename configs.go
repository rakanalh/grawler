package goscrape

import (
	"log"
	"net/http"
	"runtime"
	"time"
)

// Configs is an object that holds properties of
// the scraping process
type Configs struct {
	// Defines list of start URLS for the scraper
	StartUrls []string

	// Enables customizing request
	Request http.Request

	// Number of workers
	WorkersCount int

	// Link failure retry duration
	RetryDuration time.Duration

	// Link failure maximum retries
	RetryMaxCount uint8
}

// GetRequest instantiates a default request object
func (configs *Configs) GetRequest() http.Request {
	if configs.Request.Method == "" {
		request, err := http.NewRequest("GET", "", nil)
		if err != nil {
			log.Println("Could not create request")
		}
		return *request
	}
	return configs.Request
}

// GetWorkersCount returns the number of workers as the number
// of available CPUs by default if not provided
func (configs *Configs) GetWorkersCount() int {
	if configs.WorkersCount == 0 {
		return runtime.GOMAXPROCS(0)
	}
	return configs.WorkersCount
}

// GetRetryDuration returns the default retry wait duration
// if not provided
func (configs *Configs) GetRetryDuration() time.Duration {
	if configs.RetryDuration == 0 {
		return time.Second
	}
	return configs.RetryDuration
}

// GetRetryMaxCount returns the max number of retries
// if not provided
func (configs *Configs) GetRetryMaxCount() uint8 {
	if configs.RetryMaxCount == 0 {
		return 3
	}
	return configs.RetryMaxCount
}
