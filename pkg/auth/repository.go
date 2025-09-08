package auth

import (
	"errors"

	"github.com/ebipenman/go-otp-auth-service/internal/database"
	"github.com/ebipenman/go-otp-auth-service/internal/model"
	"github.com/ebipenman/go-otp-auth-service/pkg/otp"
	"github.com/ebipenman/go-otp-auth-service/pkg/user"
)

var ErrUserNotFound = errors.New("user not found")

// CHANGE 1: Define a RateLimiter interface.
// This decouples the auth repository from any specific rate limiter implementation.
// Any struct that has an `Allow(key string) bool` method will satisfy this interface.
type RateLimiter interface {
	Allow(key string) bool
}

// Repository defines the interface for authentication-related data operations.
type Repository interface {
	GetUserByPhoneNumber(phoneNumber string) (model.User, error)
	CreateUser(user model.User) (model.User, error)
	StoreOTP(otp model.OTP) error
	GetOTP(phoneNumber string) (model.OTP, error)
	DeleteOTP(phoneNumber string) error
	AllowOTPRate(phoneNumber string) bool
}

type authRepository struct {
	userRepo user.Repository
	otpRepo  otp.Repository
	// CHANGE 2: Depend on the interface, not the concrete type.
	rateLimiter RateLimiter
}

// CHANGE 3: The function now accepts the interface.
// This makes it more flexible and testable.
func NewRepository(userRepo user.Repository, otpRepo otp.Repository, rateLimiter RateLimiter) Repository {
	return &authRepository{
		userRepo:    userRepo,
		otpRepo:     otpRepo,
		rateLimiter: rateLimiter,
	}
}

func (r *authRepository) GetUserByPhoneNumber(phoneNumber string) (model.User, error) {
	u, err := r.userRepo.GetUserByPhoneNumber(phoneNumber)
	if errors.Is(err, database.ErrNotFound) {
		return model.User{}, ErrUserNotFound // Translate internal error to a domain-specific one
	}
	return u, err
}

func (r *authRepository) CreateUser(user model.User) (model.User, error) {
	return r.userRepo.CreateUser(user)
}

func (r *authRepository) StoreOTP(otp model.OTP) error {
	return r.otpRepo.StoreOTP(otp)
}

func (r *authRepository) GetOTP(phoneNumber string) (model.OTP, error) {
	return r.otpRepo.GetOTP(phoneNumber)
}

func (r *authRepository) DeleteOTP(phoneNumber string) error {
	return r.otpRepo.DeleteOTP(phoneNumber)
}

// This method works exactly as before because the interface guarantees
// that a `.Allow()` method exists.
func (r *authRepository) AllowOTPRate(phoneNumber string) bool {
	return r.rateLimiter.Allow(phoneNumber)
}
