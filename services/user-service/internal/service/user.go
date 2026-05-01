package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jason/ecommerce/user-service/internal/model"
	"github.com/jason/ecommerce/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func NewUserService(repo *repository.UserRepository) *UserService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}
	return &UserService{repo: repo, jwtSecret: secret}
}

func (s *UserService) Register(ctx context.Context, email, password, name string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user := &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", fmt.Errorf("invalid credentials")
	}

	accessToken, err = s.generateToken(user.ID, 15*time.Minute)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = s.generateToken(user.ID, 7*24*time.Hour)
	return accessToken, refreshToken, err
}

func (s *UserService) GetUser(ctx context.Context, id string) (*model.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) ValidateToken(token string) (userID string, valid bool) {
	claims := &jwt.RegisteredClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !t.Valid {
		return "", false
	}
	return claims.Subject, true
}

func (s *UserService) generateToken(userID string, expiry time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.jwtSecret))
}
