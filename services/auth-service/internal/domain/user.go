package domain

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Roles        []string
	CreatedAt    time.Time
}
