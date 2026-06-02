package auth

import (
	"testing"
	"time"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	mgr := NewJWTManager("test-secret", time.Minute, time.Hour)
	pair, err := mgr.GeneratePair("user-1", "test@example.com", []string{"user"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := mgr.Validate(pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user-1" || claims.Email != "test@example.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}
