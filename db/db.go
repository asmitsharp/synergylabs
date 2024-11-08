package db

import (
	"log"
	"synergylabs/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(dsn string) *gorm.DB {
	var db *gorm.DB
	var err error

	// Retry loop to wait for DB readiness
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("Connected to the database successfully!")
			break
		}

		log.Printf("Database connection failed. Retrying in 5 seconds... (%d/5)\n", i+1)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after multiple attempts: %v", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.Job{}, &models.Profile{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Migration Successful")

	return db
}
