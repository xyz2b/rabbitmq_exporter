package main

import (
	"context"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	RegisterExporter("queue", newExporterQueue)
}

var (
	queueLabels    = []string{"vhost", "queue", "durable", "policy", "self"}
	queueLabelKeys = []string{"vhost", "name", "durable", "policy", "state", "node"}

	queueGaugeVec = map[string]*prometheus.GaugeVec{
		"messages_ready":               newGaugeVec("queue_messages_ready", "Number of messages ready to be delivered to clients.", queueLabels),
		"messages_unacknowledged":      newGaugeVec("queue_messages_unacknowledged", "Number of messages delivered to clients but not yet acknowledged.", queueLabels),
		"messages":                     newGaugeVec("queue_messages", "Sum of ready and unacknowledged messages (queue depth).", queueLabels),
		"messages_ready_ram":           newGaugeVec("queue_messages_ready_ram", "Number of messages from messages_ready which are resident in ram.", queueLabels),
		"messages_unacknowledged_ram":  newGaugeVec("queue_messages_unacknowledged_ram", "Number of messages from messages_unacknowledged which are resident in ram.", queueLabels),
		"messages_ram":                 newGaugeVec("queue_messages_ram", "Total number of messages which are resident in ram.", queueLabels),
		"messages_persistent":          newGaugeVec("queue_messages_persistent", "Total number of persistent messages in the queue (will always be 0 for transient queues).", queueLabels),
		"message_bytes":                newGaugeVec("queue_message_bytes", "Sum of the size of all message bodies in the queue. This does not include the message properties (including headers) or any overhead.", queueLabels),
		"message_bytes_ready":          newGaugeVec("queue_message_bytes_ready", "Like message_bytes but counting only those messages ready to be delivered to clients.", queueLabels),
		"message_bytes_unacknowledged": newGaugeVec("queue_message_bytes_unacknowledged", "Like message_bytes but counting only those messages delivered to clients but not yet acknowledged.", queueLabels),
		"message_bytes_ram":            newGaugeVec("queue_message_bytes_ram", "Like message_bytes but counting only those messages which are in RAM.", queueLabels),
		"message_bytes_persistent":     newGaugeVec("queue_message_bytes_persistent", "Like message_bytes but counting only those messages which are persistent.", queueLabels),
		"consumers":                    newGaugeVec("queue_consumers", "Number of consumers.", queueLabels),
		"consumer_utilisation":         newGaugeVec("queue_consumer_utilisation", "Fraction of the time (between 0.0 and 1.0) that the queue is able to immediately deliver messages to consumers. This can be less than 1.0 if consumers are limited by network congestion or prefetch count.", queueLabels),
		"memory":                       newGaugeVec("queue_memory", "Bytes of memory consumed by the Erlang process associated with the queue, including stack, heap and internal structures.", queueLabels),
		"head_message_timestamp":       newGaugeVec("queue_head_message_timestamp", "The timestamp property of the first message in the queue, if present. Timestamps of messages only appear when they are in the paged-in state.", queueLabels), //https://github.com/rabbitmq/rabbitmq-server/pull/54
	}

	queueCounterVec = map[string]*prometheus.Desc{
		"disk_reads":                   newDesc("queue_disk_reads_total", "Total number of times messages have been read from disk by this queue since it started.", queueLabels),
		"disk_writes":                  newDesc("queue_disk_writes_total", "Total number of times messages have been written to disk by this queue since it started.", queueLabels),
		"message_stats.publish":        newDesc("queue_messages_published_total", "Count of messages published.", queueLabels),
		"message_stats.confirm":        newDesc("queue_messages_confirmed_total", "Count of messages confirmed. ", queueLabels),
		"message_stats.deliver":        newDesc("queue_messages_delivered_total", "Count of messages delivered in acknowledgement mode to consumers.", queueLabels),
		"message_stats.deliver_no_ack": newDesc("queue_messages_delivered_noack_total", "Count of messages delivered in no-acknowledgement mode to consumers. ", queueLabels),
		"message_stats.get":            newDesc("queue_messages_get_total", "Count of messages delivered in acknowledgement mode in response to basic.get.", queueLabels),
		"message_stats.get_no_ack":     newDesc("queue_messages_get_noack_total", "Count of messages delivered in no-acknowledgement mode in response to basic.get.", queueLabels),
		"message_stats.redeliver":      newDesc("queue_messages_redelivered_total", "Count of subset of messages in deliver_get which had the redelivered flag set.", queueLabels),
		"message_stats.return":         newDesc("queue_messages_returned_total", "Count of messages returned to publisher as unroutable.", queueLabels),
		"message_stats.ack":            newDesc("queue_messages_ack_total", "Count of messages delivered in acknowledgement mode in response to basic.get.", queueLabels),
	}
)

type exporterQueue struct {
	queueMetricsGauge   map[string]*prometheus.GaugeVec
	queueMetricsCounter map[string]*prometheus.Desc
	stateMetric         *prometheus.GaugeVec
}

func newExporterQueue() Exporter {
	queueGaugeVecActual := queueGaugeVec
	queueCounterVecActual := queueCounterVec

	if len(config.ExcludeMetrics) > 0 {
		for _, metric := range config.ExcludeMetrics {
			if queueGaugeVecActual[metric] != nil {
				delete(queueGaugeVecActual, metric)
			}
			if queueCounterVecActual[metric] != nil {
				delete(queueCounterVecActual, metric)
			}
		}
	}

	return exporterQueue{
		queueMetricsGauge:   queueGaugeVecActual,
		queueMetricsCounter: queueCounterVecActual,
		stateMetric:         newGaugeVec("queue_state", "A metric with a value of constant '1' if the queue is in a certain state", append(queueLabels, "state")),
	}
}

func (e exporterQueue) String() string {
	return "Exporter queue"
}

func (e exporterQueue) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	for _, gaugevec := range e.queueMetricsGauge {
		gaugevec.Reset()
	}
	e.stateMetric.Reset()

	if config.MaxQueues > 0 {
		// Get overview info to check total queues
		rabbitMqOverviewData, err := getMetricMap(config, "overview") //TODO: Replace by context value

		if err != nil {
			return err
		}

		totalQueues := rabbitMqOverviewData["object_totals.queues"]

		if int(totalQueues) > config.MaxQueues {
			log.WithFields(log.Fields{
				"MaxQueues":   config.MaxQueues,
				"TotalQueues": totalQueues,
			}).Debug("MaxQueues exceeded.")
			return nil
		}
	}

	selfNode := ""
	if n, ok := ctx.Value(nodeName).(string); ok {
		selfNode = n
	}

	rabbitMqQueueData, err := getStatsInfo(config, "queues", queueLabelKeys)

	if err != nil {
		return err
	}

	log.WithField("queueData", rabbitMqQueueData).Debug("Queue data")
	for key, gaugevec := range e.queueMetricsGauge {
		for _, queue := range rabbitMqQueueData {
			qname := queue.labels["name"]
			vname := queue.labels["vhost"]
			if value, ok := queue.metrics[key]; ok {

				if matchVhost := config.IncludeVHost.MatchString(strings.ToLower(vname)); matchVhost {
					if skipVhost := config.SkipVHost.MatchString(strings.ToLower(vname)); !skipVhost {
						if matchInclude := config.IncludeQueues.MatchString(strings.ToLower(qname)); matchInclude {
							if matchSkip := config.SkipQueues.MatchString(strings.ToLower(qname)); !matchSkip {
								self := "0"
								if queue.labels["node"] == selfNode {
									self = "1"
								}
								// log.WithFields(log.Fields{"vhost": queue.labels["vhost"], "queue": queue.labels["name"], "key": key, "value": value}).Info("Set queue metric for key")
								gaugevec.WithLabelValues(queue.labels["vhost"], queue.labels["name"], queue.labels["durable"], queue.labels["policy"], self).Set(value)
							}
						}
					}
				}
			}
		}
	}

	for _, queue := range rabbitMqQueueData {
		qname := queue.labels["name"]
		vname := queue.labels["vhost"]
		if _, ok := queue.metrics["messages"]; ok { // "messages" is used to retrieve one record per queue for setting the queue state

			if matchVhost := config.IncludeVHost.MatchString(strings.ToLower(vname)); matchVhost {
				if skipVhost := config.SkipVHost.MatchString(strings.ToLower(vname)); !skipVhost {
					if matchInclude := config.IncludeQueues.MatchString(strings.ToLower(qname)); matchInclude {
						if matchSkip := config.SkipQueues.MatchString(strings.ToLower(qname)); !matchSkip {
							self := "0"
							if queue.labels["node"] == selfNode {
								self = "1"
							}
							e.stateMetric.WithLabelValues(queue.labels["vhost"], queue.labels["name"], queue.labels["durable"], queue.labels["policy"], self, queue.labels["state"]).Set(1)
						}
					}
				}
			}
		}
	}

	for key, countvec := range e.queueMetricsCounter {
		for _, queue := range rabbitMqQueueData {
			qname := queue.labels["name"]
			vname := queue.labels["vhost"]

			if matchVhost := config.IncludeVHost.MatchString(strings.ToLower(vname)); matchVhost {
				if skipVhost := config.SkipVHost.MatchString(strings.ToLower(vname)); !skipVhost {
					if matchInclude := config.IncludeQueues.MatchString(strings.ToLower(qname)); matchInclude {
						if matchSkip := config.SkipQueues.MatchString(strings.ToLower(qname)); !matchSkip {
							self := "0"
							if queue.labels["node"] == selfNode {
								self = "1"
							}
							if value, ok := queue.metrics[key]; ok {
								ch <- prometheus.MustNewConstMetric(countvec, prometheus.CounterValue, value, queue.labels["vhost"], queue.labels["name"], queue.labels["durable"], queue.labels["policy"], self)
							} else {
								ch <- prometheus.MustNewConstMetric(countvec, prometheus.CounterValue, 0, queue.labels["vhost"], queue.labels["name"], queue.labels["durable"], queue.labels["policy"], self)
							}
						}
					}
				}
			}
		}
	}

	for _, gaugevec := range e.queueMetricsGauge {
		gaugevec.Collect(ch)
	}
	e.stateMetric.Collect(ch)

	return nil
}

func (e exporterQueue) Describe(ch chan<- *prometheus.Desc) {
	for _, gaugevec := range e.queueMetricsGauge {
		gaugevec.Describe(ch)
	}
	e.stateMetric.Describe(ch)
	for _, countervec := range e.queueMetricsCounter {
		ch <- countervec
	}
}
