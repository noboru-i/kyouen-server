package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port           string
	ProjectID      string
	Environment    string
	FirebaseConfig FirebaseConfig
}

type FirebaseConfig struct {
	CredentialsFile string
}

func Load() *Config {
	// Determine default project ID based on environment
	defaultProjectID := "my-android-server" // Production default
	if env := getEnv("ENVIRONMENT", ""); env == "dev" {
		defaultProjectID = "api-project-732262258565"
	}
	
	config := &Config{
		Port:        getEnv("PORT", "8080"),
		ProjectID:   getEnv("GOOGLE_CLOUD_PROJECT", defaultProjectID),
		Environment: getEnv("GIN_MODE", "debug"),
		FirebaseConfig: FirebaseConfig{
			CredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", ""),
		},
	}

	if err := validateConfig(config); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	return config
}

func validateConfig(config *Config) error {
	// 必須設定の確認
	if config.ProjectID == "" {
		log.Printf("Warning: GOOGLE_CLOUD_PROJECT is not set, using default: my-android-server")
	}


	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}