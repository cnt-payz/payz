package postgres

import (
	"fmt"

	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/config"
	userpg "github.com/cnt-payz/payz/sso-service/internal/infrastructure/persistence/postgres/user"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Secrets.Postgres.Host,
		cfg.Secrets.Postgres.Port,
		cfg.Secrets.Postgres.User,
		cfg.Secrets.Postgres.Password,
		cfg.Secrets.Postgres.DBName,
		cfg.Secrets.Postgres.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(nil, logger.Config{
			LogLevel: logger.Silent,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	sqlDB.SetConnMaxIdleTime(cfg.Secrets.Postgres.ConnMaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.Secrets.Postgres.ConnMaxLifetime)
	sqlDB.SetMaxIdleConns(cfg.Secrets.Postgres.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Secrets.Postgres.MaxOpenConns)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&userpg.UserModel{}); err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}

	return nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql db: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close connection with postgres: %w", err)
	}

	return nil
}
