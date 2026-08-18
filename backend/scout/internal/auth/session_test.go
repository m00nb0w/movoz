package auth

import (
	"encoding/base64"
	"strings"
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

// TestSessionTokenTamperedSignature verifies that flipping a single
// character in a well-formed token's signature (as opposed to signing
// with the wrong secret, or passing a token that isn't validly
// formatted at all) is still rejected.
func TestSessionTokenTamperedSignature(t *testing.T) {
	now := time.Now()
	token := NewSessionToken("s3cr3t", now)

	raw, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("failed to decode token fixture: %v", err)
	}
	parts := strings.SplitN(string(raw), ".", 2)
	if len(parts) != 2 {
		t.Fatalf("expected token to decode into payload.signature, got %q", raw)
	}
	payload, sig := parts[0], parts[1]

	// Flip the first hex character of the signature, keeping the
	// payload (expiry) untouched, and re-encode.
	flipped := byte('a')
	if sig[0] == 'a' {
		flipped = 'b'
	}
	tamperedSig := string(flipped) + sig[1:]
	tampered := base64.URLEncoding.EncodeToString([]byte(payload + "." + tamperedSig))

	if ValidateSessionToken("s3cr3t", tampered, now) {
		t.Fatal("expected token with tampered signature to fail validation")
	}
}
