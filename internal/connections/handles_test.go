package connections

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSignerBindsOwnerAndRejectsTampering(t *testing.T) {
	signer, err := NewSigner([]byte(strings.Repeat("k", 32)))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	handle, err := signer.Sign(HandleClaims{UserID: "user-1", ScanID: "scan-1", NodeUUID: "d9428888-122b-11e1-b85c-61cd3cbb3210", IP: "203.0.113.8", Expires: now.Add(HandleTTL)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := signer.Verify(handle, "user-2", now); !errors.Is(err, ErrInvalidHandle) {
		t.Fatalf("expected owner mismatch, got %v", err)
	}
	altered := handle[:len(handle)-1] + "A"
	if _, err := signer.Verify(altered, "user-1", now); !errors.Is(err, ErrInvalidHandle) {
		t.Fatalf("expected tamper rejection, got %v", err)
	}
	if _, err := signer.Verify(handle, "user-1", now.Add(HandleTTL)); !errors.Is(err, ErrInvalidHandle) {
		t.Fatalf("expected expiry rejection, got %v", err)
	}
}
