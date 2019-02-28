package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	overviewTestData      = `{"management_version":"3.5.1","rates_mode":"basic","exchange_types":[{"name":"topic","description":"AMQP topic exchange, as per the AMQP specification","enabled":true},{"name":"fanout","description":"AMQP fanout exchange, as per the AMQP specification","enabled":true},{"name":"direct","description":"AMQP direct exchange, as per the AMQP specification","enabled":true},{"name":"headers","description":"AMQP headers exchange, as per the AMQP specification","enabled":true}],"rabbitmq_version":"3.5.1","cluster_name":"my-rabbit@ae74c041248b","erlang_version":"17.5","erlang_full_version":"Erlang/OTP 17 [erts-6.4] [source] [64-bit] [smp:2:2] [async-threads:30] [kernel-poll:true]","message_stats":{},"queue_totals":{"messages":48,"messages_details":{"rate":0.0},"messages_ready":48,"messages_ready_details":{"rate":0.0},"messages_unacknowledged":0,"messages_unacknowledged_details":{"rate":0.0}},"object_totals":{"consumers":0,"queues":4,"exchanges":8,"connections":0,"channels":0},"statistics_db_event_queue":0,"node":"my-rabbit@ae74c041248b","statistics_db_node":"my-rabbit@ae74c041248b","listeners":[{"node":"my-rabbit@ae74c041248b","protocol":"amqp","ip_address":"::","port":5672},{"node":"my-rabbit@ae74c041248b","protocol":"clustering","ip_address":"::","port":25672}],"contexts":[{"node":"my-rabbit@ae74c041248b","description":"RabbitMQ Management","path":"/","port":"15672"}]}`
	queuesTestData        = `[{"memory":16056,"message_stats":{"disk_writes":6,"disk_writes_details":{"rate":0.4},"publish":6,"publish_details":{"rate":0.4}},"messages":6,"messages_details":{"rate":0.4},"messages_ready":6,"messages_ready_details":{"rate":0.4},"messages_unacknowledged":0,"messages_unacknowledged_details":{"rate":0.0},"idle_since":"2015-07-07 19:02:19","consumer_utilisation":"","policy":"","exclusive_consumer_tag":"","consumers":0,"recoverable_slaves":"","state":"flow","messages_ram":6,"messages_ready_ram":6,"messages_unacknowledged_ram":0,"messages_persistent":6,"message_bytes":30,"message_bytes_ready":30,"message_bytes_unacknowledged":0,"message_bytes_ram":30,"message_bytes_persistent":30,"disk_reads":0,"disk_writes":6,"backing_queue_status":{"q1":0,"q2":0,"delta":["delta","undefined",0,"undefined"],"q3":0,"q4":6,"len":6,"target_ram_count":"infinity","next_seq_id":6,"avg_ingress_rate":0.007658533940556533,"avg_egress_rate":0.0,"avg_ack_ingress_rate":0.0,"avg_ack_egress_rate":0.0},"name":"myQueue1","vhost":"/","durable":true,"auto_delete":false,"arguments":{},"node":"my-rabbit@ae74c041248b"},{"memory":55344,"message_stats":{"disk_reads":25,"disk_reads_details":{"rate":0.0}},"messages":25,"messages_details":{"rate":0.0},"messages_ready":25,"messages_ready_details":{"rate":0.0},"messages_unacknowledged":0,"messages_unacknowledged_details":{"rate":0.0},"idle_since":"2015-07-07 18:57:52","consumer_utilisation":"","policy":"ha-2","exclusive_consumer_tag":"","consumers":0,"recoverable_slaves":"","state":"running","messages_ram":25,"messages_ready_ram":25,"messages_unacknowledged_ram":0,"messages_persistent":25,"message_bytes":75,"message_bytes_ready":75,"message_bytes_unacknowledged":0,"message_bytes_ram":75,"message_bytes_persistent":75,"disk_reads":25,"disk_writes":0,"backing_queue_status":{"q1":0,"q2":0,"delta":["delta","undefined",0,"undefined"],"q3":24,"q4":1,"len":25,"target_ram_count":"infinity","next_seq_id":16384,"avg_ingress_rate":0.0,"avg_egress_rate":0.0,"avg_ack_ingress_rate":0.0,"avg_ack_egress_rate":0.0},"name":"myQueue2","vhost":"/","durable":true,"auto_delete":false,"arguments":{},"node":"my-rabbit@ae74c041248b"},{"memory":34648,"message_stats":{"disk_reads":23,"disk_reads_details":{"rate":0.0}},"messages":23,"messages_details":{"rate":0.0},"messages_ready":23,"messages_ready_details":{"rate":0.0},"messages_unacknowledged":0,"messages_unacknowledged_details":{"rate":0.0},"idle_since":"2015-07-07 18:57:52","consumer_utilisation":"","policy":"","exclusive_consumer_tag":"","consumers":0,"recoverable_slaves":"","state":"running","messages_ram":23,"messages_ready_ram":23,"messages_unacknowledged_ram":0,"messages_persistent":23,"message_bytes":207,"message_bytes_ready":207,"message_bytes_unacknowledged":0,"message_bytes_ram":207,"message_bytes_persistent":207,"disk_reads":23,"disk_writes":0,"backing_queue_status":{"q1":0,"q2":0,"delta":["delta","undefined",0,"undefined"],"q3":22,"q4":1,"len":23,"target_ram_count":"infinity","next_seq_id":16384,"avg_ingress_rate":0.0,"avg_egress_rate":0.0,"avg_ack_ingress_rate":0.0,"avg_ack_egress_rate":0.0},"name":"myQueue3","vhost":"/","durable":true,"auto_delete":false,"arguments":{},"node":"my-rabbit@ae74c041248b"},{"memory":13912,"messages":0,"messages_details":{"rate":0.0},"messages_ready":0,"messages_ready_details":{"rate":0.0},"messages_unacknowledged":0,"messages_unacknowledged_details":{"rate":0.0},"idle_since":"2015-07-07 18:57:52","consumer_utilisation":"","policy":"","exclusive_consumer_tag":"","consumers":0,"recoverable_slaves":"","state":"running","messages_ram":0,"messages_ready_ram":0,"messages_unacknowledged_ram":0,"messages_persistent":0,"message_bytes":0,"message_bytes_ready":0,"message_bytes_unacknowledged":0,"message_bytes_ram":0,"message_bytes_persistent":0,"disk_reads":0,"disk_writes":0,"backing_queue_status":{"q1":0,"q2":0,"delta":["delta","undefined",0,"undefined"],"q3":0,"q4":0,"len":0,"target_ram_count":"infinity","next_seq_id":0,"avg_ingress_rate":0.0,"avg_egress_rate":0.0,"avg_ack_ingress_rate":0.0,"avg_ack_egress_rate":0.0},"name":"myQueue4","vhost":"vhost4","durable":true,"auto_delete":false,"arguments":{},"node":"my-rabbit@ae74c041248b"}]`
	exchangeAPIResponse   = `[{"name":"","vhost":"/","type":"direct","durable":true,"auto_delete":false,"internal":false,"arguments":{}},{"name":"amq.direct","vhost":"/","type":"direct","durable":true,"auto_delete":false,"internal":false,"arguments":{}},{"name":"amq.fanout","vhost":"/","type":"fanout","durable":true,"auto_delete":false,"internal":false,"arguments":{}},{"name":"amq.headers","vhost":"/","type":"headers","durable":true,"auto_delete":false,"internal":false,"arguments":{}},{"name":"amq.match","vhost":"/","type":"headers","durable":true,"auto_delete":false,"internal":false,"arguments":{}},{"name":"amq.rabbitmq.log","vhost":"/","type":"topic","durable":true,"auto_delete":false,"internal":true,"arguments":{}},{"name":"amq.rabbitmq.trace","vhost":"/","type":"topic","durable":true,"auto_delete":false,"internal":true,"arguments":{}},{"name":"amq.topic","vhost":"/","type":"topic","durable":true,"auto_delete":false,"internal":false,"arguments":{}},{"message_stats":{"publish":0,"publish_details":{"rate":0.0},"publish_in":5,"publish_in_details":{"rate":0.0},"publish_out":0,"publish_out_details":{"rate":0.0},"ack":0,"ack_details":{"rate":0.0},"deliver_get":0,"deliver_get_details":{"rate":0.0},"confirm":5,"confirm_details":{"rate":0.0},"return_unroutable":5,"return_unroutable_details":{"rate":0.0},"redeliver":0,"redeliver_details":{"rate":0.0}},"name":"myExchange","vhost":"/","type":"fanout","durable":true,"auto_delete":false,"internal":false,"arguments":{}}]`
	nodesAPIResponse      = `[{"mem_used":150456032,"mem_used_details":{"rate":25176},"fd_used":55,"fd_used_details":{"rate":0},"sockets_used":0,"sockets_used_details":{"rate":0},"proc_used":226,"proc_used_details":{"rate":0},"disk_free":189045161984,"disk_free_details":{"rate":0},"partitions":["rabbit@ribbit-0","rabbit@ribbit-1","rabbit@ribbit-3","rabbit@ribbit-4"],"os_pid":"113","fd_total":1048576,"sockets_total":943626,"mem_limit":838395494,"mem_alarm":false,"disk_free_limit":50000000,"disk_free_alarm":false,"proc_total":1048576,"rates_mode":"basic","uptime":3772165,"run_queue":0,"processors":4,"name":"my-rabbit@5a00cd8fe2f4","type":"disc","running":true}]`
	connectionAPIResponse = `[{"auth_mechanism": "PLAIN","channel_max": 65535,"channels": 1,"client_properties": {"copyright": "Copyright (c) 2007-2014 VMWare Inc, Tony Garnock-Jones, and Alan Antonuk.","information": "See https://github.com/alanxz/rabbitmq-c","platform": "linux-gn","product": "rabbitmq-c","version": "0.5.3-pre"},"connected_at": 1501868641834,"frame_max": 131072,"garbage_collection": {"fullsweep_after": 65535,"min_bin_vheap_size": 46422,"min_heap_size": 233,"minor_gcs": 3},"host": "172.31.15.10","name": "172.31.0.130:32769 -> 172.31.15.10:5672","node": "my-rabbit@ae74c041248b","peer_cert_issuer": null,"peer_cert_subject": null,"peer_cert_validity": null,"peer_host": "172.31.0.130","peer_port": 32769,"port": 5672,"protocol": "AMQP 0-9-1","recv_cnt": 22708,"recv_oct": 8905713,"recv_oct_details": {"rate": 169.6},"reductions": 6257210,"reductions_details": {"rate": 148.8},"send_cnt": 6,"send_oct": 573,"send_oct_details": {"rate": 0.0},"send_pend": 0,"ssl": false,"ssl_cipher": null,"ssl_hash": null,"ssl_key_exchange": null,"ssl_protocol": null,"state": "running","timeout": 0,"type": "network","user": "rmq_oms","vhost": "/"},{"auth_mechanism": "PLAIN","channel_max": 65535,"channels": 1,"client_properties": {"copyright": "Copyright (c) 2007-2014 VMWare Inc, Tony Garnock-Jones, and Alan Antonuk.","information": "See https://github.com/alanxz/rabbitmq-c","platform": "linux-gn","product": "rabbitmq-c","version": "0.5.3-pre"},"connected_at": 1501868641834,"frame_max": 131072,"garbage_collection": {"fullsweep_after": 65535,"min_bin_vheap_size": 46422,"min_heap_size": 233,"minor_gcs": 3},"host": "172.31.15.10","name": "172.31.0.130:32769 -> 172.31.15.10:5672","node": "rabbit@rmq-cluster-node-04","peer_cert_issuer": null,"peer_cert_subject": null,"peer_cert_validity": null,"peer_host": "172.31.0.130","peer_port": 32769,"port": 5672,"protocol": "AMQP 0-9-1","recv_cnt": 22708,"recv_oct": 8905713,"recv_oct_details": {"rate": 169.6},"reductions": 6257210,"reductions_details": {"rate": 148.8},"send_cnt": 6,"send_oct": 573,"send_oct_details": {"rate": 0.0},"send_pend": 0,"ssl": false,"ssl_cipher": null,"ssl_hash": null,"ssl_key_exchange": null,"ssl_protocol": null,"state": "running","timeout": 0,"type": "network","user": "rmq_oms","vhost": "/"}]`
)

func expectSubstring(t *testing.T, body string, substr string) {
	t.Helper()
	if !strings.Contains(body, substr) {
		t.Errorf("Substring expected but not found. Substr=%v", substr)
	}
}

func dontExpectSubstring(t *testing.T, body string, substr string) {
	t.Helper()
	if strings.Contains(body, substr) {
		t.Errorf("Substring not expected but found. Substr=%v", substr)
	}
}

func setupServer(t *testing.T, overview, queues, exchange, nodes, connections string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if r.RequestURI == "/api/overview" {
			fmt.Fprintln(w, overview)
		} else if r.RequestURI == "/api/queues" {
			fmt.Fprintln(w, queues)
		} else if r.RequestURI == "/api/exchanges" {
			fmt.Fprintln(w, exchange)
		} else if r.RequestURI == "/api/nodes" {
			fmt.Fprintln(w, nodes)
		} else if r.RequestURI == "/api/connections" {
			fmt.Fprintln(w, connections)
		} else {
			t.Errorf("Invalid request. URI=%v", r.RequestURI)
			fmt.Fprintf(w, "Invalid request. URI=%v", r.RequestURI)
		}

	}))
	return server
}

func TestWholeApp(t *testing.T) {
	server := setupServer(t, overviewTestData, queuesTestData, exchangeAPIResponse, nodesAPIResponse, connectionAPIResponse)
	defer server.Close()

	os.Setenv("RABBIT_URL", server.URL)
	defer os.Unsetenv("RABBIT_URL")
	os.Setenv("SKIP_QUEUES", "^.*3$")
	defer os.Unsetenv("SKIP_QUEUES")
	os.Setenv("SKIP_QUEUES", "^.*3$")
	defer os.Unsetenv("SKIP_QUEUES")
	os.Setenv("RABBIT_CAPABILITIES", " ")
	defer os.Unsetenv("RABBIT_CAPABILITIES")
	initConfig()

	exporter := newExporter()
	prometheus.MustRegister(exporter)
	defer prometheus.Unregister(exporter)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	prometheus.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Home page didn't return %v", http.StatusOK)
	}
	body := w.Body.String()
	//t.Log(body)
	expectSubstring(t, body, `rabbitmq_up 1`)

	// overview
	expectSubstring(t, body, `rabbitmq_exchanges 8`)
	expectSubstring(t, body, `rabbitmq_queues 4`)
	expectSubstring(t, body, `rabbitmq_queue_messages_global 48`)
	expectSubstring(t, body, `rabbitmq_queue_messages_ready_global 48`)
	expectSubstring(t, body, `rabbitmq_queue_messages_unacknowledged_global 0`)

	// node
	expectSubstring(t, body, `rabbitmq_running{node="my-rabbit@5a00cd8fe2f4",self="0"} 1`)
	expectSubstring(t, body, `rabbitmq_partitions{node="my-rabbit@5a00cd8fe2f4",self="0"} 4`)

	// queue
	expectSubstring(t, body, `rabbitmq_queue_messages_ready{durable="true",policy="ha-2",queue="myQueue2",self="1",vhost="/"} 25`)
	expectSubstring(t, body, `rabbitmq_queue_memory{durable="true",policy="",queue="myQueue4",self="1",vhost="vhost4"} 13912`)
	expectSubstring(t, body, `rabbitmq_queue_messages_published_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 6`)
	expectSubstring(t, body, `rabbitmq_queue_disk_writes_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 6`)
	expectSubstring(t, body, `rabbitmq_queue_messages_delivered_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 0`)
	// exchange
	expectSubstring(t, body, `rabbitmq_exchange_messages_published_in_total{exchange="myExchange",vhost="/"} 5`)
	// connection
	dontExpectSubstring(t, body, `rabbitmq_connection_channels{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="1",user="rmq_oms",vhost="/"} 2`)
	dontExpectSubstring(t, body, `rabbitmq_connection_received_packets{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="1",user="rmq_oms",vhost="/"} 45416`)
}

func TestWholeAppInverted(t *testing.T) {
	server := setupServer(t, overviewTestData, queuesTestData, exchangeAPIResponse, nodesAPIResponse, connectionAPIResponse)
	defer server.Close()

	os.Setenv("RABBIT_URL", server.URL)
	os.Setenv("SKIP_QUEUES", "^.*3$")
	defer os.Unsetenv("SKIP_QUEUES")
	os.Setenv("RABBIT_CAPABILITIES", " ")
	defer os.Unsetenv("RABBIT_CAPABILITIES")
	initConfig()
	config.EnabledExporters = []string{"connections"}

	exporter := newExporter()
	prometheus.MustRegister(exporter)
	defer prometheus.Unregister(exporter)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	prometheus.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Home page didn't return %v", http.StatusOK)
	}
	body := w.Body.String()
	t.Log(body)
	expectSubstring(t, body, `rabbitmq_up 1`)

	// overview is always scraped and exported
	expectSubstring(t, body, `rabbitmq_exchanges 8`)
	expectSubstring(t, body, `rabbitmq_queues 4`)
	expectSubstring(t, body, `rabbitmq_queue_messages_global 48`)
	expectSubstring(t, body, `rabbitmq_queue_messages_ready_global 48`)
	expectSubstring(t, body, `rabbitmq_queue_messages_unacknowledged_global 0`)

	// node
	dontExpectSubstring(t, body, `rabbitmq_running{node="my-rabbit@5a00cd8fe2f4"} 1`)
	dontExpectSubstring(t, body, `rabbitmq_partitions{node="my-rabbit@5a00cd8fe2f4"} 4`)

	// queue
	dontExpectSubstring(t, body, `rabbitmq_queue_messages_ready{durable="true",policy="ha-2",queue="myQueue2",self="1",vhost="/"} 25`)
	dontExpectSubstring(t, body, `rabbitmq_queue_memory{durable="true",policy="",queue="myQueue4",self="1",vhost="vhost4"} 13912`)
	dontExpectSubstring(t, body, `rabbitmq_queue_messages_published_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 6`)
	dontExpectSubstring(t, body, `rabbitmq_queue_disk_writes_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 6`)
	dontExpectSubstring(t, body, `rabbitmq_queue_messages_delivered_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 0`)
	// exchange
	dontExpectSubstring(t, body, `rabbitmq_exchange_messages_published_in_total{exchange="myExchange",vhost="/"} 5`)
	// connection
	expectSubstring(t, body, `rabbitmq_connection_channels{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="0",user="rmq_oms",vhost="/"} 1`)
	expectSubstring(t, body, `rabbitmq_connection_received_packets{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="0",user="rmq_oms",vhost="/"} 22708`)
}

func TestAppMaxQueues(t *testing.T) {
	server := setupServer(t, overviewTestData, queuesTestData, exchangeAPIResponse, nodesAPIResponse, connectionAPIResponse)
	defer server.Close()

	os.Setenv("RABBIT_URL", server.URL)
	os.Setenv("SKIP_QUEUES", "^.*3$")
	defer os.Unsetenv("SKIP_QUEUES")
	os.Setenv("MAX_QUEUES", "3")
	defer os.Unsetenv("MAX_QUEUES")
	os.Setenv("RABBIT_CAPABILITIES", " ")
	defer os.Unsetenv("RABBIT_CAPABILITIES")
	initConfig()

	exporter := newExporter()
	prometheus.MustRegister(exporter)
	defer prometheus.Unregister(exporter)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	prometheus.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Home page didn't return %v", http.StatusOK)
	}
	body := w.Body.String()

	expectSubstring(t, body, `rabbitmq_up 1`)

	// overview
	expectSubstring(t, body, `rabbitmq_exchanges 8`)
	expectSubstring(t, body, `rabbitmq_queues 4`)
	expectSubstring(t, body, `rabbitmq_queue_messages_global 48`)
	expectSubstring(t, body, `rabbitmq_queue_messages_ready_global 48`)
	expectSubstring(t, body, `rabbitmq_queue_messages_unacknowledged_global 0`)

	// node
	expectSubstring(t, body, `rabbitmq_running{node="my-rabbit@5a00cd8fe2f4",self="0"} 1`)
	expectSubstring(t, body, `rabbitmq_partitions{node="my-rabbit@5a00cd8fe2f4",self="0"} 4`)

	// queue
	dontExpectSubstring(t, body, `rabbitmq_queue_messages_ready{durable="true",policy="ha-2",queue="myQueue2",self="1",vhost="/"} 25`)
	dontExpectSubstring(t, body, `rabbitmq_queue_memory{durable="true",policy="",queue="myQueue4",self="1",vhost="vhost4"} 13912`)
	dontExpectSubstring(t, body, `rabbitmq_queue_messages_published_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 6`)
	dontExpectSubstring(t, body, `rabbitmq_queue_disk_writes_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 6`)
	dontExpectSubstring(t, body, `rabbitmq_queue_messages_delivered_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 0`)

	// exchange
	expectSubstring(t, body, `rabbitmq_exchange_messages_published_in_total{exchange="myExchange",vhost="/"} 5`)

	// connection
	dontExpectSubstring(t, body, `rabbitmq_connection_channels{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="1",user="rmq_oms",vhost="/"} 2`)
	dontExpectSubstring(t, body, `rabbitmq_connection_received_packets{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="1",user="rmq_oms",vhost="/"} 45416`)
}

func TestRabbitError(t *testing.T) {
	server := createTestserver(500, http.StatusText(500))
	defer server.Close()
	os.Setenv("RABBIT_URL", server.URL)
	initConfig()

	exporter := newExporter()
	prometheus.MustRegister(exporter)
	defer prometheus.Unregister(exporter)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	prometheus.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Home page didn't return %v", http.StatusOK)
	}
	body := w.Body.String()
	// fmt.Println(body)
	expectSubstring(t, body, `rabbitmq_up 0`)
	if strings.Contains(body, "rabbitmq_channelsTotal") {
		t.Errorf("Metric 'rabbitmq_channelsTotal' unexpected as the server  did not respond")
	}
}

//TestResetMetricsOnRabbitFailure verifies the behaviour of the exporter if the rabbitmq fails after one successfull retrieval of the data
// List of metrics should be empty except rabbitmq_up should be 0
func TestResetMetricsOnRabbitFailure(t *testing.T) {
	rabbitUP := true
	rabbitQueuesUp := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rabbitUP {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, http.StatusText(http.StatusInternalServerError))
			return
		}
		if !rabbitQueuesUp && r.RequestURI == "/api/queues" {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, http.StatusText(http.StatusInternalServerError))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if r.RequestURI == "/api/overview" {
			fmt.Fprintln(w, overviewTestData)
		} else if r.RequestURI == "/api/queues" {
			fmt.Fprintln(w, queuesTestData)
		} else if r.RequestURI == "/api/exchanges" {
			fmt.Fprintln(w, exchangeAPIResponse)
		} else if r.RequestURI == "/api/nodes" {
			fmt.Fprintln(w, nodesAPIResponse)
		} else if r.RequestURI == "/api/connections" {
			fmt.Fprintln(w, connectionAPIResponse)
		} else {
			t.Errorf("Invalid request. URI=%v", r.RequestURI)
			fmt.Fprintf(w, "Invalid request. URI=%v", r.RequestURI)
		}

	}))
	defer server.Close()
	os.Setenv("RABBIT_URL", server.URL)
	os.Setenv("RABBIT_CAPABILITIES", " ")
	defer os.Unsetenv("RABBIT_CAPABILITIES")
	os.Setenv("RABBIT_EXPORTERS", "exchange,node,overview,queue,connections")
	defer os.Unsetenv("RABBIT_EXPORTERS")

	initConfig()

	exporter := newExporter()
	prometheus.MustRegister(exporter)
	defer prometheus.Unregister(exporter)

	t.Run("RabbitMQ is up -> all metrics are ok", func(t *testing.T) {
		rabbitUP = true
		rabbitQueuesUp = true
		req, _ := http.NewRequest("GET", "", nil)
		w := httptest.NewRecorder()
		prometheus.Handler().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Home page didn't return %v", http.StatusOK)
		}
		body := w.Body.String()

		expectSubstring(t, body, `rabbitmq_up 1`)
		expectSubstring(t, body, `rabbitmq_module_up{module="exchange"} 1`)
		expectSubstring(t, body, `rabbitmq_module_up{module="node"} 1`)
		expectSubstring(t, body, `rabbitmq_module_up{module="overview"} 1`)
		expectSubstring(t, body, `rabbitmq_module_up{module="queue"} 1`)
		expectSubstring(t, body, `rabbitmq_module_up{module="connections"} 1`)

		// overview
		expectSubstring(t, body, `rabbitmq_exchanges 8`)
		expectSubstring(t, body, `rabbitmq_queues 4`)
		expectSubstring(t, body, `rabbitmq_queue_messages_global 48`)
		expectSubstring(t, body, `rabbitmq_queue_messages_ready_global 48`)
		expectSubstring(t, body, `rabbitmq_queue_messages_unacknowledged_global 0`)

		// node
		expectSubstring(t, body, `rabbitmq_running{node="my-rabbit@5a00cd8fe2f4",self="0"} 1`)
		expectSubstring(t, body, `rabbitmq_partitions{node="my-rabbit@5a00cd8fe2f4",self="0"} 4`)

		// queue
		expectSubstring(t, body, `rabbitmq_queue_messages_ready{durable="true",policy="ha-2",queue="myQueue2",self="1",vhost="/"} 25`)
		expectSubstring(t, body, `rabbitmq_queue_memory{durable="true",policy="",queue="myQueue4",self="1",vhost="vhost4"} 13912`)
		expectSubstring(t, body, `rabbitmq_queue_messages_published_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 6`)
		expectSubstring(t, body, `rabbitmq_queue_disk_writes_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 6`)
		expectSubstring(t, body, `rabbitmq_queue_messages_delivered_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"} 0`)

		// exchange
		expectSubstring(t, body, `rabbitmq_exchange_messages_published_in_total{exchange="myExchange",vhost="/"} 5`)

		// connection
		expectSubstring(t, body, `rabbitmq_connection_channels{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="0",user="rmq_oms",vhost="/"} 1`)
		expectSubstring(t, body, `rabbitmq_connection_received_packets{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="0",user="rmq_oms",vhost="/"} 22708`)
	})

	t.Run("Rabbit Queue Endpoint down -> 'rabbitmq_up 0' and queue metrics are missing", func(t *testing.T) {
		rabbitQueuesUp = false
		req, _ := http.NewRequest("GET", "", nil)
		w := httptest.NewRecorder()
		prometheus.Handler().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Home page didn't return %v", http.StatusOK)
		}
		body := w.Body.String()

		expectSubstring(t, body, `rabbitmq_up 0`)
		expectSubstring(t, body, `rabbitmq_module_up{module="exchange"} 1`)
		expectSubstring(t, body, `rabbitmq_module_up{module="node"} 1`)
		expectSubstring(t, body, `rabbitmq_module_up{module="overview"} 1`)
		expectSubstring(t, body, `rabbitmq_module_up{module="queue"} 0`) //down
		expectSubstring(t, body, `rabbitmq_module_up{module="connections"} 1`)

		// overview
		expectSubstring(t, body, `rabbitmq_exchanges 8`)
		expectSubstring(t, body, `rabbitmq_queues 4`)
		expectSubstring(t, body, `rabbitmq_queue_messages_global 48`)
		expectSubstring(t, body, `rabbitmq_queue_messages_ready_global 48`)
		expectSubstring(t, body, `rabbitmq_queue_messages_unacknowledged_global 0`)

		// node
		expectSubstring(t, body, `rabbitmq_running{node="my-rabbit@5a00cd8fe2f4",self="0"} 1`)
		expectSubstring(t, body, `rabbitmq_partitions{node="my-rabbit@5a00cd8fe2f4",self="0"} 4`)

		// queue
		dontExpectSubstring(t, body, `rabbitmq_queue_messages_ready{durable="true",policy="ha-2",queue="myQueue2",self="1",vhost="/"}`)
		dontExpectSubstring(t, body, `rabbitmq_queue_memory{durable="true",policy="",queue="myQueue4",self="1",vhost="vhost4"}`)
		dontExpectSubstring(t, body, `rabbitmq_queue_messages_published_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"}`)
		dontExpectSubstring(t, body, `rabbitmq_queue_disk_writes_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"}`)
		dontExpectSubstring(t, body, `rabbitmq_queue_messages_delivered_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"}`)

		// exchange
		expectSubstring(t, body, `rabbitmq_exchange_messages_published_in_total{exchange="myExchange",vhost="/"} 5`)

		// connection
		expectSubstring(t, body, `rabbitmq_connection_channels{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="0",user="rmq_oms",vhost="/"} 1`)
		expectSubstring(t, body, `rabbitmq_connection_received_packets{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="0",user="rmq_oms",vhost="/"} 22708`)
	})

	t.Run("RabbitMQ is down -> all metrics are missing except 'rabbitmq_up 0'", func(t *testing.T) {
		rabbitUP = false
		req, _ := http.NewRequest("GET", "", nil)
		w := httptest.NewRecorder()
		prometheus.Handler().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Home page didn't return %v", http.StatusOK)
		}
		body := w.Body.String()

		expectSubstring(t, body, `rabbitmq_up 0`)
		expectSubstring(t, body, `rabbitmq_module_up{module="exchange"} 0`)
		expectSubstring(t, body, `rabbitmq_module_up{module="node"} 0`)
		expectSubstring(t, body, `rabbitmq_module_up{module="overview"} 0`)
		expectSubstring(t, body, `rabbitmq_module_up{module="queue"} 0`)
		expectSubstring(t, body, `rabbitmq_module_up{module="connections"} 0`)

		// overview
		dontExpectSubstring(t, body, `rabbitmq_exchangesTotal`)
		dontExpectSubstring(t, body, `rabbitmq_queuesTotal`)
		dontExpectSubstring(t, body, `rabbitmq_queue_messages_global`)
		dontExpectSubstring(t, body, `rabbitmq_queue_messages_ready_global`)
		dontExpectSubstring(t, body, `rabbitmq_queue_messages_unacknowledged_global`)

		// node
		dontExpectSubstring(t, body, `rabbitmq_running{node="my-rabbit@5a00cd8fe2f4",self="0"} 1`)
		dontExpectSubstring(t, body, `rabbitmq_partitions{node="my-rabbit@5a00cd8fe2f4",self="0"} 4`)

		// queue
		dontExpectSubstring(t, body, `rabbitmq_queue_messages_ready{durable="true",policy="ha-2",queue="myQueue2",self="1",vhost="/"}`)
		dontExpectSubstring(t, body, `rabbitmq_queue_memory{durable="true",policy="",queue="myQueue4",self="1",vhost="vhost4"}`)
		dontExpectSubstring(t, body, `rabbitmq_queue_messages_published_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"}`)
		dontExpectSubstring(t, body, `rabbitmq_queue_disk_writes_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"}`)
		dontExpectSubstring(t, body, `rabbitmq_queue_messages_delivered_total{durable="true",policy="",queue="myQueue1",self="1",vhost="/"}`)

		// exchange
		dontExpectSubstring(t, body, `rabbitmq_exchange_messages_published_in_total{exchange="myExchange",vhost="/"}`)

		// connection
		dontExpectSubstring(t, body, `rabbitmq_connection_channels{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="1",user="rmq_oms",vhost="/"}`)
		dontExpectSubstring(t, body, `rabbitmq_connection_received_packets{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="1",user="rmq_oms",vhost="/"}`)
	})

}

func TestQueueState(t *testing.T) {
	server := setupServer(t, overviewTestData, queuesTestData, exchangeAPIResponse, nodesAPIResponse, connectionAPIResponse)
	defer server.Close()

	os.Setenv("RABBIT_URL", server.URL)
	os.Setenv("RABBIT_CAPABILITIES", " ")
	defer os.Unsetenv("RABBIT_CAPABILITIES")
	os.Setenv("RABBIT_EXPORTERS", "queue,connections")
	defer os.Unsetenv("RABBIT_EXPORTERS")
	initConfig()

	exporter := newExporter()
	prometheus.MustRegister(exporter)
	defer prometheus.Unregister(exporter)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	prometheus.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Home page didn't return %v", http.StatusOK)
	}
	body := w.Body.String()
	t.Log(body)

	expectSubstring(t, body, `rabbitmq_up 1`)

	// queue
	expectSubstring(t, body, `rabbitmq_queue_state{durable="true",policy="",queue="myQueue1",self="1",state="flow",vhost="/"} 1`)
	expectSubstring(t, body, `rabbitmq_queue_state{durable="true",policy="",queue="myQueue3",self="1",state="running",vhost="/"} 1`)
	expectSubstring(t, body, `rabbitmq_queue_state{durable="true",policy="ha-2",queue="myQueue2",self="1",state="running",vhost="/"} 1`)

	// connections
	expectSubstring(t, body, `rabbitmq_connection_status{node="rabbit@rmq-cluster-node-04",peer_host="172.31.0.130",self="0",state="running",user="rmq_oms",vhost="/"} 1`)
	expectSubstring(t, body, `rabbitmq_connection_status{node="my-rabbit@ae74c041248b",peer_host="172.31.0.130",self="1",state="running",user="rmq_oms",vhost="/"} 1`)

}

func TestQueueLength(t *testing.T) {
	testdataFile := "testdata/queue-max-length.json"
	queuedata, err := ioutil.ReadFile(testdataFile)
	if err != nil {
		t.Fatalf("Error reading %s", testdataFile)
	}
	server := setupServer(t, "", string(queuedata), "", "", "")
	defer server.Close()

	os.Setenv("RABBIT_URL", server.URL)
	os.Setenv("RABBIT_CAPABILITIES", " ")
	defer os.Unsetenv("RABBIT_CAPABILITIES")
	os.Setenv("RABBIT_EXPORTERS", "queue")
	defer os.Unsetenv("RABBIT_EXPORTERS")
	initConfig()

	exporter := newExporter()
	prometheus.MustRegister(exporter)
	defer prometheus.Unregister(exporter)

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	prometheus.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Home page didn't return %v", http.StatusOK)
	}
	body := w.Body.String()
	t.Log(body)

	// queue
	expectSubstring(t, body, `rabbitmq_queue_max_length{durable="true",policy="",queue="QueueWithMaxLength55",self="0",vhost="/"} 55`)
	expectSubstring(t, body, `rabbitmq_queue_max_length_bytes{durable="true",policy="",queue="QueueWithMaxBytes99",self="0",vhost="/"} 99`)

}
