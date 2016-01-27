package benchmark

import (
	"time"
)

// CollectStats takes measures and send these measure to an output channel.
type CollectStats struct {
	output chan *Metric
}

// NewCollector returns a new instance of CollectStats.
func NewCollectStats(n int) *CollectStats {
	c := &CollectStats{output: make(chan *Metric, n)}
	return c
}

// NewMeasure starts a new stats measure and returns a function to finalize
// the measure. The result will be sent to the CollectStats output channel.
func (c *CollectStats) NewMeasure(op *Operation) func() {
	startTime := time.Now()

	final := func() {
		finalTime := time.Now()

		duration := time.Since(startTime)

		stat := &Metric{
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
