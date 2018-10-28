package main

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	exportersMu       sync.RWMutex
	exporterFactories = make(map[string]func() Exporter)
)

//RegisterExporter makes an exporter available by the provided name.
func RegisterExporter(name string, f func() Exporter) {
	exportersMu.Lock()
	defer exportersMu.Unlock()
	if f == nil {
		panic("exporterFactory is nil")
	}
	exporterFactories[name] = f
}

type exporter struct {
	mutex                        sync.RWMutex
	upMetric                     prometheus.Gauge
	endpointUpMetric             *prometheus.GaugeVec
	exporter                     map[string]Exporter
}

//Exporter interface for prometheus metrics. Collect is fetching the data and therefore can return an error
type Exporter interface {
	Collect(ch chan<- prometheus.Metric) error
	Describe(ch chan<- *prometheus.Desc)
}

func newExporter() *exporter {
	enabledExporter := make(map[string]Exporter)
	for _, e := range config.EnabledExporters {
		enabledExporter[e] = exporterFactories[e]()
	}

	return &exporter{
		upMetric:                     newGauge("up", "Was the last scrape of rabbitmq successful."),
		endpointUpMetric:             newGaugeVec("module_up", "Was the last scrape of rabbitmq successful per module.", []string{"module"}),
		exporter:                     enabledExporter,
	}
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {

	for _, ex := range e.exporter {
		ex.Describe(ch)
	}

	e.upMetric.Describe(ch)
	e.endpointUpMetric.Describe(ch)
	BuildInfo.Describe(ch)
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	start := time.Now()
	allUp := true

	for name, ex := range e.exporter {
		err := ex.Collect(ch)
		if err != nil {
			allUp = false
			e.endpointUpMetric.WithLabelValues(name).Set(0)
		} else {
			e.endpointUpMetric.WithLabelValues(name).Set(1)
		}
	}

	BuildInfo.Collect(ch)

	if allUp {
		e.upMetric.Set(1)
	} else {
		e.upMetric.Set(0)
	}
	e.upMetric.Collect(ch)
	e.endpointUpMetric.Collect(ch)
	log.WithField("duration", time.Since(start)).Info("Metrics updated")

}
