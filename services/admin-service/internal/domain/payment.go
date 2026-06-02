package domain

import "time"

type Payment struct {
	ID        string
	UserID    string
	Amount    float64
	Currency  string
	Status    string
	Provider  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AuditEvent struct {
	ID           string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     []byte
	CreatedAt    time.Time
}
