package airbrake

import "github.com/blend/go-sdk/env"

// Config is a config for the airbrake notifier.
type Config struct {
	ProjectID int64  `json:"projectID" yaml:"projectID" env:"AIRBRAKE_PROJECT_ID"`
	APIKey    string `json:"apiKey" yaml:"apiKey" env:"AIRBRAKE_API_KEY"`
}

// NewConfigFromEnv returns a new airbrake from the environment.
func NewConfigFromEnv() *Config {
	var cfg Config
	env.Env().ReadInto(&cfg)
	return &cfg
}
