package auth

import (
	"testing"
	"time"
)

func TestSessionTokenRoundTrip(t *testing.T) {
	now := time.Now()
	token := NewSessionToken("secret", now)

	if !ValidateSessionToken("secret", token, now.Add(time.Hour)) {
		t.Fatal("expected valid token to validate")
	}
}

func TestSessionTokenExpired(t *testing.T) {
	now := time.Now()
	token := NewSessionToken("secret", now)

	if ValidateSessionToken("secret", token, now.Add(48*time.Hour)) {
		t.Fatal("expected expired token to fail validation")
	}
}

func TestSessionTokenWrongSecret(t *testing.T) {
	now := time.Now()
	token := NewSessionToken("secret", now)

	if ValidateSessionToken("wrong-secret", token, now) {
		t.Fatal("expected token signed with a different secret to fail validation")
	}
}

func TestSessionTokenGarbage(t *testing.T) {
	if ValidateSessionToken("secret", "not-a-valid-token", time.Now()) {
		t.Fatal("expected garbage token to fail validation")
	}
}
