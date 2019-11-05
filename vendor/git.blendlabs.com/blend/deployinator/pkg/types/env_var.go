package types

// EnvVar represents an environment variable present in a Container.
type EnvVar struct {
	Name      string        `json:"name" yaml:"name"`
	Value     string        `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

// EnvVarSource represents a source for the value of an EnvVar.
type EnvVarSource struct {
	FieldRef         *ObjectFieldSelector   `json:"fieldRef,omitempty" yaml:"fieldRef,omitempty"`
	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty" yaml:"resourceFieldRef,omitempty"`
	ConfigMapKeyRef  *ConfigMapKeySelector  `json:"configMapKeyRef,omitempty" yaml:"configMapKeyRef,omitempty"`
	SecretKeyRef     *SecretKeySelector     `json:"secretKeyRef,omitempty" yaml:"secretKeyRef,omitempty"`
}

// ObjectFieldSelector selects an APIVersioned field of an object.
type ObjectFieldSelector struct {
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	FieldPath  string `json:"fieldPath" yaml:"fieldPath"`
}

// ResourceFieldSelector represents container resources (cpu, memory) and their output format
type ResourceFieldSelector struct {
	ContainerName string   `json:"containerName,omitempty" yaml:"containerName,omitempty"`
	Resource      string   `json:"resource" yaml:"resource"`
	Divisor       Quantity `json:"divisor,omitempty" yaml:"divisor,omitempty"`
}

// ConfigMapKeySelector Selects a key from a ConfigMap.
type ConfigMapKeySelector struct {
	LocalObjectReference `json:",inline" yaml:",inline"`
	Key                  string `json:"key" yaml:"key"`
	Optional             *bool  `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// SecretKeySelector selects a key of a Secret.
type SecretKeySelector struct {
	LocalObjectReference `json:",inline" yaml:",inline"`
	Key                  string `json:"key" yaml:"key"`
	Optional             *bool  `json:"optional,omitempty" yaml:"optional,omitempty"`
}
