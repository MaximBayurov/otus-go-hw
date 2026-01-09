package contracts

import (
	"context"
	"time"

	storagecontracts "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
)

type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Logger interface {
	Debug(msg string)
	Warn(msg string)
	Info(msg string)
	Error(msg string)
	Fatal(msg string)
}

type Application interface {
	CreateEvent(
		ctx context.Context,
		title string,
		from,
		to time.Time,
		desc,
		ownerID string,
		notify time.Time,
	) (storagecontracts.Event, error)

	UpdateEvent(
		ctx context.Context,
		id string,
		title string,
		from,
		to time.Time,
		desc string,
		notify time.Time,
	) (storagecontracts.Event, error)

	DeleteEvent(ctx context.Context, id string) error
	GetEventsForDay(ctx context.Context, day time.Time) ([]storagecontracts.Event, error)
	GetEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]storagecontracts.Event, error)
	GetEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]storagecontracts.Event, error)
}
