package main

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

type exporter struct {
	mutex    sync.RWMutex
	upMetric prometheus.Gauge
	exporter []Exporter
}

type Exporter interface {
	Collect(ch chan<- prometheus.Metric) error
	Describe(ch chan<- *prometheus.Desc)
}

func newExporter() *exporter {
	return &exporter{

		upMetric: newGauge("up", "Was the last scrape of rabbitmq successful."),

		exporter: []Exporter{
			NewExporterNode(),
			NewExporterQueue(),
			NewExporterOverview(),
			NewExporterExchange(),
			// NewExporterConnections(),
		},
	}
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {

	for _, ex := range e.exporter {
		ex.Describe(ch)
	}

	e.upMetric.Describe(ch)
	BuildInfo.Describe(ch)
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	start := time.Now()
	allUp := true

	for _, ex := range e.exporter {
		err := ex.Collect(ch)
		if err != nil {
			allUp = false
		}
	}

	BuildInfo.Collect(ch)

	if allUp {
		e.upMetric.Set(1)
	} else {
		e.upMetric.Set(0)
	}
	e.upMetric.Collect(ch)
	log.WithField("duration", time.Since(start)).Info("Metrics updated")

}
