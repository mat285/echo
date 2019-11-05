package airbrake

import (
	"fmt"

	"github.com/airbrake/gobrake"
)

// NewClientFromConfig initializes airbrake.
func NewClientFromConfig(cfg *Config) (*gobrake.Notifier, error) {
	if cfg == nil {
		return nil, nil
	}
	if cfg.ProjectID == 0 || len(cfg.APIKey) == 0 {
		return nil, fmt.Errorf("`%s` and `%s` are empty, cannot create airbrake client", EnvVarAirbrakeProjectID, EnvVarAirbrakeKey)
	}

	return gobrake.NewNotifier(cfg.ProjectID, cfg.APIKey), nil
}
