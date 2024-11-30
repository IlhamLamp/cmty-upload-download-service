package utils

import (
	"fmt"
	"go-upload-download-service/config"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
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

const idleTimeout = 5 * time.Minute

func NewRabbitMQClient(amqpURL, queueName string) (*RabbitMQClient, error) {
	conn, err := connectWithRetries(5)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open channel: %v", err)
		conn.Close()
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
		log.Printf("Failed to declare queue: %v", err)
		channel.Close()
		conn.Close()
		return nil, err
	}

	client := &RabbitMQClient{
		conn:      conn,
		channel:   channel,
		queue:     queue,
		closed:    false,
		lastUsed:  time.Now(),
		idleTimer: time.NewTimer(idleTimeout),
	}

	go client.monitorIdleState()
	return client, nil
}

func connectWithRetries(retries int) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	conf := config.LoadConfig()
	url := conf.GetRabbitMQUrl()

	config := amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
	}
	for retries > 0 {
		conn, err = amqp.DialConfig(url, config)
		if err == nil {
			return conn, nil
		}
		log.Printf("Failed to connect to RabbitMQ, retrying... (%d retries left)", retries-1)
		time.Sleep(5 * time.Second)
		retries--
	}

	log.Printf("Failed to connect to RabbitMQ after retries: %v", err)
	return nil, err
}

func (r *RabbitMQClient) monitorIdleState() {
	for {
		select {
		case <-r.idleTimer.C:
			if time.Since(r.lastUsed) > idleTimeout && !r.IsClosed() {
				log.Println("Idle timeout reached, closing RabbitMQ connection.")
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
	r.idleTimer.Reset(idleTimeout)
}

func (r *RabbitMQClient) PublishDeleteImageMessage(publicId string) error {
	log.Printf("Attempting to publish message for public ID: %s", publicId)
	if r.IsClosed() {
		log.Println("Channel is closed, attempting to reconnect...")
		if err := r.Reconnect(); err != nil {
			return fmt.Errorf("reconnection failed: %w", err)
		}
	}

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
		if err := r.Reconnect(); err != nil {
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
	return r.closed || r.conn.IsClosed()
}

func (r *RabbitMQClient) Reconnect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.closed {
		return nil
	}

	conn, err := connectWithRetries(5)
	if err != nil {
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
		log.Println("Closing RabbitMQ connection...")
		r.channel.Close()
		r.conn.Close()
		r.closed = true
		log.Println("RabbitMQ connection closed.")
	}
}
