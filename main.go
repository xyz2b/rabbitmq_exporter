package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

func init() {
	if config.OUTPUT_FORMAT == "JSON" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		// The TextFormatter is default, you don't actually have to do this.
		log.SetFormatter(&log.TextFormatter{})
	}
}

func main() {
	exporter := newExporter()
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>RabbitMQ Exporter</title></head>
             <body>
             <h1>RabbitMQ Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})

	log.WithFields(log.Fields{
		"PUBLISH_PORT":  config.PUBLISH_PORT,
		"RABBIT_URL":    config.RABBIT_URL,
		"RABBIT_USER":   config.RABBIT_USER,
		"OUTPUT_FORMAT": config.OUTPUT_FORMAT,
		//		"RABBIT_PASSWORD": config.RABBIT_PASSWORD,
	}).Info("Starting RabbitMQ exporter")

	log.Fatal(http.ListenAndServe(":"+config.PUBLISH_PORT, nil))
}
