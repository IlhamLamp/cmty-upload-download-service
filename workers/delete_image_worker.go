package workers

import (
	"go-upload-download-service/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func StartDeleteImageWorker(rmqClient *utils.RabbitMQClient, cldClient *utils.CloudinaryClient) {

	err := rmqClient.GetChannel().Qos(1, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	msgs, err := rmqClient.StartConsumer()
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for msg := range msgs {
			publicId := string(msg.Body)
			log.Printf("Received message to delete image with public ID: %s", publicId)

			err := cldClient.DeleteFile(publicId)
			if err != nil {
				log.Printf("Error deleting image with public ID %s: %v", publicId, err)
				msg.Nack(false, true)
				continue
			}

			err = msg.Ack(false)
			if err != nil {
				log.Printf("Failed to ack message: %v", err)
			}
		}
		log.Println("RabbitMQ consumer channel closed, no more messages will be received.")
	}()
	log.Printf("Started delete image worker for queue: %s", rmqClient.GetQueueName())
	<-quit
	log.Println("Shutting down worker gracefully...")
}
