// cmd/sender/main.go
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/broker"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/notification"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/configs/sender/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	config := configuration.NewConfigFrom(configFile)
	logg := logger.New(config.Logger)

	client := broker.NewClient(config.Broker, logg)
	if err := client.Connect(ctx); err != nil {
		logg.Fatal("broker connect:" + err.Error())
	}

	go func() {
		<-ctx.Done()

		if err := client.Close(); err != nil {
			logg.Error("failed to broker disconnection: " + err.Error())
		}
	}()

	sender := notification.NewSender(client, logg, config.Sender)
	if err := sender.Run(ctx); err != nil {
		logg.Error(err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
