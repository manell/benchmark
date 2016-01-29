package benchmark

import (
	"net/http"
	"testing"
)

func TestNewMeasure(t *testing.T) {
	cs := NewCollectStats(10)

	tests := []struct {
		name   string
		method string
		host   string
		path   string
	}{
		{"test", "GET", "var:8080", "/foo"},
		{"test", "POST", "var", "/var/foo"},
		{"test", "DELETE", "var:9090", ""},
	}

	for _, test := range tests {

		req, err := http.NewRequest(test.method, "http://"+test.host+test.path, nil)
		if err != nil {
			t.Fatal(err)
		}

		done := cs.NewMeasure(test.name, req)
		done()

		res := <-cs.output
		if !res.FinalTime.After(res.StartTime) {
			t.Fatal("Invalid time")
		}

		if res.Name != test.name {
			t.Fatal("Invalid name")
		}

		if res.Operation.Host != test.host {
			t.Fatal("Invalid host")
		}

		if res.Operation.Method != test.method {
			t.Fatal("Invalid method")
		}

		if res.Operation.Path != test.path {
			t.Fatal("Invalid path")
		}
	}
}
