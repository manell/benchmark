package benchmark

import (
	"time"
)

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
}

func NewBenchmark(flow FlowRunner) *Benchmark {

}

// FlowRunner is an interface that represents the ability to run a workflow
type FlowRunner interface {
	RunFlow(int, int) error
	InitCollector(chan *Metric)
}
