package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dev-32/auth-service/internal/domain"
	jwtpkg "github.com/dev-32/auth-service/pkg/jwt"
	"github.com/dev-32/auth-service/pkg/password"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo   domain.UserRepository
	tokenRepo  domain.TokenRepository
	tokenCache domain.TokenCache
	jwt        *jwtpkg.Manager
}

func NewAuthService(userRepo domain.UserRepository, tokenRepo domain.TokenRepository, tokenCache domain.TokenCache, jwt *jwtpkg.Manager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		tokenCache: tokenCache,
		jwt:        jwt,
	}
}

func (s *AuthService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.TokenPair, error) {
	// checking email is already registered
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, fmt.Errorf("register: %w", err)
	}

	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// hash password
	hashPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	user := &domain.User{
		Name:          req.Name,
		Email:         req.Email,
		Password:      hashPassword,
		Role:          domain.RoleUser,
		Provider:      domain.ProviderLocal,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	slog.Info("user registered", "user_id", user.ID, "email", user.Email)

	return s.generateTokenPair(ctx, user)

}

func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest) (*domain.TokenPair, error) {
	// finding user
	user, err := s.userRepo.FindByEmail(ctx, req.Email)

	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.Provider != domain.ProviderLocal {
		return nil, domain.ErrInvalidCredentials
	}

	// password verify
	if !password.Verify(req.Password, user.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	slog.Info("user logged in", "user_id", user.ID, "email", user.Email)

	return s.generateTokenPair(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.jwt.ParseAccessToken(accessToken)
	if err != nil {
		return domain.ErrInvalidToken
	}

	remaining := time.Until(claims.ExpiresAt.Time)

	if remaining > 0 {
		if err := s.tokenCache.BlacklistToken(ctx, accessToken, remaining); err != nil {
			return fmt.Errorf("logout: %w", err)
		}
	}

	userId, err := uuid.Parse(claims.UserID)
	if err != nil {
		return domain.ErrInvalidToken
	}

	if err := s.tokenRepo.DeleteByUserID(ctx, userId); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	slog.Info("user logged out", "user_id", claims.UserID)
	return nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, id uuid.UUID, req domain.UpdateProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	user.Name = req.Name
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}

	slog.Info("profile updated", "user_id", id)
	return user, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	// parsing refresh token
	claims, err := s.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	stored, err := s.tokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if time.Now().After(stored.ExpiresAt) {
		_ = s.tokenRepo.DeleteByToken(ctx, refreshToken)
		return nil, domain.ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	if err := s.tokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("refresh: %w", err)
	}

	slog.Info("token refreshed", "user_id", user.ID)

	return s.generateTokenPair(ctx, user)

}

func (s *AuthService) generateTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {

	accessToken, err := s.jwt.GenerateAccessToken(user.ID.String(), user.Email, string(user.Role))

	if err != nil {
		return nil, fmt.Errorf("error generating access Token: %w", err)
	}

	refershToken, err := s.jwt.GenerateRefreshToken(user.ID.String(), user.Email, string(user.Role))

	if err != nil {
		return nil, fmt.Errorf("error generating refresh Token: %w", err)
	}

	storedToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refershToken,
		ExpiresAt: time.Now().Add(s.jwt.GetRefreshExpiry()),
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, storedToken); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	tokenPair := &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refershToken,
	}

	return tokenPair, nil

}
