package model

import (
	"time"

	"github.com/google/uuid" // Assuming you'll use UUIDs for IDs
)

// User represents a user in the system.
type User struct {
	ID          uuid.UUID `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserCreateRequest is used for creating a new user (implicitly during OTP login/reg).
type UserCreateRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// UserResponse is a DTO for user details, possibly omitting sensitive fields.
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToUserResponse converts a User model to a UserResponse DTO.
func (u *User) ToUserResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		PhoneNumber: u.PhoneNumber,
		CreatedAt:   u.CreatedAt,
	}
}
