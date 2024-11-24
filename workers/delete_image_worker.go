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

	stopChan := make(chan struct{})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			if rmqClient.IsClosed() {
				log.Println("RabbitMQ connection is closed, attempting to reconnect...")
				if err := rmqClient.Reconnect(); err != nil {
					log.Printf("Failed to reconnect to RabbitMQ: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}
			}

			if err := rmqClient.GetChannel().Qos(1, 0, false); err != nil {
				log.Printf("Failed to set QoS: %v", err)
				rmqClient.Close()
				time.Sleep(5 * time.Second)
				continue
			}

			log.Println("Started consuming messages from RabbitMQ.")

			msgs, err := rmqClient.StartConsumer()
			if err != nil {
				log.Printf("Failed to start consumer: %v", err)
				rmqClient.Close()
				time.Sleep(5 * time.Second)
				continue
			}

			workerDone := make(chan struct{})
			go func() {
				defer close(workerDone)
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
						log.Printf("Failed to delete image with public ID %s after %d attempts: %v", publicId, maxRetries, deleteErr)
						_ = msg.Nack(false, true)
					}
				}
			}()

			select {
			case <-stopChan:
				log.Println("Stop signal received, shutting down worker...")
				rmqClient.Close()
				return
			case <-workerDone:
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
