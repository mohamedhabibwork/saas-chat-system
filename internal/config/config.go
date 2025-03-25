package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Storage  StorageConfig  `json:"storage"`
	AI       AIConfig       `json:"ai"`
	JWT      JWTConfig      `json:"jwt"`
}

// ServerConfig represents server-related configuration
type ServerConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	BaseURL  string `json:"base_url"`
	LogLevel string `json:"log_level"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Type      string      `json:"type"` // "local" or "s3"
	Local     LocalConfig `json:"local"`
	S3        S3Config    `json:"s3"`
	MaxSizeMB int64       `json:"max_size_mb"`
}

// LocalConfig represents local storage configuration
type LocalConfig struct {
	BasePath string `json:"base_path"`
}

// S3Config represents S3 storage configuration
type S3Config struct {
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Endpoint        string `json:"endpoint"` // Optional, for custom S3-compatible services
}

// AIConfig represents AI service configuration
type AIConfig struct {
	DefaultModel string            `json:"default_model"`
	Models       map[string]ModelConfig `json:"models"`
}

// ModelConfig represents configuration for a specific AI model
type ModelConfig struct {
	Type       string `json:"type"` // "gpt4", "claude", "custom"
	APIKey     string `json:"api_key"`
	APIURL     string `json:"api_url"`
	MaxTokens  int    `json:"max_tokens"`
	ModelName  string `json:"model_name"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	SecretKey     string `json:"secret_key"`
	ExpiryHours   int    `json:"expiry_hours"`
	RefreshHours  int    `json:"refresh_hours"`
}

var globalConfig *Config

// LoadConfig loads the configuration from a JSON file
func LoadConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	globalConfig = &Config{}
	if err := json.NewDecoder(file).Decode(globalConfig); err != nil {
		return err
	}

	return nil
}

// GetConfig returns the global configuration
func GetConfig() *Config {
	return globalConfig
}

// GetStoragePath returns the appropriate storage path based on configuration
func GetStoragePath() string {
	config := GetConfig()
	if config.Storage.Type == "local" {
		return config.Storage.Local.BasePath
	}
	return "" // For S3, we don't need a local path
}

// GetMaxFileSize returns the maximum allowed file size in bytes
func GetMaxFileSize() int64 {
	return GetConfig().Storage.MaxSizeMB * 1024 * 1024
}

// EnsureStoragePath ensures the storage directory exists
func EnsureStoragePath() error {
	config := GetConfig()
	if config.Storage.Type == "local" {
		return os.MkdirAll(config.Storage.Local.BasePath, 0755)
	}
	return nil
}

// GetUploadPath returns the path for file uploads
func GetUploadPath(userID int) string {
	config := GetConfig()
	if config.Storage.Type == "local" {
		return filepath.Join(config.Storage.Local.BasePath, "uploads", fmt.Sprintf("user_%d", userID))
	}
	return fmt.Sprintf("uploads/user_%d", userID)
} 