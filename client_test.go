package benchmark

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDo(t *testing.T) {
	coll := make(chan *Metric, 1)
	cli := NewClient(coll, false)

	response := "I'm the backend"
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(response))
	}))

	tests := []struct {
		Name   string
		Method string
		Path   string
	}{
		{"List", "GET", "/var"},
		{"Create", "POST", "/foo"},
		{"Delete", "DELETE", "/var/foo"},
	}

	for _, test := range tests {
		req, err := http.NewRequest(test.Method, backend.URL+test.Path, nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := cli.Do(test.Name, req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if string(body) != response {
			if err != nil {
				t.Fatal("Invalid response")
			}
		}

		metric := <-coll
		if metric.Operation.Method != test.Method {
			t.Fatal("Invalid metric method. Got: " + metric.Operation.Method)
		}
		if metric.Operation.Name != test.Name {
			t.Fatal("Invalid metric Name. Got: " + metric.Operation.Name)
		}
		if metric.Operation.Path != test.Path {
			t.Fatal("Invalid metric Path. Got: " + metric.Operation.Path)
		}
		if !metric.FinalTime.After(metric.StartTime) {
			t.Fatal("Invalid time")
		}
	}
}
