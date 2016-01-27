package benchmark

import (
	"errors"
	"time"
)

// Consumer is an interface that represents the ability to consume metrics and do
// some stuff with them.
type Consumer interface {
	Loaded() bool
	Run(chan *Metric, int, int)
	Finalize(time.Duration)
}

// Consumers handles the registration and execution of multiple Consumer.
type Consumers struct {
	registry        map[string]Consumer
	consumersInputs []chan *Metric
	workersOutput   chan *Metric
}

// NewConsumers returns a new instance of Consumers
func NewConsumers() *Consumers {
	registry := make(map[string]Consumer)
	consumersInputs := []chan *Metric{}
	c := &Consumers{
		registry:        registry,
		consumersInputs: consumersInputs,
	}

	return c
}

// Initialize asks to each consumer if it must be loaded. If so, creates a new
// input channel for the consumer, and launches the consumer in a new goroutine.
func (c *Consumers) Initialize(number, concurrency int) {
	for _, consumer := range c.registry {
		if consumer.Loaded() {
			input := make(chan *Metric, number)
			c.consumersInputs = append(c.consumersInputs, input)
			go consumer.Run(input, number, concurrency)
		}
	}
}

// Pipe sends the recieved data of the input channel to each consumer. It returns
// a channel that indicates whether al data has been sent.
func (c *Consumers) Pipe(input chan *Metric) chan int {
	done := make(chan int)
	go c.feedConsumers(input, done)

	return done
}

// Fan-out channel
func (c *Consumers) feedConsumers(input chan *Metric, done chan int) {
	for metric := range input {
		for _, consumerInput := range c.consumersInputs {
			consumerInput <- metric
		}
	}

	for _, consumerInput := range c.consumersInputs {
		close(consumerInput)
	}

	done <- 1
}

// Register registers a new unique consumer.
func (c *Consumers) Register(name string, consumer Consumer) error {
	if consumer == nil {
		return errors.New("consumer: Register consumer is nil")
	}
	if _, dup := c.registry[name]; dup {
		return errors.New("consumer: Register called twice for consumer " + name)
	}
	c.registry[name] = consumer

	return nil
}

// Finalize tells to each consumer that no more data will be sent and calls the
// final action for each consumer.
func (c *Consumers) Finalize(duration time.Duration) {
	for _, consumer := range c.registry {
		if consumer.Loaded() {
			consumer.Finalize(duration)
		}
	}
}
