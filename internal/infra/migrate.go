package infra

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	pgdriver "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

func RunMigrations(ctx context.Context, db *gorm.DB, migrationsDir string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := pgdriver.WithInstance(sqlDB, &pgdriver.Config{})
	if err != nil {
		return err
	}

	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(migrationsDir))
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- m.Up()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		if err == nil || errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return err
	}
}
