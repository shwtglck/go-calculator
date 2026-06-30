package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	storagekafka "newstart/internal/storage/kafka"
	storagehandler "newstart/internal/storage/handler"
	"newstart/internal/storage/repository"
)

func main() {
	ctx := context.Background()

	repo, err := repository.NewPostgres(ctx)
	if err != nil {
		log.Fatalf("ошибка базы данных: %v", err)
	}
	defer repo.Close()

	consumer := storagekafka.NewConsumer("localhost:9092", repo)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("ошибка закрытия kafka consumer: %v", err)
		}
	}()

	go func() {
		log.Printf("kafka consumer запущен: broker=localhost:9092, topic=calculations")
		if err := consumer.Run(ctx); err != nil {
			log.Printf("kafka consumer остановился: %v", err)
		}
	}()

	h := storagehandler.New(repo)

	mux := http.NewServeMux()
	h.Register(mux)

	port := os.Getenv("STORAGE_PORT")
	if port == "" {
		port = "8081"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("storage-сервис запущен на http://localhost%s", addr)
	log.Printf("POST http://localhost%s/calculations", addr)
	log.Printf("POST http://localhost%s/register", addr)
	log.Printf("GET  http://localhost%s/calculations", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("storage-сервис остановился: %v", err)
	}
}
