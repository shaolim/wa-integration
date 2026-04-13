package models

import "time"

type Chat struct {
	ID        int64
	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
