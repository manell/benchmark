package benchmark

import (
	"net/http"
)

// BenchClient contains helpers for measuring metrics and executing HTTP requests.
type BenchClient struct {
	Collector collector
	Client    *http.Client
}

//  NewClient returns a new instance of a BenchClient
func NewClient(collector collector, keepAlive bool) *BenchClient {
	tp := &http.Transport{
		MaxIdleConnsPerHost: 16,
		DisableKeepAlives:   keepAlive,
	}

	client := &http.Client{Transport: tp}

	return &BenchClient{Client: client, Collector: collector}
}
