package api

import (
	"github.com/ebipenman/go-otp-auth-service/internal/middleware"
	"github.com/ebipenman/go-otp-auth-service/pkg/auth"
	"github.com/ebipenman/go-otp-auth-service/pkg/user"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	authHandler *auth.Handler,
	userHandler *user.Handler,
	jwtSecret string,
	otpRateLimiter middleware.RateLimiterStore,
) {
	// Public routes (no authentication required)
	public := router.Group("/")
	{
		// Health check endpoint
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "UP"})
		})
	}

	// Authentication routes
	authRoutes := router.Group("/otp")
	{
		authRoutes.POST("/send", middleware.OTPRateLimiter(otpRateLimiter), authHandler.SendOTP)
		authRoutes.POST("/verify", authHandler.VerifyOTP)
	}

	// Protected routes (JWT authentication required)
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// User management endpoints
		userRoutes := protected.Group("/users")
		{
			userRoutes.GET("", userHandler.ListUsers)
			userRoutes.GET("/:id", userHandler.GetUserByID)
			// Add other user management routes here (e.g., PUT, DELETE) if needed
		}

		// Example of a protected endpoint that uses the user from context
		protected.GET("/me", func(c *gin.Context) {
			user, exists := c.Get(middleware.ContextKeyUser)
			if !exists {
				c.JSON(401, gin.H{"error": "User not found in context"})
				return
			}
			c.JSON(200, user)
		})
	}
}
