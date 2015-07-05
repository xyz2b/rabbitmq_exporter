package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	config        rabbitExporterConfig
	defaultConfig = rabbitExporterConfig{
		RabbitURL:      "http://localhost:15672",
		RabbitUsername: "guest",
		RabbitPassword: "guest",
		PublishPort:    "9090",
		OutputFormat:   "TTY", //JSON
	}
)

type rabbitExporterConfig struct {
	RabbitURL      string
	RabbitUsername string
	RabbitPassword string
	PublishPort    string
	OutputFormat   string
}

func initConfig() {
	config = defaultConfig
	if url := os.Getenv("RABBIT_URL"); url != "" {
		if valid, _ := regexp.MatchString("https?://[a-zA-Z.0-9]+", strings.ToLower(url)); valid {
			config.RabbitURL = url
		}

	}

	if user := os.Getenv("RABBIT_USER"); user != "" {
		config.RabbitUsername = user
	}

	if pass := os.Getenv("RABBIT_PASSWORD"); pass != "" {
		config.RabbitPassword = pass
	}

	if port := os.Getenv("PUBLISH_PORT"); port != "" {
		if _, err := strconv.Atoi(port); err == nil {
			config.PublishPort = port
		}

	}
	if output := os.Getenv("OUTPUT_FORMAT"); output != "" {
		config.OutputFormat = output
	}
}
