package benchmark

import (
	"flag"
	"log"
	"time"
)

var (
	n = flag.Int("n", 1, "Number of iterations")
	c = flag.Int("c", 1, "Number of concurrent workers")
	k = flag.Bool("k", true, "Reuse TCP connections")
)

// consumersRegistry contains the available consumers
var consumersRegistry = NewConsumers()

// Register allows registering new consumers with an unique name
func Register(name string, consumer Consumer) {
	if err := consumersRegistry.Register(name, consumer); err != nil {
		panic(err)
	}
}

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
	Duration  float64
}

type Benchmark struct {
	// N is the number of flow executions.
	N int

	// C is the number of concurrent workers to run the flow.
	C int

	// Dissable TCP connections re-use.
	DisableKeepAlives bool

	consumers      *Consumers // Change to interface
	syncFeed       chan int
	statsCollector chan *Metric
	flow           FlowRunner
}

// NewBenchmark returns a new instance of Benchmark.
func NewBenchmark(flow FlowRunner) *Benchmark {
	flag.Parse()

	statsCollector := make(chan *Metric, 4096)

	bench := &Benchmark{
		statsCollector:    statsCollector,
		flow:              flow,
		C:                 *c,
		N:                 *n,
		DisableKeepAlives: !*k,
		consumers:         consumersRegistry,
		syncFeed:          make(chan int),
	}

	return bench
}

// Run executes the benchmark.
func (b *Benchmark) Run() {
	b.consumers.Initialize(b.N, b.C)

	// Connect the collector with consumers.
	dataSent := b.consumers.Pipe(b.statsCollector)

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

	// There are no more metrics to send, so we need to notify the consumers
	// that no more data will be sent.
	close(b.statsCollector)

	// Wait until the remaining data is sent to the consumers
	<-dataSent

	// All channels are closed and no more data will be generated. Lets do some
	// stuff with the data.
	b.consumers.Finalize()
}

func (b *Benchmark) runWorker(iterations chan int, waitSync chan int) {
	// Lets create a new client for each worker.
	cli := NewClient(b.statsCollector, b.DisableKeepAlives)
	for _ = range iterations {
		if err := b.flow.RunFlow(cli); err != nil {
			log.Fatal(err)
		}
		waitSync <- 1
	}
}
