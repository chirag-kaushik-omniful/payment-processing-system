package model

type WebhookPayload struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	PaymentID string                 `json:"payment_id"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type WebhookResponse struct {
	EventID   string `json:"event_id"`
	Processed bool   `json:"processed"`
	Replay    bool   `json:"replay,omitempty"`
}
