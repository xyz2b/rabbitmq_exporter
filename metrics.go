package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "rabbitmq"
)

var (
	extraLabels = []string{"hostname", "subsystemName", "subsystemID"}
)

func newGaugeVec(metricName string, docString string, labels []string) *prometheus.GaugeVec {
	if labels != nil {
		labels = append(labels, extraLabels...)
	} else {
		labels = extraLabels
	}

	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metricName,
			Help:      docString,
		},
		labels,
	)
}

func newGauge(metricName string, docString string) prometheus.Gauge {
	return prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metricName,
			Help:      docString,
		},
	)
}

func newDesc(metricName string, docString string, labels []string) *prometheus.Desc {
	if labels != nil {
		labels = append(labels, extraLabels...)
	} else {
		labels = extraLabels
	}

	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", metricName),
		docString,
		labels,
		nil)
}


func counterVecWithLabelValues(ctx *context.Context,v *prometheus.CounterVec, lvs ...string) prometheus.Counter {
	subsystemID := ""
	if id, ok := (*ctx).Value(subSystemID).(string); ok {
		subsystemID = id
	}
	subsystemName := ""
	if name, ok := (*ctx).Value(subSystemName).(string); ok {
		subsystemName = name
	}
	hostname := ""
	if n, ok := (*ctx).Value(hostInfo).(string); ok {
		hostname = n
	}

	if lvs != nil {
		extraLabels := []string{hostname, subsystemName, subsystemID}
		lvs = append(lvs, extraLabels...)
	} else {
		lvs = []string{hostname, subsystemName, subsystemID}
	}


	counter := v.WithLabelValues(lvs...)
	return counter
}

func gaugeVecWithLabelValues(ctx *context.Context, v *prometheus.GaugeVec, lvs ...string) prometheus.Gauge {
	subsystemID := ""
	if id, ok := (*ctx).Value(subSystemID).(string); ok {
		subsystemID = id
	}
	subsystemName := ""
	if name, ok := (*ctx).Value(subSystemName).(string); ok {
		subsystemName = name
	}
	hostname := ""
	if n, ok := (*ctx).Value(hostInfo).(string); ok {
		hostname = n
	}

	if lvs != nil {
		extraLabels := []string{hostname, subsystemName, subsystemID}
		lvs = append(lvs, extraLabels...)
	} else {
		lvs = []string{hostname, subsystemName, subsystemID}
	}

	counter := v.WithLabelValues(lvs...)
	return counter
}

func mustNewConstHistogram(
	ctx *context.Context,
	desc *prometheus.Desc,
	count uint64,
	sum float64,
	buckets map[float64]uint64,
	labelValues ...string,
) prometheus.Metric {
	subsystemID := ""
	if id, ok := (*ctx).Value(subSystemID).(string); ok {
		subsystemID = id
	}
	subsystemName := ""
	if name, ok := (*ctx).Value(subSystemName).(string); ok {
		subsystemName = name
	}
	hostname := ""
	if n, ok := (*ctx).Value(hostInfo).(string); ok {
		hostname = n
	}

	if labelValues != nil {
		extraLabels := []string{hostname, subsystemName, subsystemID}
		labelValues = append(labelValues, extraLabels...)
	} else {
		labelValues = []string{hostname, subsystemName, subsystemID}
	}

	metric := prometheus.MustNewConstHistogram(
		desc, count, sum, buckets, labelValues...
	)
	return metric
}

func mustNewConstMetric(ctx *context.Context, desc *prometheus.Desc, valueType prometheus.ValueType, value float64, labelValues ...string) prometheus.Metric {
	subsystemID := ""
	if id, ok := (*ctx).Value(subSystemID).(string); ok {
		subsystemID = id
	}
	subsystemName := ""
	if name, ok := (*ctx).Value(subSystemName).(string); ok {
		subsystemName = name
	}
	hostname := ""
	if n, ok := (*ctx).Value(hostInfo).(string); ok {
		hostname = n
	}

	if labelValues != nil {
		extraLabels := []string{hostname, subsystemName, subsystemID}
		labelValues = append(labelValues, extraLabels...)
	} else {
		labelValues = []string{hostname, subsystemName, subsystemID}
	}

	metric := prometheus.MustNewConstMetric(desc, valueType, value, labelValues...)

	return metric
}
