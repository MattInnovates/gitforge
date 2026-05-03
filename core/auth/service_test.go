package auth

import (
	"testing"
	"time"
)

func TestHashAndVerifyPassword(t *testing.T) {
	svc := NewService("secret", time.Hour)

	hash, err := svc.HashPassword("p@ssw0rd")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if !svc.VerifyPassword("p@ssw0rd", hash) {
		t.Fatal("expected password verification to succeed")
	}
	if svc.VerifyPassword("wrong", hash) {
		t.Fatal("expected wrong password verification to fail")
	}
}

func TestIssueAndParseToken(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	svc := NewService("secret", 30*time.Minute)

	token, err := svc.IssueToken(42, "matt", now)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	claims, err := svc.ParseToken(token, now.Add(10*time.Minute))
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.UserID != 42 || claims.Username != "matt" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseTokenExpired(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	svc := NewService("secret", 1*time.Minute)

	token, err := svc.IssueToken(1, "u", now)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	if _, err := svc.ParseToken(token, now.Add(2*time.Minute)); err == nil {
		t.Fatal("expected expired token error")
	}
}
