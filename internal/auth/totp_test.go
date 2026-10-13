package auth

import (
	"testing"
	"time"
)

func TestVerifyTOTPAcceptsCurrentWindow(t *testing.T) {
	// RFC 6238 test secret encoded as base32.
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	if VerifyTOTP(secret, "", time.Unix(59, 0)) {
		t.Fatal("empty code must fail")
	}
	if VerifyTOTP(secret, "287082", time.Unix(59, 0)) == false {
		t.Fatal("expected RFC 6238 token")
	}
}
