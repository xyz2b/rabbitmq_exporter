package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	RegisterExporter("shovel", newExporterShovel)
}

var (
	//shovelLabels are the labels for all shovel mertrics
	shovelLabels = []string{"cluster", "host", "subsystemName", "subsystemID", "vhost", "shovel", "type", "self", "state"}
	//shovelLabelKeys are the important keys to be extracted from json
	shovelLabelKeys = []string{"vhost", "name", "type", "node", "state"}
)

type exporterShovel struct {
	stateMetric *prometheus.GaugeVec
}

func newExporterShovel() Exporter {
	return exporterShovel{
		stateMetric: newGaugeVec("shovel_state", "A metric with a value of constant '1' for each shovel in a certain state", shovelLabels),
	}
}

func (e exporterShovel) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	e.stateMetric.Reset()

	shovelData, err := getStatsInfo(config, "shovels", shovelLabelKeys)
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

	for _, shovel := range shovelData {
		self := "0"
		if shovel.labels["node"] == selfNode {
			self = "1"
		}
		e.stateMetric.WithLabelValues(cluster, host, subsystemName, subsystemID, shovel.labels["vhost"], shovel.labels["name"], shovel.labels["type"], self, shovel.labels["state"]).Set(1)
	}

	e.stateMetric.Collect(ch)
	return nil
}

func (e exporterShovel) Describe(ch chan<- *prometheus.Desc) {
	e.stateMetric.Describe(ch)
}
