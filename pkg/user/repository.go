package user

import (
	"github.com/ebipenman/go-otp-auth-service/internal/model"

	"github.com/google/uuid"
)

// Repository defines the interface for user data operations.
type Repository interface {
	CreateUser(user model.User) (model.User, error)
	GetUserByID(id uuid.UUID) (model.User, error)
	GetUserByPhoneNumber(phoneNumber string) (model.User, error)
	ListUsers(limit, offset int, search string) ([]model.User, int, error)
	// Add UpdateUser, DeleteUser if needed
}

type userRepository struct {
	store UserStore // Using the internal database interface
}

func NewRepository(store UserStore) Repository {
	return &userRepository{store: store}
}

func (r *userRepository) CreateUser(user model.User) (model.User, error) {
	return r.store.CreateUser(user)
}

func (r *userRepository) GetUserByID(id uuid.UUID) (model.User, error) {
	return r.store.GetUserByID(id)
}

func (r *userRepository) GetUserByPhoneNumber(phoneNumber string) (model.User, error) {
	return r.store.GetUserByPhoneNumber(phoneNumber)
}

func (r *userRepository) ListUsers(limit, offset int, search string) ([]model.User, int, error) {
	return r.store.ListUsers(limit, offset, search)
}

// UserStore is the interface that the database implementation must satisfy.
// It's defined here for the service layer to depend on an interface from its own package.
type UserStore interface {
	CreateUser(user model.User) (model.User, error)
	GetUserByID(id uuid.UUID) (model.User, error)
	GetUserByPhoneNumber(phoneNumber string) (model.User, error)
	ListUsers(limit, offset int, search string) ([]model.User, int, error)
}
