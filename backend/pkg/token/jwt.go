package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Payload contains the payload data of the token
type Payload struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// NewPayload creates a new token payload
func NewPayload(userID uuid.UUID, email string, tokenType TokenType, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Payload{
		ID:        tokenID,
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}, nil
}

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) *JWTMaker {
	return &JWTMaker{secretKey: secretKey}
}

// CreateToken creates a new token for a specific user and duration
func (maker *JWTMaker) CreateToken(userID uuid.UUID, email string, tokenType TokenType, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(userID, email, tokenType, duration)
	if err != nil {
		return "", nil, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := jwtToken.SignedString([]byte(maker.secretKey))
	if err != nil {
		return "", nil, err
	}

	return tokenString, payload, nil
}

// VerifyToken checks if the token is valid or not
func (maker *JWTMaker) VerifyToken(tokenString string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(tokenString, &Payload{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok || !jwtToken.Valid {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
