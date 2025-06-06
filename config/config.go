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
	TwitterConfig  TwitterConfig
	FirebaseConfig FirebaseConfig
}

type TwitterConfig struct {
	ConsumerKey    string
	ConsumerSecret string
}

type FirebaseConfig struct {
	CredentialsFile string
}

func Load() *Config {
	config := &Config{
		Port:        getEnv("PORT", "8080"),
		ProjectID:   getEnv("GOOGLE_CLOUD_PROJECT", "my-android-server"),
		Environment: getEnv("GIN_MODE", "debug"),
		TwitterConfig: TwitterConfig{
			ConsumerKey:    getEnv("CONSUMER_KEY", ""),
			ConsumerSecret: getEnv("CONSUMER_SECRET", ""),
		},
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
	required := map[string]string{
		"GOOGLE_CLOUD_PROJECT": config.ProjectID,
		"CONSUMER_KEY":         config.TwitterConfig.ConsumerKey,
		"CONSUMER_SECRET":      config.TwitterConfig.ConsumerSecret,
	}

	for key, value := range required {
		if value == "" {
			log.Printf("Warning: Environment variable %s is not set", key)
		}
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