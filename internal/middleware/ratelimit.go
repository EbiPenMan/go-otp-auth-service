package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ebipenman/go-otp-auth-service/internal/model"

	"github.com/gin-gonic/gin"
)

// RateLimiterStore defines the interface for a rate limiter's underlying storage.
// This allows for easy swapping between in-memory, Redis, etc.
type RateLimiterStore interface {
	Allow(key string) bool
}

// InMemoryRateLimiter implements RateLimiterStore using a simple in-memory map.
// It tracks request timestamps for each key (e.g., phone number).
type InMemoryRateLimiter struct {
	requests   map[string][]time.Time
	mu         sync.RWMutex
	maxReq     int
	timeWindow time.Duration
}

// NewInMemoryRateLimiter creates and returns a new InMemoryRateLimiter.
// maxReq: Maximum number of requests allowed.
// timeWindow: The duration of the time window.
func NewInMemoryRateLimiter(maxReq int, timeWindow time.Duration) *InMemoryRateLimiter {
	limiter := &InMemoryRateLimiter{
		requests:   make(map[string][]time.Time),
		maxReq:     maxReq,
		timeWindow: timeWindow,
	}

	// Start a background goroutine to periodically clean up old entries
	go limiter.cleanup()

	return limiter
}

// Allow checks if a request for a given key is permitted.
func (r *InMemoryRateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentTime := time.Now()
	// Filter out requests that are older than the time window
	var recentRequests []time.Time
	for _, t := range r.requests[key] {
		if currentTime.Sub(t) <= r.timeWindow {
			recentRequests = append(recentRequests, t)
		}
	}

	// Check if the number of recent requests has reached the maximum
	if len(recentRequests) >= r.maxReq {
		log.Printf("Rate limit exceeded for key: %s", key)
		r.requests[key] = recentRequests // Update with the filtered list
		return false                     // Rate limit exceeded
	}

	// Add the current request timestamp and allow the request
	recentRequests = append(recentRequests, currentTime)
	r.requests[key] = recentRequests
	return true
}

// cleanup periodically iterates through the map and removes keys with no recent requests.
func (r *InMemoryRateLimiter) cleanup() {
	// Run cleanup every 10 minutes (the same as our time window)
	for range time.Tick(10 * time.Minute) {
		r.mu.Lock()
		currentTime := time.Now()
		for key, timestamps := range r.requests {
			var recentTimestamps []time.Time
			for _, t := range timestamps {
				if currentTime.Sub(t) <= r.timeWindow {
					recentTimestamps = append(recentTimestamps, t)
				}
			}
			if len(recentTimestamps) == 0 {
				delete(r.requests, key)
			} else {
				r.requests[key] = recentTimestamps
			}
		}
		r.mu.Unlock()
		log.Println("Rate limiter cleanup finished.")
	}
}

// OTPRateLimiter creates a Gin middleware to rate limit OTP requests based on phone number.
func OTPRateLimiter(store RateLimiterStore) gin.HandlerFunc {

	return func(c *gin.Context) {
		var req model.SendOTPRequest

		// Step 1: Bind the JSON.
		// If binding fails, it's a malformed request. We should stop here.
		if err := c.ShouldBindJSON(&req); err != nil {
			// Abort the request with a 400 Bad Request.
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Step 2: Use the phone number from the successfully bound request for rate limiting.
		if !store.Allow(req.PhoneNumber) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "You have made too many requests. Please try again after rate limit time.",
			})
			return
		}

		// Step 3: IMPORTANT - Store the bound request object in the context
		// for the final handler to use.
		c.Set("otp_request", req)

		// Step 4: Proceed to the next handler.
		c.Next()
	}
}
