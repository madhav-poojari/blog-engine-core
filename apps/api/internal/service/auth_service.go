package service

import (
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/typewriter-blog/api/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// AuthService authenticates the single admin.
//
// Credentials come from configuration rather than a database: there is no
// database any more, and the content repo is the only durable state. This also
// removes the old seeding footgun, where changing ADMIN_PASSWORD had no effect
// once the admin row existed.
type AuthService struct {
	admin        model.Admin
	passwordHash []byte
	jwtSecret    []byte
}

func NewAuthService(username, passwordHash, jwtSecret string) *AuthService {
	return &AuthService{
		admin: model.Admin{
			ID:        "admin",
			Username:  username,
			CreatedAt: time.Time{},
		},
		passwordHash: []byte(passwordHash),
		jwtSecret:    []byte(jwtSecret),
	}
}

type LoginResult struct {
	Token string      `json:"token"`
	Admin model.Admin `json:"admin"`
}

func (s *AuthService) Login(username, password string) (*LoginResult, error) {
	// Compare the username in constant time and always run bcrypt, so a wrong
	// username and a wrong password take the same path.
	userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.admin.Username)) == 1
	passErr := bcrypt.CompareHashAndPassword(s.passwordHash, []byte(password))

	if !userMatch || passErr != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.generateToken(s.admin.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	return &LoginResult{Token: token, Admin: s.admin}, nil
}

// ValidateToken parses and validates a JWT token string.
func (s *AuthService) ValidateToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

func (s *AuthService) generateToken(userID string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    "typewriter-blog",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
