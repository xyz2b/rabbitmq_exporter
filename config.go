package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	config = rabbitExporterConfig{
		RABBIT_URL:      "http://localhost:15672",
		RABBIT_USER:     "guest",
		RABBIT_PASSWORD: "guest",
		PUBLISH_PORT:    "9090",
		OUTPUT_FORMAT:   "TTY", //JSON
	}
)

type rabbitExporterConfig struct {
	RABBIT_URL      string
	RABBIT_USER     string
	RABBIT_PASSWORD string
	PUBLISH_PORT    string
	OUTPUT_FORMAT   string
}

func initConfig() {

	if url := os.Getenv("RABBIT_URL"); url != "" {
		if valid, _ := regexp.MatchString("https?://[a-zA-Z.]+", strings.ToLower(url)); valid {
			config.RABBIT_URL = url
		}

	}

	if user := os.Getenv("RABBIT_USER"); user != "" {
		config.RABBIT_USER = user
	}

	if pass := os.Getenv("RABBIT_PASSWORD"); pass != "" {
		config.RABBIT_PASSWORD = pass
	}

	if port := os.Getenv("PUBLISH_PORT"); port != "" {
		if _, err := strconv.Atoi(port); err == nil {
			config.PUBLISH_PORT = port
		}

	}
	if output := os.Getenv("OUTPUT_FORMAT"); output != "" {
		config.OUTPUT_FORMAT = output
	}
}
