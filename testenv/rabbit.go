package testenv

import (
	"log"

	"time"

	"github.com/streadway/amqp"
)

type rabbit struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func (r *rabbit) connect(url string) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ:%s", err)
	}
	r.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	r.channel = ch
}

func (r *rabbit) DeclareQueue(name string, durable bool) {
	_, err := r.channel.QueueDeclare(
		name,    // name
		durable, // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}
}

func (r *rabbit) SendMessageToQ(body string, routingKey string, timestamp *time.Time) {
	pub := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	}
	if timestamp != nil {
		pub.Timestamp = *timestamp
	}
	err := r.channel.Publish(
		"",         // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		pub)
	if err != nil {
		log.Fatalf("Failed to publish a message:%s . Error:%s", body, err)
	}
}
