package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgdriver "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"project-template/internal/user"
	userrepo "project-template/internal/user/repository"
	pkgcrypto "project-template/pkg/crypto"
)

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: go run ./cmd/migrate [create|up|down|drop|version|seed]")
	}

	action := strings.ToLower(os.Args[1])
	if action == "create" {
		if len(os.Args) < 3 {
			fatalf("usage: go run ./cmd/migrate create <name>")
		}
		if err := createMigration(os.Args[2]); err != nil {
			fatalf("create: %v", err)
		}
		fmt.Printf("migration %s created\n", os.Args[2])
		return
	}

	_ = godotenv.Load()
	cfg := loadConfig()

	db, err := gorm.Open(postgres.Open(cfg.dsn()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		fatalf("open db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fatalf("db handle: %v", err)
	}

	switch action {
	case "seed":
		if err := seedDemoUser(context.Background(), db, cfg); err != nil {
			fatalf("seed: %v", err)
		}
		fmt.Println("seed complete")
		return
	}

	driver, err := pgdriver.WithInstance(sqlDB, &pgdriver.Config{})
	if err != nil {
		fatalf("migrate driver: %v", err)
	}

	migrationsDir, err := filepath.Abs("migrations")
	if err != nil {
		fatalf("migrations dir: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(migrationsDir), "postgres", driver)
	if err != nil {
		fatalf("create migrator: %v", err)
	}
	defer func() { _, _ = m.Close() }()

	switch action {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	case "drop":
		err = m.Drop()
	case "version":
		version, dirty, versionErr := m.Version()
		if versionErr != nil {
			fatalf("version: %v", versionErr)
		}
		fmt.Printf("version=%d dirty=%t\n", version, dirty)
		return
	default:
		fatalf("unknown action %q", action)
	}

	if err != nil && err != migrate.ErrNoChange {
		fatalf("%s: %v", action, err)
	}

	fmt.Printf("migration %s complete\n", action)
}

func createMigration(name string) error {
	migrationsDir, err := filepath.Abs("migrations")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		return err
	}

	next, err := nextMigrationPrefix(migrationsDir)
	if err != nil {
		return err
	}

	slug := migrationSlug(name)
	upPath := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.up.sql", next, slug))
	downPath := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.down.sql", next, slug))

	if err := os.WriteFile(upPath, []byte("-- TODO: add migration SQL\n"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(downPath, []byte("-- TODO: add rollback SQL\n"), 0o644); err != nil {
		_ = os.Remove(upPath)
		return err
	}

	return nil
}

func nextMigrationPrefix(migrationsDir string) (string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "000001", nil
		}
		return "", err
	}

	re := regexp.MustCompile(`^(\d+)_.*\.up\.sql$`)
	maxSeq := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := re.FindStringSubmatch(entry.Name())
		if len(matches) != 2 {
			continue
		}
		value, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		if value > maxSeq {
			maxSeq = value
		}
	}

	return fmt.Sprintf("%06d", maxSeq+1), nil
}

func migrationSlug(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(value, "")
	if value == "" {
		return "migration"
	}
	return value
}

type config struct {
	DBHost           string
	DBPort           string
	DBName           string
	DBUser           string
	DBPassword       string
	DBSSLMode        string
	DemoUserEmail    string
	DemoUserPassword string
}

func loadConfig() config {
	return config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBName:           getEnv("DB_NAME", "develop"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		DemoUserEmail:    getEnv("DEMO_USER_EMAIL", "admin@example.com"),
		DemoUserPassword: getEnv("DEMO_USER_PASSWORD", "admin123"),
	}
}

func (c config) dsn() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func seedDemoUser(ctx context.Context, db *gorm.DB, cfg config) error {
	repo := userrepo.New(db, nil)
	hasher := pkgcrypto.NewBcryptHasher(0)

	_, err := repo.GetByEmail(ctx, cfg.DemoUserEmail)
	if err == nil {
		return nil
	}
	if !errors.Is(err, user.ErrNotFound) {
		return err
	}

	passwordHash, err := hasher.Hash(cfg.DemoUserPassword)
	if err != nil {
		return err
	}

	return repo.Create(ctx, &user.User{
		ID:           uuid.NewString(),
		Email:        cfg.DemoUserEmail,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	})
}
