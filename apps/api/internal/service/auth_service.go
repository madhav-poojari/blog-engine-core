package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/typewriter-blog/api/internal/model"
	"github.com/typewriter-blog/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      *repository.AdminRepository
	jwtSecret []byte
}

func NewAuthService(repo *repository.AdminRepository, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

type LoginResult struct {
	Token string      `json:"token"`
	Admin model.Admin `json:"admin"`
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	admin, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("lookup admin: %w", err)
	}
	if admin == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.generateToken(admin.ID, admin.Username)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResult{Token: token, Admin: *admin}, nil
}

// SeedAdmin creates the admin account if it doesn't exist.
func (s *AuthService) SeedAdmin(ctx context.Context, username, password string) error {
	existing, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil // already seeded
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	admin := &model.Admin{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	return s.repo.Create(ctx, admin)
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

func (s *AuthService) generateToken(userID, username string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    "typewriter-blog",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
