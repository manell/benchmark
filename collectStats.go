package benchmark

import (
	"net/http"
	"time"
)

// CollectStats takes measures and send these measure to an output channel.
type CollectStats struct {
	output chan *Metric
}

// Initialize sets its channel
func (c *CollectStats) Initialize(ouput chan *Metric) {
	c.output = ouput
}

// NewMeasure starts a new stats measure and returns a function to finalize
// the measure. The result will be sent to the CollectStats output channel.
func (c *CollectStats) NewMeasure(name string, req *http.Request) func() {
	startTime := time.Now()

	op := &Operation{}
	if req != nil {
		op = &Operation{
			Host:   req.URL.Host,
			Method: req.Method,
			Path:   req.URL.Path,
		}
	}

	final := func() {
		finalTime := time.Now()

		duration := time.Since(startTime)

		stat := &Metric{
			Name:      name,
			StartTime: startTime,
			Operation: op,
			FinalTime: finalTime,
			Duration:  duration,
		}

		c.output <- stat
	}

	return final
}

// Close simply close the output channel
func (c *CollectStats) close() {
	close(c.output)
}
