package broker

import (
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/broker/contracts"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/broker/rabbitmq"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/logger"
)

// NewClient создает новый клиент.
func NewClient(config configuration.BrokerConf, logger *logger.Logger) contracts.Client {
	var client contracts.Client = rabbitmq.NewRabbitMQClient(config, logger)

	return client
}
