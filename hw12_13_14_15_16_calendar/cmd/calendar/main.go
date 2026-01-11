package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	config := configuration.NewConfigFrom(configFile)
	logg := logger.New(config.Logger)

	store, err := storage.NewContext(ctx, config.Storage)
	if err != nil {
		logg.Error("failed to init store: " + err.Error())
	}

	calendar := app.New(logg, *store)

	serv := server.NewServer(logg, calendar, config.Server)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := serv.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	if err := serv.Start(ctx); err != nil {
		logg.Error(err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
