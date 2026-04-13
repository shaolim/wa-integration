package models

import "time"

type SubmitterStatus string

const (
	SubmitterStatusInvited     SubmitterStatus = "invite"
	SubmitterStatusActive      SubmitterStatus = "active"
	SubmitterStatusRejected    SubmitterStatus = "rejected"
	SubmitterStatusDeactivated SubmitterStatus = "deactivated"
)

type Submitter struct {
	ID          int64
	FirstName   string
	LastName    string
	PhoneNumber string
	Status      SubmitterStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
