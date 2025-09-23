package main

import (
	"log"

	"github.com/F3dosik/metalert.git/internal/config"
	"github.com/F3dosik/metalert.git/internal/server"
)


func main() {
	port := config.DefaultPort()
	if err := server.Run(port); err != nil {
		log.Fatal(err)
	}
}
