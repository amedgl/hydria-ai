// internal/config/config.go
// Application configuration — loaded from config.yaml with built-in defaults.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all configurable values for HydrIA AI.
type Config struct {
	Gemini struct {
		Model     string `yaml:"model"`
		MaxTokens int    `yaml:"max_tokens"`
	} `yaml:"gemini"`
	Wordlist struct {
		MaxSize             int  `yaml:"max_size"`
		MinLength           int  `yaml:"min_length"`
		MaxLength           int  `yaml:"max_length"`
		IncludeLeet         bool `yaml:"include_leet"`
		IncludeReverse      bool `yaml:"include_reverse"`
		IncludeCombinations bool `yaml:"include_combinations"`
	} `yaml:"wordlist"`
	Hydra struct {
		Threads   int `yaml:"threads"`
		Timeout   int `yaml:"timeout"`
		BatchSize int `yaml:"batch_size"`
	} `yaml:"hydra"`
	Session struct {
		AutoResume bool `yaml:"auto_resume"`
	} `yaml:"session"`
}

// Load reads config.yaml and returns a Config with defaults pre-applied.
func Load() Config {
	cfg := Config{}

	// Built-in defaults
	cfg.Gemini.Model = "gemini-2.0-flash"
	cfg.Wordlist.MaxSize = 50000
	cfg.Wordlist.MinLength = 4
	cfg.Wordlist.MaxLength = 20
	cfg.Wordlist.IncludeLeet = true
	cfg.Wordlist.IncludeReverse = true
	cfg.Wordlist.IncludeCombinations = true
	cfg.Hydra.Threads = 4
	cfg.Hydra.Timeout = 30
	cfg.Hydra.BatchSize = 50
	cfg.Session.AutoResume = true

	data, err := os.ReadFile("config.yaml")
	if err == nil {
		yaml.Unmarshal(data, &cfg) //nolint:errcheck
	}
	return cfg
}
