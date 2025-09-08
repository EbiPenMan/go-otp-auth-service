package main

import (
	"log"
	"time"

	"github.com/ebipenman/go-otp-auth-service/config"
	"github.com/ebipenman/go-otp-auth-service/internal/api"
	"github.com/ebipenman/go-otp-auth-service/internal/database"
	"github.com/ebipenman/go-otp-auth-service/internal/middleware"
	"github.com/ebipenman/go-otp-auth-service/pkg/auth"
	"github.com/ebipenman/go-otp-auth-service/pkg/otp"
	"github.com/ebipenman/go-otp-auth-service/pkg/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	// Swagger docs (generated)
	_ "github.com/ebipenman/go-otp-auth-service/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title OTP Auth GoLang API
// @version 1.0
// @description This is a sample server for OTP based authentication and user management.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.LoadConfig()

	// Declare variables for our stores using their INTERFACE types.
	var userStore user.UserStore
	var otpStore otp.OTPStore

	// Decide which concrete implementation to create based on the config.
	if cfg.StorageType == "postgres" {
		log.Println("Initializing PostgreSQL database store...")
		postgresStore, err := database.NewPostgresStore(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("FATAL: could not connect to postgres database: %v", err)
		}
		// The single PostgresStore object implements BOTH interfaces.
		userStore = postgresStore
		otpStore = postgresStore
	} else {
		log.Println("Initializing in-memory database store...")
		// For in-memory, we have separate store objects.
		userStore = database.NewInMemoryUserStore()
		otpStore = database.NewInMemoryOTPStore()
	}

	// NOTE: We now use the middleware's rate limiter, not the one from the database package
	// as it contains the cleanup logic.
	otpRateLimiter := middleware.NewInMemoryRateLimiter(3, 2*time.Minute)

	// Initialize OTP components
	otpGenerator := otp.NewSimpleOTPGenerator()

	// Initialize Repositories
	userRepo := user.NewRepository(userStore)
	otpRepo := otp.NewRepository(otpStore)
	authRepo := auth.NewRepository(userRepo, otpRepo, otpRateLimiter)

	// The auth service now correctly receives all its dependencies via the authRepo.
	authService := auth.NewService(authRepo, otpGenerator, cfg.JWTSecret)
	userService := user.NewService(userRepo)

	// Initialize Handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)

	// Setup Gin router
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Global Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// The router setup function needs this to apply the rate limiting middleware
	api.SetupRoutes(router, authHandler, userHandler, cfg.JWTSecret, otpRateLimiter)

	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
