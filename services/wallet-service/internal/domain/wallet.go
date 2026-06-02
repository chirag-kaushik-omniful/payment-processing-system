package domain

import "time"

type Wallet struct {
	UserID    string
	Balance   float64
	Version   int
	UpdatedAt time.Time
}
