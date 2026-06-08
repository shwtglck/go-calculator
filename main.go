package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()

	storage, err := NewStorage(ctx)
	if err != nil {
		log.Fatalf("ошибка базы данных: %v", err)
	}
	defer storage.Close()

	h := &handler{storage: storage}

	mux := http.NewServeMux()
	mux.HandleFunc("/calculate", h.calculate)
	mux.HandleFunc("/calculations", h.listCalculations)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("сервер запущен на http://localhost%s", addr)
	log.Printf("POST http://localhost%s/calculate", addr)
	log.Printf("GET  http://localhost%s/calculations", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("сервер остановился: %v", err)
	}
}
