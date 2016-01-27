package benchmark

import (
	"flag"
	"log"
	"sync"
	"time"
)

var (
	n = flag.Int("n", 0, "Number of iterations.")
	c = flag.Int("c", 1, "Number of concurrent workers.")
	k = flag.Bool("k", true, "Reuse TCP connections.")
)

// consumersRegistry contains the available Consumer.
var consumersRegistry = NewConsumers()

// Register allows registering new consumers with an unique name
func Register(name string, consumer Consumer) {
	if err := consumersRegistry.Register(name, consumer); err != nil {
		panic(err)
	}
}

// FlowRunner is an interface that represents the ability to run a workflow.
type FlowRunner interface {
	RunFlow(*BenchClient) error
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

	Consumers *Consumers // Change to interface?
}

// NewBenchmark returns a new instance of Benchmark.
func NewBenchmark() *Benchmark {
	flag.Parse()

	bench := &Benchmark{
		C:                 *c,
		N:                 *n,
		DisableKeepAlives: !*k,
		Consumers:         consumersRegistry, //Should be interface!,
	}

	return bench
}

// Run executes the benchmark.
func (b *Benchmark) Run(flow FlowRunner) {
	b.Consumers.Initialize(b.N, b.C)

	collector := NewCollectStats(b.N)

	// Connect the collector with consumers.
	dataSent := b.Consumers.Pipe(collector.output)

	b.runNCBenchmark(flow, collector)

	// There are no more metrics to send, so we need to notify the Consumers
	// that no more data will be sent.
	collector.close()

	// Wait until the remaining data is sent to the consumers
	<-dataSent

	// All channels are closed and no more data will be generated. Lets do some
	// stuff with the data.
	b.Consumers.Finalize()
}

func (b *Benchmark) runNCBenchmark(flow FlowRunner, collector *CollectStats) {
	var fi sync.WaitGroup
	fi.Add(b.N)

	iterations := make(chan int) // Not sure if it must be buffered

	for j := 0; j < b.C; j++ {
		go b.runWorker(flow, collector, iterations, &fi)
	}

	for i := 0; i < b.N; i++ {
		iterations <- 1
	}
	close(iterations)

	fi.Wait()
}

func (b *Benchmark) runWorker(flow FlowRunner, collector *CollectStats, iterations chan int, waitSync *sync.WaitGroup) {
	// Lets create a new client for each worker.
	cli := NewClient(collector, b.DisableKeepAlives)
	for _ = range iterations {
		measure := collector.NewMeasure(&Operation{Name: "Flow"})

		if err := flow.RunFlow(cli); err != nil {
			log.Fatal(err)
		}

		measure()

		waitSync.Done()
	}
}
