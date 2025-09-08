package auth

import (
	"errors"
	"net/http"

	"github.com/ebipenman/go-otp-auth-service/internal/model"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService Service
}

func NewHandler(authService Service) *Handler {
	return &Handler{authService: authService}
}

type verifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
	OTP         string `json:"otp" binding:"required,len=6,numeric"`
}

// @Summary Send OTP
// @Description Sends an OTP to the provided phone number for login or registration.
// @Description Rate limit: 3 requests per phone number within 10 minutes.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body model.SendOTPRequest true "Phone Number"
// @Success 200 {object} map[string]string "message: OTP sent successfully (check console)"
// @Failure 400 {object} map[string]string "error: Invalid phone number format"
// @Failure 429 {object} map[string]string "error: Rate limit exceeded"
// @Failure 500 {object} map[string]string "error: Failed to process OTP request"
// @Router /otp/send [post]
func (h *Handler) SendOTP(c *gin.Context) {
	// Step 1: Retrieve the pre-bound request object from the context.
	val, exists := c.Get("otp_request")
	if !exists {
		// This should not happen if the middleware is applied correctly.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve request from context"})
		return
	}

	// Step 2: Perform a type assertion to get the correct struct type.
	req, ok := val.(model.SendOTPRequest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid request type in context"})
		return
	}

	// Step 3: The rest of the handler logic remains the same.
	err := h.authService.SendOTP(req.PhoneNumber)
	if err != nil {
		if errors.Is(err, ErrRateLimitExceeded) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully (check console)"})
}

// @Summary Verify OTP and Login/Register
// @Description Submits a phone number and OTP to get a JWT token.
// @Description If the user doesn't exist, they will be registered.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body verifyOTPRequest true "Phone Number and OTP"
// @Success 200 {object} map[string]string "token: <jwt_token>"
// @Failure 400 {object} map[string]string "error: Invalid request format"
// @Failure 401 {object} map[string]string "error: Invalid or expired OTP"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /otp/verify [post]
func (h *Handler) VerifyOTP(c *gin.Context) {
	var req verifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	token, err := h.authService.VerifyOTPAndAuthenticate(req.PhoneNumber, req.OTP)
	if err != nil {
		if errors.Is(err, ErrInvalidOTP) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// Other errors from the service layer are likely 500s
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
