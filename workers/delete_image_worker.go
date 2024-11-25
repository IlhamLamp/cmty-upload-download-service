package workers

import (
	"go-upload-download-service/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/streadway/amqp"
)

type WorkerParams struct {
	Msgs       <-chan utils.Message
	RmqClient  *utils.RabbitMQClient
	CldClient  *utils.CloudinaryClient
	StopChan   <-chan struct{}
	WorkerDone chan struct{}
}

func StartDeleteImageWorker(rmqClient *utils.RabbitMQClient, cldClient *utils.CloudinaryClient) {
	stopChan := make(chan struct{})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			if !ensureRabbitMQConnection(rmqClient) {
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

			workerParams := WorkerParams{
				Msgs:       convertDeliveryChannel(msgs),
				RmqClient:  rmqClient,
				CldClient:  cldClient,
				StopChan:   stopChan,
				WorkerDone: make(chan struct{}),
			}

			if !processMessages(workerParams) {
				break
			}
		}

		log.Println("RabbitMQ consumer channel closed, no more messages will be received.")
	}()
	log.Printf("Started delete image worker for queue: %s", rmqClient.GetQueueName())
	<-quit
	log.Println("Shutting down worker gracefully...")
	rmqClient.Close()
}

func ensureRabbitMQConnection(rmqClient *utils.RabbitMQClient) bool {
	if rmqClient.IsClosed() {
		log.Println("RabbitMQ connection is closed, attempting to reconnect...")
		if err := rmqClient.Reconnect(); err != nil {
			log.Printf("Failed to reconnect to RabbitMQ: %v", err)
			return false
		}
	}

	if err := rmqClient.GetChannel().Qos(1, 0, false); err != nil {
		log.Printf("Failed to set QoS: %v", err)
		rmqClient.Close()
		return false
	}

	log.Println("RabbitMQ connection established and QoS set.")
	return true
}

// processMessages processes incoming messages and handles retries.
func processMessages(params WorkerParams) bool {
	workerDone := make(chan struct{})

	go func() {
		defer close(workerDone)
		for msg := range params.Msgs {
			params.RmqClient.UpdateLastUsed()
			handleMessage(msg, params.CldClient)
		}
	}()

	select {
	case <-params.StopChan:
		log.Println("Stop signal received, shutting down worker...")
		return false
	case <-params.WorkerDone:
		log.Println("Consumer stopped unexpectedly, restarting...")
		return true
	}
}

func handleMessage(msg utils.Message, cldClient *utils.CloudinaryClient) {
	publicId := string(msg.Body)
	log.Printf("Received message to delete image with public ID: %s", publicId)

	const maxRetries = 3
	retries := 0

	for retries < maxRetries {
		err := cldClient.DeleteFile(publicId)
		if err == nil {
			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("Failed to ack message: %v", ackErr)
			} else {
				log.Printf("Successfully deleted image with public ID: %s", publicId)
			}
			return
		}

		retries++
		log.Printf("Error deleting image with public ID %s (attempt %d/%d): %v", publicId, retries, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	log.Printf("Failed to delete image with public ID %s after %d attempts.", publicId, maxRetries)
	_ = msg.Nack(false, true)
}

func convertDeliveryChannel(deliveryChan <-chan amqp.Delivery) <-chan utils.Message {
	messageChan := make(chan utils.Message)

	go func() {
		defer close(messageChan)
		for delivery := range deliveryChan {
			messageChan <- utils.Message{
				Body: delivery.Body,
				Ack:  delivery.Ack,
				Nack: delivery.Nack,
			}
		}
	}()

	return messageChan
}
