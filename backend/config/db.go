package config

import (
	"backend/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes and returns a GORM database connection
func InitDB() *gorm.DB {
	// Database connection string
	dsn := "host=localhost user=srini password=srini@4614 dbname=realtime_chat port=5432 sslmode=disable"

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Successfully connected to the database!")

	// Auto migrate database schema
	if err := db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.ChatMessage{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed!")

	return db
}

