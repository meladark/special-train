package main

import (
	"log"
	"os"

	"github.com/meladark/special-train/internal/api"
	"github.com/meladark/special-train/internal/app"
	"github.com/meladark/special-train/internal/service"
	"github.com/meladark/special-train/internal/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := storage.NewInMemoryStorage()
	svc := service.New(store)
	router := api.NewRouter(svc)

	srv := app.NewServer(":"+port, router)

	log.Printf("Starting server on :%s\n", port)
	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
