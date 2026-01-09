package storagecontracts

import "time"

type Event struct {
	ID          string    `db:"id" json:"id,omitempty"`
	Title       string    `db:"title" json:"title"`
	From        time.Time `db:"start_time" json:"from"`
	To          time.Time `db:"end_time" json:"to"`
	Description string    `db:"description" json:"description"`
	OwnerID     string    `db:"owner_id" json:"ownerID"`
	Notify      time.Time `db:"notify_time" json:"notify"`
}
