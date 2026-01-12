package notification

import (
	"context"
	"time"

	storagecontracts "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
)

type Scheduled interface {
	GetEventsForNotification(context.Context, time.Duration) ([]storagecontracts.Event, error)
	DeleteEvents(startsFrom time.Time) (int, error)
}
