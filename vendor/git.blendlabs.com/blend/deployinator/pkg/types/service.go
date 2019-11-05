package types

import (
	"time"

	uuid "github.com/blend/go-sdk/uuid"
)

// Service is the metadata for a service in deployinator.
type Service struct {
	// Name is the name of the service.
	Name string `json:"name" yaml:"name"`
	// Env is the environment the service is deployed to.
	Env string `json:"env" yaml:"env"`
	// Namespace is the kube namespace the resource exists in.
	Namespace string `json:"namespace" yaml:"namespace"`

	// Created is the timestamp when the service was created.
	Created time.Time `json:"created" yaml:"created"`
	// CreatedBy is the username that created the service.
	CreatedBy string `json:"createdBy" yaml:"createdBy"`

	// SchemaVersion represents the current version of the schema.
	SchemaVersion string `json:"schemaVersion" yaml:"schemaVersion"`

	// Config contains the working configuration options that may or may not have been provisioned / configured.
	Defaults Config `json:"defaults" yaml:"defaults"`

	// Current represents the currently provisioned config.
	Current Config `json:"current" yaml:"current"`

	// Last represents the last config for any build type.
	Last *Config `json:"last,omitempty" yaml:"last,omitempty"`

	// EnvVars is used in exports for the environment variables.
	EnvVars map[string]string `json:"envVars,omitempty" yaml:"envVars,omitempty"`

	// Certs is used in exports and represents TLS certificate and key pairs.
	Certs []File `json:"certs,omitempty" yaml:"certs,omitempty"`

	// Files is used in exports for the secret file contents.
	Files []File `json:"files,omitempty" yaml:"files,omitempty"`
}

// Config returns a unified config for the service.
func (s Service) Config() *Config {
	service := &Config{
		ServiceName:   s.Name,
		ServiceEnv:    s.Env,
		SchemaVersion: s.SchemaVersion,
	}
	s.Defaults.DeployID = ""
	return ComposeConfigs(service, &s.Defaults)
}

// NewTestService return new test service.
func NewTestService() Service {
	return Service{
		Name:          uuid.V4().String(),
		Env:           "test",
		Namespace:     BlendNamespace,
		Created:       time.Now().UTC(),
		CreatedBy:     UserSystemTarget(),
		SchemaVersion: CurrentSchemaVersion,
	}
}
