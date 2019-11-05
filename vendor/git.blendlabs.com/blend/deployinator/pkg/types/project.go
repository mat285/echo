package types

import (
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
)

// Project is the metadata for a project in deployinator
type Project struct {
	// Name is the name of the project
	Name string `json:"name" yaml:"name"`

	// Created is the timestamp when the service was created.
	Created time.Time `json:"created" yaml:"created"`
	// CreatedBy is the username that created the service.
	CreatedBy string `json:"createdBy" yaml:"createdBy"`

	// Config contains the working configuration options that may or may not have been provisioned / configured.
	Defaults Config `json:"defaults" yaml:"defaults"`

	// Last represents the last config for any build type.
	Last *Config `json:"last,omitempty" yaml:"last,omitempty"`

	// Triggers are the project run triggers
	Triggers []Trigger `json:"-"`

	// ServiceNames are the names of the steps in the project
	ServiceNames []string `json:"-"`
	// Services are the services in this project, populated from ServiceNames
	Services []*Service `json:"-"`

	// StagedServices is a map of stage number to services to run in that
	StagedServices map[int][]string `json:"-"`
	// ServiceNamesToPointers is a map of service names to object pointers
	ServiceNamesToPointers map[string]*Service `json:"-"`

	// Env is the environment of the project
	Env string `json:"env" yaml:"env"`
}

// ProjectStep is a struct containing the information for a step in the project
type ProjectStep struct {
	Name  string `json:"name"`
	Stage int    `json:"stage"`
}

// Config returns the config for this project
func (p *Project) Config() *Config {
	c := &Config{
		ProjectName: p.Name,
		BuildMode:   BuildModeRunProject,
		DeployID:    rand.String(RandomStringLength),
	}
	return ComposeConfigs(c, &p.Defaults)
}
