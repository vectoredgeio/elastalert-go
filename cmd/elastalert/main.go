package main

import (
	"elastalert-go/config"
	"elastalert-go/processor"
	"log"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Start the main processor
	processor.Start(cfg)
}
