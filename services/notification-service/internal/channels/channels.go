package channels

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

type Notifier struct {
	log *zap.Logger
}

func New(log *zap.Logger) *Notifier {
	return &Notifier{log: log}
}

func (n *Notifier) SendEmail(to, subject, body string) {
	if os.Getenv("NOTIFICATION_EMAIL_ENABLED") != "true" {
		return
	}
	n.log.Info("email notification",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("body", body),
	)
}

func (n *Notifier) SendSMS(phone, message string) {
	if os.Getenv("NOTIFICATION_SMS_ENABLED") != "true" {
		return
	}
	n.log.Info("sms notification",
		zap.String("phone", phone),
		zap.String("message", message),
	)
}

func (n *Notifier) NotifyPayment(userID, paymentID, status string) {
	email := fmt.Sprintf("%s@payments.local", userID)
	n.SendEmail(email, "Payment "+status, fmt.Sprintf("Payment %s is now %s", paymentID, status))
	n.SendSMS(userID, fmt.Sprintf("Payment %s: %s", paymentID, status))
}
