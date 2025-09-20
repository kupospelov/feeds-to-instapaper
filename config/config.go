package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Instapaper struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type Hooks struct {
	NewArticle []Hook `toml:"new_article"`
}

type Hook struct {
	Spawn []string `toml:"spawn"`
}

type Feeds struct {
	URLs []string `toml:"urls"`
}

type Config struct {
	Instapaper `toml:"instapaper"`
	Hooks      `toml:"hooks"`
	Feeds      `toml:"feeds"`
}

func Load() (*Config, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	configPath := filepath.Join(configDir, "feeds-to-instapaper", "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	_, err = toml.Decode(string(data), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Instapaper.Username == "" || config.Instapaper.Password == "" {
		return nil, fmt.Errorf("Instapaper username and password are required")
	}
	if len(config.Feeds.URLs) == 0 {
		return nil, fmt.Errorf("at least one feed URL is required")
	}
	return &config, nil
}
