package main

import (
	"github.com/F3dosik/metalert.git/internal/agent"
)

const serverURL = "http://localhost:8080"

func main() {
	agent.Run(serverURL)
}