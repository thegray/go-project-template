package infra

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type PostgresConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	SSLMode  string
	MaxConns int32
	MinConns int32
}

func NewPostgresPool(ctx context.Context, cfg PostgresConfig) (*gorm.DB, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxConns > 0 {
		sqlDB.SetMaxOpenConns(int(cfg.MaxConns))
	}
	if cfg.MinConns > 0 {
		sqlDB.SetMaxIdleConns(int(cfg.MinConns))
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	return db.WithContext(ctx), nil
}
