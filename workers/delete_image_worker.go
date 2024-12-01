package workers

import (
	"fmt"
	"go-upload-download-service/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/streadway/amqp"
)

func StartDeleteImageWorker(rmqClient *utils.RabbitMQClient, cldClient *utils.CloudinaryClient) {
	stopChan := make(chan struct{})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Goroutine panicked: %v", r)
			}
			log.Println("Worker goroutine exited.")
		}()

		for {

			if !prepareRabbitMQ(rmqClient) {
				time.Sleep(5 * time.Second)
				continue
			}

			msgs, err := rmqClient.StartConsumer()
			if err != nil {
				log.Printf("Failed to start consumer: %v", err)
				rmqClient.Close()
				time.Sleep(5 * time.Second)
				continue
			}

			if !processMessages(msgs, rmqClient, cldClient, stopChan) {
				log.Println("Consumer stopped unexpectedly, restarting...")
			}
		}
	}()

	log.Printf("Started delete image worker for queue: %s", rmqClient.GetQueueName())
	<-quit
	log.Println("Shutting down worker gracefully...")
	close(stopChan)
	rmqClient.Close()
}

func deleteImageWithRetries(cldClient *utils.CloudinaryClient, publicId string) error {
	maxRetries := 3
	for retries := 0; retries < maxRetries; retries++ {
		deleteErr := cldClient.DeleteFile(publicId)
		if deleteErr == nil {
			return nil
		}
		log.Printf("Error deleting image with public ID %s (attempt %d/%d): %v", publicId, retries+1, maxRetries, deleteErr)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("failed to delete image after %d attempts", maxRetries)
}

func prepareRabbitMQ(rmqClient *utils.RabbitMQClient) bool {
	if rmqClient.GetConnecionClosed() {
		log.Println("RabbitMQ connection is closed, attempting to reconnect...")
		if err := rmqClient.Reconnect(); err != nil {
			log.Printf("Failed to reconnect to RabbitMQ: %v", err)
			return false
		} else {
			log.Println("Successfully reconnected to RabbitMQ.")
		}
	}

	channel := rmqClient.GetChannel()
	if channel == nil {
		log.Println("RabbitMQ channel is closed, attempting to recreate...")
		rmqClient.Close()
		return false
	}

	if err := channel.Qos(1, 0, false); err != nil {
		log.Printf("Failed to set QoS, reconnecting RabbitMQ: %v", err)
		if err := rmqClient.Reconnect(); err != nil {
			log.Printf("Failed to reconnect to RabbitMQ: %v", err)
			return false
		}
		return prepareRabbitMQ(rmqClient)
	}

	log.Println("Started consuming messages from RabbitMQ.")
	return true
}

func processMessages(msgs <-chan amqp.Delivery, rmqClient *utils.RabbitMQClient, cldClient *utils.CloudinaryClient, stopChan chan struct{}) bool {
	log.Println("Started processing messages.")
	workerDone := make(chan struct{})
	go func() {
		defer func() {
			log.Println("Worker done processing messages.")
			close(workerDone)
		}()
		for msg := range msgs {
			log.Printf("Processing message: %s", msg.Body)
			rmqClient.UpdateLastUsed()
			publicId := string(msg.Body)
			log.Printf("Received message to delete image with public ID: %s", publicId)

			if err := deleteImageWithRetries(cldClient, publicId); err == nil {
				if err := msg.Ack(false); err != nil {
					log.Printf("Failed to ack message: %v", err)
				} else {
					log.Printf("Successfully deleted image with public ID: %s", publicId)
				}
			} else {
				log.Printf("Failed to delete image with public ID %s: %v", publicId, err)
				_ = msg.Nack(false, true)
			}
		}
	}()

	select {
	case <-stopChan:
		log.Println("Stop signal received, shutting down worker...")
		rmqClient.Close()
		return false
	case <-workerDone:
		return false
	}
}
