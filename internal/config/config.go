package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	OpenAI   OpenAIConfig
	Telegram TelegramConfig
	JWT      JWTConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host string
	Port int
	Env  string
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	URL      string
	PoolSize int
}

// OpenAIConfig holds OpenAI API settings.
type OpenAIConfig struct {
	APIKey            string
	EmbeddingModel    string
	EmbeddingDim      int
	EmbeddingURL      string
}

// TelegramConfig holds Telegram bot settings.
type TelegramConfig struct {
	BotToken    string
	WebhookURL  string
	WebhookPort int
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret string
}

// Load reads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Environment variable support
	v.SetEnvPrefix("AI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Default values
	setDefaults(v)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
		// Config file not found, use defaults + env vars
	}

	cfg := &Config{}

	// Server
	cfg.Server.Host = v.GetString("server.host")
	cfg.Server.Port = v.GetInt("server.port")
	cfg.Server.Env = v.GetString("server.env")

	// Database
	cfg.Database.URL = v.GetString("database.url")
	cfg.Database.PoolSize = v.GetInt("database.pool_size")

	// OpenAI
	cfg.OpenAI.APIKey = v.GetString("openai.api_key")
	cfg.OpenAI.EmbeddingModel = v.GetString("openai.embedding_model")
	cfg.OpenAI.EmbeddingDim = v.GetInt("openai.embedding_dimensions")
	cfg.OpenAI.EmbeddingURL = v.GetString("openai.embedding_url")

	// Telegram
	cfg.Telegram.BotToken = v.GetString("telegram.bot_token")
	cfg.Telegram.WebhookURL = v.GetString("telegram.webhook_url")
	cfg.Telegram.WebhookPort = v.GetInt("telegram.webhook_port")

	// JWT
	cfg.JWT.Secret = v.GetString("jwt.secret")

	return cfg, nil
}

// Address returns the server address in host:port format.
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// setDefaults sets default configuration values.
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.env", "development")

	// Database defaults
	v.SetDefault("database.pool_size", 10)

	// OpenAI defaults
	v.SetDefault("openai.embedding_model", "text-embedding-3-small")
	v.SetDefault("openai.embedding_dimensions", 1536)
	v.SetDefault("openai.embedding_url", "https://api.openai.com/v1/embeddings")

	// JWT defaults
	v.SetDefault("jwt.secret", "change-me-in-production")
}
