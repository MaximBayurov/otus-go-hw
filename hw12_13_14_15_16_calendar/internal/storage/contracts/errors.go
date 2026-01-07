package storagecontracts

import "errors"

var (
	ErrEventNotFound         = errors.New("the event is not found")
	ErrDateBusy              = errors.New("the time is already busy by another event")
	ErrInvalidDuration       = errors.New("the incorrect event duration")
	ErrPastEvent             = errors.New("cannot create an event in the past")
	ErrMissingRequiredFields = errors.New("missing required fields")
)
