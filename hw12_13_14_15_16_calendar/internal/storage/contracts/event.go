package storagecontracts

import "time"

type Event struct {
	ID          string    `db:"id"`
	Title       string    `db:"title"`
	From        time.Time `db:"start_time"`
	To          time.Time `db:"end_time"`
	Description string    `db:"desc"`
	OwnerID     string    `db:"owner_id"`
	Notify      time.Time `db:"notify_time"`
}
