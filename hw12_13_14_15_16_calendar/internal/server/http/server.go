package internalhttp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/http/handlers"
)

type Server struct {
	configs configuration.ServerConf
	logger  Logger
	app     Application
}

type Logger interface {
	Debug(msg string)
	Warn(msg string)
	Info(msg string)
	Error(msg string)
	Fatal(msg string)
}

type Application interface{}

func NewServer(
	logger Logger,
	app Application,
	configs configuration.ServerConf,
) *Server {
	return &Server{
		logger:  logger,
		app:     app,
		configs: configs,
	}
}

func (s *Server) Start(_ context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", handlers.Hello)

	handler := loggingMiddleware(&s.logger)(mux)

	addr := fmt.Sprintf("%s:%d", s.configs.Host, s.configs.Port)
	s.logger.Info(fmt.Sprintf("Starting server on %s\n", addr))

	return http.ListenAndServe(addr, handler) //nolint:gosec
}

func (s *Server) Stop(_ context.Context) error {
	return nil
}
