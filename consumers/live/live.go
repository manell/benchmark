package live

import (
	"flag"
	"fmt"
	"time"

	"github.com/manell/benchmark"
)

var (
	loaded = flag.Bool("live", false, "Use live module.")
)

func init() {
	benchmark.Register("live", &Live{
		count:    make(map[string]int),
		ticker:   time.NewTicker(time.Second * 1),
		countLat: make(map[string]float64),
		sync:     make(chan int, 1),
	})
}

type Live struct {
	count       map[string]int
	countLat    map[string]float64
	ticker      *time.Ticker
	concurrency int
	sync        chan int
}

func (s *Live) Loaded() bool { return *loaded }

func (s *Live) Run(collector chan *benchmark.Metric, iterations, concurrency int) {
	s.concurrency = concurrency
	s.sync <- 1

	go s.Print()
	for metric := range collector {
		<-s.sync
		s.count[metric.Name] += 1
		s.countLat[metric.Name] += float64(metric.Duration.Nanoseconds())
		s.sync <- 1
	}
	s.ticker.Stop()
}

func (s *Live) Print() {
	go func() {
		for _ = range s.ticker.C {
			for op, _ := range s.count {
				<-s.sync
				lat := (s.countLat[op] / float64(s.count[op])) / 1e9
				fmt.Printf("%d TPS   %d TPS  %d TPS   %f\n",
					s.count[op], int((1/lat)*float64(s.concurrency)),
					int((1/lat)*float64(s.concurrency))-s.count[op],
					lat*1e9)
				s.count[op] = 0
				s.countLat[op] = 0
				s.sync <- 1
			}
		}
	}()

}

func (s *Live) Finalize(d time.Duration) {}
