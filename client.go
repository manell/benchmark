package benchmark

import (
	"net/http"
	"time"
)

// Client runs HTTP requests and collects metrics
type Client struct {
	collector chan *Metric
}

// InitCollector configures the collector that will be used to send statistics
func (c *Client) InitCollector(collector chan *Metric) {
	c.collector = collector
}

// Do is a wrap in top of http.Client that execute requests and collects metrics
func (c *Client) Do(name string, req *http.Request, client *http.Client) (*http.Response, error) {
	startTime := time.Now()

	resp, err := client.Do(req)

	stat := &Metric{
		StartTime: startTime,
		Operation: &Operation{
			Name:   name,
			Method: req.Method,
			Path:   req.URL.Path,
		},
		FinalTime: time.Now(),
	}

	c.collector <- stat

	return resp, err
}
