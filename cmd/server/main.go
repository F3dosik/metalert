package main

import (
	"log"

	"github.com/F3dosik/metalert.git/internal/server"
)

func main() {
	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
