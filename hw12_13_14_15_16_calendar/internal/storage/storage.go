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
	Update(ID string, event storagecontracts.Event) (storagecontracts.Event, error)
	// Delete удаляет событие.
	Delete(ID string) error
	// GetEventsForDay возвращает события на конкретный день
	GetEventsForDay(day time.Time) ([]storagecontracts.Event, error)
	// GetEventsForWeek возвращает события на неделю
	GetEventsForWeek(startOfWeek time.Time) ([]storagecontracts.Event, error)
	// GetEventsForMonth возвращает события на месяц
	GetEventsForMonth(startOfMonth time.Time) ([]storagecontracts.Event, error)
}

func New(configs configuration.StorageConf) *EventStorage {
	var storage EventStorage
	switch configs.Type {
	case "database":
		storage = sqlstorage.New(configs.Database)
		break
	case "in-memory":
	default:
		storage = memorystorage.New()
		break
	}
	return &storage
}

func NewContext(ctx context.Context, configs configuration.StorageConf) (*EventStorage, error) {
	store := New(configs)

	var val interface{} = *store
	switch val := val.(type) {
	case sqlstorage.Storage:
		sqlStore, ok := val
		if !ok {
			break
		}

		if err := sqlStore.Connect(ctx); err != nil {
			return nil, err
		}

		go func() {
			<-ctx.Done()

			if err := sqlStore.Close(ctx); err != nil {
				panic(err)
			}
		}()
	default:
		break
	}
	return store, nil
}
