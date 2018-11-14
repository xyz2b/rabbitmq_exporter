package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	//RegisterExporter("overview", newExporterOverview)
}

var overviewMetricDescription = map[string]prometheus.Gauge{
	"object_totals.channels":               newGauge("channels", "Number of channels."),
	"object_totals.connections":            newGauge("connections", "Number of connections."),
	"object_totals.consumers":              newGauge("consumers", "Number of message consumers."),
	"object_totals.queues":                 newGauge("queues", "Number of queues in use."),
	"object_totals.exchanges":              newGauge("exchanges", "Number of exchanges in use."),
	"queue_totals.messages":                newGauge("queue_messages_global", "Number ready and unacknowledged messages in cluster."),
	"queue_totals.messages_ready":          newGauge("queue_messages_ready_global", "Number of messages ready to be delivered to clients."),
	"queue_totals.messages_unacknowledged": newGauge("queue_messages_unacknowledged_global", "Number of messages delivered to clients but not yet acknowledged."),
}

type exporterOverview struct {
	overviewMetrics map[string]prometheus.Gauge
	nodeInfo        NodeInfo
}

//NodeInfo presents the name and version of fetched rabbitmq
type NodeInfo struct {
	Node            string `json:"node"`
	RabbitmqVersion string `json:"rabbitmq_version"`
	ErlangVersion   string `json:"erlang_version"`
}

func newExporterOverview() *exporterOverview {
	return &exporterOverview{
		overviewMetrics: overviewMetricDescription,
		nodeInfo:        NodeInfo{},
	}
}

func (e exporterOverview) String() string {
	return "overview"
}

func (e exporterOverview) NodeInfo() NodeInfo {
	return e.nodeInfo
}

func (e *exporterOverview) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	body, err := apiRequest(config, "overview")
	if err != nil {
		return err
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	var nodeInfo NodeInfo
	if err := dec.Decode(&nodeInfo); err == io.EOF {
		return err
	}
	e.nodeInfo = nodeInfo

	reply, err := MakeReply(config, body)
	if err != nil {
		return err
	}
	rabbitMqOverviewData := reply.MakeMap()

	log.WithField("overviewData", rabbitMqOverviewData).Debug("Overview data")
	for key, gauge := range e.overviewMetrics {
		if value, ok := rabbitMqOverviewData[key]; ok {
			log.WithFields(log.Fields{"key": key, "value": value}).Debug("Set overview metric for key")
			gauge.Set(value)
		}
	}

	for _, gauge := range e.overviewMetrics {
		gauge.Collect(ch)
	}
	return nil
}

func (e exporterOverview) Describe(ch chan<- *prometheus.Desc) {
	for _, gauge := range e.overviewMetrics {
		gauge.Describe(ch)
	}

}
