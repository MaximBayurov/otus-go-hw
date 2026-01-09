package internalhttp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	srvcontr "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/contracts"
	httphandlers "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/http/handlers"
)

type Server struct {
	srv     *http.Server
	configs configuration.ServerConf
	logger  srvcontr.Logger
	app     srvcontr.Application
}

func NewServer(
	logger srvcontr.Logger,
	app srvcontr.Application,
	configs configuration.ServerConf,
) *Server {
	return &Server{
		srv:     &http.Server{},
		logger:  logger,
		app:     app,
		configs: configs,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", httphandlers.Hello)

	mux.HandleFunc("PUT /events/create", httphandlers.CreateEvent(s.app, s.logger))
	mux.HandleFunc("PATCH /events/{id}", httphandlers.UpdateEvent(s.app, s.logger))
	mux.HandleFunc("DELETE /events/{id}", httphandlers.DeleteEvent(s.app, s.logger))
	mux.HandleFunc("GET /events/list/daily", httphandlers.GetEventsForDay(s.app, s.logger))
	mux.HandleFunc("GET /events/list/weekly", httphandlers.GetEventsForWeek(s.app, s.logger))
	mux.HandleFunc("GET /events/list/monthly", httphandlers.GetEventsForMonth(s.app, s.logger))

	handler := loggingMiddleware(&s.logger)(mux)

	addr := fmt.Sprintf("%s:%d", s.configs.Host, s.configs.Port)
	s.logger.Info(fmt.Sprintf("Starting HTTP server on %s\n", addr))

	s.srv.Addr = addr
	s.srv.Handler = handler
	return s.srv.ListenAndServe() //nolint:gosec
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
