package statistics

import (
	"flag"
	"fmt"
	"math"
	"sort"

	"time"

	"github.com/manell/benchmark"
)

var (
	loaded = flag.Bool("report", false, "Use report module.")
)

func init() {
	sync := make(chan int, 1)
	benchmark.Register("report", &Report{
		sync:    sync,
		metrics: make(map[string]*MetricList),
	})
}

type MetricList struct {
	latencies    []float64
	avgLatsTotal float64
	avgLats      float64
	tps          float64
	op           *benchmark.Operation
	name         string
	concurrency  int
	requests     int
	timeTaken    time.Duration
}

func (m *MetricList) Print() {
	fmt.Printf("Name: %s\nMethod: %s\nHost: %s\nPath: %s\n\n", m.name, m.op.Method, m.op.Host, m.op.Path)

	fmt.Printf("Concurrency Level: %d\n", m.concurrency)
	fmt.Printf("Time taken for tests %f seconds\n", m.timeTaken.Seconds())
	fmt.Printf("Complete iterations: %d\n", m.requests)

	fmt.Printf("Requests per second: %f [#/sec] (mean)\n", m.tps)
	fmt.Printf("Time per request: %f [ms] (mean)\n", m.avgLats)
	fmt.Printf("Time per request: %f [ms] (mean across all concurrent requests)\n\n",
		m.avgLats/float64(m.concurrency))

	fmt.Printf("Fastest request: %f [ms]\n", m.latencies[0])
	fmt.Printf("Slowest request: %f [ms]\n\n", m.latencies[len(m.latencies)-1])

	m.PrintPercentiles()

}

func (m *MetricList) PrintPercentiles() {
	percent := []int{50, 66, 75, 80, 90, 95, 98, 99, 100}
	size := len(m.latencies)

	fmt.Println("Percentage of the requests served within a certain time (ms)")

	for i := 0; i < len(percent); i++ {
		posf := (float64(size) / 100) * float64(percent[i])
		pos := math.Ceil(posf)
		fmt.Printf("%d%%  %f\n", percent[i], m.latencies[int(pos)-1])
	}
}

type Report struct {
	sync    chan int
	n       int
	c       int
	metrics map[string]*MetricList
}

func (r *Report) Loaded() bool { return *loaded }

func (r *Report) Run(collector chan *benchmark.Metric, iterations, concurrency int) {
	r.n = iterations
	r.c = concurrency

	for m := range collector {
		if r.metrics[m.Name] == nil {
			r.metrics[m.Name] = &MetricList{
				op:   m.Operation,
				name: m.Name,
			}
		}

		durationMs := float64(m.Duration.Nanoseconds()) / 1e6
		r.metrics[m.Name].latencies = append(r.metrics[m.Name].latencies, durationMs)
		r.metrics[m.Name].avgLatsTotal += durationMs
	}

	r.sync <- 1
}

func (r *Report) Finalize(d time.Duration) {
	<-r.sync

	for _, m := range r.metrics {
		m.avgLats = m.avgLatsTotal / float64(len(m.latencies))
		m.tps = float64(len(m.latencies)) / d.Seconds()
		sort.Float64s(m.latencies)
		m.requests = r.n
		m.concurrency = r.c
		m.timeTaken = d

		m.Print()
	}
}
