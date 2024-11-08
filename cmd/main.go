package main

import (
	"log"
	"os"
	"synergylabs/api"
	"synergylabs/db"
	"synergylabs/services/cache"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize database
	dsn := os.Getenv("DATABASE_URL") // Set your database URL in environment variable
	database := db.InitDB(dsn)

	// Initialize Redis cache
	redisAddr := os.Getenv("REDIS_ADDR") // Set your Redis address in environment variable
	redisCache := cache.NewCache(redisAddr)

	// Initialize Echo framework
	e := echo.New()

	// Set up API routes
	api.SetupRoutes(e, database, redisCache, logger)

	// Start the server
	e.Logger.Fatal(e.Start(":3000")) // Change the port as needed
}
