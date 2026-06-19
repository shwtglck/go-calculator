package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"newstart/internal/api/client"
	apihandler "newstart/internal/api/handler"
	apikafka "newstart/internal/api/kafka"
)

func main() {
	storageClient, err := client.New(client.Config{
		Transport:  os.Getenv("STORAGE_TRANSPORT"),
		StorageURL: os.Getenv("STORAGE_URL"),
	})
	if err != nil {
		log.Fatalf("ошибка клиента storage-сервиса: %v", err)
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	producer := apikafka.NewProducer(broker)
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("ошибка закрытия kafka producer: %v", err)
		}
	}()

	h := apihandler.New(storageClient, producer)

	mux := http.NewServeMux()
	h.Register(mux)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("api-сервис запущен на http://localhost%s", addr)
	log.Printf("POST http://localhost%s/calculate", addr)
	log.Printf("GET  http://localhost%s/calculations", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("api-сервис остановился: %v", err)
	}
}
