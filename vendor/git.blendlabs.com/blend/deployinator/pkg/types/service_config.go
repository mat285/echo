package types

// ServiceConfig represents the config variables service owners can tweak for their services.
// It is typically read out of a `service.yml` file in the root of the git repo.
type ServiceConfig struct {
	// InheritsFrom is a list of `service.yml` files that will be merged into this configuration.
	// They will be merged in the order they appear, so the "last" one wins in the case of a field shared
	// between multiple configuration files.
	InheritsFrom      []string      `json:"inheritsFrom,omitempty" yaml:"inheritsFrom,omitempty"`
	ServiceName       string        `json:"serviceName,omitempty" yaml:"serviceName,omitempty"`
	ECRAuthentication Flag          `json:"ecrAuthentication,omitempty" yaml:"ecrAuthentication,omitempty"`
	FileMountPath     string        `json:"fileMountPath,omitempty" yaml:"fileMountPath,omitempty"`
	Dockerfile        string        `json:"dockerfile,omitempty" yaml:"dockerfile,omitempty"`
	DockerRegistry    string        `json:"dockerRegistry,omitempty" yaml:"dockerRegistry,omitempty"`
	FQDN              string        `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`
	SANs              []string      `json:"sans,omitempty" yaml:"sans,omitempty"`
	Accessibility     Accessibility `json:"accessibility,omitempty" yaml:"accessibility,omitempty"`
	// LoadBalancerSourceRanges will limit the ips of our accessibility:public service if defined
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty" yaml:"loadBalancerSourceRanges,omitempty"`
	ContainerImage           string   `json:"containerImage,omitempty" yaml:"containerImage,omitempty"`
	ContainerPort            string   `json:"containerPort,omitempty" yaml:"containerPort,omitempty"`
	ContainerProto           string   `json:"containerProto,omitempty" yaml:"containerProto,omitempty"`
	// SlackChannel is a slack channel to use for notifications.
	SlackChannel string `json:"slackChannel,omitempty" yaml:"slackChannel,omitempty"`
	//Labels are the the user custom labels
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// AWSRegion is the current aws region.
	AWSRegion string `json:"awsRegion,omitempty" yaml:"awsRegion,omitempty"`
	// AWSPolicyARN sets an aws role arn to be provisioned for the running service.
	AWSPolicyARN string `json:"awsPolicyArn,omitempty" yaml:"awsPolicyArn,omitempty"`
	// AWSCredentialsMountPath is where we mount aws credentials.
	AWSCredentialsMountPath string `json:"awsCredentialsMountPath,omitempty" yaml:"awsCredentialsMountPath,omitempty"`

	/// Autoscale determines if we should provision an autoscaling group for the service.
	Autoscale          Flag              `json:"autoscale,omitempty" yaml:"autoscale,omitempty"`
	Replicas           string            `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	MaxReplicas        string            `json:"maxReplicas,omitempty" yaml:"maxReplicas,omitempty"`
	MinReplicas        string            `json:"minReplicas,omitempty" yaml:"minReplicas,omitempty"`
	CPUThreshold       string            `json:"cpuThreshold,omitempty" yaml:"cpuThreshold,omitempty"`
	MemoryThreshold    string            `json:"memoryThreshold,omitempty" yaml:"memoryThreshold,omitempty"`
	AutoscaleMetrics   []AutoscaleMetric `json:"autoscaleMetrics,omitempty" yaml:"autoscaleMetrics,omitempty"`
	DeploymentStrategy string            `json:"deploymentStrategy,omitEmpty" yaml:"deploymentStrategy,omitempty"`

	// ActiveDeadlineSeconds is the grace period allowed for service start.
	ActiveDeadlineSeconds string `json:"activeDeadlineSeconds,omitempty" yaml:"activeDeadlineSeconds,omitempty"`
	// ImagePullSecrets represent secrets used during image pulls like docker registry auth tokens.
	ImagePullSecrets []LocalObjectReference `json:"imagePullSecrets,omitempty" yaml:"imagePullSecrets,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	// RestartPolicy determines if the app should be restarted in the event of failure.
	RestartPolicy string `json:"restartPolicy,omitempty" yaml:"restartPolicy,omitempty"`
	// Volumes represent pod wide volumes attachments.
	Volumes []Volume `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	// TerminationGracePeriodSeconds is the grace period for termination of a service
	TerminationGracePeriodSeconds string `json:"terminationGracePeriodSeconds,omitempty" yaml:"terminationGracePeriodSeconds,omitempty"`
	// HostAliases represents additional entries to the pod's hosts file.
	HostAliases []HostAlias `json:"hostAliases,omitempty" yaml:"hostAliases,omitempty"`

	// Storage configures a persistent volume for the service
	Storage []StorageConfig `json:"storage,omitempty" yaml:"storage,omitempty"`

	// ServiceLBConnectionIdleTimeout is the time for the load balancer to close connection when no data is sent over the connection, default is 60 seconds
	ServiceLBConnectionIdleTimeout string `json:"serviceLBConnectionIdleTimeout,omitempty" yaml:"serviceLBConnectionIdleTimeout,omitempty"`

	// Protocol is the protocol of the server
	Protocol Protocol `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	// ProxyProtocol tells whether proxy protocol should be enabled or not
	ProxyProtocol Flag `json:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty"`
	// HTTPRedirect tells whether to include the http redirect server
	HTTPRedirect Flag `json:"httpRedirect,omitempty" yaml:"httpRedirect"`
	// BrowserAuthentication tells whether or not the service should authenticate visitors as Blend employees
	BrowserAuthentication Flag `json:"browserAuthentication,omitempty" yaml:"browserAuthentication,omitempty"`

	// Job level options
	// JobActiveDeadlineSeconds is the amount of time a job may be active before being terminated.
	JobActiveDeadlineSeconds string `json:"jobActiveDeadlineSeconds,omitempty" yaml:"jobActiveDeadlineSeconds,omitempty"`

	// Deployment level options
	// ProgressDeadlineSeconds is the amount of time a deployment has to make progress before being counted as a failure on update
	ProgressDeadlineSeconds string `json:"progressDeadlineSeconds,omitempty" yaml:"progressDeadlineSeconds,omitempty"`

	// Container level options

	// Arguments to the entrypoint or `Command`.
	Args []string `json:"args,omitempty" yaml:"args,omitempty"`
	// Command represents an Entrypoint array. Not executed within a shell.
	// The docker image's ENTRYPOINT is used if this is not provided.
	Command []string `json:"command,omitempty" yaml:"command,omitempty"`
	// Env represents environment variables to be injected into running pod containers.
	Env []EnvVar `json:"env,omitempty" yaml:"env,omitempty"`
	// LivenessProbe represents a probe to determine if the app is healthy.
	LivenessProbe *Probe `json:"livenessProbe,omitempty" yaml:"livenessProbe,omitempty"`
	// Ports defines ports to expose on the container; it is coalesced with `ContainerPort`.
	Ports []ContainerPort `json:"ports,omitempty" yaml:"ports,omitempty"`
	// ReadinessProbe represents a probe to determine if the app is ready.
	ReadinessProbe *Probe `json:"readinessProbe,omitempty" yaml:"readinessProbe,omitempty"`
	// Resources represents limits and requirements on resources for the running app.
	Resources *ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
	// VolumeMounts represent container specifci mounts for volums provisioned in `Volumes`
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
	// WorkingDir is the container's working directory.
	WorkingDir string `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	// SmokeSignalCheckPath is the relative path to the check yaml
	SmokeSignalCheckPath string `json:"smokeSignalCheckPath,omitempty" yaml:"smokeSignalCheckPath,omitempty"`
	// ELBAccessLogsBucket is the bucket name for the elb access logs
	ELBAccessLogsBucket string `json:"elbAccessLogsBucket,omitempty" yaml:"elbAccessLogsBucket,omitempty"`
	// SidecarResources are the resource requirements for the sidecar containers
	SidecarResources map[string]*ResourceRequirements `json:"sidecarResources,omitempty" yaml:"sidecarResources,omitempty"`
}

// Config returns the service config as a full config.
func (sc ServiceConfig) Config() *Config {
	return &Config{
		ServiceName:              sc.ServiceName,
		ECRAuthentication:        sc.ECRAuthentication,
		FileMountPath:            sc.FileMountPath,
		Dockerfile:               sc.Dockerfile,
		DockerRegistry:           sc.DockerRegistry,
		Replicas:                 sc.Replicas,
		FQDN:                     sc.FQDN,
		SANs:                     sc.SANs,
		Accessibility:            sc.Accessibility,
		LoadBalancerSourceRanges: sc.LoadBalancerSourceRanges,
		ContainerImage:           sc.ContainerImage,
		ContainerPort:            sc.ContainerPort,
		ContainerProto:           sc.ContainerProto,
		SlackChannel:             sc.SlackChannel,
		Labels:                   sc.Labels,
		AWSRegion:                sc.AWSRegion,
		AWSPolicyARN:             sc.AWSPolicyARN,
		AWSCredentialsMountPath:  sc.AWSCredentialsMountPath,

		Autoscale:          sc.Autoscale,
		MaxReplicas:        sc.MaxReplicas,
		MinReplicas:        sc.MinReplicas,
		CPUThreshold:       sc.CPUThreshold,
		MemoryThreshold:    sc.MemoryThreshold,
		AutoscaleMetrics:   sc.AutoscaleMetrics,
		DeploymentStrategy: sc.DeploymentStrategy,

		ActiveDeadlineSeconds:         sc.ActiveDeadlineSeconds,
		NodeSelector:                  sc.NodeSelector,
		RestartPolicy:                 sc.RestartPolicy,
		Volumes:                       sc.Volumes,
		TerminationGracePeriodSeconds: sc.TerminationGracePeriodSeconds,
		HostAliases:                   sc.HostAliases,

		Storage:                        sc.Storage,
		ServiceLBConnectionIdleTimeout: sc.ServiceLBConnectionIdleTimeout,
		Protocol:                       sc.Protocol,
		ProxyProtocol:                  sc.ProxyProtocol,
		HTTPRedirect:                   sc.HTTPRedirect,

		JobActiveDeadlineSeconds: sc.JobActiveDeadlineSeconds,

		ProgressDeadlineSeconds: sc.ProgressDeadlineSeconds,

		Args:             sc.Args,
		Command:          sc.Command,
		Env:              sc.Env,
		LivenessProbe:    sc.LivenessProbe,
		Ports:            sc.Ports,
		ReadinessProbe:   sc.ReadinessProbe,
		Resources:        sc.Resources,
		SidecarResources: sc.SidecarResources,
		VolumeMounts:     sc.VolumeMounts,
		WorkingDir:       sc.WorkingDir,

		SmokeSignalCheckPath: sc.SmokeSignalCheckPath,
		ELBAccessLogsBucket:  sc.ELBAccessLogsBucket,

		BrowserAuthentication: sc.BrowserAuthentication,
	}
}
