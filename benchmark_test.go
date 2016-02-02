package benchmark

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

type mockConsumers struct{}

func (m *mockConsumers) Initialize(n, c int) {}
func (m *mockConsumers) Pipe(input chan *Metric) chan int {
	c := make(chan int, 1)
	c <- 1
	return c
}
func (m *mockConsumers) Finalize(d time.Duration) {}

type mockCollector struct{}

func (m *mockCollector) Initialize(chan *Metric)                 {}
func (m *mockCollector) NewMeasure(string, *http.Request) func() { return func() {} }

type mockFlow struct {
	n     int
	mutex *sync.Mutex
	init  bool
}

func (m *mockFlow) Initialize(opts *InitParameters) error {
	m.init = true

	return nil
}
func (m *mockFlow) RunFlow(cli *BenchClient) error {
	m.mutex.Lock()
	m.n++
	m.mutex.Unlock()

	return nil
}

func TestBencharkNC(t *testing.T) {
	tests := []struct {
		n int
		c int
	}{
		{0, 0},
		{0, 1},
		{1, 1},
		{1, 10},
		{10, 1},
		{10, 10},
		{100, 10},
	}

	for _, test := range tests {
		b := &Benchmark{
			N:                 test.n,
			C:                 test.c,
			DisableKeepAlives: false,
			Consumers:         &mockConsumers{},
			Collector:         &mockCollector{},
		}

		flow := &mockFlow{0, &sync.Mutex{}, false}
		b.Run(flow)

		if flow.n != test.n {
			t.Fatal("Incorrect iterations value")
		}

		if !flow.init {
			t.Fatal("Initialize not executed")
		}
	}
}
