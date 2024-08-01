package main

import (
	"log"
	"voting-service/internal/server"

	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()

	if err := server.Run(); err != nil {
		log.Fatalf("Server failure. %s", err.Error())
	}
}
