package benchmark

import (
	"net/http"
)

// BenchClient contains helpers for measuring metrics and executing HTTP requests.
type BenchClient struct {
	Collector *CollectStats
	Client    *http.Client
}

//  NewClient returns a new instance of a BenchClient
func NewClient(collector *CollectStats, keepAlive bool) *BenchClient {
	tp := &http.Transport{
		MaxIdleConnsPerHost: 16,
		DisableKeepAlives:   keepAlive,
	}

	client := &http.Client{Transport: tp}

	return &BenchClient{Client: client, Collector: collector}
}
