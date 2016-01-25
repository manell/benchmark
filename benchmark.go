package benchmark

import (
	"flag"
	"log"
	"time"
)

var (
	n = flag.Int("n", 0, "number of requests")
	c = flag.Int("c", 1, "number of concurrent workers")
	k = flag.Bool("k", false, "reuse TCP connections")
)

// FlowRunner is an interface that represents the ability to run a workflow.
type FlowRunner interface {
	RunFlow(*Client) error
}

// Operation contains information about an HTTP request.
type Operation struct {
	Name   string
	Method string
	Path   string
}

// Metric contains information about an HTTP request.
type Metric struct {
	StartTime time.Time
	Operation *Operation
	FinalTime time.Time
}

type Benchmark struct {
	// N is the number of flow executions.
	N int

	// C is the number of concurrent workers to run the flow.
	C int

	// Dissable TCP connections re-use.
	DisableKeepAlives bool

	syncFeed       chan int
	statsCollector chan *Metric
	statsConsumers []chan *Metric
	flow           FlowRunner
}

// NewBenchmark returns a new instance of Benchmark.
func NewBenchmark(flow FlowRunner) *Benchmark {
	flag.Parse()
	statsConsumers := []chan *Metric{}

	statsCollector := make(chan *Metric, 4096)

	bench := &Benchmark{
		statsCollector:    statsCollector,
		statsConsumers:    statsConsumers,
		flow:              flow,
		C:                 *c,
		N:                 *n,
		DisableKeepAlives: *k,
		syncFeed:          make(chan int),
	}

	return bench
}

// 1 to n channel
func (b *Benchmark) feedConsumers() {
	// Basically it sends the output of a channel to the input of n-channels.
	for metric := range b.statsCollector {
		for _, consumer := range b.statsConsumers {
			consumer <- metric
		}
	}

	// Notify that all data has been properly sent to the consumers.
	b.syncFeed <- 1
}

// Run executes the benchmark.
func (b *Benchmark) Run() {
	for _, consumer := range regConsumers {
		c := make(chan *Metric, 4096)
		b.statsConsumers = append(b.statsConsumers, c)
		go consumer.Run(c, b.C)
	}

	// Connect the collector with consumers.
	go b.feedConsumers()

	fi := make(chan int, b.N)

	workerFeed := make(chan int, b.N)

	// Start C workers.
	for j := 0; j < b.C; j++ {
		go b.runWorker(workerFeed, fi)
	}

	for i := 0; i < b.N; i++ {
		workerFeed <- 1
	}
	close(workerFeed)

	// Wait until all requests finalize.
	for j := 0; j < b.N; j++ {
		<-fi
	}

	// There are no more metrics to send, so we need to notify the feeder
	// that no more data will be sent.
	close(b.statsCollector)

	// Wait until the feeder sends all the remaining metrics to the consumers.
	<-b.syncFeed

	// Notify the consumers that no more data will be sent.
	for _, channel := range b.statsConsumers {
		close(channel)
	}

	// All channels are closed and no more data will be generated.
	for _, consumer := range regConsumers {
		consumer.Finalize()
	}
}

func (b *Benchmark) runWorker(iterations chan int, waitSync chan int) {
	for _ = range iterations {
		// Lets create a new client for each worker.
		cli := NewClient(b.statsCollector, b.DisableKeepAlives)
		if err := b.flow.RunFlow(cli); err != nil {
			log.Fatal(err)
		}
		waitSync <- 1
	}
}
