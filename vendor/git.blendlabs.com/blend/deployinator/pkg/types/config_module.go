package types

import (
	"fmt"

	"git.blendlabs.com/blend/deployinator/pkg/taskrunner"
	exception "github.com/blend/go-sdk/exception"
)

// ConfigModule is a base module for builder and deployer
type ConfigModule struct {
	Config *Config
}

// ValidateConfig validates the config
func (m *ConfigModule) ValidateConfig(tr *taskrunner.TaskRunner) error {
	config, ok := tr.GetConfig().(*Config)
	if !ok {
		return exception.New(fmt.Sprintf("Incorrect type for config"))
	}
	if config == nil {
		return exception.New(fmt.Sprintf("Nil config"))
	}
	m.Config = config
	return nil
}
