package storage

import (
	"context"
	"time"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	storagecontracts "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
	memorystorage "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/sql"
)

// EventStorage интерфейс хранилища событий.
type EventStorage interface {
	// Create создает событие.
	Create(event storagecontracts.Event) (storagecontracts.Event, error)
	// Update обновляет событие.
	Update(id string, event storagecontracts.Event) (storagecontracts.Event, error)
	// Delete удаляет событие.
	Delete(id string) error
	// GetEventsForDay возвращает события на конкретный день
	GetEventsForDay(day time.Time) ([]storagecontracts.Event, error)
	// GetEventsForWeek возвращает события на неделю
	GetEventsForWeek(startOfWeek time.Time) ([]storagecontracts.Event, error)
	// GetEventsForMonth возвращает события на месяц
	GetEventsForMonth(startOfMonth time.Time) ([]storagecontracts.Event, error)
}

func NewContext(ctx context.Context, configs configuration.StorageConf) (*EventStorage, error) {
	var storage EventStorage
	switch configs.Type {
	case "database":
		s := sqlstorage.New(configs.Database)

		if err := s.Connect(ctx); err != nil {
			return nil, err
		}

		go func() {
			<-ctx.Done()

			if err := s.Close(ctx); err != nil {
				panic(err)
			}
		}()

		storage = s
	case "in-memory":
		storage = memorystorage.New()
	default:
		storage = memorystorage.New()
	}
	return &storage, nil
}
