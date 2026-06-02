package usecase

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/omniful/payment-platform/shared/featureflags"
	"github.com/omniful/payment-platform/services/reconciliation-service/internal/model"
)

type ReconciliationUsecase struct {
	ledgerURL string
	client    *http.Client
}

func NewReconciliationUsecase() *ReconciliationUsecase {
	return &ReconciliationUsecase{
		ledgerURL: getEnv("LEDGER_SERVICE_URL", "http://localhost:8005"),
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (u *ReconciliationUsecase) Reconcile(req model.ReconcileRequest) model.ReconcileResponse {
	ledgerByPayment := make(map[string]float64, len(req.Ledger))
	for _, entry := range req.Ledger {
		ledgerByPayment[entry.PaymentID] = entry.Amount
	}
	return u.compare(req.Settlements, ledgerByPayment)
}

func (u *ReconciliationUsecase) AutoReconcile(settlements []model.SettlementRecord) (model.ReconcileResponse, error) {
	ledger, err := u.fetchLedger()
	if err != nil {
		return model.ReconcileResponse{}, err
	}
	return u.Reconcile(model.ReconcileRequest{Settlements: settlements, Ledger: ledger}), nil
}

func (u *ReconciliationUsecase) fetchLedger() ([]model.LedgerRecord, error) {
	resp, err := u.client.Get(u.ledgerURL + "/ledger/export?limit=5000")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Ledger []struct {
			PaymentID string  `json:"payment_id"`
			Amount    float64 `json:"amount"`
		} `json:"ledger"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	var records []model.LedgerRecord
	for _, e := range out.Ledger {
		records = append(records, model.LedgerRecord{PaymentID: e.PaymentID, Amount: e.Amount})
	}
	return records, nil
}

func (u *ReconciliationUsecase) compare(settlements []model.SettlementRecord, ledgerByPayment map[string]float64) model.ReconcileResponse {
	var resp model.ReconcileResponse
	seen := make(map[string]bool)

	for _, settlement := range settlements {
		seen[settlement.PaymentID] = true
		ledgerAmount, ok := ledgerByPayment[settlement.PaymentID]
		if !ok {
			resp.Mismatched++
			resp.Mismatches = append(resp.Mismatches, model.Mismatch{
				PaymentID:        settlement.PaymentID,
				SettlementAmount: settlement.Amount,
				Reason:           "missing in ledger",
			})
			continue
		}
		if ledgerAmount != settlement.Amount {
			resp.Mismatched++
			resp.Mismatches = append(resp.Mismatches, model.Mismatch{
				PaymentID:        settlement.PaymentID,
				SettlementAmount: settlement.Amount,
				LedgerAmount:     ledgerAmount,
				Reason:           "amount mismatch",
			})
			continue
		}
		resp.Matched++
	}

	for paymentID, ledgerAmount := range ledgerByPayment {
		if seen[paymentID] {
			continue
		}
		resp.Mismatched++
		resp.Mismatches = append(resp.Mismatches, model.Mismatch{
			PaymentID:    paymentID,
			LedgerAmount: ledgerAmount,
			Reason:       "missing in settlement",
		})
	}
	return resp
}

func (u *ReconciliationUsecase) RunScheduled(settlements []model.SettlementRecord) model.ReconcileResponse {
	if !featureflags.AutoReconcileEnabled() {
		return model.ReconcileResponse{}
	}
	resp, err := u.AutoReconcile(settlements)
	if err != nil {
		return model.ReconcileResponse{Mismatched: -1, Mismatches: []model.Mismatch{{Reason: fmt.Sprintf("auto reconcile failed: %v", err)}}}
	}
	return resp
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
