package config

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes and returns a GORM database connection
func InitDB() *gorm.DB {
	// Ensure environment is initialized
	if EnvironmentData == nil {
		InitializeEnv()
	}

	// Get DSN from environment config
	dsn := EnvironmentData.GetDSN()

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Successfully connected to the database!")

	return db
}
