package app

import (
	"context"
	"time"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/contracts"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage"
	storagecontracts "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
)

type App struct {
	logger Logger
	store  Storage
	contracts.Application
}

type Logger interface {
	contracts.Logger
}

type Storage interface {
	storage.EventStorage
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger: logger,
		store:  storage,
	}
}

func (a *App) CreateEvent(
	_ context.Context,
	title string,
	from,
	to time.Time,
	desc,
	ownerID string,
	notify time.Time,
) (storagecontracts.Event, error) {
	var event storagecontracts.Event
	var err error
	if event, err = a.store.Create(storagecontracts.Event{
		ID:          "",
		Title:       title,
		From:        from,
		To:          to,
		Description: desc,
		OwnerID:     ownerID,
		Notify:      notify,
	}); err != nil {
		return storagecontracts.Event{}, err
	}
	return event, nil
}

func (a *App) UpdateEvent(
	_ context.Context,
	id string,
	title string,
	from,
	to time.Time,
	desc string,
	notify time.Time,
) (storagecontracts.Event, error) {
	var event storagecontracts.Event
	var err error
	if event, err = a.store.Update(
		id,
		storagecontracts.Event{
			Title:       title,
			From:        from,
			To:          to,
			Description: desc,
			Notify:      notify,
		},
	); err != nil {
		return storagecontracts.Event{}, err
	}
	return event, nil
}

func (a *App) DeleteEvent(_ context.Context, id string) error {
	if err := a.store.Delete(id); err != nil {
		return err
	}
	return nil
}

func (a *App) GetEventsForDay(_ context.Context, day time.Time) ([]storagecontracts.Event, error) {
	var err error
	var events []storagecontracts.Event
	if events, err = a.store.GetEventsForDay(day); err != nil {
		return make([]storagecontracts.Event, 0), err
	}
	return events, nil
}

func (a *App) GetEventsForWeek(_ context.Context, startOfWeek time.Time) ([]storagecontracts.Event, error) {
	var err error
	var events []storagecontracts.Event
	if events, err = a.store.GetEventsForWeek(startOfWeek); err != nil {
		return make([]storagecontracts.Event, 0), err
	}
	return events, nil
}

func (a *App) GetEventsForMonth(_ context.Context, startOfMonth time.Time) ([]storagecontracts.Event, error) {
	var err error
	var events []storagecontracts.Event
	if events, err = a.store.GetEventsForMonth(startOfMonth); err != nil {
		return make([]storagecontracts.Event, 0), err
	}
	return events, nil
}
