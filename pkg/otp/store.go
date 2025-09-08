package otp

import "github.com/ebipenman/go-otp-auth-service/internal/model"

// Repository defines the interface for OTP data operations.
type Repository interface {
	StoreOTP(otp model.OTP) error
	GetOTP(phoneNumber string) (model.OTP, error)
	DeleteOTP(phoneNumber string) error
}

type otpRepository struct {
	store OTPStore // Using the internal database interface
}

func NewRepository(store OTPStore) Repository {
	return &otpRepository{store: store}
}

func (r *otpRepository) StoreOTP(otp model.OTP) error {
	return r.store.StoreOTP(otp)
}

func (r *otpRepository) GetOTP(phoneNumber string) (model.OTP, error) {
	return r.store.GetOTP(phoneNumber)
}

func (r *otpRepository) DeleteOTP(phoneNumber string) error {
	return r.store.DeleteOTP(phoneNumber)
}

// OTPStore is the interface that the database implementation must satisfy.
// It's defined here for the service layer to depend on an interface from its own package.
type OTPStore interface {
	StoreOTP(otp model.OTP) error
	GetOTP(phoneNumber string) (model.OTP, error)
	DeleteOTP(phoneNumber string) error
}
