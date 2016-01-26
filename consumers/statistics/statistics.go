package statistics

import (
	"flag"
	"fmt"

	"github.com/manell/benchmark"
	"github.com/montanaflynn/stats"
)

var (
	loaded = flag.Bool("stats", false, "Use stats module.")
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

func (s *Statistics) Loaded() bool { return *loaded }

func (s *Statistics) Run(collector chan *benchmark.Metric, iterations, concurrency int) {
	s.concurrency = concurrency

	for metric := range collector {
		s.data[*metric.Operation] = append(s.data[*metric.Operation], metric.Duration)
	}
	s.sync <- 1
}

func (s *Statistics) Finalize() {
	<-s.sync
	for key, times := range s.data {
		fmt.Printf("Operation: %s\nMethod: %s\nPath: %s\n\n", key.Name, key.Method, key.Path)

		mean, _ := stats.Mean(times)
		meanConc := mean / float64(s.concurrency)
		rps := 1 / (meanConc / 1e3)

		fmt.Printf("Requests per second: %f [#/ms] (mean)\n", rps)

		fmt.Printf("Time per request: %f [ms] (mean)\n", mean)

		fmt.Printf("Time per request: %f [ms] (mean across all concurrent requests)\n\n", meanConc)
		fmt.Println("Percentage of the requests served within a certain time (ms)")
		per50, _ := stats.PercentileNearestRank(times, 50)
		per65, _ := stats.PercentileNearestRank(times, 65)
		per75, _ := stats.PercentileNearestRank(times, 75)
		per85, _ := stats.PercentileNearestRank(times, 85)
		per90, _ := stats.PercentileNearestRank(times, 90)
		per95, _ := stats.PercentileNearestRank(times, 95)
		per99, _ := stats.PercentileNearestRank(times, 99)
		per100, _ := stats.PercentileNearestRank(times, 100)
		fmt.Printf("  50%%  %f \n  65%%  %f\n  75%%  %f\n  85%%  %f\n  90%%  %f\n  95%%  %f\n  99%%  %f\n  100%% %f\n",
			per50, per65, per75, per85, per90, per95, per99, per100)
	}
}
