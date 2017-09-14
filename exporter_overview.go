package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	RegisterExporter("overview", newExporterOverview)
}

var overviewMetricDescription = map[string]prometheus.Gauge{
	"object_totals.channels":    newGauge("channelsTotal", "Total number of open channels."),
	"object_totals.connections": newGauge("connectionsTotal", "Total number of open connections."),
	"object_totals.consumers":   newGauge("consumersTotal", "Total number of message consumers."),
	"object_totals.queues":      newGauge("queuesTotal", "Total number of queues in use."),
	"object_totals.exchanges":   newGauge("exchangesTotal", "Total number of exchanges in use."),
	"partitions_len":            newGauge("partitions", "Current Number of network partitions. 0 is ok. If the cluster is splitted the value is at least 2"),
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
