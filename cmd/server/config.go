package main

import (
	"os"
	"strconv"
)

type config struct {
	ServerHost string
	ServerPort string
	LogEnv     string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string
	MaxConns   int32
	MinConns   int32
	AppEnv     string
}

func loadConfig() config {
	return config{
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		LogEnv:     getEnv("LOG_ENV", "production"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "develop"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		MaxConns:   getEnvInt32("DB_MAX_CONNS", 10),
		MinConns:   getEnvInt32("DB_MIN_CONNS", 1),
		AppEnv:     getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt32(key string, fallback int32) int32 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return int32(parsed)
}
