package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github/go/gobromq/config"
	"github/go/gobromq/internal/broker"
)

func main() {
	var configPath = flag.String("config", "config.yml", "Path to the config file")
	flag.Parse()

	fmt.Println("Starting server with config:", *configPath)

	cfg, err := config.LoadConfig(*configPath)

	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	fmt.Println("Loaded config:", cfg)
	broker := broker.NewBroker(cfg)

	if err := broker.Start(); err != nil {
		fmt.Printf("Error starting broker: %v\n", err)
		os.Exit(1)
	}

	configureShutdown := make(chan os.Signal, 1)
	signal.Notify(configureShutdown, os.Interrupt, syscall.SIGTERM)
	fmt.Println(" GoBroMQ is running... Press Ctrl+C to stop")

	<-configureShutdown

	fmt.Println("Shutting down broker...")
	if err := broker.Stop(); err != nil {
		fmt.Printf("Error stopping broker: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("GoBroker stopped successfully.")
}
