package workers

import (
	"go-upload-download-service/utils"
	"log"
)

func StartDeleteImageWorker(rmqClient *utils.RabbitMQClient, cldClient *utils.CloudinaryClient) {
	msgs, err := rmqClient.StartConsumer()
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			publicId := string(msg.Body)
			log.Printf("Received message to delete image with public ID: %s", publicId)

			err := cldClient.DeleteFile(publicId)
			if err != nil {
				log.Printf("Error deleting image with public ID %s: %v", publicId, err)
			}
		}
	}()
	log.Printf("Started delete image worker for queue: %s", rmqClient.GetQueueName())
}
