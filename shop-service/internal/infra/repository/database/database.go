package dbrepo

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/cnt-payz/payz/shop-service/config"
	domainrepo "github.com/cnt-payz/payz/shop-service/internal/domain/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseRepo struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewDatabaseRepo(cfg *config.Config, log *slog.Logger) (domainrepo.DatabaseRepo, error) {
	dsn := fmt.Sprintf("dbname=%s user=%s password=%s host=%s port=%s sslmode=disable",
		cfg.DB.Name, cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(&shop{}); err != nil {
		return nil, err
	}

	return &DatabaseRepo{
		db:  db,
		log: log,
	}, nil
}

func (dr *DatabaseRepo) Close() error {
	sqlDB, err := dr.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %v", err)
	}

	dr.log.Info("database connection closed")
	return nil
}
