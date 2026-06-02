package domain

import "time"

type LedgerEntry struct {
	ID            string
	TransactionID string
	DebitAccount  string
	CreditAccount string
	Amount        float64
	CreatedAt     time.Time
}
