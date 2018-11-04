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
	endpointScrapeDurationMetric *prometheus.GaugeVec
	exporter                     map[string]Exporter
	overviewExporter             *exporterOverview
	self                         string
}

//Exporter interface for prometheus metrics. Collect is fetching the data and therefore can return an error
type Exporter interface {
	Collect(ch chan<- prometheus.Metric) error
	Describe(ch chan<- *prometheus.Desc)
}

//NodeAwareExporter are exporters where the data is located on one rabbit node. This is indicated by a self label
type NodeAwareExporter interface {
	SetSelfNode(string)
}

func newExporter() *exporter {
	enabledExporter := make(map[string]Exporter)
	for _, e := range config.EnabledExporters {
		if _, ok := exporterFactories[e]; ok {
			enabledExporter[e] = exporterFactories[e]()
		}
	}

	return &exporter{
		upMetric:                     newGauge("up", "Was the last scrape of rabbitmq successful."),
		endpointUpMetric:             newGaugeVec("module_up", "Was the last scrape of rabbitmq successful per module.", []string{"module"}),
		endpointScrapeDurationMetric: newGaugeVec("module_scrape_duration_seconds", "Duration of the last scrape in seconds", []string{"module"}),
		exporter:                     enabledExporter,
		overviewExporter:             newExporterOverview(),
	}
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {

	e.overviewExporter.Describe(ch)
	for _, ex := range e.exporter {
		ex.Describe(ch)
	}

	e.upMetric.Describe(ch)
	e.endpointUpMetric.Describe(ch)
	e.endpointScrapeDurationMetric.Describe(ch)
	BuildInfo.Describe(ch)
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	start := time.Now()
	allUp := true

	if success := collectWithDuration(e.overviewExporter, "overview", ch, e.endpointScrapeDurationMetric, e.endpointUpMetric); !success {
		allUp = false
	}
	nodeName := e.overviewExporter.NodeInfo().Node

	for name, ex := range e.exporter {
		if nex, ok := ex.(NodeAwareExporter); ok {
			nex.SetSelfNode(nodeName)
		}

		if success := collectWithDuration(ex, name, ch, e.endpointScrapeDurationMetric, e.endpointUpMetric); !success {
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
	e.endpointUpMetric.Collect(ch)
	e.endpointScrapeDurationMetric.Collect(ch)
	log.WithField("duration", time.Since(start)).Info("Metrics updated")

}

func collectWithDuration(ex Exporter, name string, ch chan<- prometheus.Metric, scrapeDuration, up *prometheus.GaugeVec) bool {
	startModule := time.Now()
	err := ex.Collect(ch)
	scrapeDuration.WithLabelValues(name).Set(time.Since(startModule).Seconds())
	if err != nil {
		up.WithLabelValues(name).Set(0)
		return false
	}
	up.WithLabelValues(name).Set(1)
	return true

}
