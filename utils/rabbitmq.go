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
		Heartbeat: 1 * time.Second,
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

	return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
}

func (r *RabbitMQClient) monitorIdleState() {
	ticker := time.NewTicker(idleTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.mu.Lock()
			if time.Since(r.lastUsed) > idleTimeout && !r.conn.IsClosed() {
				log.Println("Idle timeout reached, closing RabbitMQ connection.")
				r.Close()
			}
			if r.conn.IsClosed() {
				log.Println("RabbitMQ connection closed, attempting to reconnect...")
				if err := r.Reconnect(); err != nil {
					log.Printf("Reconnect failed: %v", err)
				} else {
					log.Println("Reconnected to RabbitMQ.")
				}
			}
			r.mu.Unlock()
		}
	}
}

func (r *RabbitMQClient) UpdateLastUsed() {
	r.lastUsed = time.Now()
	if !r.idleTimer.Stop() {
		<-r.idleTimer.C
	}
	r.idleTimer.Reset(idleTimeout)
}

func (r *RabbitMQClient) PublicMessage(message string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn == nil || r.conn.IsClosed() {
		log.Println("Connection is closed, attempting to reconnect...")
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
			Body:         []byte(message),
		},
	)
	if err != nil {
		log.Printf("Failed to publish image message: %v", err)
		return err
	}
	r.UpdateLastUsed()
	log.Printf("Message published to queue: %s", message)
	return nil
}

func (r *RabbitMQClient) StartConsumer() (<-chan amqp.Delivery, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn == nil || r.conn.IsClosed() {
		log.Println("Connection is closed, attempting to reconnect...")
		if err := r.Reconnect(); err != nil {
			return nil, fmt.Errorf("reconnection failed: %w", err)
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
		return nil, fmt.Errorf("failed to start consumer: %w", err)
	}

	r.UpdateLastUsed()
	return msgs, nil
}

func (r *RabbitMQClient) Reconnect() error {

	r.mu.Lock()
	defer r.mu.Unlock()

	oldConn := r.conn

	conn, err := connectWithRetries(5)
	if err != nil {
		return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	queue, err := channel.QueueDeclare(
		r.queue.Name,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	r.conn = conn
	r.channel = channel
	r.queue = queue

	if oldConn != nil && !oldConn.IsClosed() {
		oldConn.Close()
	}

	log.Println("Reconnected to RabbitMQ")
	return nil
}

func (r *RabbitMQClient) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.conn.IsClosed() || r.conn != nil {
		log.Println("Closing RabbitMQ connection...")
		r.channel.Close()
		r.conn.Close()
		r.conn = nil
		log.Println("RabbitMQ connection closed.")
	}
}

func (r *RabbitMQClient) GetQueueName() string {
	return r.queue.Name
}

func (r *RabbitMQClient) GetChannel() *amqp.Channel {
	return r.channel
}

func (r *RabbitMQClient) GetConnecionClosed() bool {
	return r.conn.IsClosed()
}
