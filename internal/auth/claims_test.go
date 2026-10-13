package auth

import (
	"encoding/json"
	"testing"
)

func TestClaimsDecodePhoneNumberAndLegacyPhone(t *testing.T) {
	var standard Claims
	if err := json.Unmarshal([]byte(`{"phone_number":"+255700000001"}`), &standard); err != nil {
		t.Fatal(err)
	}
	if standard.Phone != "+255700000001" {
		t.Fatalf("phone_number claim parsed as %q", standard.Phone)
	}

	var legacy Claims
	if err := json.Unmarshal([]byte(`{"phone":"+255700000002"}`), &legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.PhoneAlt != "+255700000002" {
		t.Fatalf("legacy phone claim parsed as %q", legacy.PhoneAlt)
	}
}
