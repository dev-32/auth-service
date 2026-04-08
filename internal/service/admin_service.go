package service

import (
	"context"

	"github.com/dev-32/auth-service/internal/domain"
	"github.com/google/uuid"
)

type AdminService struct {
	userRepo domain.UserRepository
}

func NewAdminService(userRepo domain.UserRepository) *AdminService {
	return &AdminService{userRepo: userRepo}
}

func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.userRepo.List(ctx, limit, offset)
}

func (s *AdminService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return domain.ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, id)
}
