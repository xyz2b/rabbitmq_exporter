package main

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "rabbitmq"
)

var (
	queueLabelNames = []string{"vhost", "queue"}

	overviewMetricDescription = map[string]prometheus.Gauge{
		"object_totals.channels":    newMetric("channelsTotal", "Total number of open channels."),
		"object_totals.connections": newMetric("connectionsTotal", "Total number of open connections."),
		"object_totals.consumers":   newMetric("consumersTotal", "Total number of message consumers."),
		"object_totals.queues":      newMetric("queuesTotal", "Total number of queues in use."),
		"object_totals.exchanges":   newMetric("exchangesTotal", "Total number of exchanges in use."),
	}

	queueGaugeVec = map[string]*prometheus.GaugeVec{
		"messages_ready":               newQueueGaugeVec("messages_ready", "Number of messages ready to be delivered to clients."),
		"messages_unacknowledged":      newQueueGaugeVec("messages_unacknowledged", "Number of messages delivered to clients but not yet acknowledged."),
		"messages":                     newQueueGaugeVec("messages", "Sum of ready and unacknowledged messages (queue depth)."),
		"messages_ready_ram":           newQueueGaugeVec("messages_ready_ram", "Number of messages from messages_ready which are resident in ram."),
		"messages_unacknowledged_ram":  newQueueGaugeVec("messages_unacknowledged_ram", "Number of messages from messages_unacknowledged which are resident in ram."),
		"messages_ram":                 newQueueGaugeVec("messages_ram", "Total number of messages which are resident in ram."),
		"messages_persistent":          newQueueGaugeVec("messages_persistent", "Total number of persistent messages in the queue (will always be 0 for transient queues)."),
		"message_bytes":                newQueueGaugeVec("message_bytes", "Sum of the size of all message bodies in the queue. This does not include the message properties (including headers) or any overhead."),
		"message_bytes_ready":          newQueueGaugeVec("message_bytes_ready", "Like message_bytes but counting only those messages ready to be delivered to clients."),
		"message_bytes_unacknowledged": newQueueGaugeVec("message_bytes_unacknowledged", "Like message_bytes but counting only those messages delivered to clients but not yet acknowledged."),
		"message_bytes_ram":            newQueueGaugeVec("message_bytes_ram", "Like message_bytes but counting only those messages which are in RAM."),
		"message_bytes_persistent":     newQueueGaugeVec("message_bytes_persistent", "Like message_bytes but counting only those messages which are persistent."),
		"consumers":                    newQueueGaugeVec("consumers", "Number of consumers."),
		"consumer_utilisation":         newQueueGaugeVec("consumer_utilisation", "Fraction of the time (between 0.0 and 1.0) that the queue is able to immediately deliver messages to consumers. This can be less than 1.0 if consumers are limited by network congestion or prefetch count."),
		"memory":                       newQueueGaugeVec("memory", "Bytes of memory consumed by the Erlang process associated with the queue, including stack, heap and internal structures."),
	}

	queueCounterVec = map[string]*prometheus.CounterVec{
		"disk_reads":                  newQueueCounterVec("disk_reads", "Total number of times messages have been read from disk by this queue since it started."),
		"disk_writes":                 newQueueCounterVec("disk_writes", "Total number of times messages have been written to disk by this queue since it started."),
		"message_stats.publish":       newQueueCounterVec("messages_published_total", "Count of messages published."),
		"message_stats.confirm":       newQueueCounterVec("messages_confirmed_total", "Count of messages confirmed. "),
		"message_stats.deliver":       newQueueCounterVec("messages_delivered_total", "Count of messages delivered in acknowledgement mode to consumers."),
		"message_stats.deliver_noack": newQueueCounterVec("messages_delivered_noack_total", "Count of messages delivered in no-acknowledgement mode to consumers. "),
		"message_stats.get":           newQueueCounterVec("messages_get_total", "Count of messages delivered in acknowledgement mode in response to basic.get."),
		"message_stats.get_noack":     newQueueCounterVec("messages_get_noack_total", "Count of messages delivered in no-acknowledgement mode in response to basic.get."),
		"message_stats.redeliver":     newQueueCounterVec("messages_redelivered_total", "Count of subset of messages in deliver_get which had the redelivered flag set."),
		"message_stats.return":        newQueueCounterVec("messages_returned_total", "Count of messages returned to publisher as unroutable."),
	}
)

func newQueueCounterVec(metricName string, docString string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "queue_" + metricName,
			Help:      docString,
		},
		queueLabelNames,
	)
}
func newQueueGaugeVec(metricName string, docString string) *prometheus.GaugeVec {
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
