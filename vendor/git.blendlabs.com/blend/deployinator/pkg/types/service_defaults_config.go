package types

import (
	"time"

	"github.com/blend/go-sdk/env"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	// DefaultNamespace is the default k8s namespace.
	DefaultNamespace = BlendNamespace
	// DefaultDockerfile is the default dockerfile to build.
	DefaultDockerfile = "Dockerfile"
	// DefaultReplicas is the default replica count.
	DefaultReplicas = "2"
	// DefaultContainerPort is the default container port.
	DefaultContainerPort = "5000"
	// DefaultContainerProto is the default container proto.
	DefaultContainerProto = "TCP"
	// DefaultHTTPPort is the default HTTP port.
	DefaultHTTPPort = "80"
	// DefaultHTTPSPort is the default HTTPS port.
	DefaultHTTPSPort = "443"
	// DefaultAccessibility is the default accessibility level.
	DefaultAccessibility = "private"
	// DefaultServiceConfigPath is the default service config path.
	DefaultServiceConfigPath = "service.yml"
	// DefaultServiceEnv is the default service env.
	DefaultServiceEnv = "sandbox"
	// DefaultDeployedBy is the default deployed by name.
	DefaultDeployedBy = "system"
	// DefaultRestartPolicy is the default restart policy.
	DefaultRestartPolicy = RestartPolicyAlways
	// DefaultFileMountPath is the default file mount path.
	DefaultFileMountPath = "/var/secrets"
	// DefaultProxyPort is the default port the proxy listens on.
	DefaultProxyPort = "8443"
	// DefaultAlternateProxyPort is the alternate port the proxy can listen on.
	DefaultAlternateProxyPort = "8444"
	// DefaultGitRef is a placeholder git ref.
	DefaultGitRef = "master"
	// DefaultAWSRegion is the default aws region to provisoin roles from.
	DefaultAWSRegion = "us-east-1"
	// DefaultCPUThreshold is the default cpu threshold to create new replicas at.
	DefaultCPUThreshold = "60"
	// DefaultMemoryThreshold is the default memory threshold to create new replicas at.
	DefaultMemoryThreshold = "60"
	// DefaultMaxReplicas is a default maximum number of replicas in an autoscale group.
	DefaultMaxReplicas = "10"
)

// NewServiceDefaultsConfig returns a new defaults config from the environment.
func NewServiceDefaultsConfig() *ServiceDefaultsConfig {
	config := &ServiceDefaultsConfig{}
	env.Env().ReadInto(config)
	return config
}

// ServiceDefaultsConfig represents inferred values from the environment and default values.
type ServiceDefaultsConfig struct {
	ServiceEnv    string `json:"serviceEnv" yaml:"serviceEnv" env:"SERVICE_ENV"`
	ClusterName   string `json:"clusterName" yaml:"clusterName" env:"CLUSTER_NAME"`
	Cluster       string `json:"cluster" yaml:"cluster" env:"CLUSTER"`
	KubectlConfig string `json:"kubectlConfig" yaml:"kubectlConfig" env:"KUBECTL_CONFIG"`
	AWSAccessKey  string `json:"-" yaml:"-" env:"AWS_ACCESS_KEY"`
	AWSSecretKey  string `json:"-" yaml:"-" env:"AWS_SECRET_KEY"`
	RegistryHost  string `json:"registryHost" yaml:"registryHost" env:"REGISTRY_HOST"`
}

// Config returns relevant config values from the base config.
func (sdc ServiceDefaultsConfig) Config() *Config {
	return &Config{
		SchemaVersion:     CurrentSchemaVersion,
		DeployID:          rand.String(RandomStringLength),
		DeployedAt:        time.Now().UTC(),
		DeployedBy:        DefaultDeployedBy,
		FileMountPath:     DefaultFileMountPath,
		ServiceEnv:        sdc.ServiceEnv,
		ClusterName:       sdc.ClusterName,
		Cluster:           sdc.Cluster,
		KubectlConfig:     sdc.KubectlConfig,
		Namespace:         DefaultNamespace,
		GitRef:            DefaultGitRef,
		Dockerfile:        DefaultDockerfile,
		ServiceConfigPath: DefaultServiceConfigPath,
		Replicas:          DefaultReplicas,
		AWSRegion:         DefaultAWSRegion,
		Autoscale:         FlagDisabled,
		Accessibility:     DefaultAccessibility,
		RestartPolicy:     string(DefaultRestartPolicy),
		AWSAccessKey:      sdc.AWSAccessKey,
		AWSSecretKey:      sdc.AWSSecretKey,
		RegistryHost:      sdc.RegistryHost,
		Protocol:          ProtocolHTTP,
	}
}

// EnvConfig returns the defaults from the env only
func (sdc ServiceDefaultsConfig) EnvConfig() *Config {
	return &Config{
		DeployID:      rand.String(RandomStringLength),
		ServiceEnv:    sdc.ServiceEnv,
		ClusterName:   sdc.ClusterName,
		Cluster:       sdc.Cluster,
		KubectlConfig: sdc.KubectlConfig,
		AWSAccessKey:  sdc.AWSAccessKey,
		AWSSecretKey:  sdc.AWSSecretKey,
		RegistryHost:  sdc.RegistryHost,
	}
}
