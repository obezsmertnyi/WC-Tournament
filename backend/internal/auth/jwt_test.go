package auth

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

// setTestSecret installs an ephemeral random 32+ byte JWT_SECRET for the test.
func setTestSecret(t *testing.T) {
	t.Helper()
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand: %v", err)
	}
	t.Setenv("JWT_SECRET", hex.EncodeToString(b)) // 64 hex chars
}

// @trace: FR-030
func TestValidateSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if err := ValidateSecret(); err == nil {
		t.Fatalf("expected error when JWT_SECRET unset")
	}
	t.Setenv("JWT_SECRET", "short")
	if err := ValidateSecret(); err == nil {
		t.Fatalf("expected error when JWT_SECRET too short")
	}
	t.Setenv("JWT_SECRET", strings.Repeat("x", 32))
	if err := ValidateSecret(); err != nil {
		t.Fatalf("expected valid 32-byte secret, got %v", err)
	}
}

func TestSecret_NoDefault(t *testing.T) {
	if err := os.Unsetenv("JWT_SECRET"); err != nil {
		t.Fatalf("unsetenv: %v", err)
	}
	if len(secret()) != 0 {
		t.Fatalf("secret() must not fall back to a default")
	}
}

func TestIssueAndParse_RoundTrip(t *testing.T) {
	setTestSecret(t)
	tok, err := IssueToken(42, "admin", "alice")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := ParseToken(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Sub != 42 || claims.Role != "admin" || claims.Nickname != "alice" {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

func TestParse_RejectsTampered(t *testing.T) {
	setTestSecret(t)
	tok, _ := IssueToken(1, "player", "bob")
	// Flip a character in the signature segment.
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token shape")
	}
	tampered := parts[0] + "." + parts[1] + ".AAAA" + parts[2][4:]
	if _, err := ParseToken(tampered); err == nil {
		t.Fatalf("expected error for tampered signature")
	}
}

func TestParse_RejectsGarbage(t *testing.T) {
	setTestSecret(t)
	if _, err := ParseToken("not-a-token"); err == nil {
		t.Fatalf("expected error for garbage token")
	}
}
