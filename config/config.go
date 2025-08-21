package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config main configuration structure
type Config struct {
	Server    ServerConfig    `yaml:"server" mapstructure:"server"`
	Redirects []RedirectRule  `yaml:"redirects" mapstructure:"redirects"`
	Logging   LoggingConfig   `yaml:"logging" mapstructure:"logging"`
}

// ServerConfig server configuration
type ServerConfig struct {
	Host         string `yaml:"host" mapstructure:"host"`
	Port         int    `yaml:"port" mapstructure:"port"`
	ReadTimeout  string `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout  string `yaml:"idle_timeout" mapstructure:"idle_timeout"`
}

// RedirectRule redirect rule
type RedirectRule struct {
	From string `yaml:"from" mapstructure:"from"`
	To   string `yaml:"to" mapstructure:"to"`
}

// LoggingConfig logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level" mapstructure:"level"`
	Format string `yaml:"format" mapstructure:"format"`
}

// GetAddress returns server address
func (s ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// GetReadTimeout returns read timeout as time.Duration
func (s ServerConfig) GetReadTimeout() (time.Duration, error) {
	return time.ParseDuration(s.ReadTimeout)
}

// GetWriteTimeout returns write timeout as time.Duration
func (s ServerConfig) GetWriteTimeout() (time.Duration, error) {
	return time.ParseDuration(s.WriteTimeout)
}

// GetIdleTimeout returns idle timeout as time.Duration
func (s ServerConfig) GetIdleTimeout() (time.Duration, error) {
	return time.ParseDuration(s.IdleTimeout)
}

// LoadConfig loads configuration
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	
	viper.AddConfigPath("/etc/iletken/")   // path to look for the config file in
	viper.AddConfigPath("$HOME/.iletken")  // call multiple times to add many search paths
	viper.AddConfigPath(".")

	// Default values
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to deserialize config: %w", err)
	}
	
	return &config, nil
}

// Validate validates configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Server.Port)
	}
	
	if len(c.Redirects) == 0 {
		return fmt.Errorf("at least one redirect rule must be defined")
	}
	
	for i, redirect := range c.Redirects {
		if redirect.From == "" {
			return fmt.Errorf("redirect rule %d: 'from' field cannot be empty", i+1)
		}
		if redirect.To == "" {
			return fmt.Errorf("redirect rule %d: 'to' field cannot be empty", i+1)
		}
	}
	
	return nil
}
