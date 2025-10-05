package main

import (
	"log"

	"github.com/Xeway/amedee/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not found or couldn't be loaded: %v (continuing)", err)
	}

	server.Run()
}
