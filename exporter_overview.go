package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

var overviewMetricDescription = map[string]prometheus.Gauge{
	"object_totals.channels":    newGauge("channelsTotal", "Total number of open channels."),
	"object_totals.connections": newGauge("connectionsTotal", "Total number of open connections."),
	"object_totals.consumers":   newGauge("consumersTotal", "Total number of message consumers."),
	"object_totals.queues":      newGauge("queuesTotal", "Total number of queues in use."),
	"object_totals.exchanges":   newGauge("exchangesTotal", "Total number of exchanges in use."),
}

type exporterOverview struct {
	overviewMetrics map[string]prometheus.Gauge
}

func NewExporterOverview() exporterOverview {
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
		} else {
			log.WithFields(log.Fields{"key": key}).Warn("Overview data not found")
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
