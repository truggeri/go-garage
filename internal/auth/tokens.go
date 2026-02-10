package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenKind represents the purpose of a JWT token
type TokenKind string

const (
	// AccessTokenKind is used for authenticating API requests
	AccessTokenKind TokenKind = "access"
	// RefreshTokenKind is used to obtain new access tokens
	RefreshTokenKind TokenKind = "refresh"
)

// GarageClaims extends JWT standard claims with app-specific fields
type GarageClaims struct {
	jwt.RegisteredClaims
	AccountID   string    `json:"account_id"`
	AccountName string    `json:"account_name"`
	TokenKind   TokenKind `json:"token_kind"`
}

// TokenDurations defines how long each token type remains valid
type TokenDurations struct {
	AccessValidity  time.Duration
	RefreshValidity time.Duration
}

// StandardTokenDurations returns the default token validity periods
func StandardTokenDurations() TokenDurations {
	return TokenDurations{
		AccessValidity:  15 * time.Minute,
		RefreshValidity: 7 * 24 * time.Hour,
	}
}

// TokenManager handles JWT creation and verification
type TokenManager struct {
	signingKey []byte
	durations  TokenDurations
}

// BuildTokenManager creates a new TokenManager with the given secret and durations
func BuildTokenManager(secretKey string, durations TokenDurations) (*TokenManager, error) {
	if secretKey == "" {
		return nil, errors.New("JWT signing key cannot be empty")
	}
	return &TokenManager{
		signingKey: []byte(secretKey),
		durations:  durations,
	}, nil
}

// TokenPayload contains the data needed to generate tokens
type TokenPayload struct {
	AccountID   string
	AccountName string
}

// TokenBundle contains both access and refresh tokens
type TokenBundle struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

// GenerateTokenBundle creates a new pair of access and refresh tokens
func (m *TokenManager) GenerateTokenBundle(payload TokenPayload) (*TokenBundle, error) {
	currentTime := time.Now()

	accessExpiry := currentTime.Add(m.durations.AccessValidity)
	accessToken, err := m.buildToken(payload, AccessTokenKind, accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshExpiry := currentTime.Add(m.durations.RefreshValidity)
	refreshToken, err := m.buildToken(payload, RefreshTokenKind, refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &TokenBundle{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiry,
		RefreshExpiresAt: refreshExpiry,
	}, nil
}

// buildToken creates a signed JWT with the given parameters
func (m *TokenManager) buildToken(payload TokenPayload, kind TokenKind, expiresAt time.Time) (string, error) {
	currentTime := time.Now()

	claims := GarageClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   payload.AccountID,
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(currentTime),
			Issuer:    "go-garage",
		},
		AccountID:   payload.AccountID,
		AccountName: payload.AccountName,
		TokenKind:   kind,
	}

	tokenInstance := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := tokenInstance.SignedString(m.signingKey)
	if err != nil {
		return "", err
	}
	return signedString, nil
}

// VerifiedClaims contains the validated claims from a token
type VerifiedClaims struct {
	AccountID   string
	AccountName string
	TokenKind   TokenKind
	ExpiresAt   time.Time
}

// ValidateToken parses and validates a JWT token string
func (m *TokenManager) ValidateToken(tokenString string) (*VerifiedClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &GarageClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, valid := t.Method.(*jwt.SigningMethodHMAC); !valid {
			return nil, errors.New("unexpected signing algorithm")
		}
		return m.signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	garageClaims, valid := parsedToken.Claims.(*GarageClaims)
	if !valid || !parsedToken.Valid {
		return nil, errors.New("invalid token claims")
	}

	return &VerifiedClaims{
		AccountID:   garageClaims.AccountID,
		AccountName: garageClaims.AccountName,
		TokenKind:   garageClaims.TokenKind,
		ExpiresAt:   garageClaims.ExpiresAt.Time,
	}, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (m *TokenManager) RefreshAccessToken(refreshTokenString string) (*TokenBundle, error) {
	verified, err := m.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	if verified.TokenKind != RefreshTokenKind {
		return nil, errors.New("provided token is not a refresh token")
	}

	payload := TokenPayload{
		AccountID:   verified.AccountID,
		AccountName: verified.AccountName,
	}

	return m.GenerateTokenBundle(payload)
}
