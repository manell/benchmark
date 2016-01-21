package logmetrics

import (
	"fmt"

	"github.com/manell/benchmark"
)

func init() {
	sync := make(chan int, 1)
	benchmark.Register("log", &LogMetric{sync})
}

type LogMetric struct {
	sync chan int
}

func (l *LogMetric) Run(collector chan *benchmark.Metric) {
	for metric := range collector {
		fmt.Println(metric)
	}
	fmt.Println("pre sync")
	l.sync <- 1
	fmt.Println("post sync")
}

func (l *LogMetric) Finalize() {
	fmt.Println("pre final")
	<-l.sync
	fmt.Println("post final")
}
