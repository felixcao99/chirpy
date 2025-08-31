package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestJwt(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecret"
	token, err := MakeJWT(userID, tokenSecret)
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

func TestFreshtoken(t *testing.T) {
	token, err := MakeRefreshToken()
	if err != nil {
		t.Fatalf("Failed to generate random token: %v", err)
	}

	if len(token) != 32 {
		t.Fatalf("Token %s length is incorrect. Got %d, want 32", token, len(token))
	}
}
