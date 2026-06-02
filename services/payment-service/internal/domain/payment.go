package domain

import "time"

type PaymentStatus string

const (
	StatusCreated     PaymentStatus = "CREATED"
	StatusAuthorized  PaymentStatus = "AUTHORIZED"
	StatusProcessing  PaymentStatus = "PROCESSING"
	StatusSuccess     PaymentStatus = "SUCCESS"
	StatusCaptured    PaymentStatus = "CAPTURED"
	StatusFailed      PaymentStatus = "FAILED"
	StatusRefunded    PaymentStatus = "REFUNDED"
	StatusDisputed    PaymentStatus = "DISPUTED"
)

type Payment struct {
	ID              string
	UserID          string
	Amount          float64
	AmountCents     int64
	Currency        string
	BaseCurrency    string
	FXRate          float64
	Status          PaymentStatus
	Provider        string
	IdempotencyKey  string
	PaymentMethodID string
	HoldID          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Refund struct {
	ID        string
	PaymentID string
	Amount    float64
	Status    string
	CreatedAt time.Time
}

type OutboxEvent struct {
	ID          string
	AggregateID string
	EventType   string
	Payload     []byte
	Processed   bool
	CreatedAt   time.Time
}

type PaymentMethod struct {
	ID       string
	UserID   string
	Provider string
	Token    string
	LastFour string
	Brand    string
}

type Hold struct {
	ID        string
	PaymentID string
	UserID    string
	Amount    float64
	Currency  string
	Status    string
}

type Dispute struct {
	ID        string
	PaymentID string
	Provider  string
	Status    string
	Reason    string
}

type AuditEvent struct {
	ID           string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     []byte
}
