package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

type QueueMetrics struct {
	messages_count                float64
	messages_ready_count          float64
	messages_unacknowledged_count float64
	consumers_count               float64
	message_bytes                 float64
	disk_reads_count              float64
	disk_writes_count             float64
}

func unpackQueueMetrics(d *json.Decoder) map[string]*QueueMetrics {
	var output []map[string]interface{}

	if err := d.Decode(&output); err != nil {
		log.Error(err)
	}
	metrics := make(map[string]*QueueMetrics)

	for _, v := range output {
		metrics[v["name"].(string)] = &QueueMetrics{
			messages_count:                v["messages"].(float64),
			messages_ready_count:          v["messages_ready"].(float64),
			messages_unacknowledged_count: v["messages_unacknowledged"].(float64),
			consumers_count:               v["consumers"].(float64),
			message_bytes:                 v["message_bytes"].(float64),
			//disk_reads_count:              v["disk_reads"].(float64),
			//disk_writes_count:             v["disk_writes"].(float64),
		}
	}
	return metrics
}

func unpackOverviewMetrics(d *json.Decoder) map[string]float64 {
	var output map[string]interface{}

	if err := d.Decode(&output); err != nil {
		log.Error(err)
	}
	metrics := make(map[string]float64)

	for k, v := range output["object_totals"].(map[string]interface{}) {
		metrics[k] = v.(float64)
	}
	return metrics
}

func getMetrics(config rabbitExporterConfig, endpoint string) *json.Decoder {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.RABBIT_URL+"/api/"+endpoint, nil)
	req.SetBasicAuth(config.RABBIT_USER, config.RABBIT_PASSWORD)

	resp, err := client.Do(req)

	if err != nil {
		log.Error(err)
	}
	return json.NewDecoder(resp.Body)
}
