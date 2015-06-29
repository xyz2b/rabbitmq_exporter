package main

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "rabbitmq"
)

var (
	queueLabelNames = []string{"queue"}

	overviewMetricDescription = map[string]prometheus.Gauge{
		"object_totals.channels":    newMetric("channelsTotal", "Total number of open channels."),
		"object_totals.connections": newMetric("connectionsTotal", "Total number of open connections."),
		"object_totals.consumers":   newMetric("consumersTotal", "Total number of message consumers."),
		"object_totals.queues":      newMetric("queuesTotal", "Total number of queues in use."),
		"object_totals.exchanges":   newMetric("exchangesTotal", "Total number of exchanges in use."),
	}

	queueMetricDescription = map[string]*prometheus.GaugeVec{
		"messages_ready":               newQueueMetric("messages_ready", "Number of messages ready to be delivered to clients."),
		"messages_unacknowledged":      newQueueMetric("messages_unacknowledged", "Number of messages delivered to clients but not yet acknowledged."),
		"messages":                     newQueueMetric("messages", "Sum of ready and unacknowledged messages (queue depth)."),
		"messages_ready_ram":           newQueueMetric("messages_ready_ram", "Number of messages from messages_ready which are resident in ram."),
		"messages_unacknowledged_ram":  newQueueMetric("messages_unacknowledged_ram", "Number of messages from messages_unacknowledged which are resident in ram."),
		"messages_ram":                 newQueueMetric("messages_ram", "Total number of messages which are resident in ram."),
		"messages_persistent":          newQueueMetric("messages_persistent", "Total number of persistent messages in the queue (will always be 0 for transient queues)."),
		"message_bytes":                newQueueMetric("message_bytes", "Sum of the size of all message bodies in the queue. This does not include the message properties (including headers) or any overhead."),
		"message_bytes_ready":          newQueueMetric("message_bytes_ready", "Like message_bytes but counting only those messages ready to be delivered to clients."),
		"message_bytes_unacknowledged": newQueueMetric("message_bytes_unacknowledged", "Like message_bytes but counting only those messages delivered to clients but not yet acknowledged."),
		"message_bytes_ram":            newQueueMetric("message_bytes_ram", "Like message_bytes but counting only those messages which are in RAM."),
		"message_bytes_persistent":     newQueueMetric("message_bytes_persistent", "Like message_bytes but counting only those messages which are persistent."),
		"disk_reads":                   newQueueMetric("disk_reads", "Total number of times messages have been read from disk by this queue since it started."),
		"disk_writes":                  newQueueMetric("disk_writes", "Total number of times messages have been written to disk by this queue since it started."),
		"consumers":                    newQueueMetric("consumers", "Number of consumers."),
		"consumer_utilisation":         newQueueMetric("consumer_utilisation", "Fraction of the time (between 0.0 and 1.0) that the queue is able to immediately deliver messages to consumers. This can be less than 1.0 if consumers are limited by network congestion or prefetch count."),
		"memory":                       newQueueMetric("memory", "Bytes of memory consumed by the Erlang process associated with the queue, including stack, heap and internal structures.")}
)

func newQueueMetric(metricName string, docString string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "queue_" + metricName,
			Help:      docString,
		},
		queueLabelNames,
	)
}

func newMetric(metricName string, docString string) prometheus.Gauge {
	return prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metricName,
			Help:      docString,
		},
	)
}
