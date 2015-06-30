# RabbitMQ Exporter

Prometheus exporter for RabbitMQ metrics, based on RabbitMQ HTTP API.

### Dependencies

* Prometheus [client](https://github.com/prometheus/client_golang) for Golang
* [Logging](https://github.com/Sirupsen/logrus)

### Setting up locally rabbitMQ and exporter with docker

1. Start rabbitMQ

        $ docker run -d -e RABBITMQ_NODENAME=my-rabbit --name my-rabbit -p 15672:15672 -p 9090:9090 rabbitmq:3-management

2. Start rabbitmq_exporter in container.

        $ docker run -d --net=container:my-rabbit kbudde/rabbitmq-exporter

Now your metrics are available through [http://localhost:9090/metrics](http://localhost:9090/metrics).

The rabbitmq_exporter is sharing the network interface with the rabbitmq container -> it is possible to use localhost and default user/password.
Disadvantage: you have to publish the port (9090) in the rabbitmq container.

### Configuration

Rabbitmq_exporter uses environment variables for configuration.
Settings:

* RABBIT_URL:      "http://localhost:15672",
* RABBIT_USER:     "guest",
* RABBIT_PASSWORD: "guest",
* PUBLISH_PORT:    "9090",
* OUTPUT_FORMAT:   "TTY", //change to JSON if needed

Example

    OUTPUT_FORMAT=JSON PUBLISH_PORT=9099 ./rabbitmq_exporter

### Metrics

#### Overview

Total number of:

* channels
* connections
* consumers
* exchanges
* queues

#### Queues

For each queue the number of:

* messages_ready
* messages_unacknowledged
* messages
* messages_ready_ram
* messages_unacknowledged_ram
* messages_ram
* messages_persistent
* message_bytes
* message_bytes_ready
* message_bytes_unacknowledged
* message_bytes_ram
* message_bytes_persistent
* disk_reads
* disk_writes
* consumers
* consumer_utilisation
* memory
