package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "rabbitmq"
)

type Exporter struct {
	mutex                         sync.RWMutex
	lastSeen                      prometheus.Counter
	connections_total             prometheus.Gauge
	channels_total                prometheus.Gauge
	queues_total                  prometheus.Gauge
	consumers_total               prometheus.Gauge
	exchanges_total               prometheus.Gauge
	messages_count                *prometheus.GaugeVec
	messages_ready_count          *prometheus.GaugeVec
	messages_unacknowledged_count *prometheus.GaugeVec
	consumers_count               *prometheus.GaugeVec
	message_bytes                 *prometheus.GaugeVec
	disk_reads_count              *prometheus.GaugeVec
	disk_writes_count             *prometheus.GaugeVec
}

var (
	log             = logrus.New()
	queueLabelNames = []string{"queue"}
)

// Listed available metrics
func newExporter() *Exporter {
	return &Exporter{
		lastSeen: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "last_seen",
			Help:      "Last time rabbitmq was seen by the exporter",
		}),
		connections_total: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "connections_total",
				Help:      "Total number of open connections.",
			}),
		channels_total: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "channels_total",
				Help:      "Total number of open channels.",
			}),
		queues_total: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "queues_total",
				Help:      "Total number of queues in use.",
			}),
		consumers_total: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "consumers_total",
				Help:      "Total number of message consumers.",
			}),
		exchanges_total: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "exchanges_total",
				Help:      "Total number of exchanges in use.",
			}),
		messages_count: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "messages_count",
			Help:      "Current length of queue.",
		},
			queueLabelNames,
		),
		messages_ready_count: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "messages_ready_count",
			Help:      "Number of messages ready to be delivered to clients.",
		},
			queueLabelNames,
		),
		messages_unacknowledged_count: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "messages_unacknowledged_count",
			Help:      "Number of messages delivered to clients but not yet acknowledged.",
		},
			queueLabelNames,
		),
		consumers_count: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "consumers_count",
			Help:      "Number of consumers subscribed to queue",
		},
			queueLabelNames,
		),
		message_bytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "message_bytes",
			Help:      "Current bytes to store queue.",
		},
			queueLabelNames,
		),
		disk_reads_count: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "disk_reads_count",
			Help:      "Total number of times messages have been read from disk by this queue since it started.",
		},
			queueLabelNames,
		),
		disk_writes_count: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "disk_writes_count",
			Help:      "Total number of times messages have been written to disk by this queue since it started.",
		},
			queueLabelNames,
		),
	}
}

type QueueMetrics struct {
	messages_count                float64
	messages_ready_count          float64
	messages_unacknowledged_count float64
	consumers_count               float64
	message_bytes                 float64
	disk_reads_count              float64
	disk_writes_count             float64
}

func unpackQueueMetrics(d *json.Decoder) map[string]*QueueMetrics {
	var output []map[string]interface{}

	if err := d.Decode(&output); err != nil {
		log.Error(err)
	}
	metrics := make(map[string]*QueueMetrics)

	for _, v := range output {
		metrics[v["name"].(string)] = &QueueMetrics{
			messages_count:                v["messages"].(float64),
			messages_ready_count:          v["messages_ready"].(float64),
			messages_unacknowledged_count: v["messages_unacknowledged"].(float64),
			consumers_count:               v["consumers"].(float64),
			message_bytes:                 v["message_bytes"].(float64),
			disk_reads_count:              v["disk_reads"].(float64),
			disk_writes_count:             v["disk_writes"].(float64),
		}
	}
	return metrics
}

func unpackOverviewMetrics(d *json.Decoder) map[string]float64 {
	var output map[string]interface{}

	if err := d.Decode(&output); err != nil {
		log.Error(err)
	}
	metrics := make(map[string]float64)

	for k, v := range output["object_totals"].(map[string]interface{}) {
		metrics[k] = v.(float64)
	}
	return metrics
}

func getMetrics(hostname, username, password, endpoint string) *json.Decoder {
	client := &http.Client{}
	req, err := http.NewRequest("GET", hostname+"/api/"+endpoint, nil)
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)

	if err != nil {
		log.Error(err)
	}
	return json.NewDecoder(resp.Body)
}

func updateMetrics(overviewMetrics map[string]float64, queueMetrics map[string]*QueueMetrics, exporter *Exporter) {

	exporter.lastSeen.Set(float64(time.Now().Unix()))

	exporter.channels_total.Set(overviewMetrics["channels"])
	exporter.connections_total.Set(overviewMetrics["connections"])
	exporter.consumers_total.Set(overviewMetrics["consumers"])
	exporter.queues_total.Set(overviewMetrics["queues"])
	exporter.exchanges_total.Set(overviewMetrics["exchanges"])

	for queue, stat := range queueMetrics {
		exporter.messages_count.WithLabelValues(queue).Set(stat.messages_count)
		exporter.messages_ready_count.WithLabelValues(queue).Set(stat.messages_ready_count)
		exporter.messages_unacknowledged_count.WithLabelValues(queue).Set(stat.messages_unacknowledged_count)
		exporter.consumers_count.WithLabelValues(queue).Set(stat.consumers_count)
		exporter.message_bytes.WithLabelValues(queue).Set(stat.message_bytes)
		exporter.disk_reads_count.WithLabelValues(queue).Set(stat.disk_reads_count)
		exporter.disk_writes_count.WithLabelValues(queue).Set(stat.disk_writes_count)
	}
}

func main() {
	log.Out = os.Stdout
	exporter := newExporter()
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", prometheus.Handler())
	log.Infof("Starting RabbitMQ exporter on port: %s.", os.Getenv("PUBLISH_PORT"))
	http.ListenAndServe(":"+os.Getenv("PUBLISH_PORT"), nil)
}

func (e *Exporter) fetchRabbit() {

	Url := os.Getenv("RABBIT_URL")
	Uname := os.Getenv("RABBIT_USER")
	Password := os.Getenv("RABBIT_PASSWORD")

	overviewDecoder := getMetrics(Url, Uname, Password, "overview")
	overviewMetrics := unpackOverviewMetrics(overviewDecoder)
	queueDecoder := getMetrics(Url, Uname, Password, "queues")
	queueMetrics := unpackQueueMetrics(queueDecoder)

	updateMetrics(overviewMetrics, queueMetrics, e)
	log.Info("Metrics updated successfully.")
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.lastSeen.Describe(ch)
	e.connections_total.Describe(ch)
	e.channels_total.Describe(ch)
	e.queues_total.Describe(ch)
	e.consumers_total.Describe(ch)
	e.exchanges_total.Describe(ch)
	e.messages_count.Describe(ch)
	e.messages_ready_count.Describe(ch)
	e.messages_unacknowledged_count.Describe(ch)
	e.consumers_count.Describe(ch)
	e.message_bytes.Describe(ch)
	e.disk_reads_count.Describe(ch)
	e.disk_writes_count.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	e.messages_count.Reset()
	e.messages_ready_count.Reset()
	e.messages_unacknowledged_count.Reset()
	e.consumers_count.Reset()
	e.message_bytes.Reset()
	e.disk_reads_count.Reset()
	e.disk_writes_count.Reset()

	e.fetchRabbit()

	e.connections_total.Collect(ch)
	e.channels_total.Collect(ch)
	e.queues_total.Collect(ch)
	e.consumers_total.Collect(ch)
	e.exchanges_total.Collect(ch)
	e.messages_count.Collect(ch)
	e.messages_ready_count.Collect(ch)
	e.messages_unacknowledged_count.Collect(ch)
	e.consumers_count.Collect(ch)
	e.message_bytes.Collect(ch)
	e.disk_reads_count.Collect(ch)
	e.disk_writes_count.Collect(ch)
}
