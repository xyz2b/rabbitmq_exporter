// +build integration

package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"regexp"

	"github.com/kbudde/rabbitmq_exporter/testenv"
)

func TestQueueCount(t *testing.T) {
	env := testenv.NewEnvironment(t, testenv.RabbitMQ3Latest)
	defer env.CleanUp() // do not panic or exit fatally or the container will stay up

	var exporterURL = fmt.Sprintf("http://localhost:%s/metrics", defaultConfig.PublishPort)
	var rabbitManagementURL = env.ManagementURL()
	os.Setenv("RABBIT_URL", rabbitManagementURL)
	defer os.Unsetenv("RABBIT_URL")

	go main()
	time.Sleep(2 * time.Second)
	t.Run("Ensure there are no queues", func(t *testing.T) {
		body := testenv.GetOrDie(exporterURL, 5*time.Second)

		r := regexp.MustCompile("rabbitmq_queuesTotal 0")
		if s := r.FindString(body); s == "" {
			t.Fatalf("QueueCount not found in body: %v", body)
		}
	})

	t.Run("Add one queue and check again", func(t *testing.T) {
		env.Rabbit.DeclareQueue("QueueForCheckCount", false)

		body := testenv.GetOrDie(exporterURL, 5*time.Second)

		r := regexp.MustCompile("rabbitmq_queuesTotal 1")
		if s := r.FindString(body); s == "" {
			t.Logf("body: %s", body)
			t.Fatalf("QueueCount not found ")
		}
	})
	// not implemented
	// t.Run("Add message with timestamp", func(t *testing.T) {
	// 	queue := "timestamp"
	// 	env.Rabbit.DeclareQueue(queue, true)
	// 	timestamp := time.Now()
	// 	env.Rabbit.SendMessageToQ("Test timestamp", queue, &timestamp)

	// 	// log.Println(timestamp.Unix())
	// 	body := testenv.GetOrDie(exporterURL, 5*time.Second)

	// 	r := regexp.MustCompile("rabbitmq_queue_headmessage_time{queue=\"QueueForCheckCount\",vhost=\"/\"} 123456")
	// 	if s := r.FindString(body); s == "" {
	// 		t.Logf("body: %s", body)
	// 		log.Println(rabbitManagementURL + "/api/queues")
	// 		// time.Sleep(30 * time.Minute)
	// 		// t.Fatalf("QueueCount not found ")
	// 	}
	// })
}
