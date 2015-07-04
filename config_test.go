package main

import (
	"os"
	"testing"
)

func TestEnvironmentSettingURL_HTTPS(t *testing.T) {
	newValue := "https://testURL"
	os.Setenv("RABBIT_URL", newValue)
	initConfig()
	if config.RABBIT_URL != newValue {
		t.Errorf("Expected config.RABBIT_URL to be modified. Found=%v, expected=%v", config.RABBIT_URL, newValue)
	}
}

func TestEnvironmentSettingURL_HTTP(t *testing.T) {
	newValue := "http://testURL"
	os.Setenv("RABBIT_URL", newValue)
	initConfig()
	if config.RABBIT_URL != newValue {
		t.Errorf("Expected config.RABBIT_URL to be modified. Found=%v, expected=%v", config.RABBIT_URL, newValue)
	}
}

func TestEnvironmentSettingUser(t *testing.T) {
	newValue := "username"
	os.Setenv("RABBIT_USER", newValue)
	initConfig()
	if config.RABBIT_USER != newValue {
		t.Errorf("Expected config.RABBIT_USER to be modified. Found=%v, expected=%v", config.RABBIT_USER, newValue)
	}
}

func TestEnvironmentSettingPassword(t *testing.T) {
	newValue := "password"
	os.Setenv("RABBIT_PASSWORD", newValue)
	initConfig()
	if config.RABBIT_PASSWORD != newValue {
		t.Errorf("Expected config.RABBIT_PASSWORD to be modified. Found=%v, expected=%v", config.RABBIT_PASSWORD, newValue)
	}
}

func TestEnvironmentSettingPort(t *testing.T) {
	newValue := "9091"
	os.Setenv("PUBLISH_PORT", newValue)
	initConfig()
	if config.PUBLISH_PORT != newValue {
		t.Errorf("Expected config.PUBLISH_PORT to be modified. Found=%v, expected=%v", config.PUBLISH_PORT, newValue)
	}
}

func TestEnvironmentSettingFormat(t *testing.T) {
	newValue := "json"
	os.Setenv("OUTPUT_FORMAT", newValue)
	initConfig()
	if config.OUTPUT_FORMAT != newValue {
		t.Errorf("Expected config.OUTPUT_FORMAT to be modified. Found=%v, expected=%v", config.OUTPUT_FORMAT, newValue)
	}
}

func TestConfig_Port(t *testing.T) {
	port := config.PUBLISH_PORT
	os.Setenv("PUBLISH_PORT", "noNumber")
	initConfig()
	if config.PUBLISH_PORT != port {
		t.Errorf("Invalid Portnumber. It should not be set. expected=%v,got=%v", port, config.PUBLISH_PORT)
	}
}

func TestConfig_Http_URL(t *testing.T) {
	url := config.RABBIT_URL
	os.Setenv("RABBIT_URL", "ftp://test")
	initConfig()
	if config.RABBIT_URL != url {
		t.Errorf("Invalid URL. It should start with http(s)://. expected=%v,got=%v", url, config.RABBIT_URL)
	}
}
