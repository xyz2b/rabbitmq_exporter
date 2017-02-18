// Package testenv provides a rabbitmq test environment in docker for a full set of integration tests.
// Some usefull helper functions for rabbitmq interaction are included as well
package testenv

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"os"

	"log"

	"gopkg.in/ory-am/dockertest.v3"
)

//list of docker tags with rabbitmq versions
const (
	RabbitMQ3_5     = "3.5-management"
	RabbitMQ3Latest = "3-management-alpine"
)

// MaxWait is time before the docker setup will fail with timeout
var MaxWait = 20 * time.Second

//TestEnvironment contains all necessars
type TestEnvironment struct {
	t        *testing.T
	docker   *dockertest.Pool
	resource *dockertest.Resource
	Rabbit   rabbit
}

//NewEnvironment sets up a new environment. It will nlog fatal if something goes wrong
func NewEnvironment(t *testing.T, dockerTag string) TestEnvironment {
	tenv := TestEnvironment{t: t}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	tenv.docker = pool
	tenv.docker.MaxWait = MaxWait

	// pulls an image, creates a container based on it and runs it
	resource, err := tenv.docker.Run("rabbitmq", dockerTag, []string{})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	tenv.resource = resource

	checkManagementWebsite := func() error {
		_, err := GetURL(tenv.ManagementURL(), 5*time.Second)
		return err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := tenv.docker.Retry(checkManagementWebsite); err != nil {
		perr := tenv.docker.Purge(resource)
		log.Fatalf("Could not connect to docker: %s; Purge Error: %s", err, perr)
	}

	r := rabbit{}
	r.connect(tenv.AmqpURL(true))
	tenv.Rabbit = r
	return tenv
}

// CleanUp removes the container. If not called the container will run forever
func (tenv *TestEnvironment) CleanUp() {
	if err := tenv.docker.Purge(tenv.resource); err != nil {
		fmt.Fprintf(os.Stderr, "Could not purge resource: %s", err)
	}
}

//ManagementURL returns the full http url including username/password to the management api in the docker environment.
// e.g. http://guest:guest@localhost:15672
func (tenv *TestEnvironment) ManagementURL() string {
	return fmt.Sprintf("http://guest:guest@localhost:%s", tenv.resource.GetPort("15672/tcp"))
}

//AmqpURL returns the url to the rabbitmq server
// e.g. amqp://localhost:5672
func (tenv *TestEnvironment) AmqpURL(withCred bool) string {
	return fmt.Sprintf("amqp://localhost:%s", tenv.resource.GetPort("5672/tcp"))
}

// GetURL fetches the url. Will fail after timeout.
func GetURL(url string, timeout time.Duration) (string, error) {
	maxTime := time.Duration(timeout)
	client := http.Client{
		Timeout: maxTime,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return string(body), err
}

func GetOrDie(url string, timeout time.Duration) string {
	body, err := GetURL(url, timeout)
	if err != nil {
		log.Fatalf("Failed to get url in time: %s", err)
	}
	return body
}
