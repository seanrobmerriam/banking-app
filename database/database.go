package database

import (
	"banking-app/models"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDatabase establishes connection to SQLite database and handles migrations
// Uses SQLite for simplicity - easily replaceable with PostgreSQL/MySQL
func InitDatabase() (*gorm.DB, error) {
	// Database configuration - SQLite file for development
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "banking.db" // Default SQLite file
	}

	// Open database connection with logging enabled for development
	// Silent mode can be used in production for better performance
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Log SQL queries during development
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Enable foreign key constraints - critical for data integrity in banking systems
	// SQLite requires explicit enabling of foreign key constraints
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	
	sqlDB.SetMaxIdleConns(10)                    // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)                   // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour)          // Connection maximum lifetime

	// Auto migrate database schema
	// Automatically creates/updates tables based on model definitions
	// Critical for maintaining database schema consistency
	err = db.AutoMigrate(
		&models.Customer{},  // Customer table
		&models.Account{},   // Account table
		&models.Transaction{}, // Transaction table
		&models.Loan{},      // Loan table
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connection established and migrations completed successfully")
	return db, nil
}