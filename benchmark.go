package benchmark

import (
	"log"
	"time"
)

// FlowRunner is an interface that represents the ability to run a workflow
type FlowRunner interface {
	RunFlow(int, int) error
	InitCollector(chan *Metric)
}

// Operation contains information about an HTTP request
type Operation struct {
	Name   string
	Method string
	Path   string
}

// Metric contains information about an HTTP request
type Metric struct {
	StartTime time.Time
	Operation *Operation
	FinalTime time.Time
}

type Benchmark struct {
	statsCollector chan *Metric
	statsConsumers []chan *Metric
	flow           FlowRunner
	C              int
	N              int
	syncFeed       chan int
}

// NewBenchmark returns a new instance of Benchmark
func NewBenchmark(flow FlowRunner) *Benchmark {
	statsConsumers := []chan *Metric{}

	statsCollector := make(chan *Metric, 4096)

	bench := &Benchmark{
		statsCollector: statsCollector,
		statsConsumers: statsConsumers,
		flow:           flow,
		C:              1,
		N:              1,
		syncFeed:       make(chan int),
	}

	return bench
}

// 1 to n channel
func (b *Benchmark) feedConsumers() {
	// Basically it sends the output of a channel to the input of n-channels
	for metric := range b.statsCollector {
		for _, consumer := range b.statsConsumers {
			consumer <- metric
		}
	}

	// Notify that all data has been properly sent to the consumers
	b.syncFeed <- 1
}

// Run executes the benchmark
func (b *Benchmark) Run() {
	for _, consumer := range regConsumers {
		c := make(chan *Metric, 4096)
		b.statsConsumers = append(b.statsConsumers, c)
		go consumer.Run(c, b.C)
	}

	// Connect the collector with the flow
	b.flow.InitCollector(b.statsCollector)

	// Connect the collector with consumers
	go b.feedConsumers()

	fi := make(chan int)

	for j := 0; j < b.C; j++ {
		go func(j, number int) {
			for i := 0; i < number; i++ {
				if err := b.flow.RunFlow(i, j); err != nil {
					log.Fatal(err)
				}
			}
			fi <- 1
		}(j, b.N)
	}

	// Wait until all requests finalize
	for j := 0; j < b.C; j++ {
		<-fi
	}

	// There are no more metrics to send, so we need to notify the feeder
	// that no more data will be sent
	close(b.statsCollector)

	// Wait until the feeder sends all the remaining metrics to the consumers
	<-b.syncFeed

	// Notify the consumers that no more data will be sent
	for _, channel := range b.statsConsumers {
		close(channel)
	}

	// All channels are closed and no more data will be generated
	for _, consumer := range regConsumers {
		consumer.Finalize()
	}
}
