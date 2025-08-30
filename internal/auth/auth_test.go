package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJwt(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecret"
	expiresIn := time.Minute * 5
	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}
	parsedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}
	if parsedUserID != userID {
		t.Fatalf("Parsed userID does not match original. Got %v, want %v", parsedUserID, userID)
	}
}
