package database

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ebipenman/go-otp-auth-service/internal/model"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// In-memory User Store
type InMemoryUserStore struct {
	users      map[uuid.UUID]model.User
	phoneIndex map[string]uuid.UUID // For fast lookup by phone number
	mu         sync.RWMutex
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users:      make(map[uuid.UUID]model.User),
		phoneIndex: make(map[string]uuid.UUID),
	}
}

func (s *InMemoryUserStore) CreateUser(user model.User) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.phoneIndex[user.PhoneNumber]; exists {
		return model.User{}, fmt.Errorf("%w: user with phone number %s", ErrAlreadyExists, user.PhoneNumber)
	}

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	s.users[user.ID] = user
	s.phoneIndex[user.PhoneNumber] = user.ID
	return user, nil
}

func (s *InMemoryUserStore) GetUserByID(id uuid.UUID) (model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	if !ok {
		return model.User{}, fmt.Errorf("%w: user with ID %s", ErrNotFound, id)
	}
	return user, nil
}

func (s *InMemoryUserStore) GetUserByPhoneNumber(phoneNumber string) (model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.phoneIndex[phoneNumber]
	if !ok {
		return model.User{}, fmt.Errorf("%w: user with phone number %s", ErrNotFound, phoneNumber)
	}
	user, ok := s.users[id]
	if !ok { // Should not happen if index is consistent
		return model.User{}, fmt.Errorf("%w: user with ID %s (from phone index)", ErrNotFound, id)
	}
	return user, nil
}

func (s *InMemoryUserStore) ListUsers(limit, offset int, search string) ([]model.User, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filteredUsers []model.User
	for _, user := range s.users {
		if search == "" || user.PhoneNumber == search { // Simple search by phone number
			filteredUsers = append(filteredUsers, user)
		}
	}

	total := len(filteredUsers)
	if offset >= total {
		return []model.User{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filteredUsers[offset:end], total, nil
}

// In-memory OTP Store
type InMemoryOTPStore struct {
	otps map[string]model.OTP // Keyed by phone number
	mu   sync.RWMutex
}

func NewInMemoryOTPStore() *InMemoryOTPStore {
	return &InMemoryOTPStore{
		otps: make(map[string]model.OTP),
	}
}

func (s *InMemoryOTPStore) StoreOTP(otp model.OTP) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	otp.ID = uuid.New() // Assign an ID, though not used as key
	otp.CreatedAt = time.Now()
	s.otps[otp.PhoneNumber] = otp
	return nil
}

func (s *InMemoryOTPStore) GetOTP(phoneNumber string) (model.OTP, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	otp, ok := s.otps[phoneNumber]
	if !ok {
		return model.OTP{}, fmt.Errorf("%w: OTP for phone number %s", ErrNotFound, phoneNumber)
	}
	return otp, nil
}

func (s *InMemoryOTPStore) DeleteOTP(phoneNumber string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.otps, phoneNumber)
	return nil
}

// In-memory Rate Limiter Store (for OTP requests)
type InMemoryRateLimiter struct {
	requests map[string][]time.Time // phone_number -> list of request timestamps
	mu       sync.RWMutex
}

func NewInMemoryRateLimiter() *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		requests: make(map[string][]time.Time),
	}
}

// Allow checks if a request is allowed based on rate limits.
func (r *InMemoryRateLimiter) Allow(phoneNumber string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentTime := time.Now()
	// Filter out requests older than 10 minutes
	var recentRequests []time.Time
	for _, t := range r.requests[phoneNumber] {
		if currentTime.Sub(t) <= 10*time.Minute {
			recentRequests = append(recentRequests, t)
		}
	}

	if len(recentRequests) >= 3 {
		return false // Rate limit exceeded
	}

	recentRequests = append(recentRequests, currentTime)
	r.requests[phoneNumber] = recentRequests
	return true
}
