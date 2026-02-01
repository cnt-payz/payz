package postgres

import (
	"fmt"

	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/config"
	extransactionpg "github.com/cnt-payz/payz/payment-service/internal/infrastructure/persistence/postgres/extransaction"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d sslmode=%s user=%s password=%s dbname=%s",
		cfg.Secrets.Postgres.Host,
		cfg.Secrets.Postgres.Port,
		cfg.Secrets.Postgres.SSLMode,
		cfg.Secrets.Postgres.User,
		cfg.Secrets.Postgres.Password,
		cfg.Secrets.Postgres.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to open sql db: %w", err)
	}

	sqlDB.SetConnMaxIdleTime(cfg.Secrets.Postgres.Conn.MaxIdletime)
	sqlDB.SetConnMaxLifetime(cfg.Secrets.Postgres.Conn.MaxLifetime)
	sqlDB.SetMaxIdleConns(cfg.Secrets.Postgres.Conn.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.Secrets.Postgres.Conn.MaxOpen)

	return db, nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql db: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close db: %w", err)
	}

	return nil
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&extransactionpg.ExTransactionModel{}); err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}

	return nil
}
