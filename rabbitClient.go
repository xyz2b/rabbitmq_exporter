package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

var client = &http.Client{Timeout: 10 * time.Second}

func InitClient() {
	roots := x509.NewCertPool()

	if data, err := ioutil.ReadFile(config.CAFile); err == nil {
		if !roots.AppendCertsFromPEM(data) {
			log.WithField("filename", config.CAFile).Error("Adding certificate to rootCAs failed")
		}
	} else {
		log.Info("Using default certificate pool")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
			RootCAs:            roots,
		},
	}

	client = &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

}

func loadMetrics(config rabbitExporterConfig, endpoint string) (*json.Decoder, error) {
	req, err := http.NewRequest("GET", config.RabbitURL+"/api/"+endpoint, nil)
	req.SetBasicAuth(config.RabbitUsername, config.RabbitPassword)

	resp, err := client.Do(req)

	if err != nil || resp == nil || resp.StatusCode != 200 {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		log.WithFields(log.Fields{"error": err, "host": config.RabbitURL, "statusCode": status}).Error("Error while retrieving data from rabbitHost")
		return nil, errors.New("Error while retrieving data from rabbitHost")
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"body": string(body), "endpoint": endpoint}).Debug("Metrics loaded")

	return json.NewDecoder(bytes.NewBuffer(body)), nil
}

func getStatsInfo(config rabbitExporterConfig, apiEndpoint string) ([]StatsInfo, error) {
	var q []StatsInfo

	d, err := loadMetrics(config, apiEndpoint)
	if err != nil {
		return q, err
	}

	q = MakeStatsInfo(d)

	return q, nil
}

func getMetricMap(config rabbitExporterConfig, apiEndpoint string) (MetricMap, error) {
	var overview MetricMap

	d, err := loadMetrics(config, apiEndpoint)
	if err != nil {
		return overview, err
	}

	return MakeMap(d), nil
}
