// +build integration

package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"regexp"

	"strings"

	"github.com/kbudde/rabbitmq_exporter/testenv"
)

func TestQueueCount(t *testing.T) {
	var env testenv.TestEnvironment
	var exporterURL string
	var rabbitManagementURL string

	t.Run("Setup test environment", func(t *testing.T) {
		env = testenv.NewEnvironment(t, testenv.RabbitMQ3Latest)
	})

	defer env.CleanUp() // do not panic or exit fatally or the container will stay up

	t.Run("Preparation", func(t *testing.T) {
		exporterURL = fmt.Sprintf("http://localhost:%s/metrics", defaultConfig.PublishPort)
		rabbitManagementURL = env.ManagementURL()
		os.Setenv("RABBIT_URL", rabbitManagementURL)
		defer os.Unsetenv("RABBIT_URL")

		go main()
		time.Sleep(2 * time.Second)
	})

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
			// t.Logf("body: %s", body)
			t.Fatalf("QueueCount not found ")
		}
	})

	t.Run("Add message with timestamp", func(t *testing.T) {
		queue := "timestamp"
		env.Rabbit.DeclareQueue(queue, true)
		timestamp := time.Now()
		env.Rabbit.SendMessageToQ("Test timestamp", queue, &timestamp)
		time.Sleep(1 * time.Second) // give rabbitmq management plugin a bit of time
		// log.Println(testenv.GetOrDie(env.ManagementURL()+"/api/queues", 5*time.Second))
		body := testenv.GetOrDie(exporterURL, 5*time.Second)

		search := fmt.Sprintf(`rabbitmq_queue_head_message_timestamp{queue="%s",vhost="/"} %1.9e`, queue, float64(timestamp.Unix()))
		i := strings.Index(body, search)

		if i == -1 {
			// t.Logf("body: %s", body, search)
			t.Fatalf("Timestamp not found")
		}
	})
}
