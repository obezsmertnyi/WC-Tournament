package auth

import (
	"strings"
	"testing"
)

func TestIssueAndParse_RoundTrip(t *testing.T) {
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
	if _, err := ParseToken("not-a-token"); err == nil {
		t.Fatalf("expected error for garbage token")
	}
}
