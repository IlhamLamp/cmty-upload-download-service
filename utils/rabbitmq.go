package utils

import (
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewRabbitMQClient(amqpURL, queueName string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
		return nil, err
	}

	return &RabbitMQClient{
		conn:    conn,
		channel: channel,
		queue:   queue,
	}, nil
}

func (r *RabbitMQClient) PublishDeleteImageMessage(publicId string) error {
	err := r.channel.Publish(
		"",
		r.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent,
			Body:         []byte(publicId),
		},
	)
	if err != nil {
		log.Printf("Failed to publish delete image message: %v", err)
		return err
	}
	log.Printf("Published delete image message with public ID: %s", publicId)
	return nil
}

func (r *RabbitMQClient) StartConsumer() (<-chan amqp.Delivery, error) {
	return r.channel.Consume(
		r.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQClient) GetQueueName() string {
	return r.queue.Name
}

func (r *RabbitMQClient) GetChannel() *amqp.Channel {
	return r.channel
}

func (r *RabbitMQClient) Close() {
	r.channel.Close()
	r.conn.Close()
}
