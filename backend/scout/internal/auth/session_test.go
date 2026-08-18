package auth

import (
	"testing"
	"time"
)

func TestSessionTokenRoundTrip(t *testing.T) {
	now := time.Now()
	token := NewSessionToken("s3cr3t", now)

	if !ValidateSessionToken("s3cr3t", token, now.Add(1*time.Hour)) {
		t.Fatal("expected token to validate within session duration")
	}
	if ValidateSessionToken("s3cr3t", token, now.Add(SessionDuration+time.Minute)) {
		t.Fatal("expected token to be expired past session duration")
	}
	if ValidateSessionToken("wrong-secret", token, now) {
		t.Fatal("expected token to fail validation with wrong secret")
	}
	if ValidateSessionToken("s3cr3t", "garbage", now) {
		t.Fatal("expected garbage token to fail validation")
	}
}
