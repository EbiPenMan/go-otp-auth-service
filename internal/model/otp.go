package model

import (
	"time"

	"github.com/google/uuid"
)

// OTP represents an One-Time Password record.
type OTP struct {
	ID          uuid.UUID `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	OTPCode     string    `json:"otp_code"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// IsExpired checks if the OTP has expired.
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

type SendOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
}
