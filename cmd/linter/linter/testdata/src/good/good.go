package main

import (
	"log"
	"os"
)

func main() {
	// os.Exit и log.Fatal внутри main.main — допустимо, ошибок нет.
	if len(os.Args) == 0 {
		log.Fatal("no args")
	}
	os.Exit(0)
}
