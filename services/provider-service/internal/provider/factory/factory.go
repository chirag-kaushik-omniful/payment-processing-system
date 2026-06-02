package factory

import (
	"fmt"
	"strings"

	"github.com/omniful/payment-platform/services/provider-service/internal/provider"
	"github.com/omniful/payment-platform/services/provider-service/internal/provider/paypal"
	"github.com/omniful/payment-platform/services/provider-service/internal/provider/razorpay"
	"github.com/omniful/payment-platform/services/provider-service/internal/provider/stripe"
)

// Factory selects a PaymentProvider implementation by name.
type Factory struct {
	providers map[string]provider.PaymentProvider
	defaultName string
}

func New() *Factory {
	stripeProvider := stripe.New()
	razorpayProvider := razorpay.New()
	paypalProvider := paypal.New()
	providers := map[string]provider.PaymentProvider{
		stripeProvider.Name():     stripeProvider,
		razorpayProvider.Name():  razorpayProvider,
		paypalProvider.Name():    paypalProvider,
	}
	return &Factory{providers: providers, defaultName: stripeProvider.Name()}
}

func (f *Factory) Get(name string) (provider.PaymentProvider, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		key = f.defaultName
	}
	p, ok := f.providers[key]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return p, nil
}
