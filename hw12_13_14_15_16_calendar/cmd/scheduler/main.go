package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/broker"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/notification"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/configs/scheduler/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	config := configuration.NewConfigFrom(configFile)
	logg := logger.New(config.Logger)

	store, err := storage.NewContext(ctx, config.Storage)
	if err != nil {
		logg.Error("failed to init store: " + err.Error())
	}

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

	var application notification.Scheduled = app.New(logg, *store)
	scheduler := notification.NewScheduler(application, client, logg, config.Scheduler)
	if err := scheduler.Run(ctx); err != nil {
		logg.Error(err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
