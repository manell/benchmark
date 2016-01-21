package benchmark

import (
	"time"
)

// FlowRunner is an interface that represents the ability to run a workflow
type FlowRunner interface {
	RunFlow(int, int) error
	InitCollector(chan *Metric)
}

// Operation conatins info about an HTTP request
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
}

// NewBenchmark retuns a new instance of Benchmark
func NewBenchmark(flow FlowRunner) *Benchmark {
	statsConsumers := []chan *Metric{}

	for _, consumer := range regConsumers {
		c := make(chan *Metric, 4096)
		statsConsumers = append(statsConsumers, c)
		consumer.Run(c)
	}

	statsCollector := make(chan *Metric, 4096)
	flow.InitCollector(statsCollector)

	bench := &Benchmark{
		statsCollector: statsCollector,
		statsConsumers: statsConsumers,
		flow:           flow,
		C:              1,
		N:              1,
	}

	return bench
}

func (b *Benchmark) feedConsumers() {
	for metric := range b.statsCollector {
		for _, consumer := range b.statsConsumers {
			consumer <- metric
		}
	}
}

// Run executes the benchmark
func (b *Benchmark) Run() {
	// Connect channels
	go b.feedConsumers()

	fi := make(chan int)

	for j := 0; j < b.C; j++ {
		go func(j, number int) {
			for i := 0; i < number; i++ {
				b.flow.RunFlow(i, j)
			}
			fi <- 1
		}(j, b.N)
	}

	// wait
	for j := 0; j < b.C; j++ {
		<-fi
	}

	for _, consumer := range regConsumers {
		consumer.Finalize()
	}
}
