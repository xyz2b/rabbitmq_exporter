package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func init() {
	//RegisterExporter("overview", newExporterOverview)
}

var (
	overviewLabels = []string{"cluster"}

	overviewMetricDescription = map[string]*prometheus.GaugeVec{
		"object_totals.channels":               newGaugeVec("channels", "Number of channels.", overviewLabels),
		"object_totals.connections":            newGaugeVec("connections", "Number of connections.", overviewLabels),
		"object_totals.consumers":              newGaugeVec("consumers", "Number of message consumers.", overviewLabels),
		"object_totals.queues":                 newGaugeVec("queues", "Number of queues in use.", overviewLabels),
		"object_totals.exchanges":              newGaugeVec("exchanges", "Number of exchanges in use.", overviewLabels),
		"queue_totals.messages":                newGaugeVec("queue_messages_global", "Number ready and unacknowledged messages in cluster.", overviewLabels),
		"queue_totals.messages_ready":          newGaugeVec("queue_messages_ready_global", "Number of messages ready to be delivered to clients.", overviewLabels),
		"queue_totals.messages_unacknowledged": newGaugeVec("queue_messages_unacknowledged_global", "Number of messages delivered to clients but not yet acknowledged.", overviewLabels),
	}

	rabbitmqVersionMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rabbitmq_version_info",
			Help: "A metric with a constant '1' value labeled by rabbitmq version, erlang version, node, cluster.",
		},
		[]string{"rabbitmq", "erlang", "node", "cluster"},
	)
)

type exporterOverview struct {
	overviewMetrics map[string]*prometheus.GaugeVec
	nodeInfo        NodeInfo
}

//NodeInfo presents the name and version of fetched rabbitmq
type NodeInfo struct {
	Node            string `json:"node"`
	RabbitmqVersion string `json:"rabbitmq_version"`
	ErlangVersion   string `json:"erlang_version"`
	ClusterName     string `json:"cluster_name"`
}

func newExporterOverview() *exporterOverview {
	overviewMetricDescriptionActual := overviewMetricDescription

	if len(config.ExcludeMetrics) > 0 {
		for _, metric := range config.ExcludeMetrics {
			if overviewMetricDescriptionActual[metric] != nil {
				delete(overviewMetricDescriptionActual, metric)
			}
		}
	}

	return &exporterOverview{
		overviewMetrics: overviewMetricDescriptionActual,
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

	reply, err := MakeReply(config, body)
	if err != nil {
		return err
	}

	e.nodeInfo.Node, _ = reply.GetString("node")
	e.nodeInfo.ErlangVersion, _ = reply.GetString("erlang_version")
	e.nodeInfo.RabbitmqVersion, _ = reply.GetString("rabbitmq_version")
	e.nodeInfo.ClusterName, _ = reply.GetString("cluster_name")

	rabbitmqVersionMetric.Reset()
	rabbitmqVersionMetric.WithLabelValues(e.nodeInfo.RabbitmqVersion, e.nodeInfo.ErlangVersion, e.nodeInfo.Node, e.nodeInfo.ClusterName).Set(1)

	rabbitMqOverviewData := reply.MakeMap()

	log.WithField("overviewData", rabbitMqOverviewData).Debug("Overview data")
	for key, gauge := range e.overviewMetrics {
		if value, ok := rabbitMqOverviewData[key]; ok {
			log.WithFields(log.Fields{"key": key, "value": value}).Debug("Set overview metric for key")
			gauge.WithLabelValues(e.nodeInfo.ClusterName).Set(value)
		}
	}

	rabbitmqVersionMetric.Collect(ch)
	for _, gauge := range e.overviewMetrics {
		gauge.Collect(ch)
	}
	return nil
}

func (e exporterOverview) Describe(ch chan<- *prometheus.Desc) {
	rabbitmqVersionMetric.Describe(ch)

	for _, gauge := range e.overviewMetrics {
		gauge.Describe(ch)
	}

}
