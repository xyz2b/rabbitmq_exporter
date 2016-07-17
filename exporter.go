package main

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

type exporter struct {
	mutex               sync.RWMutex
	queueMetricsGauge   map[string]*prometheus.GaugeVec
	queueMetricsCounter map[string]*prometheus.CounterVec
	overviewMetrics     map[string]prometheus.Gauge
}

func newExporter() *exporter {
	return &exporter{
		queueMetricsGauge:   queueGaugeVec,
		queueMetricsCounter: queueCounterVec,
		overviewMetrics:     overviewMetricDescription,
	}
}

func (e *exporter) fetchRabbit() {
	rabbitMqOverviewData := getOverviewMap(config)
	rabbitMqQueueData := getQueueInfo(config)

	log.WithField("overviewData", rabbitMqOverviewData).Debug("Overview data")
	for key, gauge := range e.overviewMetrics {
		if value, ok := rabbitMqOverviewData[key]; ok {
			log.WithFields(log.Fields{"key": key, "value": value}).Debug("Set overview metric for key")
			gauge.Set(value)
		} else {
			log.WithFields(log.Fields{"key": key}).Warn("Overview data not found")
		}
	}

	log.WithField("queueData", rabbitMqQueueData).Debug("Queue data")
	for key, gaugevec := range e.queueMetricsGauge {
		for _, queue := range rabbitMqQueueData {
			if value, ok := queue.metrics[key]; ok {
				log.WithFields(log.Fields{"vhost": queue.vhost, "queue": queue.name, "key": key, "value": value}).Debug("Set queue metric for key")
				gaugevec.WithLabelValues(queue.vhost, queue.name).Set(value)
			} else {
				//log.WithFields(log.Fields{"queue": queue, "key": key}).Warn("Queue data not found")
			}
		}
	}
	for key, countvec := range e.queueMetricsCounter {
		for _, queue := range rabbitMqQueueData {
			if value, ok := queue.metrics[key]; ok {
				log.WithFields(log.Fields{"vhost": queue.vhost, "queue": queue.name, "key": key, "value": value}).Debug("Set queue metric for key")
				countvec.WithLabelValues(queue.vhost, queue.name).Set(value)
			} else {
				//log.WithFields(log.Fields{"queue": queue, "key": key}).Warn("Queue data not found")
			}
		}
	}

	log.Info("Metrics updated successfully.")
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, gauge := range e.overviewMetrics {
		gauge.Describe(ch)
	}

	for _, gaugevec := range e.queueMetricsGauge {
		gaugevec.Describe(ch)
	}
	for _, countervec := range e.queueMetricsCounter {
		countervec.Describe(ch)
	}
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	for _, gaugevec := range e.queueMetricsGauge {
		gaugevec.Reset()
	}
	for _, countvec := range e.queueMetricsCounter {
		countvec.Reset()
	}

	e.fetchRabbit()

	for _, gauge := range e.overviewMetrics {
		gauge.Collect(ch)
	}

	for _, gaugevec := range e.queueMetricsGauge {
		gaugevec.Collect(ch)
	}
	for _, countervec := range e.queueMetricsCounter {
		countervec.Collect(ch)
	}

	BuildInfo.Collect(ch)
}
