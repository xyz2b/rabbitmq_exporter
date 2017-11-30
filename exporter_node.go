package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	RegisterExporter("node", newExporterNode)
}

var (
	nodeGauge = map[string]prometheus.Gauge{
		"running":         newGauge("running", "number of running nodes"),
		"mem_used":        newGauge("node_mem_used", "Memory used in bytes"),
		"mem_limit":       newGauge("node_mem_limit", "Point at which the memory alarm will go off"),
		"mem_alarm":       newGauge("node_mem_alarm", "Whether the memory alarm has gone off"),
		"disk_free":       newGauge("node_disk_free", "Disk free space in bytes."),
		"disk_free_alarm": newGauge("node_disk_free_alarm", "Whether the disk alarm has gone off."),
		"disk_free_limit": newGauge("node_disk_free_limit", "Point at which the disk alarm will go off."),
		"fd_used":         newGauge("fd_used", "Used File descriptors"),
		"fd_total":        newGauge("fd_total", "File descriptors available"),
		"sockets_used":    newGauge("sockets_used", "File descriptors used as sockets."),
		"sockets_total":   newGauge("sockets_total", "File descriptors available for use as sockets"),
		"partitions_len":  newGauge("partitions", "Current Number of network partitions. 0 is ok. If the cluster is splitted the value is at least 2"),
	}
)

type exporterNode struct {
	nodeMetricsGauge map[string]prometheus.Gauge
}

func newExporterNode() Exporter {
	return exporterNode{
		nodeMetricsGauge: nodeGauge,
	}
}

func (e exporterNode) String() string {
	return "Exporter node"
}

func (e exporterNode) Collect(ch chan<- prometheus.Metric) error {
	nodeData, err := getStatsInfo(config, "nodes", []string{})

	if err != nil {
		return err
	}

	for key, gauge := range e.nodeMetricsGauge {
		for _, node := range nodeData {
			if value, ok := node.metrics[key]; ok {
				gauge.Set(value)
			}
		}
	}

	for _, gauge := range e.nodeMetricsGauge {
		gauge.Collect(ch)
	}
	return nil
}

func (e exporterNode) Describe(ch chan<- *prometheus.Desc) {
	for _, nodeMetric := range e.nodeMetricsGauge {
		nodeMetric.Describe(ch)
	}

}
