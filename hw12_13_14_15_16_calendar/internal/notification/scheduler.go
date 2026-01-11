package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/broker/contracts"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/logger"
	storagecontracts "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
)

// Scheduler - планировщик уведомлений.
type Scheduler struct {
	app    Scheduled
	broker contracts.Client
	logger *logger.Logger
	config configuration.SchedulerConf
}

// NewScheduler создает новый планировщик.
func NewScheduler(
	app Scheduled,
	client contracts.Client,
	logger *logger.Logger,
	config configuration.SchedulerConf,
) *Scheduler {
	return &Scheduler{
		app:    app,
		broker: client,
		logger: logger,
		config: config,
	}
}

// Run запускает планировщик.
func (s *Scheduler) Run(ctx context.Context) error {
	s.logger.Info(
		"run event send scheduler." +
			fmt.Sprintf("\ninterval: %d", s.config.GetInterval()),
	)

	// Таймеры для периодических задач
	scheduleTicker := time.NewTicker(s.config.GetInterval())
	cleanupTicker := time.NewTicker(s.config.GetCleanupInterval())

	defer scheduleTicker.Stop()
	defer cleanupTicker.Stop()

	// Первый запуск сразу
	if err := s.processNotifications(ctx); err != nil {
		s.logger.Error("first time scheduler run error: " + err.Error())
	}

	if err := s.cleanupOldEvents(ctx); err != nil {
		s.logger.Error("first time old events clear error: " + err.Error())
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler stopped")
			return nil

		case <-scheduleTicker.C:
			if err := s.processNotifications(ctx); err != nil {
				s.logger.Error("scheduler run error: " + err.Error())
			}

		case <-cleanupTicker.C:
			if err := s.cleanupOldEvents(ctx); err != nil {
				s.logger.Error("old events clear error: " + err.Error())
			}
		}
	}
}

// processNotifications обрабатывает уведомления.
func (s *Scheduler) processNotifications(ctx context.Context) error {
	s.logger.Info("process notifications begin")

	// Получаем события для уведомлений
	events, err := s.app.GetEventsForNotification(ctx, s.config.GetInterval())
	if err != nil {
		return fmt.Errorf("events obtain: %w", err)
	}

	s.logger.Info(fmt.Sprintf("found events for notifications: %d", len(events)))

	// Создаем уведомления и публикуем в очередь
	notificationsCreated := 0
	for _, event := range events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := s.publishNotificationFor(ctx, event); err != nil {
			s.logger.Error(
				"create and pub event error." +
					fmt.Sprintf("\nevent_id: %s", event.ID) +
					fmt.Sprintf("\nerror: %s", err.Error()),
			)
			continue
		}

		notificationsCreated++
	}

	s.logger.Info(
		"process notifications completed" +
			fmt.Sprintf("\nevents_processed: %d", len(events)) +
			fmt.Sprintf("\nnotifications_created: %d", notificationsCreated),
	)

	return nil
}

// publishNotificationFor публикует уведомление о событии в очередь.
func (s *Scheduler) publishNotificationFor(ctx context.Context, event storagecontracts.Event) error {
	queueMessage := Notification{
		EventID: event.ID,
		Title:   event.Title,
		From:    event.From,
		OwnerID: event.OwnerID,
	}

	messageBytes, err := json.Marshal(queueMessage)
	if err != nil {
		return fmt.Errorf("message serialization: %w", err)
	}

	// Публикуем в очередь
	err = s.broker.PublishWithRetry(ctx, "notification", "notification", messageBytes, 3)
	if err != nil {
		return fmt.Errorf("notification publish: %w", err)
	}

	s.logger.Info(
		"event send to queue" +
			fmt.Sprintf("\nevent_id: %s", queueMessage.EventID) +
			fmt.Sprintf("\ntitle: %s", queueMessage.Title) +
			fmt.Sprintf("\nfrom: %s", queueMessage.From.Format(time.RFC3339)) +
			fmt.Sprintf("\nuser_id: %s", queueMessage.OwnerID),
	)

	return nil
}

// cleanupOldEvents очищает старые события.
func (s *Scheduler) cleanupOldEvents(_ context.Context) error {
	s.logger.Info("start deleting old events")

	from := time.Now().Add(-s.config.GetCleanupThreshold())
	deletedCount, err := s.app.DeleteEvents(from)
	if err != nil {
		return fmt.Errorf("old events clean: %w", err)
	}

	s.logger.Info(
		"old events deleted" +
			fmt.Sprintf("\ndeleted_count: %d", deletedCount) +
			fmt.Sprintf("\nfrom: %s", from.Format("2006-01-02")),
	)

	return nil
}
