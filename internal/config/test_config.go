package config

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestConfig returns a configuration specifically for testing
func TestConfig(t *testing.T) *Config {
	// Create a temporary directory for the test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Mode: "test",
		},
		CORS: CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Requested-With"},
			AllowCredentials: true,
		},
		Database: DatabaseConfig{
			Driver:          "sqlite",
			DSN:             dbPath,
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: "5m",
			LogLevel:        "silent",
		},
		JWT: JWTConfig{
			SecretKey:      "test-secret-key-for-jwt-tokens",
			AccessTokenExp: "1h",
			Issuer:         "test-ssp",
			Audience:       "test-ssp-web",
		},
	}
}

// TestDB creates a test database connection
func TestDB(t *testing.T) *gorm.DB {
	cfg := TestConfig(t)
	
	// Connect to SQLite database
	db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	// Set connection pool settings for tests
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)

	// Clean up function to be called after test
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing test database: %v", err)
		}
	})

	return db
}