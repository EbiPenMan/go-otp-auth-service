package auth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ebipenman/go-otp-auth-service/internal/model"
	"github.com/ebipenman/go-otp-auth-service/pkg/otp"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrInvalidOTP        = errors.New("invalid or expired OTP")
	ErrUserRegistration  = errors.New("failed to register new user")
	ErrJWTGeneration     = errors.New("failed to generate JWT token")
)

// Service defines the business logic for authentication.
type Service interface {
	SendOTP(phoneNumber string) error
	VerifyOTPAndAuthenticate(phoneNumber, receivedOTP string) (string, error)
}

type authService struct {
	authRepo     Repository
	otpGenerator otp.OTPGenerator
	jwtSecret    string
}

func NewService(authRepo Repository, otpGenerator otp.OTPGenerator, jwtSecret string) Service {
	return &authService{
		authRepo:     authRepo,
		otpGenerator: otpGenerator,
		jwtSecret:    jwtSecret,
	}
}

func (s *authService) SendOTP(phoneNumber string) error {
	// 1. Check Rate Limit
	if !s.authRepo.AllowOTPRate(phoneNumber) {
		return ErrRateLimitExceeded
	}

	// 2. Generate OTP
	otpCode := s.otpGenerator.GenerateOTP()
	expiresAt := time.Now().Add(2 * time.Minute) // As per requirement

	// 3. Store OTP
	otpModel := model.OTP{
		PhoneNumber: phoneNumber,
		OTPCode:     otpCode,
		ExpiresAt:   expiresAt,
	}
	if err := s.authRepo.StoreOTP(otpModel); err != nil {
		// Log the internal error
		log.Printf("ERROR: Failed to store OTP for %s: %v", phoneNumber, err)
		return fmt.Errorf("failed to process OTP request")
	}

	// 4. Print to console (as per requirement, no SMS sending)
	log.Printf("---- OTP for %s: %s (Expires in 2 minutes) ----", phoneNumber, otpCode)

	return nil
}

func (s *authService) VerifyOTPAndAuthenticate(phoneNumber, receivedOTP string) (string, error) {
	// 1. Retrieve and Validate OTP
	storedOTP, err := s.authRepo.GetOTP(phoneNumber)
	if err != nil || storedOTP.OTPCode != receivedOTP || storedOTP.IsExpired() {
		return "", ErrInvalidOTP
	}

	// 2. OTP is valid, delete it to prevent reuse
	// We can ignore the error here for now, as the main flow can continue.
	_ = s.authRepo.DeleteOTP(phoneNumber)

	// 3. Find or Create User
	user, err := s.authRepo.GetUserByPhoneNumber(phoneNumber)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// User does not exist, register them
			newUser := model.User{PhoneNumber: phoneNumber}
			createdUser, createErr := s.authRepo.CreateUser(newUser)
			if createErr != nil {
				log.Printf("ERROR: Failed to create user for %s: %v", phoneNumber, createErr)
				return "", ErrUserRegistration
			}
			user = createdUser
			log.Printf("New user registered: %s (ID: %s)", user.PhoneNumber, user.ID)
		} else {
			// A different database error occurred
			log.Printf("ERROR: Failed to get user by phone %s: %v", phoneNumber, err)
			return "", err
		}
	} else {
		log.Printf("Existing user logged in: %s (ID: %s)", user.PhoneNumber, user.ID)
	}

	// 4. Generate JWT Token
	token, err := s.generateJWT(user.ID, user.PhoneNumber)
	if err != nil {
		log.Printf("ERROR: Failed to generate JWT for user %s: %v", user.ID, err)
		return "", ErrJWTGeneration
	}

	return token, nil
}

// generateJWT creates a new JWT token for a given user.
func (s *authService) generateJWT(userID uuid.UUID, phoneNumber string) (string, error) {
	// Create the claims
	claims := jwt.MapClaims{
		"sub":   userID.String(),                       // Subject (user ID)
		"phone": phoneNumber,                           // Custom claim
		"iat":   time.Now().Unix(),                     // Issued At
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Expiration Time (24 hours)
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
