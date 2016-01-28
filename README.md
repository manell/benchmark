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
package main

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





