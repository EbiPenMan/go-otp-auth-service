package otp

import (
	"crypto/rand"
	"fmt"
	"io"
)

// OTPGenerator defines the interface for generating OTPs.
type OTPGenerator interface {
	GenerateOTP() string
}

// SimpleOTPGenerator generates a 6-digit numeric OTP.
type SimpleOTPGenerator struct{}

func NewSimpleOTPGenerator() *SimpleOTPGenerator {
	return &SimpleOTPGenerator{}
}

func (g *SimpleOTPGenerator) GenerateOTP() string {
	b := make([]byte, 3) // Need 3 bytes for 6 digits (approx. 2 digits per byte in base 100)
	_, err := io.ReadAtLeast(rand.Reader, b, 3)
	if err != nil {
		// In a real application, you'd log this and potentially handle it
		// For simplicity, falling back to a non-crypto random or panicking is an option here
		// but for production, this should be robust.
		fmt.Println("WARNING: Failed to read from crypto/rand, using fallback for OTP. Error:", err)
		return "000000" // Fallback to a fixed OTP for demonstration, NOT for production
	}
	return fmt.Sprintf("%06d", (int(b[0])<<16|int(b[1])<<8|int(b[2]))%1000000)
}
