package server

import (
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	srvcontr "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/contracts"
	grpcserver "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/http"
)

func NewServer(logg srvcontr.Logger, calendar srvcontr.Application, configs configuration.ServerConf) srvcontr.Server {
	var serv srvcontr.Server
	switch configs.Type {
	case "grpc":
		serv = grpcserver.NewServer(logg, calendar, configs)
	case "http":
		serv = internalhttp.NewServer(logg, calendar, configs)
	default:
		serv = internalhttp.NewServer(logg, calendar, configs)
	}
	return serv
}
