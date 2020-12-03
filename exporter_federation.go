package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	RegisterExporter("federation", newExporterFederation)
}

var (
	federationLabels     = []string{"cluster","host", "subsystemName", "subsystemID", "vhost", "node", "queue", "exchange", "self", "status"}
	federationLabelsKeys = []string{"vhost", "status", "node", "queue", "exchange"}
)

type exporterFederation struct {
	stateMetric *prometheus.GaugeVec
}

func newExporterFederation() Exporter {
	return exporterFederation{
		stateMetric: newGaugeVec("federation_state", "A metric with a value of constant '1' for each federation in a certain state", federationLabels),
	}
}

func (e exporterFederation) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	e.stateMetric.Reset()

	federationData, err := getStatsInfo(config, "federation-links", federationLabelsKeys)
	if err != nil {
		return err
	}

	cluster := ""
	if n, ok := ctx.Value(clusterName).(string); ok {
		cluster = n
	}
	selfNode := ""
	if n, ok := ctx.Value(nodeName).(string); ok {
		selfNode = n
	}
	host := ""
	if n, ok := ctx.Value(hostInfo).(string); ok {
		host = n
	}
	subsystemName := ""
	if n, ok := ctx.Value(subSystemName).(string); ok {
		subsystemName = n
	}
	subsystemID := ""
	if n, ok := ctx.Value(subSystemID).(string); ok {
		subsystemID = n
	}

	for _, federation := range federationData {
		self := "0"
		if federation.labels["node"] == selfNode {
			self = "1"
		}
		e.stateMetric.WithLabelValues(cluster, host, subsystemName, subsystemID, federation.labels["vhost"], federation.labels["node"], federation.labels["queue"], federation.labels["exchange"], self, federation.labels["status"]).Set(1)
	}

	e.stateMetric.Collect(ch)
	return nil
}

func (e exporterFederation) Describe(ch chan<- *prometheus.Desc) {
	e.stateMetric.Describe(ch)
}
