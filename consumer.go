package benchmark

// Consumer is an interface that represnets the ability to consume metrics and do
// some stuff with them
type Consumer interface {
	Run(chan *Metric, int, int)
	Finalize()
}

// regConsumers contains the available consumers
var regConsumers = make(map[string]Consumer)

// Register allows registering new consumers with an unique name
func Register(name string, consumer Consumer) {
	if consumer == nil {
		panic("consumer: Register consumer is nil")
	}
	if _, dup := regConsumers[name]; dup {
		panic("consumer: Register called twice for consumer " + name)
	}
	regConsumers[name] = consumer
}
