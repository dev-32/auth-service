package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	accessSecret  string
	refreshSecret string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

type CustomClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewManager(accessSecret, refreshSecret, accessExpiry, refershExpiry string) (*Manager, error) {
	accessDur, err := time.ParseDuration(accessExpiry)

	if err != nil {
		return nil, errors.New("invalid JWT_ACCESS_EXPIRY format")
	}

	refreshDur, err := time.ParseDuration(refershExpiry)

	if err != nil {
		return nil, errors.New("invalid JWT_REFRESH_EXPIRY format")

	}

	return &Manager{
		accessSecret:  accessSecret,
		accessExpiry:  accessDur,
		refreshExpiry: refreshDur,
		refreshSecret: refreshSecret,
	}, nil

}

func (m *Manager) GenerateAccessToken(userId, email, role string) (string, error) {
	claims := CustomClaims{
		UserID: userId,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(m.accessSecret))

}

func (m *Manager) GenerateRefreshToken(userId, email, role string) (string, error) {
	claims := CustomClaims{
		UserID: userId,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.refreshSecret))
}

func (m *Manager) parseToken(tokenStr, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(*CustomClaims)

	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (m *Manager) ParseAccessToken(tokenStr string) (*CustomClaims, error) {
	return m.parseToken(tokenStr, m.accessSecret)
}

func (m *Manager) ParseRefreshToken(tokenStr string) (*CustomClaims, error) {
	return m.parseToken(tokenStr, m.refreshSecret)
}

func (m *Manager) GetAccessExpiry() time.Duration {
	return m.accessExpiry
}

func (m *Manager) GetRefreshExpiry() time.Duration {
	return m.refreshExpiry
}
