package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/broker/contracts"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/logger"
)

type Sender struct {
	broker contracts.Client
	logger *logger.Logger
	config configuration.SenderConf
}

func NewSender(client contracts.Client, logger *logger.Logger, config configuration.SenderConf) *Sender {
	return &Sender{
		broker: client,
		logger: logger,
		config: config,
	}
}

func (s *Sender) Run(ctx context.Context) error {
	s.logger.Info(
		"run event send sender." +
			fmt.Sprintf("\nworkers: %d", s.config.Workers),
	)

	// Создаем воркеры
	for i := 0; i < s.config.Workers; i++ {
		go s.worker(ctx, i)
	}

	// Ждем завершения контекста
	<-ctx.Done()
	s.logger.Info("event sender stop")
	return nil
}

func (s *Sender) worker(ctx context.Context, workerID int) {
	s.logger.Info(
		"starting a worker." +
			fmt.Sprintf("\nworker_id: %d", workerID),
	)

	handler := func(ctx context.Context, body []byte) error {
		return s.processNotification(ctx, body, workerID)
	}

	// Подписываемся на очередь
	err := s.broker.Consume(ctx, "notifications.queue", handler)
	if err != nil {
		s.logger.Error(
			"queue subscription error." +
				fmt.Sprintf("\nworker_id: %d", workerID) +
				fmt.Sprintf("\nerror: %s", err.Error()),
		)
	}
}

// processNotification обрабатывает одно уведомление.
func (s *Sender) processNotification(ctx context.Context, body []byte, workerID int) error {
	var message Notification
	if err := json.Unmarshal(body, &message); err != nil {
		return fmt.Errorf("message unserialization: %w", err)
	}

	s.logger.Info(
		"notification handling." +
			fmt.Sprintf("\nworker_id: %d", workerID) +
			fmt.Sprintf("\nevent_id: %s", message.EventID) +
			fmt.Sprintf("\nuser_id: %s", message.OwnerID),
	)

	// Отправляем уведомление
	err := s.sendNotification(ctx, message)
	if err != nil {
		s.logger.Error(
			"notification sending error." +
				fmt.Sprintf("\nworker_id: %d", workerID) +
				fmt.Sprintf("\nevent_id: %s", message.EventID) +
				fmt.Sprintf("\nerror: %s", err.Error()),
		)

		return err
	}

	s.logger.Info(
		"notification successfully sent." +
			fmt.Sprintf("\nworker_id: %d", workerID) +
			fmt.Sprintf("\nevent_id: %s", message.EventID),
	)

	return nil
}

// sendNotification отправляет уведомление.
func (s *Sender) sendNotification(_ context.Context, message Notification) error {
	s.logger.Info(
		"notification sending:" +
			fmt.Sprintf("\nevent_id: %s", message.EventID) +
			fmt.Sprintf("\ntitle: %s", message.Title) +
			fmt.Sprintf("\nfrom: %s", message.From.Format("02/Jan/2006:15:04:05 -0700")) +
			fmt.Sprintf("\nowner_id: %s", message.OwnerID),
	)

	// Имитация отправки
	time.Sleep(100 * time.Millisecond)

	return nil
}
