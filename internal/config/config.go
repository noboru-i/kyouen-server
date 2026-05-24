package config

import (
	"log"
	"os"
)

type Config struct {
	Port           string
	ProjectID      string
	DatabaseID     string
	Environment    string
	WriteFreeze    bool
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
		DatabaseID:  getEnv("DATASTORE_DATABASE_ID", ""),
		Environment: getEnv("GIN_MODE", "debug"),
		WriteFreeze: getEnv("WRITE_FREEZE", "") == "true",
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
