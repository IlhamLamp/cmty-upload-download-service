package config

import (
	"log"

	"github.com/streadway/amqp"
)

func InitRabbitMQ(url string) *amqp.Connection {
    conn, err := amqp.Dial(url)
    if err != nil {
        log.Fatalf("Failed to connect RabbitMQ: %v", err)
    }
    return conn;
}