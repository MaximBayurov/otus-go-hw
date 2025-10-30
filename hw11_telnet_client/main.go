package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Timeout time.Duration
	Host    string
	Port    string
}

func parseArgs() *Config {
	config := &Config{
		Timeout: 10 * time.Second,
	}

	flag.DurationVar(&config.Timeout, "timeout", config.Timeout, "Timeout for connection")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		log.Fatalf("Usage: %s [--timeout=10s] host port\n", os.Args[0])
	}

	config.Host = args[0]
	config.Port = args[1]

	return config
}

func main() {
	config := parseArgs()

	client := NewTelnetClient(
		net.JoinHostPort(config.Host, config.Port),
		config.Timeout,
		os.Stdin,
		os.Stdout,
	)
	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)

	// Чтение из сокета
	go func() {
		err := client.Receive()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Запись в сокет
	go func() {
		err := client.Send()
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-sigCh
}
