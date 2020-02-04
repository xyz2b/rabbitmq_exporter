package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// curl like check for an http url.
// used inside docker image for checking health endpoint
// will os.Exit(0) for http response code 200. Otherwise os.Exit(1)
func curl(url string) {
	var client = &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error checking url: %v\n", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error checking url: Unexpected http code %v\n", resp.StatusCode)
		os.Exit(1)
	}
	os.Exit(0)
}
