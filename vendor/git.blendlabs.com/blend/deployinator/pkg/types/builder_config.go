package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blend/go-sdk/env"
	exception "github.com/blend/go-sdk/exception"
	logger "github.com/blend/go-sdk/logger"
)

// NewBuilderConfig returns a new builder config from the os environment.
func NewBuilderConfig() *Config {
	config := &BuilderConfig{}
	env.Env().ReadInto(config)
	return config.Config()
}

// BuilderConfig represents the options for the builder.
// These options override defaults set in the service config(s).
type BuilderConfig struct {
	KubectlConfig       string      `json:"kubectlConfig,omitempty" yaml:"kubectlConfig,omitempty" env:"KUBECTL_CONFIG"`
	ServiceType         ServiceType `json:"serviceType,omitempty" yaml:"serviceType,omitempty" env:"SERVICE_TYPE"`
	DeployID            string      `json:"deployID,omitempty" yaml:"deployID,omitempty" env:"DEPLOY_ID"`
	DeployedBy          string      `json:"deployedBy,omitempty" yaml:"deployedBy,omitempty" env:"DEPLOYED_BY"`
	BuildMode           BuildMode   `json:"buildMode,omitempty" yaml:"buildMode,omitempty" env:"BUILD_MODE"`
	SourcePath          string      `json:"sourcePath,omitempty" yaml:"sourcePath,omitempty" env:"SOURCE_PATH"`
	ServiceConfigPath   string      `json:"serviceConfigPath,omitempty" yaml:"serviceConfigPath,omitempty" env:"SERVICE_CONFIG_PATH"`
	ServiceName         string      `json:"serviceName,omitempty" yaml:"serviceName,omitempty" env:"SERVICE_NAME"`
	ProjectName         string      `json:"projectName,omitempty" yaml:"projectName,omitempty" env:"PROJECT_NAME"`
	DatabaseName        string      `json:"databaseName,omitempty" yaml:"databaseName,omitempty" env:"DATABASE_NAME"`
	ServiceEnv          string      `json:"serviceEnv,omitempty" yaml:"serviceEnv,omitempty" env:"SERVICE_ENV"`
	GitRemote           string      `json:"gitRemote,omitempty" yaml:"gitRemote,omitempty" env:"GIT_REMOTE"`
	GitRefHash          string      `json:"gitRefHash,omitempty" yaml:"gitRefHash,omitempty" env:"GIT_REF_HASH"`
	GitRef              string      `json:"gitRef,omitempty" yaml:"gitRef,omitempty" env:"GIT_REF"`
	GitBase             string      `json:"gitBase,omitempty" yaml:"gitBase,omitempty" env:"GIT_BASE"`
	Dockerfile          string      `json:"dockerfile,omitempty" yaml:"dockerfile,omitempty" env:"DOCKERFILE"`
	DockerRegistry      string      `json:"dockerRegistry,omitempty" yaml:"dockerRegistry,omitempty" env:"DOCKER_REGISTRY"`
	ContainerImage      string      `json:"containerImage,omitempty" yaml:"containerImage,omitempty" env:"CONTAINER_IMAGE"`
	FQDN                string      `json:"fqdn,omitempty" yaml:"fqdn,omitempty" env:"FQDN"`
	AWSRegion           string      `json:"awsRegion,omitempty" yaml:"awsRegion,omitempty" env:"AWS_REGION"`
	ProjectSlackChannel string      `json:"projectSlackChannel,omitempty" yaml:"projectSlackChannel" env:"PROJECT_SLACK_CHANNEL"`
	ProjectRun          bool        `json:"projectRun,omitempty" yaml:"projectRun,omitempty" env:"PROJECT_RUN"`
	SlackChannel        string      `json:"slackChannel,omitempty" yaml:"slackChannel" env:"SLACK_CHANNEL"`

	Args       string `json:"args,omitempty" yaml:"args,omitempty" env:"CONTAINER_ARGS"`
	Command    string `json:"command,omitempty" yaml:"command,omitempty" env:"CONTAINER_COMMAND"`
	WorkingDir string `json:"workingDir,omitempy" yaml:"workingDir,omitempy" env:"CONTAINER_WORKING_DIR"`
	EnvVars    string `json:"envVars,omitempy" yaml:"envVars,omitempy" env:"CONTAINER_ENV_VARS"`
}

// Validate validates the build config.
func (bc BuilderConfig) Validate() error {
	if len(bc.ServiceName) == 0 && len(bc.ProjectName) == 0 && len(bc.DatabaseName) == 0 {
		return exception.New(fmt.Sprintf("`SERVICE_NAME`, `PROJECT_NAME`,  or `DATABASE_NAME` must be set, cannot continue"))
	}
	return nil
}

// Config returns a default set of service config variables from the builder config set.
func (bc BuilderConfig) Config() *Config {
	var args, command []string
	envVars := []EnvVar{}
	logger, _ := logger.NewFromEnv()
	if len(bc.Args) > 0 {
		args = strings.Split(bc.Args, " ")
	}
	if len(bc.Command) > 0 {
		command = strings.Split(bc.Command, " ")
	}
	if len(bc.EnvVars) > 0 {
		err := json.Unmarshal([]byte(bc.EnvVars), &envVars)
		if err != nil {
			logger.SyncError(exception.New(err))
		}
	}
	return &Config{
		KubectlConfig:       bc.KubectlConfig,
		ServiceType:         bc.ServiceType,
		BuildMode:           bc.BuildMode,
		SourcePath:          bc.SourcePath,
		DeployID:            bc.DeployID,
		DeployedBy:          bc.DeployedBy,
		ServiceConfigPath:   bc.ServiceConfigPath,
		ServiceEnv:          bc.ServiceEnv,
		ServiceName:         bc.ServiceName,
		ProjectName:         bc.ProjectName,
		DatabaseName:        bc.DatabaseName,
		ProjectSlackChannel: bc.ProjectSlackChannel,
		ProjectRun:          bc.ProjectRun,
		SlackChannel:        bc.SlackChannel,
		GitRemote:           bc.GitRemote,
		GitRefHash:          bc.GitRefHash,
		GitRef:              bc.GitRef,
		GitBase:             bc.GitBase,
		Dockerfile:          bc.Dockerfile,
		DockerRegistry:      bc.DockerRegistry,
		ContainerImage:      bc.ContainerImage,
		FQDN:                bc.FQDN,
		AWSRegion:           bc.AWSRegion,
		Args:                args,
		Command:             command,
		WorkingDir:          bc.WorkingDir,
		Env:                 envVars,
	}
}
