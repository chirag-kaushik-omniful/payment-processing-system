package kafka

type Config struct {
	Brokers []string
	GroupID string
}

const (
	TopicPaymentCreated          = "payment.created"
	TopicPaymentProcessing       = "payment.processing"
	TopicPaymentCompleted        = "payment.completed"
	TopicPaymentFailed           = "payment.failed"
	TopicRefundRequested         = "refund.requested"
	TopicRefundCompleted         = "refund.completed"
	TopicWalletDebited           = "wallet.debited"
	TopicWalletCredited          = "wallet.credited"
	TopicLedgerEntryCreated      = "ledger.entry.created"
	TopicLedgerCompensation      = "ledger.compensation"
	TopicNotificationSend        = "notification.send"
	TopicReconciliationStarted   = "reconciliation.started"
	TopicReconciliationCompleted = "reconciliation.completed"
	TopicFraudValidate           = "fraud.validate"
	TopicFraudValidated          = "fraud.validated"
	TopicProviderCharge          = "provider.charge"
	TopicProviderRefund          = "provider.refund"
	TopicProviderSuccess           = "provider.success"
	TopicProviderFailed            = "provider.failed"
	TopicWebhookReceived           = "webhook.received"
	TopicDisputeOpened             = "dispute.opened"
	TopicSubscriptionCharge        = "subscription.charge"
	TopicPaymentHoldCreated        = "payment.hold.created"
	TopicPaymentHoldCaptured       = "payment.hold.captured"
)

func RetryTopic(topic string) string { return topic + ".retry" }
func DLQTopic(topic string) string   { return topic + ".dlq" }
