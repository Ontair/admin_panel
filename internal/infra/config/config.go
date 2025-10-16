package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config represents application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Cookie   CookieConfig   `mapstructure:"cookie"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	Environment  string `mapstructure:"environment"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	SecretKey     string `mapstructure:"secret_key"`
	RefreshSecret string `mapstructure:"refresh_secret"`
	AccessExpiry  int    `mapstructure:"access_expiry"`  // minutes
	RefreshExpiry int    `mapstructure:"refresh_expiry"` // minutes
}

// CookieConfig represents cookie configuration
type CookieConfig struct {
	Domain        string `mapstructure:"domain"`
	Secure        bool   `mapstructure:"secure"`
	SameSite      string `mapstructure:"same_site"`
	Path          string `mapstructure:"path"`
	AccessExpiry  int    `mapstructure:"access_expiry"`  // minutes
	RefreshExpiry int    `mapstructure:"refresh_expiry"` // minutes
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	File   string `mapstructure:"file"`
}

// Load reads configuration from files and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// Set default values
	setDefaults()

	// Enable reading from environment variables
	viper.AutomaticEnv()

	// Bind environment variables
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.username", "DATABASE_USERNAME")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.name", "DATABASE_NAME")
	viper.BindEnv("database.sslmode", "DATABASE_SSLMODE")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.name", "admin_panel")
	viper.SetDefault("database.sslmode", "disable")

	// JWT defaults
	viper.SetDefault("jwt.secret_key", "your-secret-key")
	viper.SetDefault("jwt.refresh_secret", "your-refresh-secret")
	viper.SetDefault("jwt.access_expiry", 15)    // 15 minutes
	viper.SetDefault("jwt.refresh_expiry", 1440) // 24 hours

	// Cookie defaults
	viper.SetDefault("cookie.domain", "")
	viper.SetDefault("cookie.secure", false)
	viper.SetDefault("cookie.same_site", "Lax")
	viper.SetDefault("cookie.path", "/")
	viper.SetDefault("cookie.access_expiry", 15)    // 15 minutes
	viper.SetDefault("cookie.refresh_expiry", 1440) // 24 hours

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.file", "")
}

// GetDSN returns database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.Username,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// IsProduction checks if environment is production
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// GetPort returns server port
func (c *Config) GetPort() string {
	return ":" + c.Server.Port
}

// GetPostgresURL returns PostgreSQL connection URL
func (c *Config) GetPostgresURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}
