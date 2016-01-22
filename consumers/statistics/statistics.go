package statistics

import (
	"fmt"
	"time"

	"github.com/manell/benchmark"
	"github.com/montanaflynn/stats"
)

func init() {
	sync := make(chan int, 1)
	benchmark.Register("statistics", &Statistics{
		sync: sync,
		data: make(map[benchmark.Operation][]float64),
	})
}

type Statistics struct {
	sync        chan int
	data        map[benchmark.Operation][]float64
	concurrency int
}

func (s *Statistics) Run(collector chan *benchmark.Metric, concurrency int) {
	s.concurrency = concurrency

	for metric := range collector {
		duration := metric.FinalTime.UnixNano() - metric.StartTime.UnixNano()
		durationMs := float64(duration) / float64(time.Millisecond)

		s.data[*metric.Operation] = append(s.data[*metric.Operation], durationMs)
	}
	s.sync <- 1
}

func (s *Statistics) Finalize() {
	for key, times := range s.data {
		fmt.Printf("Operation: %s\nMethod: %s\nPath: %s\n\n", key.Name, key.Method, key.Path)

		mean, _ := stats.Mean(times)
		meanConc := mean / float64(s.concurrency)
		rps := 1 / (meanConc / 1000)

		fmt.Printf("Requests per second: %f [#/ms] (mean)\n", rps)

		fmt.Printf("Time per request: %f [ms] (mean)\n", mean)

		fmt.Printf("Time per request: %f [ms] (mean across all concurrent requests)\n\n", meanConc)

		fmt.Println("Percentage of the requests served within a certain time (ms)")
		per50, _ := stats.Percentile(times, 50)
		per65, _ := stats.Percentile(times, 65)
		per75, _ := stats.Percentile(times, 75)
		per85, _ := stats.Percentile(times, 85)
		per90, _ := stats.Percentile(times, 90)
		per95, _ := stats.Percentile(times, 95)
		per99, _ := stats.Percentile(times, 99)
		per100, _ := stats.Percentile(times, 99.9999999999)
		fmt.Printf("  50%%  %f \n  65%%  %f\n  75%%  %f\n  85%%  %f\n  90%%  %f\n  95%%  %f\n  99%%  %f\n  100%% %f\n",
			per50, per65, per75, per85, per90, per95, per99, per100)
	}
	<-s.sync
}
