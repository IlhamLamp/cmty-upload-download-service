package workers

import (
	"go-upload-download-service/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartDeleteImageWorker(rmqClient *utils.RabbitMQClient, cldClient *utils.CloudinaryClient) {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := rmqClient.Reconnect(); err != nil {
			log.Fatalf("Failed to reconnect to RabbitMQ: %v", err)
		}

		if err := rmqClient.GetChannel().Qos(1, 0, false); err != nil {
			log.Fatalf("Failed to set QoS: %v", err)
		}

		log.Println("Started consuming messages from RabbitMQ.")

		msgs, err := rmqClient.StartConsumer()
		if err != nil {
			log.Fatalf("Failed to start consumer: %v", err)
		}

		for msg := range msgs {

			rmqClient.UpdateLastUsed()

			publicId := string(msg.Body)
			log.Printf("Received message to delete image with public ID: %s", publicId)

			maxRetries := 3
			retries := 0
			var deleteErr error

			for retries < maxRetries {
				deleteErr = cldClient.DeleteFile(publicId)
				if deleteErr == nil {
					if err := msg.Ack(false); err != nil {
						log.Printf("Failed to ack message: %v", err)
					} else {
						log.Printf("Successfully deleted image with public ID: %s", publicId)
					}
					break
				}

				retries++
				log.Printf("Error deleting image with public ID %s (attempt %d/%d): %v", publicId, retries, maxRetries, deleteErr)
				time.Sleep(2 * time.Second)
			}

			if deleteErr != nil {
				log.Printf("Error deleting image with public ID %s after %d attempts: %v", publicId, maxRetries, deleteErr)
				_ = msg.Nack(false, true)
			}
		}

		log.Println("RabbitMQ consumer channel closed, no more messages will be received.")
	}()
	log.Printf("Started delete image worker for queue: %s", rmqClient.GetQueueName())
	<-quit
	log.Println("Shutting down worker gracefully...")
	rmqClient.Close()
}
