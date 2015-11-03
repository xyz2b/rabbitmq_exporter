package main

import (
	"net/http"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

const defaultLogLevel = log.InfoLevel

func initLogger() {
	log.SetLevel(getLogLevel())
	if strings.ToUpper(config.OutputFormat) == "JSON" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		// The TextFormatter is default, you don't actually have to do this.
		log.SetFormatter(&log.TextFormatter{})
	}
}

func main() {
	initConfig()
	initLogger()
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
		"PUBLISH_PORT":  config.PublishPort,
		"RABBIT_URL":    config.RabbitURL,
		"RABBIT_USER":   config.RabbitUsername,
		"OUTPUT_FORMAT": config.OutputFormat,
		//		"RABBIT_PASSWORD": config.RABBIT_PASSWORD,
	}).Info("Starting RabbitMQ exporter")

	log.Fatal(http.ListenAndServe(":"+config.PublishPort, nil))
}

func getLogLevel() log.Level {
	var level log.Level
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	case "panic":
		level = log.PanicLevel
	case "fatal":
		level = log.FatalLevel
	default:
		level = defaultLogLevel
	}
	return level
}
