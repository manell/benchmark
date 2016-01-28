# Benchmark 
Benchmark allows you to easily write customizable benchmarks and load test.

## Why not simply use ab or a similar tool?
 Simulating multiple users login or dynamically generate HTTP requests using a certain logic are examples of what is not possible 
to do using ab or any similar load generator. The aim of this project is not to try to replace ab, but is intended to provide a solution for those cases where ab is not 
enough.

## Usage
Rather than sending and measuring HTTP requests, this project executes several iterations of what we call a Flow. A Flow can be a single
piece of code that sends a single HTTP requests or a complex code which sends several requests to multiple servers. To be valid, a Flow
must implement the FlowRunner interface:
```go
type FlowRunner interface {
	RunFlow(*BenchClient) error
}
```

The following code is the minimum what we need to create a valid Flow:
```go
import (
	"github.com/manell/benchmark"
)

type Flow struct{}

func (f *Flow) RunFlow(cli *benchmark.BenchClient) error { return nil }
```
This Flow, however, is very simple and it does nothing, so we need to code some logic to do an HTTP request. 
The following code is a Flow that does GET requests to a server.

```go
func (f *Flow) RunFlow(cli *benchmark.BenchClient) error {
	req, err := http.NewRequest("GET", "http://127.0.0.1:8081/", nil)
	if err != nil {
		return err
	}

	resp, err := cli.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	ioutil.ReadAll(resp.Body)

	return nil
}
```

The code we wrote it does not differ so much from the code we could write outside this package. The only difference is that 
we are using a custom ``` http.Client``` rather that initializing our one.

Now is time to execute the load test, so what we need to do is create a main function and put all the code together:
```go
package main

import (
	"io/ioutil"
	"net/http"

	"github.com/manell/benchmark"
)

type Flow struct{}

func (f *Flow) RunFlow(cli *benchmark.BenchClient) error {
	req, err := http.NewRequest("GET", "http://127.0.0.1:8081/", nil)
	if err != nil {
		return err
	}

	resp, err := cli.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ioutil.ReadAll(resp.Body)

	return nil
}

func main() {
	benchmark.NewBenchmark().Run(&Flow{})
}

```

Now we only need to execute the file using the flags ```-n``` to specify the number of iterations and
```-c ``` to specify the number of concurrent workers:

```$ go run main.go -n 1000 -c 8```

## Capturing metrics

The project also provides a way to capture metrics from the execution flow. To do so, we just need to use the  helper function provided by ```benchmark.BenchClient````inside our Flow.

```go
saveMetric := cli.Collector.NewMeasure("My first test", nil)
// execute here the code you want to measure
saveMetric()
```

We can save as many metrics a we want inside a Flow, however, each measure needs to be created with a distict name or the results will be incorrect.
The function also supports passing the requests as an argument to get better readability in the final report.

```go
req, err := http.NewRequest("GET", "http://127.0.0.1:8081/", nil)
if err != nil {
	return err
}

saveMetric := cli.Collector.NewMeasure("My first test", req)
resp, err := cli.Client.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()
saveMetric()
```
## Showing the results

If we want to process the metrics and show the results, we need to import a report module. 
```go
import _ "github.com/manell/benchmark/consumers/report"
```
Then we need to add an additional extra flag ```-report```, so the command will be:

```$ go run main.go -n 1000 -c 8 - report```

The final result will be something similiar to the Apache benchmark ouput:
```
Name: My first test
Method: GET
Host: 127.0.0.1:8081
Path: /

Concurrency Level: 8
Time taken for tests 0.036433 seconds
Complete iterations: 1000
Requests per second: 27447.587870 [#/sec] (mean)
Time per request: 0.261008 [ms] (mean)
Time per request: 0.032626 [ms] (mean across all concurrent requests)

Fastest request: 0.083549 [ms]
Slowest request: 2.065831 [ms]

Percentage of the requests served within a certain time (ms)
50%  0.247523
66%  0.285260
75%  0.330560
80%  0.370794
90%  0.408942
95%  0.442869
98%  0.550662
99%  0.742551
100%  2.065831
```




