package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

const sessionDuration = 24 * time.Hour

func NewSessionToken(secret string, now time.Time) string {
	expiry := now.Add(sessionDuration).Unix()
	payload := strconv.FormatInt(expiry, 10)
	sig := sign(secret, payload)
	return base64.URLEncoding.EncodeToString([]byte(payload + "." + sig))
}

func ValidateSessionToken(secret, token string, now time.Time) bool {
	raw, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false
	}
	parts := strings.SplitN(string(raw), ".", 2)
	if len(parts) != 2 {
		return false
	}
	payload, sig := parts[0], parts[1]

	if !hmac.Equal([]byte(sig), []byte(sign(secret, payload))) {
		return false
	}

	expiry, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return false
	}
	return now.Unix() < expiry
}

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
