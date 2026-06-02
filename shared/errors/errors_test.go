package errors

import (
	"errors"
	"testing"
)

func TestIsAppError(t *testing.T) {
	err := Validation("bad input")
	got, ok := IsAppError(err)
	if !ok || got.Code != CodeValidation {
		t.Fatalf("expected validation error, got %v %v", got, ok)
	}
	wrapped := errors.New("wrap")
	if _, ok := IsAppError(wrapped); ok {
		t.Fatal("expected false for generic error")
	}
}
