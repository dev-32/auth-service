package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Role string
type Provider string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

const (
	ProviderLocal  Provider = "local"
	ProviderGoogle Provider = "google"
	ProviderGitHub Provider = "github"
)

type User struct {
	ID            uuid.UUID `db:"id"             json:"id"`
	Name          string    `db:"name"           json:"name"`
	Email         string    `db:"email"          json:"email"`
	Password      string    `db:"password"       json:"-"`
	Role          Role      `db:"role"           json:"role"`
	Provider      Provider  `db:"provider"       json:"provider"`
	ProviderID    string    `db:"provider_id"    json:"-"`
	EmailVerified bool      `db:"email_verified" json:"email_verified"`
	CreatedAt     time.Time `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"     json:"updated_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:            u.ID.String(),
		Name:          u.Name,
		Email:         u.Email,
		Role:          u.Role,
		Provider:      u.Provider,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
	}
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByProviderID(ctx context.Context, provider Provider, providerID string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]User, error)
}

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*TokenPair, error)
	Login(ctx context.Context, req LoginRequest) (*TokenPair, error)
	Logout(ctx context.Context, accessToken string) error
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req UpdateProfileRequest) (*User, error)
}

type OAuthService interface {
	GetGoogleAuthURL(state string) string
	HandleGoogleCallback(ctx context.Context, code string) (*TokenPair, error)
	GetGitHubAuthURL(state string) string
	HandleGitHubCallback(ctx context.Context, code string) (*TokenPair, error)
}
