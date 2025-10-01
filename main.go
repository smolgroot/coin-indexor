package main

import (
	"log"
	"os"

	"github.com/user/coin-indexer/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}