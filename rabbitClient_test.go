package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func expect(t *testing.T, got interface{}, expected interface{}) {
	if got != expected {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", expected, reflect.TypeOf(expected), got, reflect.TypeOf(got))
	}
}

func createTestserver(result int, answer string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(result)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, answer)
	}))
}

func TestOverview(t *testing.T) {
	// Test server that always responds with 200 code, and specific payload
	server := createTestserver(200, `{"nonFloat":"bob@example.com","float1":1.23456789101112,"number":2}`)
	defer server.Close()

	config := &rabbitExporterConfig{
		RABBIT_URL: server.URL,
	}

	overview := getOverviewMap(*config)

	expect(t, len(overview), 2)
	expect(t, overview["float1"], 1.23456789101112)
	expect(t, overview["number"], 2.0)

	//Unknown error Server
	errorServer := createTestserver(500, http.StatusText(500))
	defer errorServer.Close()

	config = &rabbitExporterConfig{
		RABBIT_URL: errorServer.URL,
	}

	overview = getOverviewMap(*config)

	expect(t, len(overview), 0)
}

func TestQueues(t *testing.T) {
	// Test server that always responds with 200 code, and specific payload
	server := createTestserver(200, `[{"name":"Queue1","nonFloat":"bob@example.com","float1":1.23456789101112,"number":2},{"name":"Queue2","nonFloat":"bob@example.com","float1":3.23456789101112,"number":3}]`)
	defer server.Close()

	config := &rabbitExporterConfig{
		RABBIT_URL: server.URL,
	}

	queues := getQueueMap(*config)
	expect(t, len(queues), 2)
	expect(t, len(queues["Queue1"]), 2)
	expect(t, len(queues["Queue2"]), 2)
	expect(t, queues["Queue1"]["float1"], 1.23456789101112)
	expect(t, queues["Queue2"]["float1"], 3.23456789101112)
	expect(t, queues["Queue1"]["number"], 2.0)
	expect(t, queues["Queue2"]["number"], 3.0)

	//Unknown error Server
	errorServer := createTestserver(500, http.StatusText(500))
	defer errorServer.Close()

	config = &rabbitExporterConfig{
		RABBIT_URL: errorServer.URL,
	}

	queues = getQueueMap(*config)

	expect(t, len(queues), 0)
}
