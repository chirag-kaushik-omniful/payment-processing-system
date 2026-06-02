package featureflags

import (
	"os"
	"strings"
)

func IsEnabled(name string) bool {
	v := os.Getenv("FEATURE_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_")))
	return v == "1" || strings.EqualFold(v, "true")
}

func ProviderEnabled(provider string) bool {
	if !IsEnabled("multi_provider") {
		return provider == "" || provider == "stripe"
	}
	return true
}

func RefundSagaEnabled() bool { return !isExplicitlyDisabled("refund_saga") }
func HoldsEnabled() bool        { return IsEnabled("payment_holds") }
func AutoReconcileEnabled() bool { return IsEnabled("auto_reconcile") }

func isExplicitlyDisabled(name string) bool {
	v := os.Getenv("FEATURE_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_")))
	return v == "0" || strings.EqualFold(v, "false")
}
