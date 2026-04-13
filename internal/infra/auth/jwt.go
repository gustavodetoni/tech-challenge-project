package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrUnauthorized = errors.New("unauthorized")

type Claims struct {
	Role string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

func NewManager(secret, issuer, audience string, ttl time.Duration) *Manager {
	return &Manager{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
	}
}

func (m *Manager) NewToken(subject, role string) (string, time.Time, error) {
	if len(m.secret) == 0 {
		return "", time.Time{}, fmt.Errorf("%w: JWT secret not configured", ErrUnauthorized)
	}
	if subject == "" {
		return "", time.Time{}, fmt.Errorf("subject is required")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings(nonEmptyString(m.audience)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (m *Manager) ParseAndValidate(tokenString string) (*Claims, error) {
	if len(m.secret) == 0 {
		return nil, fmt.Errorf("%w: JWT secret not configured", ErrUnauthorized)
	}
	if tokenString == "" {
		return nil, fmt.Errorf("%w: missing token", ErrUnauthorized)
	}

	claims := &Claims{}
	parserOpts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuedAt(),
	}
	if m.issuer != "" {
		parserOpts = append(parserOpts, jwt.WithIssuer(m.issuer))
	}
	if m.audience != "" {
		parserOpts = append(parserOpts, jwt.WithAudience(m.audience))
	}

	parser := jwt.NewParser(parserOpts...)
	token, err := parser.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("%w: unexpected signing method", ErrUnauthorized)
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnauthorized, err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("%w: invalid token", ErrUnauthorized)
	}

	return claims, nil
}

func nonEmptyString(v string) []string {
	if v == "" {
		return nil
	}
	return []string{v}
}
