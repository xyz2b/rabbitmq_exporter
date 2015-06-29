package main

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

type exporter struct {
	mutex           sync.RWMutex
	queueMetrics    map[string]*prometheus.GaugeVec
	overviewMetrics map[string]prometheus.Gauge
}

func newExporter() *exporter {
	return &exporter{
		queueMetrics:    queueMetricDescription,
		overviewMetrics: overviewMetricDescription,
	}
}

func (e *exporter) fetchRabbit() {
	rabbitMqOverviewData := getOverviewMap(config)
	rabbitMqQueueData := getQueueMap(config)

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
	for key, gaugevec := range e.queueMetrics {
		for queue, data := range rabbitMqQueueData {
			if value, ok := data[key]; ok {
				log.WithFields(log.Fields{"queue": queue, "key": key, "value": value}).Debug("Set queue metric for key")
				gaugevec.WithLabelValues(queue).Set(value)
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

	for _, gaugevec := range e.queueMetrics {
		gaugevec.Describe(ch)
	}
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	for _, gaugevec := range e.queueMetrics {
		gaugevec.Reset()
	}

	e.fetchRabbit()

	for _, gauge := range e.overviewMetrics {
		gauge.Collect(ch)
	}

	for _, gaugevec := range e.queueMetrics {
		gaugevec.Collect(ch)
	}
}
