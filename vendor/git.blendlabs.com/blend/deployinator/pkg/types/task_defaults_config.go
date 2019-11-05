package types

import (
	"time"

	"github.com/blend/go-sdk/env"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	// DefaultTaskReplicas is the default replica count.
	DefaultTaskReplicas = "1"
	// DefaultTaskContainerPort is the default container port.
	DefaultTaskContainerPort = ContainerPortNone
	// DefaultTaskAccessibility is the default accessibility level.
	DefaultTaskAccessibility = AccessibilityNone
	// DefaultTaskRestartPolicy is the default restart policy.
	DefaultTaskRestartPolicy = RestartPolicyNever
)

// NewTaskDefaultsConfig returns a new defaults config from the environment.
func NewTaskDefaultsConfig() *TaskDefaultsConfig {
	config := &TaskDefaultsConfig{}
	env.Env().ReadInto(config)
	return config
}

// TaskDefaultsConfig represents inferred values from the environment and default values.
type TaskDefaultsConfig struct {
	ServiceEnv    string `json:"serviceEnv" yaml:"serviceEnv" env:"SERVICE_ENV"`
	ClusterName   string `json:"clusterName" yaml:"clusterName" env:"CLUSTER_NAME"`
	Cluster       string `json:"cluster" yaml:"cluster" env:"CLUSTER"`
	KubectlConfig string `json:"kubectlConfig" yaml:"kubectlConfig" env:"KUBECTL_CONFIG"`
}

// Config returns relevant config values from the base config.
func (dc TaskDefaultsConfig) Config() *Config {
	return &Config{
		SchemaVersion:     CurrentSchemaVersion,
		DeployID:          rand.String(RandomStringLength),
		DeployedAt:        time.Now().UTC(),
		DeployedBy:        DefaultDeployedBy,
		FileMountPath:     DefaultFileMountPath,
		ServiceEnv:        dc.ServiceEnv,
		ClusterName:       dc.ClusterName,
		Cluster:           dc.Cluster,
		KubectlConfig:     dc.KubectlConfig,
		Namespace:         DefaultNamespace,
		GitRef:            DefaultGitRef,
		Dockerfile:        DefaultDockerfile,
		ServiceConfigPath: DefaultServiceConfigPath,
		Replicas:          DefaultTaskReplicas,
		AWSRegion:         DefaultAWSRegion,
		Autoscale:         FlagDisabled,
		Accessibility:     DefaultTaskAccessibility,
		ContainerPort:     DefaultTaskContainerPort,
		RestartPolicy:     string(DefaultTaskRestartPolicy),
	}
}
