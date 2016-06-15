package main

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func loadMetrics(config rabbitExporterConfig, endpoint string, build func(d *json.Decoder)) {
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
		return
	}
	build(json.NewDecoder(resp.Body))
	resp.Body.Close()
}

func getQueueMap(config rabbitExporterConfig) map[string]MetricMap {
	var qm map[string]MetricMap
	loadMetrics(config, "queues", func(d *json.Decoder) {
		qm = MakeQueueMap(d)
	})
	return qm
}

func getOverviewMap(config rabbitExporterConfig) MetricMap {
	var overview MetricMap
	loadMetrics(config, "overview", func(d *json.Decoder) {
		overview = MakeMap(d)
	})
	return overview
}
