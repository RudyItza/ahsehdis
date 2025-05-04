package data

import "time"

type Story struct {
	ID        int
	Title     string
	Content   string
	UserID    int
	CreatedAt time.Time
	UpdatedAt time.Time
	UserEmail string
}
