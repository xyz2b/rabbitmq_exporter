package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	nodeLabels    = []string{"vhost", "node"}
	nodeLabelKeys = []string{"vhost", "name"}

	nodeCounterVec = map[string]*prometheus.Desc{
		"running":         newDesc("running", "test", nodeLabels),
		"mem_used":        newDesc("node_mem_used", "Memory used in bytes", nodeLabels),
		"mem_limit":       newDesc("node_mem_limit", "Point at which the memory alarm will go off", nodeLabels),
		"mem_alarm":       newDesc("node_mem_alarm", "Whether the memory alarm has gone off", nodeLabels),
		"disk_free":       newDesc("node_disk_free", "Disk free space in bytes.", nodeLabels),
		"disk_free_alarm": newDesc("node_disk_free_alarm", "Whether the disk alarm has gone off.", nodeLabels),
		"disk_free_limit": newDesc("node_disk_free_limit", "Point at which the disk alarm will go off.", nodeLabels),
		"fd_used":         newDesc("fd_used", "Used File descriptors", nodeLabels),
		"fd_total":        newDesc("fd_total", "File descriptors available", nodeLabels),
		"sockets_used":    newDesc("sockets_used", "File descriptors used as sockets.", nodeLabels),
		"sockets_total":   newDesc("sockets_total", "File descriptors available for use as sockets", nodeLabels),
	}
)

type exporterNode struct {
	nodeMetricsCounter map[string]*prometheus.Desc
}

func newExporterNode() exporterNode {
	return exporterNode{
		nodeMetricsCounter: nodeCounterVec,
	}
}

func (e exporterNode) String() string {
	return "Exporter node"
}

func (e exporterNode) Collect(ch chan<- prometheus.Metric) error {
	nodeData, err := getStatsInfo(config, "nodes", nodeLabelKeys)

	if err != nil {
		return err
	}

	for key, countvec := range e.nodeMetricsCounter {
		for _, node := range nodeData {
			if value, ok := node.metrics[key]; ok {
				// log.WithFields(log.Fields{"type": node.vhost, "node": node.name, "key": key, "value": value}).Debug("Set node metric for key")
				ch <- prometheus.MustNewConstMetric(countvec, prometheus.CounterValue, value, node.labels["vhost"], node.labels["name"])
			}
		}
	}
	return nil
}

func (e exporterNode) Describe(ch chan<- *prometheus.Desc) {
	for _, nodeMetric := range e.nodeMetricsCounter {
		ch <- nodeMetric
	}

}
