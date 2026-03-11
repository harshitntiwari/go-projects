package main

import (
	"fmt"
	"sync"
)

type Crawler interface {
	Crawl(url string)
}

// ---------------------------- serial crawler ----------------------------

// SerialCrawler crawls the web pages recursively in a non-concurrent fashion.
type SerialCrawler struct {
	// fetched map tracks if a url has already been crawled once
	fetched map[string]bool
	fetcher Fetcher
}

// returns a new [SerialCrawler]
func NewSerialCrawler() *SerialCrawler {
	return &SerialCrawler{
		fetched: map[string]bool{},
		fetcher: NewFakeFetcher(),
	}
}

func (c *SerialCrawler) Crawl(url string) {
	if c.fetched[url] {
		return
	}
	_, urls, err := c.fetcher.Fetch(url)

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, u := range urls {
		c.Crawl(u)
	}
}

// ----------------------- concurrent crawler with shared state --------------------------

type FetchedState struct {
	fetched map[string]bool
	mu sync.Mutex
}

func NewFetchedState() *FetchedState {
	return &FetchedState{
		fetched: map[string]bool{},
	}
}

func (fs *FetchedState) checkAndSetFetched(url string) bool {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	check := fs.fetched[url]
	fs.fetched[url] = true
	return check
}

type ConcurrentCrawlerSharedState struct {
	fetchedState *FetchedState
	fetcher Fetcher
}

func NewConcurrentCrawlerSharedState() *ConcurrentCrawlerSharedState {
	return &ConcurrentCrawlerSharedState{
		fetchedState: NewFetchedState(),
		fetcher: NewFakeFetcher(),
	}
}

func (c *ConcurrentCrawlerSharedState) Crawl(url string) {
	if c.fetchedState.checkAndSetFetched(url) {
		return
	}

	_, urls, err := c.fetcher.Fetch(url)

	if err != nil {
		fmt.Println("Failed to crawl: ", err)
		return
	}

	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Add(1)
		// spawn one go routine for each URL
		go func(url string) {
			c.Crawl(url)
			wg.Done()
		}(u)
	}

	wg.Wait() // wait till all goroutines are finished
}

// -------------- Concurrent crawler with channel ----------------

type ConcurrentCrawler struct {
	fetchedState *FetchedState
	urlCh chan []string
	fetcher Fetcher
	wg sync.WaitGroup
}

// allocates and returns a new [ConcurrentCrawler]
func NewConcurrentCrawler() *ConcurrentCrawler {
	return &ConcurrentCrawler{
		fetchedState: NewFetchedState(),
		urlCh: make(chan []string),
		fetcher: NewFakeFetcher(),
	}
}

func (c *ConcurrentCrawler) Crawl(url string) {
	
	go func () {
		// sending the root url
		c.urlCh <- []string{url}
	}()

	go func() {
		// wait for all goroutines to finish, then close urlCh channel
		c.wg.Wait()
		close(c.urlCh)
	}()
	
	for urls := range c.urlCh { // will exit when urlCh is closed
		for _, url := range urls {
			if c.fetchedState.checkAndSetFetched(url) { continue }
			c.wg.Add(1)
			go func (u string) {
				defer c.wg.Done()
				c.process(u)
			}(url)
		}
	}
}

func (c *ConcurrentCrawler) process(url string) {
	_, urls, err := c.fetcher.Fetch(url)

	if err != nil {
		fmt.Println("Failed to crawl: ", err)
		// if we return without sending this then Crawl() will block in next iteration of range c.urlCh
		c.urlCh <- []string{}
		return
	}

	c.urlCh <- urls
}

// -----------------Concurrent crawler with worker pool and max depth support-------------------------

type urlJob struct {
	url string
	depth int
}

type ConcurrentCrawlerWorkerPool struct {
	fetchedState *FetchedState
	fetcher Fetcher
	urlCh chan urlJob
	maxWorkers int
	maxDepth int
	workerWg sync.WaitGroup
	jobWg sync.WaitGroup
}

// NewConcurrentCrawlerWorkerPool alocates and returns a new [ConcurrentCrawlerWorkerPool]
func NewConcurrentCrawlerWorkerPool(maxWorkers int, maxDepth int) *ConcurrentCrawlerWorkerPool {
	return &ConcurrentCrawlerWorkerPool{
		fetchedState: NewFetchedState(),
		fetcher: NewFakeFetcher(),
		urlCh: make(chan urlJob, 1000),
		maxWorkers: maxWorkers,
		maxDepth: maxDepth,
	}
}

func (c *ConcurrentCrawlerWorkerPool) Crawl(url string) {
	// starting worker pool
	for range c.maxWorkers {
		c.workerWg.Add(1)
		go c.worker()
	}

	c.jobWg.Add(1) // for the root url
	go func() {
		c.jobWg.Wait()
		// this is called when the counter for job wg hits zero, which means all jobs are done
		close(c.urlCh)
	}()

	// sending the root url
	c.urlCh <- urlJob{url: url, depth: 1}

	c.workerWg.Wait()
}

func (c *ConcurrentCrawlerWorkerPool) worker() {
	defer c.workerWg.Done()
	for job := range c.urlCh {
		if c.fetchedState.checkAndSetFetched(job.url) { 
			c.jobWg.Done()
			continue 
		}
		c.process(job)
	}
}

func (c *ConcurrentCrawlerWorkerPool) process(job urlJob) {
	defer c.jobWg.Done() // decreasing the wait group counter when the job is taken out of the channel to process

	if job.depth > c.maxDepth {
		return
	}
	
	_, urls, err := c.fetcher.Fetch(job.url)

	if err != nil {
		return
	}

	for _, url := range urls {
		c.jobWg.Add(1) // increasing the wait group counter when sending the job
		c.urlCh <- urlJob{url: url, depth: job.depth + 1}
	}
}


// -----------------------------------------------------------------------