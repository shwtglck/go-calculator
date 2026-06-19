package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"newstart/internal/api/client"
	apihandler "newstart/internal/api/handler"
)

func main() {
	storageClient, err := client.New(client.Config{
		Transport:  os.Getenv("STORAGE_TRANSPORT"),
		StorageURL: os.Getenv("STORAGE_URL"),
	})
	if err != nil {
		log.Fatalf("ошибка клиента storage-сервиса: %v", err)
	}

	h := apihandler.New(storageClient)

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
