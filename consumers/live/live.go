package live

import (
	"flag"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/manell/benchmark"
)

var (
	loaded = flag.Bool("live", false, "Use live module.")
)

func init() {
	benchmark.Register("live", &Live{})
}

// LiveMetric calculates and prints statistics
type LiveMetric struct {
	count    map[string]int
	countLat map[string]float64
	c        int
	sync.RWMutex
}

// Read computes a new metric
func (lm *LiveMetric) Read(metric *benchmark.Metric) {
	lm.Lock()
	defer lm.Unlock()

	lm.count[metric.Name] += 1
	lm.countLat[metric.Name] += float64(metric.Duration.Nanoseconds())
}

// Print prints the available metrics and resets the counters
func (lm *LiveMetric) Print() {
	var keys []string
	for k := range lm.count {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, op := range keys {
		lm.RLock()
		lat := (lm.countLat[op] / float64(lm.count[op])) / (1e6 * float64(lm.c))

		fmt.Printf("%s: %.3f [ms]  %f [req/s]", op, lat, (1/lat)*1000)
		lm.count[op] = 0
		lm.countLat[op] = 0
		lm.RUnlock()
	}
	fmt.Println()
}

// Live is a consumer that periodically print statistics using the recieved metrics.
type Live struct{}

func (s *Live) Loaded() bool { return *loaded }

func (s *Live) Run(collector chan *benchmark.Metric, iterations, concurrency int) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	lm := &LiveMetric{
		count:    make(map[string]int),
		countLat: make(map[string]float64),
		c:        concurrency,
	}

	go func() {
		for _ = range ticker.C {
			lm.Print()
		}
	}()

	for metric := range collector {
		lm.Read(metric)
	}
}

// Finalize does nothing. (needed to implement the consumer interface)
func (s *Live) Finalize(d time.Duration) {}
