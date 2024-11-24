package utils

import (
	"github.com/streadway/amqp"
	"go-upload-download-service/config"
	"log"
	"sync"
	"time"
)

type RabbitMQClient struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queue     amqp.Queue
	closed    bool
	mu        sync.Mutex
	lastUsed  time.Time
	idleTimer *time.Timer
}

func NewRabbitMQClient(amqpURL, queueName string) (*RabbitMQClient, error) {
	var conn *amqp.Connection
	var err error

	for retries := 5; retries > 0; retries-- {
		conn, err = amqp.Dial(amqpURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ, retrying in 5 seconds... (%d retries left)", retries-1)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after retries: %v", err)
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

	client := &RabbitMQClient{
		conn:      conn,
		channel:   channel,
		queue:     queue,
		closed:    false,
		lastUsed:  time.Now(),
		idleTimer: time.NewTimer(0),
	}

	go client.monitorIdleState()

	return client, nil
}

func (r *RabbitMQClient) monitorIdleState() {
	for {
		select {
		case <-r.idleTimer.C:
			if time.Since(r.lastUsed) > 5*time.Minute && !r.IsClosed() {
				r.Close()
			}
		}
	}
}

func (r *RabbitMQClient) UpdateLastUsed() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastUsed = time.Now()

	if !r.idleTimer.Stop() {
		<-r.idleTimer.C
	}
	r.idleTimer.Reset(5 * time.Minute)
}

func (r *RabbitMQClient) PublishDeleteImageMessage(publicId string) error {
	log.Printf("Checking rabbitmq connections .... %v", r.closed)
	if r.IsClosed() {
		err := r.Reconnect()
		if err != nil {
			log.Printf("Failed to reconnect to RabbitMQ: %v", err)
			return err
		}
	}

	log.Printf("Attempting to publish delete image message with Public ID: %s", publicId)
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
	r.UpdateLastUsed()
	log.Printf("Published delete image message with public ID: %s", publicId)
	return nil
}

func (r *RabbitMQClient) StartConsumer() (<-chan amqp.Delivery, error) {
	if r.IsClosed() {
		err := r.Reconnect()
		if err != nil {
			log.Printf("Failed to reconnect to RabbitMQ: %v", err)
			return nil, err
		}
	}
	msgs, err := r.channel.Consume(
		r.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to start consumer: %v", err)
		return nil, err
	}
	r.UpdateLastUsed()
	return msgs, nil
}

func (r *RabbitMQClient) IsClosed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closed
}

func (r *RabbitMQClient) Reconnect() error {
	conf := config.LoadConfig()
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.closed {
		return nil
	}

	url := conf.GetRabbitMQUrl()
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("Failed to reconnect to RabbitMQ: %v", err)
		return err
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open a new channel: %v", err)
		conn.Close()
		return err
	}

	r.conn = conn
	r.channel = channel
	r.closed = false

	log.Println("Successfully reconnected to RabbitMQ")
	return nil
}

func (r *RabbitMQClient) GetQueueName() string {
	return r.queue.Name
}

func (r *RabbitMQClient) GetChannel() *amqp.Channel {
	return r.channel
}

func (r *RabbitMQClient) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.closed {
		r.channel.Close()
		r.conn.Close()
		r.closed = true
		log.Println("RabbitMQ connection closed")
	}
}
