package main

import (
	"bytes"
	"net/http"
	"os"
	"strings"

	"golang.org/x/sys/windows/svc"
	
	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultLogLevel = log.InfoLevel
	serviceName                  = "RabbitMQ_exporter"
)

type rmqExporterService struct {
	stopCh chan<- bool
}

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
	initClient()
	exporter := newExporter()
	prometheus.MustRegister(exporter)

	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatal(err)
	}
	
	stopCh := make(chan bool)
	if !isInteractive {
		go svc.Run(serviceName, &rmqExporterService{stopCh: stopCh})
	}
	
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
		"VERSION":    Version,
		"REVISION":   Revision,
		"BRANCH":     Branch,
		"BUILD_DATE": BuildDate,
		//		"RABBIT_PASSWORD": config.RABBIT_PASSWORD,
	}).Info("Starting RabbitMQ exporter")

	log.WithFields(log.Fields{
		"PUBLISH_ADDR":        config.PublishAddr,
		"PUBLISH_PORT":        config.PublishPort,
		"RABBIT_URL":          config.RabbitURL,
		"RABBIT_USER":         config.RabbitUsername,
		"OUTPUT_FORMAT":       config.OutputFormat,
		"RABBIT_CAPABILITIES": formatCapabilities(config.RabbitCapabilities),
		"RABBIT_EXPORTERS":    config.EnabledExporters,
		"CAFILE":              config.CAFile,
		"SKIPVERIFY":          config.InsecureSkipVerify,
		"SKIP_QUEUES":         config.SkipQueues.String(),
		"INCLUDE_QUEUES":      config.IncludeQueues,
		"SKIP_VHOST":          config.SkipVHost.String(),
		"INCLUDE_VHOST":       config.IncludeVHost,
		"RABBIT_TIMEOUT":      config.Timeout,
		"MAX_QUEUES":          config.MaxQueues,
		//		"RABBIT_PASSWORD": config.RABBIT_PASSWORD,
	}).Info("Active Configuration")

	go func() {
		log.Fatal(http.ListenAndServe(config.PublishAddr+":"+config.PublishPort, nil))
	}()
	for {
	if <-stopCh {
		log.Info("Shutting down RMQ exporter")
		break
	}
}
}

func getLogLevel() log.Level {
	lvl := strings.ToLower(os.Getenv("LOG_LEVEL"))
	level, err := log.ParseLevel(lvl)
	if err != nil {
		level = defaultLogLevel
	}
	return level
}

func formatCapabilities(caps rabbitCapabilitySet) string {
	var buffer bytes.Buffer
	first := true
	for k := range caps {
		if !first {
			buffer.WriteString(",")
		}
		first = false
		buffer.WriteString(string(k))
	}
	return buffer.String()
}

func (s *rmqExporterService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s.stopCh <- true
				break loop
			default:
				log.Error("unexpected control request ",c)
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}
