package main

import (
	"encoding/json"
	"net/http"
	"io"

	log "github.com/Sirupsen/logrus"
)

func getMetrics(config rabbitExporterConfig, endpoint string) io.ReadCloser {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.RabbitURL+"/api/"+endpoint, nil)
	req.SetBasicAuth(config.RabbitUsername, config.RabbitPassword)

	resp, err := client.Do(req)

	if err != nil || resp == nil || resp.StatusCode != 200 {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		log.WithFields(log.Fields{"error": err, "host": config.RabbitURL, "statusCode": status}).Error("Error while retrieving data from rabbitHost")
		return nil
	}
	return resp.Body
}

func getQueueMap(config rabbitExporterConfig) map[string]MetricMap {
	metric := getMetrics(config, "queues")
	qm := MakeQueueMap(json.NewDecoder(metric))
	metric.Close()
	return qm
}

func getOverviewMap(config rabbitExporterConfig) MetricMap {
	metric := getMetrics(config, "overview")
	overview := MakeMap(json.NewDecoder(metric))
	metric.Close()
	return overview
}
