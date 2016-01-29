package benchmark

import (
	"flag"
	"log"
	"net/http"
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

// Register allows registering new consumers with an unique name.
func Register(name string, consumer Consumer) {
	if err := consumersRegistry.Register(name, consumer); err != nil {
		panic(err)
	}
}

// FlowRunner is an interface that represents the ability to run a workflow.
type FlowRunner interface {
	RunFlow(*BenchClient) error
}

type consumersManager interface {
	Initialize(int, int)
	Pipe(chan *Metric) chan int
	Finalize(time.Duration)
}

type collector interface {
	Initialize(chan *Metric)
	NewMeasure(string, *http.Request) func()
}

// Operation contains information about an HTTP request.
type Operation struct {
	Method string
	Path   string
	Host   string
}

// Metric contains information about an HTTP request.
type Metric struct {
	StartTime time.Time
	Operation *Operation
	FinalTime time.Time
	Duration  time.Duration
	Name      string
}

type Benchmark struct {
	// N is the number of flow executions.
	N int

	// C is the number of concurrent workers to run the flow.
	C int

	// Dissable TCP connections re-use.
	DisableKeepAlives bool

	Consumers consumersManager
	Collector collector
}

// NewBenchmark returns a new instance of Benchmark.
func NewBenchmark() *Benchmark {
	flag.Parse()

	bench := &Benchmark{
		C:                 *c,
		N:                 *n,
		DisableKeepAlives: !*k,
		Consumers:         consumersRegistry,
		Collector:         &CollectStats{},
	}

	return bench
}

// Run executes the benchmark.
func (b *Benchmark) Run(flow FlowRunner) {
	ouput := make(chan *Metric, b.N)
	b.Collector.Initialize(ouput)

	//These will consume the metrics.
	b.Consumers.Initialize(b.N, b.C)

	// Connect the collector with consumers.
	dataSent := b.Consumers.Pipe(ouput)

	// Execute the benchmark.
	start := time.Now()
	b.runNCBenchmark(flow)
	elapsed := time.Since(start)

	// There are no more metrics to send.
	close(ouput)

	// Wait until the remaining data is sent to the consumers
	<-dataSent

	// All channels are closed and no more data will be generated. Lets do some
	// stuff with the data.
	b.Consumers.Finalize(elapsed)
}

func (b *Benchmark) runNCBenchmark(flow FlowRunner) {
	var fi sync.WaitGroup
	fi.Add(b.N)

	iterations := make(chan int, b.N)

	for j := 0; j < b.C; j++ {
		go b.runWorker(flow, iterations, &fi)
	}

	for i := 0; i < b.N; i++ {
		iterations <- 1
	}
	close(iterations)

	fi.Wait()
}

func (b *Benchmark) runWorker(flow FlowRunner, iterations chan int, waitSync *sync.WaitGroup) {
	// Lets create a new client for each worker.
	cli := NewClient(b.Collector, b.DisableKeepAlives)
	for _ = range iterations {
		if err := flow.RunFlow(cli); err != nil {
			log.Fatal(err)
		}
		waitSync.Done()
	}
}
