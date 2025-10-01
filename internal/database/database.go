package database

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
	"github.com/user/coin-indexer/internal/models"
	"github.com/spf13/viper"
)

var DB *gorm.DB

// Initialize sets up the database connection
func Initialize() error {
	driver := viper.GetString("database.driver")
	dsn := viper.GetString("database.dsn")
	
	var dialector gorm.Dialector
	
	switch driver {
	case "sqlite":
		dialector = sqlite.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}
	
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	
	var err error
	DB, err = gorm.Open(dialector, config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Auto-migrate the models
	if err := DB.AutoMigrate(
		&models.Transaction{},
		&models.Contract{},
		&models.BlockProgress{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	
	log.Println("Database initialized successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}