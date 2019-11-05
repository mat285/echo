package types

import (
	"fmt"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/stringutil"
)

// InferredConfig is a shim that sets up required default values if they aren't set.
type InferredConfig struct {
	ServiceName     string
	ServiceEnv      string
	ClusterName     string
	Cluster         string
	GitRemote       string
	GitRefHash      string
	DockerRegistry  string
	DockerTag       string
	FQDN            string
	ContainerImage  string
	Accessibility   Accessibility
	Autoscale       Flag
	Replicas        string
	MinReplicas     string
	MaxReplicas     string
	CPUThreshold    string
	MemoryThreshold string

	Ports []ContainerPort

	ContainerPort  string
	ContainerProto string
	HTTPPort       string
	HTTPSPort      string

	RegistryHost string
}

// IsProdLike returns if the config is prod like.
func (ic InferredConfig) IsProdLike() bool {
	return stringutil.EqualsCaseless(ic.ServiceEnv, env.ServiceEnvPreprod) ||
		stringutil.EqualsCaseless(ic.ServiceEnv, env.ServiceEnvBeta) ||
		stringutil.EqualsCaseless(ic.ServiceEnv, env.ServiceEnvProd)
}

// GetBaseDomain returns the last two domain segments of the FQDN.
func (ic InferredConfig) GetBaseDomain() string {
	if ic.IsProdLike() {
		return "blend.com"
	}
	return "centrio.com"
}

// GetClusterName returns a value or an inferred default if unset.
func (ic InferredConfig) GetClusterName() string {
	if len(ic.ClusterName) > 0 {
		return ic.ClusterName
	}
	switch ic.ServiceEnv {
	case env.ServiceEnvTest, env.ServiceEnvSandbox, env.ServiceEnvDev:
		return fmt.Sprintf("%s.k8s.%s", ic.ServiceEnv, ic.GetBaseDomain())
	default:
		return fmt.Sprintf("k8s.%s.%s", ic.ServiceEnv, ic.GetBaseDomain())
	}
}

// GetCluster returns a value or an inferred default if unset.
func (ic InferredConfig) GetCluster() string {
	if len(ic.Cluster) > 0 {
		return ic.Cluster
	}
	return ic.ServiceEnv
}

// GetDockerRegistry returns a value or an inferred default if unset.
func (ic InferredConfig) GetDockerRegistry() string {
	if len(ic.DockerRegistry) > 0 {
		return ic.DockerRegistry
	}
	return ic.GetRegistryHost()
}

// GetDockerTag returns a value or an inferred default if unset.
func (ic InferredConfig) GetDockerTag() string {
	if len(ic.DockerTag) > 0 {
		return ic.DockerTag
	}
	return ic.GitRefHash
}

// GetFQDN returns a value or an inferred default if unset.
func (ic InferredConfig) GetFQDN() string {
	if len(ic.FQDN) > 0 {
		return ic.FQDN
	}

	return fmt.Sprintf("%s.%s", ic.ServiceName, ic.GetClusterName())
}

// GetContainerImage returns a value or an inferred default if unset.
func (ic InferredConfig) GetContainerImage() string {
	if len(ic.ContainerImage) > 0 {
		return ic.ContainerImage
	}
	return fmt.Sprintf("%s/%s:%s", ic.GetDockerRegistry(), ic.ServiceName, ic.GetDockerTag())
}

// GetCPUThreshold returns the configured cpu threshold or a default if autoscale is enabled.
func (ic InferredConfig) GetCPUThreshold() string {
	if len(ic.CPUThreshold) > 0 {
		return ic.CPUThreshold
	}
	if ic.Autoscale == FlagEnabled {
		return DefaultCPUThreshold
	}
	return ""
}

// GetMemoryThreshold returns the memory threshold
func (ic InferredConfig) GetMemoryThreshold() string {
	if len(ic.MemoryThreshold) > 0 {
		return ic.MemoryThreshold
	}
	if ic.Autoscale == FlagEnabled {
		return DefaultMemoryThreshold
	}
	return ""
}

// GetMinReplicas returns the configured max replicas or a default if autoscale is enabled.
func (ic InferredConfig) GetMinReplicas() string {
	if len(ic.MinReplicas) > 0 {
		return ic.MinReplicas
	}
	if ic.Autoscale == FlagEnabled {
		return ic.Replicas
	}
	return ""
}

// GetMaxReplicas returns the configured max replicas or a default if autoscale is enabled.
func (ic InferredConfig) GetMaxReplicas() string {
	if len(ic.MaxReplicas) > 0 {
		return ic.MaxReplicas
	}
	if ic.Autoscale == FlagEnabled {
		return DefaultMaxReplicas
	}
	return ""
}

// GetContainerPort returns the container port for the service.
func (ic InferredConfig) GetContainerPort() string {
	if len(ic.ContainerPort) > 0 {
		return ic.ContainerPort
	}
	if len(ic.Ports) > 0 {
		return ""
	}
	if ic.Accessibility == AccessibilityNone {
		return ""
	}
	return DefaultContainerPort
}

// GetContainerProto returns the container proto for the service.
func (ic InferredConfig) GetContainerProto() string {
	if len(ic.ContainerProto) > 0 {
		return ic.ContainerProto
	}
	if len(ic.Ports) > 0 {
		return ""
	}
	if ic.Accessibility == AccessibilityNone {
		return ""
	}
	return DefaultContainerProto
}

// GetHTTPPort returns the HTTP port for the service.
func (ic InferredConfig) GetHTTPPort() string {
	if len(ic.HTTPPort) > 0 {
		return ic.HTTPPort
	}
	if len(ic.Ports) > 0 {
		return ""
	}
	if ic.Accessibility == AccessibilityNone {
		return ""
	}
	return DefaultHTTPPort
}

// GetHTTPSPort returns the HTTPS port for the service.
func (ic InferredConfig) GetHTTPSPort() string {
	if len(ic.HTTPSPort) > 0 {
		return ic.HTTPSPort
	}
	if len(ic.Ports) > 0 {
		return ""
	}
	if ic.Accessibility == AccessibilityNone {
		return ""
	}
	return DefaultHTTPSPort
}

// GetRegistryHost returns the registry host
func (ic InferredConfig) GetRegistryHost() string {
	if len(ic.RegistryHost) > 0 {
		return ic.RegistryHost
	}
	if ic.ServiceEnv == ServiceEnvMinikube {
		return env.Env().String(EnvVarRegistryHost)
	}
	return fmt.Sprintf("registry.%s", ic.GetClusterName())
}

// InheritFrom inherits the values from another config, returning the inferred config.
func (ic InferredConfig) InheritFrom(other *Config) *Config {
	ic.ServiceName = other.ServiceName
	ic.ServiceEnv = other.ServiceEnv
	ic.ClusterName = other.ClusterName
	ic.Cluster = other.Cluster
	ic.GitRemote = other.GitRemote
	ic.GitRefHash = other.GitRefHash
	ic.DockerRegistry = other.DockerRegistry
	ic.DockerTag = other.DockerTag
	ic.FQDN = other.FQDN
	ic.ContainerImage = other.ContainerImage
	ic.Autoscale = other.Autoscale
	ic.Accessibility = other.Accessibility
	ic.CPUThreshold = other.CPUThreshold
	ic.Replicas = other.Replicas
	ic.MinReplicas = other.Replicas
	ic.MaxReplicas = other.MaxReplicas
	ic.Ports = other.Ports
	ic.ContainerPort = other.ContainerPort
	ic.ContainerProto = other.ContainerProto
	ic.RegistryHost = other.RegistryHost

	return ic.Config()
}

// Config represents the configuration variables this config exports.
func (ic InferredConfig) Config() *Config {
	return &Config{
		ServiceName:     ic.ServiceName,
		ServiceEnv:      ic.ServiceEnv,
		GitRefHash:      ic.GitRefHash,
		Autoscale:       ic.Autoscale,
		ClusterName:     ic.GetClusterName(),
		Cluster:         ic.GetCluster(),
		GitRemote:       ic.GitRemote,
		DockerRegistry:  ic.GetDockerRegistry(),
		DockerTag:       ic.GetDockerTag(),
		FQDN:            ic.GetFQDN(),
		ContainerImage:  ic.GetContainerImage(),
		CPUThreshold:    ic.GetCPUThreshold(),
		MemoryThreshold: ic.GetMemoryThreshold(),
		MinReplicas:     ic.GetMinReplicas(),
		MaxReplicas:     ic.GetMaxReplicas(),
		ContainerPort:   ic.GetContainerPort(),
		ContainerProto:  ic.GetContainerProto(),
		RegistryHost:    ic.GetRegistryHost(),
	}
}
