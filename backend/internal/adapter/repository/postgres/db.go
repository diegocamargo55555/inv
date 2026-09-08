package postgres

import (
	"fmt"
	"log"

	"github.com/invest/backend/internal/config"
	"github.com/invest/backend/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB initializes PostgreSQL connection and runs GORM AutoMigrate
func NewDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	log.Println("PostgreSQL connection established successfully.")

	// AutoMigrate all domain models
	err = db.AutoMigrate(
		&domain.User{},
		&domain.RefreshToken{},
		&domain.Account{},
		&domain.Category{},
		&domain.Transaction{},
		&domain.Budget{},
		&domain.Asset{},
		&domain.Portfolio{},
		&domain.Position{},
		&domain.InvestmentTransaction{},
		&domain.Earning{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}

	log.Println("Database AutoMigrate completed.")

	// Run initial seeds for categories and common assets
	if err := SeedInitialData(db); err != nil {
		log.Printf("Warning: Seed error: %v", err)
	}

	return db, nil
}
