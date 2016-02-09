package benchmark

import (
	"testing"
	"time"
)

type MockConsumer struct {
	load     bool
	finalize bool
	n        int
	c        int
	loadRun  chan int
	duration *time.Duration
}

func (mc *MockConsumer) Loaded() bool { return mc.load }

func (mc *MockConsumer) Run(input chan *Metric, n int, c int) {
	mc.n = n
	mc.c = c
	mc.loadRun <- 1
}

func (mc *MockConsumer) Finalize(d time.Duration) { mc.duration = &d }

type MockConsumer2 struct {
	data    *Metric
	loadRun chan int
}

func (mc *MockConsumer2) Loaded() bool { return true }

func (mc *MockConsumer2) Run(input chan *Metric, n int, c int) {
	mc.data = <-input
	mc.loadRun <- 1
}
func (mc *MockConsumer2) Finalize(d time.Duration) { <-mc.loadRun }

func TestRegister(t *testing.T) {
	mc := &MockConsumer{}

	c := NewConsumers()
	if err := c.Register("mc1", mc); err != nil {
		t.Fatal(err)
	}
	if err := c.Register("mc1", mc); err == nil {
		t.Fatal("Consumers cannot be registered twice")
	}
}

func TestInitialize(t *testing.T) {
	mc := &MockConsumer{load: true, loadRun: make(chan int, 1)}
	mc2 := &MockConsumer{load: false, loadRun: make(chan int, 1)}

	cons := NewConsumers()
	cons.Register("mc", mc)
	cons.Register("mc2", mc2)

	n := 2
	c := 3

	cons.Initialize(n, c)
	<-mc.loadRun

	if mc.n != n || mc.c != c {
		t.Fatal("C and N should be properly set")
	}
	if mc2.n == n || mc2.c == c {
		t.Fatal("Should not be loaded")
	}
}

func TestFinalize(t *testing.T) {
	mc := &MockConsumer{load: false, loadRun: make(chan int, 1)}
	mc2 := &MockConsumer{load: true, loadRun: make(chan int, 1)}

	cons := NewConsumers()
	cons.Register("mc", mc)
	cons.Register("mc2", mc2)

	cons.Finalize(time.Second * 1)

	if *mc2.duration != time.Second*1 {
		t.Fatal("Duration incorrect")
	}
	if mc.duration != nil {
		t.Fatal("Duration should be nil")
	}
}

func TestPipe(t *testing.T) {
	mc1 := &MockConsumer2{loadRun: make(chan int, 1)}
	mc2 := &MockConsumer2{loadRun: make(chan int, 1)}

	cons := NewConsumers()
	cons.Register("mc1", mc1)
	cons.Register("mc2", mc2)
	cons.Initialize(1, 1)

	input := make(chan *Metric, 1)
	done := cons.Pipe(input)

	name := "test"
	m := &Metric{Name: name}
	input <- m
	close(input)

	<-done
	mc1.Finalize(time.Second * 1)
	mc2.Finalize(time.Second * 1)

	if mc1.data.Name != name {
		t.Fatal("mc1 has not recived the metric")
	}
	if mc2.data.Name != name {
		t.Fatal("mc2 has not recived the metric")
	}
}
