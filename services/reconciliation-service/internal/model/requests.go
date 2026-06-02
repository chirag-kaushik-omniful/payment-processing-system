package model

type SettlementRecord struct {
	PaymentID string  `json:"payment_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Currency  string  `json:"currency"`
}

type LedgerRecord struct {
	PaymentID string  `json:"payment_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
}

type ReconcileRequest struct {
	Settlements []SettlementRecord `json:"settlements" binding:"required"`
	Ledger      []LedgerRecord     `json:"ledger" binding:"required"`
}

type Mismatch struct {
	PaymentID        string  `json:"payment_id"`
	SettlementAmount float64 `json:"settlement_amount,omitempty"`
	LedgerAmount     float64 `json:"ledger_amount,omitempty"`
	Reason           string  `json:"reason"`
}

type ReconcileResponse struct {
	Matched   int        `json:"matched"`
	Mismatched int       `json:"mismatched"`
	Mismatches []Mismatch `json:"mismatches"`
}
