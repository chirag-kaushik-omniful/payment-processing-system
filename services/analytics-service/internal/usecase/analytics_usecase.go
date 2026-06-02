package usecase

import (
	"sync"
)

type PaymentSummary struct {
	TotalPayments   int64   `json:"total_payments"`
	CompletedCount  int64   `json:"completed_count"`
	FailedCount     int64   `json:"failed_count"`
	TotalVolume     float64 `json:"total_volume"`
	CompletedVolume float64 `json:"completed_volume"`
}

type AnalyticsUsecase struct {
	mu      sync.RWMutex
	summary PaymentSummary
}

func NewAnalyticsUsecase() *AnalyticsUsecase {
	return &AnalyticsUsecase{}
}

func (u *AnalyticsUsecase) RecordCreated(amount float64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.summary.TotalPayments++
	u.summary.TotalVolume += amount
}

func (u *AnalyticsUsecase) RecordCompleted(amount float64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.summary.CompletedCount++
	u.summary.CompletedVolume += amount
}

func (u *AnalyticsUsecase) RecordFailed() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.summary.FailedCount++
}

func (u *AnalyticsUsecase) Summary() PaymentSummary {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.summary
}
