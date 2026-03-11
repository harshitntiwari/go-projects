package main

import "fmt"

// A fetcher should implememt this interface.
type Fetcher interface {
	// Fetches and returns the data (body + URLs) from a URL.
	Fetch(url string) (body string, urls []string, err error)
}

// -------------create a fake fetcher ------------

type fakeResult struct {
	body string
	urls []string
}

type FakeFetcher struct {
	items map[string]*fakeResult
}

// returns a new [FakeFetcher]
func NewFakeFetcher() *FakeFetcher {
	return &FakeFetcher{
		items: map[string]*fakeResult {
			"https://example.org" : {
				body: "some text",
				urls: []string{
					"https://go.dev",
					"https://golang.org/pkg/os/",
				},
			},
			"https://go.dev" : {
				body: "some text",
				urls: []string{
					"https://golang.org/cmd/",
					"https://golang.org/pkg/os/",
				},
			},
		},
	}
}

func (f *FakeFetcher) Fetch(url string) (string, []string, error) {
	if data, ok := f.items[url]; ok {
		return data.body, data.urls, nil
	}
	return "", nil, fmt.Errorf("url not found: %s", url)
}

// ------------------------------------------



