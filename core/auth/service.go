package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"exp"`
}

type Service struct {
	secret []byte
	ttl    time.Duration
}

func NewService(secret string, ttl time.Duration) *Service {
	if strings.TrimSpace(secret) == "" {
		secret = "change-me-gitforge-auth-secret"
	}
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}

	return &Service{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func (s *Service) VerifyPassword(password, stored string) bool {
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
	}

	// Backward-compatible fallback for old/plaintext records during migration.
	return hmac.Equal([]byte(stored), []byte(password))
}

func (s *Service) IssueToken(userID int64, username string, now time.Time) (string, error) {
	claims := Claims{
		UserID:    userID,
		Username:  username,
		ExpiresAt: now.Add(s.ttl).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal token claims: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	sig := s.sign(encodedPayload)
	encodedSig := base64.RawURLEncoding.EncodeToString(sig)

	return encodedPayload + "." + encodedSig, nil
}

func (s *Service) ParseToken(token string, now time.Time) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, ErrInvalidToken
	}

	payloadRaw, sigRaw := parts[0], parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(sigRaw)
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	expected := s.sign(payloadRaw)
	if !hmac.Equal(sig, expected) {
		return Claims{}, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(payloadRaw)
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, ErrInvalidToken
	}

	if claims.Username == "" || claims.UserID <= 0 || claims.ExpiresAt <= 0 {
		return Claims{}, ErrInvalidToken
	}
	if now.Unix() >= claims.ExpiresAt {
		return Claims{}, ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) sign(payload string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return mac.Sum(nil)
}
