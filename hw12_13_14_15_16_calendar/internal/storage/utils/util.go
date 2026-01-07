package storageutils

import (
	"time"
)

// NormalizeDate нормализует дату (убирает время)
func NormalizeDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
