package notification

import "time"

type Notification struct {
	EventID string    `json:"eventId,omitempty"`
	Title   string    `json:"title"`
	From    time.Time `json:"from"`
	OwnerID string    `json:"ownerId"`
}
