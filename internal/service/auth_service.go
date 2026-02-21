package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dermasense/backend/internal/config"
	"github.com/dermasense/backend/internal/domain/entity"
	"github.com/dermasense/backend/internal/domain/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// GoogleTokenInfo represents the response from Google's tokeninfo endpoint.
type GoogleTokenInfo struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	EmailVerified string `json:"email_verified"`
	Aud           string `json:"aud"`
}

// AuthService handles authentication logic.
type AuthService struct {
	userRepo  repository.UserRepository
	quotaRepo repository.QuotaRepository
	cfg       *config.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository, quotaRepo repository.QuotaRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		quotaRepo: quotaRepo,
		cfg:       cfg,
	}
}

// AuthResponse is the response returned after successful authentication.
type AuthResponse struct {
	Token string       `json:"token"`
	User  *entity.User `json:"user"`
}

// GoogleLogin verifies a Google OAuth token and returns a JWT.
func (s *AuthService) GoogleLogin(ctx context.Context, idToken string) (*AuthResponse, error) {
	tokenInfo, err := s.verifyGoogleToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid google token: %w", err)
	}

	user, err := s.userRepo.GetByEmail(ctx, tokenInfo.Email)
	if err != nil {
		user = &entity.User{
			ID:      uuid.New().String(),
			Email:   tokenInfo.Email,
			Name:    tokenInfo.Name,
			Picture: tokenInfo.Picture,
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		quota := &entity.Quota{
			UserID:         user.ID,
			RemainingScans: s.cfg.App.DefaultFreeScans,
		}
		if err := s.quotaRepo.Create(ctx, quota); err != nil {
			return nil, fmt.Errorf("failed to create quota: %w", err)
		}
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// DummyLogin logs in a predefined dummy user for development environments.
func (s *AuthService) DummyLogin(ctx context.Context) (*AuthResponse, error) {
	dummyEmail := "dummy.dev@example.com"
	dummyName := "Dev User"
	dummyPicture := "https://api.dicebear.com/7.x/avataaars/svg?seed=Dummy"

	user, err := s.userRepo.GetByEmail(ctx, dummyEmail)
	if err != nil {
		user = &entity.User{
			ID:      uuid.New().String(),
			Email:   dummyEmail,
			Name:    dummyName,
			Picture: dummyPicture,
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create dummy user: %w", err)
		}

		quota := &entity.Quota{
			UserID:         user.ID,
			RemainingScans: s.cfg.App.DefaultFreeScans,
		}
		if err := s.quotaRepo.Create(ctx, quota); err != nil {
			return nil, fmt.Errorf("failed to create dummy quota: %w", err)
		}
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate dummy token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) verifyGoogleToken(ctx context.Context, idToken string) (*GoogleTokenInfo, error) {
	url := "https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google token verification failed: %s", string(body))
	}

	var tokenInfo GoogleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, err
	}

	if s.cfg.Google.ClientID != "" && tokenInfo.Aud != s.cfg.Google.ClientID {
		return nil, fmt.Errorf("token audience mismatch")
	}

	if tokenInfo.EmailVerified != "true" {
		return nil, fmt.Errorf("email not verified")
	}

	return &tokenInfo, nil
}

func (s *AuthService) generateJWT(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"name":  user.Name,
		"exp":   time.Now().Add(s.cfg.JWT.Expiration).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

// ValidateJWT validates a JWT and returns the user ID.
func (s *AuthService) ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["sub"].(string)
		if !ok {
			return "", fmt.Errorf("invalid token claims")
		}
		return userID, nil
	}

	return "", fmt.Errorf("invalid token")
}
