package benchmark

import (
	"net/http"
	"time"
)

// Client runs HTTP requests and collects metrics
type Client struct {
	Collector chan *Metric
	Client    *http.Client
}

//  NewClient returns a new instance of a Client
func NewClient(collector chan *Metric, keepAlive bool) *Client {
	tp := &http.Transport{
		MaxIdleConnsPerHost: 16,
		DisableKeepAlives:   keepAlive,
	}

	client := &http.Client{Transport: tp}

	return &Client{Client: client, Collector: collector}
}

// Do is a wrap on top of http.Client that execute requests and collects metrics
func (c *Client) Do(name string, req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	resp, err := c.Client.Do(req)

	fiTime := time.Now()

	duration := fiTime.UnixNano() - startTime.UnixNano()
	durationMs := float64(duration) / float64(time.Millisecond)

	stat := &Metric{
		StartTime: startTime,
		Operation: &Operation{
			Name:   name,
			Method: req.Method,
			Path:   req.URL.Path,
		},
		FinalTime: time.Now(),
		Duration:  durationMs,
	}

	c.Collector <- stat

	return resp, err
}
