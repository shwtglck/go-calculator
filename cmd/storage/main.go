package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
	log.Printf("GET  http://localhost%s/calculations", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("storage-сервис остановился: %v", err)
	}
}
