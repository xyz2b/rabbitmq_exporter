package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	RegisterExporter("overview", newExporterOverview)
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
}

func newExporterOverview() Exporter {
	return exporterOverview{
		overviewMetrics: overviewMetricDescription,
	}
}

func (e exporterOverview) String() string {
	return "Exporter overview"
}

func (e exporterOverview) Collect(ch chan<- prometheus.Metric) error {
	rabbitMqOverviewData, err := getMetricMap(config, "overview")

	if err != nil {
		return err
	}

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
