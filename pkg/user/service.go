package user

import (
	"errors"
	"fmt"

	"github.com/ebipenman/go-otp-auth-service/internal/database"
	"github.com/ebipenman/go-otp-auth-service/internal/model"

	"github.com/google/uuid"
)

// Service defines the business logic for user management.
type Service interface {
	GetUserByID(id uuid.UUID) (model.UserResponse, error)
	ListUsers(limit, offset int, search string) ([]model.UserResponse, int, error)
}

type userService struct {
	userRepo Repository
}

func NewService(userRepo Repository) Service {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetUserByID(id uuid.UUID) (model.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return model.UserResponse{}, fmt.Errorf("user not found: %w", err)
		}
		return model.UserResponse{}, fmt.Errorf("failed to retrieve user: %w", err)
	}
	return user.ToUserResponse(), nil
}

func (s *userService) ListUsers(limit, offset int, search string) ([]model.UserResponse, int, error) {
	users, total, err := s.userRepo.ListUsers(limit, offset, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	var userResponses []model.UserResponse
	for _, u := range users {
		userResponses = append(userResponses, u.ToUserResponse())
	}
	return userResponses, total, nil
}
