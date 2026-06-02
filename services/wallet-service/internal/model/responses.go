package model

type WalletResponse struct {
	UserID  string  `json:"user_id"`
	Balance float64 `json:"balance"`
	Version int     `json:"version"`
}
