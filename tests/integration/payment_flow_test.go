//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"
)

// TestPaymentFlow exercises the end-to-end payment path through the gateway.
// Requires DATABASE_URL and running infrastructure (Postgres, Redis, Kafka).
func TestPaymentFlow(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}

	ctx := context.Background()

	// testcontainers setup stub:
	//   req := testcontainers.ContainerRequest{
	//     Image:        "postgres:16-alpine",
	//     ExposedPorts: []string{"5432/tcp"},
	//     Env: map[string]string{
	//       "POSTGRES_USER":     "payment",
	//       "POSTGRES_PASSWORD": "payment",
	//       "POSTGRES_DB":       "payment_platform",
	//     },
	//     WaitingFor: wait.ForListeningPort("5432/tcp"),
	//   }
	//   postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
	//     ContainerRequest: req,
	//     Started:          true,
	//   })
	//   if err != nil {
	//     t.Fatalf("start postgres container: %v", err)
	//   }
	//   defer func() { _ = postgresC.Terminate(ctx) }()
	//
	//   host, _ := postgresC.Host(ctx)
	//   port, _ := postgresC.MappedPort(ctx, "5432")
	//   databaseURL := fmt.Sprintf(
	//     "postgres://payment:payment@%s:%s/payment_platform?sslmode=disable",
	//     host, port.Port(),
	//   )
	//   os.Setenv("DATABASE_URL", databaseURL)

	_ = ctx
	t.Log("payment flow integration test stub — wire gateway + payment-service calls here")
}

// TestDuplicateIdempotencyKey verifies repeated Idempotency-Key returns the same payment.
func TestDuplicateIdempotencyKey(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}

	// Stub: POST /payments twice with the same Idempotency-Key header and assert equal IDs.
	t.Log("duplicate idempotency key integration test stub")
}
