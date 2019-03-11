package main

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	exportersMu       sync.RWMutex
	exporterFactories = make(map[string]func() Exporter)
)

type contextValues string

const (
	endpointScrapeDuration contextValues = "endpointScrapeDuration"
	endpointUpMetric       contextValues = "endpointUpMetric"
	nodeName               contextValues = "node"
	clusterName            contextValues = "cluster"
	totalQueues            contextValues = "totalQueues"
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
	upMetric                     *prometheus.GaugeVec
	endpointUpMetric             *prometheus.GaugeVec
	endpointScrapeDurationMetric *prometheus.GaugeVec
	exporter                     map[string]Exporter
	overviewExporter             *exporterOverview
	self                         string
}

//Exporter interface for prometheus metrics. Collect is fetching the data and therefore can return an error
type Exporter interface {
	Collect(ctx context.Context, ch chan<- prometheus.Metric) error
	Describe(ch chan<- *prometheus.Desc)
}

func newExporter() *exporter {
	enabledExporter := make(map[string]Exporter)
	for _, e := range config.EnabledExporters {
		if _, ok := exporterFactories[e]; ok {
			enabledExporter[e] = exporterFactories[e]()
		}
	}

	return &exporter{
		upMetric:                     newGaugeVec("up", "Was the last scrape of rabbitmq successful.", []string{"cluster", "node"}),
		endpointUpMetric:             newGaugeVec("module_up", "Was the last scrape of rabbitmq successful per module.", []string{"cluster", "node", "module"}),
		endpointScrapeDurationMetric: newGaugeVec("module_scrape_duration_seconds", "Duration of the last scrape in seconds", []string{"cluster", "node", "module"}),
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

	ctx := context.Background()
	ctx = context.WithValue(ctx, endpointScrapeDuration, e.endpointScrapeDurationMetric)
	ctx = context.WithValue(ctx, endpointUpMetric, e.endpointUpMetric)
	if err := collectWithDuration(ctx, e.overviewExporter, "overview", ch); err != nil {
		log.WithError(err).Warn("retrieving overview failed")
		allUp = false
	}

	ctx = context.WithValue(ctx, nodeName, e.overviewExporter.NodeInfo().Node)
	ctx = context.WithValue(ctx, clusterName, e.overviewExporter.NodeInfo().ClusterName)
	ctx = context.WithValue(ctx, totalQueues, e.overviewExporter.NodeInfo().TotalQueues)

	for name, ex := range e.exporter {

		if err := collectWithDuration(ctx, ex, name, ch); err != nil {
			log.WithError(err).Warn("retrieving " + name + " failed")
			allUp = false
		}
	}

	BuildInfo.Collect(ch)

	if allUp {
		e.upMetric.WithLabelValues(e.overviewExporter.NodeInfo().ClusterName, e.overviewExporter.NodeInfo().Node).Set(1)
	} else {
		e.upMetric.WithLabelValues(e.overviewExporter.NodeInfo().ClusterName, e.overviewExporter.NodeInfo().Node).Set(0)
	}
	e.upMetric.Collect(ch)
	e.endpointUpMetric.Collect(ch)
	e.endpointScrapeDurationMetric.Collect(ch)
	log.WithField("duration", time.Since(start)).Info("Metrics updated")

}

func collectWithDuration(ctx context.Context, ex Exporter, name string, ch chan<- prometheus.Metric) error {
	startModule := time.Now()
	err := ex.Collect(ctx, ch)
	node := ""
	if n, ok := ctx.Value(clusterName).(string); ok {
		node = n
	}
	cluster := ""
	if n, ok := ctx.Value(clusterName).(string); ok {
		cluster = n
	}

	if scrapeDuration, ok := ctx.Value(endpointScrapeDuration).(*prometheus.GaugeVec); ok {
		if cluster != "" && node != "" { //values are not available until first scrape of overview succeeded
			scrapeDuration.WithLabelValues(cluster, node, name).Set(time.Since(startModule).Seconds())
		}
	}
	if up, ok := ctx.Value(endpointUpMetric).(*prometheus.GaugeVec); ok {
		if cluster != "" && node != "" { //values are not available until first scrape of overview succeeded
			if err != nil {
				up.WithLabelValues(cluster, node, name).Set(0)
			} else {
				up.WithLabelValues(cluster, node, name).Set(1)
			}
		}
	}
	return err
}
