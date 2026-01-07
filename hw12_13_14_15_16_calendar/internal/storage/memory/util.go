package memorystorage

import (
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
)

// eventsOverlap проверяет пересечение двух событий
func eventsOverlap(e1, e2 storagecontracts.Event) bool {
	// События одного пользователя?
	if e1.OwnerID != e2.OwnerID {
		return false
	}

	// Проверяем пересечение временных интервалов
	// События пересекаются если:
	// 1. Начало e1 внутри e2
	// 2. Конец e1 внутри e2
	// 3. e1 полностью содержит e2
	return (e1.From.Before(e2.To) && e1.To.After(e2.From)) ||
		(e2.From.Before(e1.To) && e2.To.After(e1.From))
}
