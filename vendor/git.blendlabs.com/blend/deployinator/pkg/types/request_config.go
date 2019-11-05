package types

// RequestConfig represents the configuration for a service that can be posted to the api
type RequestConfig struct {
	// ServiceName is the name of the service. This won't be exported to `types.Config` object through `Config()`.
	ServiceName string `json:"serviceName,omitempty" yaml:"serviceName,omitempty"`
	// ServiceType determines how we should provision the service, defaults to `deployment`.
	ServiceType ServiceType `json:"serviceType,omitempty" yaml:"serviceType,omitempty"`
	//Labels are the the user custom labels
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// ProjectName is the name of the project. This won't be exported to the `types.Config` object through the `Config()`
	ProjectName string `json:"projectName,omitempty" yaml:"projectName,omitempty"`

	// SourcePath determines where in the builder filesystem we should check out source code.
	SourcePath string `json:"sourcePath,omitempty" yaml:"sourcePath,omitempty"`
	// BuildMode is the builder build mode.
	BuildMode BuildMode `json:"buildMode,omitempty" yaml:"buildMode,omitempty"`
	// ServiceConfigPath is the file name to read the service configuration variables out of.
	ServiceConfigPath string `json:"serviceConfigPath,omitempty" yaml:"serviceConfigPath,omitempty"`
	// GitRef is the branch / tag / sha1 hash to clone.
	GitRef string `json:"gitRef,omitempty" yaml:"gitRef,omitempty"`
	// GitRemote is the git repo to pull code from.
	GitRemote string `json:"gitRemote,omitempty" yaml:"gitRemote,omitempty"`
	// GitBase is the base branch
	GitBase string `json:"gitBase,omitempty" yaml:"gitBase,omitempty"`
	// Dockerfile is the docker image file to build
	Dockerfile string `json:"dockerfile,omitempty" yaml:"dockerfile,omitempty"`
	// FQDN is the fully qualified domain name we should map to the app.
	// If unset, the default is ${SERVICE_NAME}.${CLUSTER_NAME}.
	FQDN string `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`
	// SANs are subject alternative names
	SANs []string `json:"sans,omitempty" yaml:"sans,omitempty"`
	// FileMountPath determines the mount path for files from secrets.
	FileMountPath string `json:"fileMountPath,omitempty" yaml:"fileMountPath,omitempty"`
	// Accessibility is the accessibility of the app, either public, private, internal or none.
	Accessibility Accessibility `json:"accessibility,omitempty" yaml:"accessibility,omitempty"`
	// LoadBalancerSourceRanges will limit the ips of our accessibility:public service
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty" yaml:"loadBalancerSourceRanges,omitempty"`
	// ContainerImage is the image we're going to associate with the kube deployment.
	ContainerImage string `json:"containerImage,omitempty" yaml:"containerImage,omitempty"`

	// ContainerPort is the port to expose on the kube deployment, it can be either an integer port value or the string "none".
	// It is a simplified configuration for the `Port` field on the container spec. If you need more options, use `Ports`.
	ContainerPort string `json:"containerPort,omitempty" yaml:"containerPort,omitempty"`
	// ContainerProto determines what proto to use for the port, the default is "TCP".
	ContainerProto string `json:"containerProto,omitempty" yaml:"containerProto,omitempty"`

	// AWSRegion is the current aws region.
	AWSRegion string `json:"awsRegion,omitempty" yaml:"awsRegion,omitempty"`
	// AWSPolicyARN sets an aws role arn to be provisioned for the running service.
	// If it is set, a sidecar container will be provisioned to mount credentials for the given role arn.
	AWSPolicyARN string `json:"awsPolicyArn,omitempty" yaml:"awsPolicyArn,omitempty"`
	// AWSCredentialsMountPath is where we mount aws credentials.
	AWSCredentialsMountPath string `json:"awsCredentialsMountPath,omitempty" yaml:"awsCredentialsMountPath,omitempty"`

	// Autoscale determines if we should provision an autoscaling group for the service.
	Autoscale Flag `json:"autoscale,omitempty" yaml:"autoscale,omitempty"`
	// Replicas determines the runtime replica count, it can be inferred.
	Replicas string `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	// MinReplicas is the minimum number of replicas to provision.
	MinReplicas string `json:"minReplicas,omitempty" yaml:"minReplicas,omitempty"`
	// MaxReplicas is the maximum number of replicas to provision.
	MaxReplicas string `json:"maxReplicas,omitempty" yaml:"maxReplicas,omitempty"`
	// CPUThreshold is the threshold to monitor to manage scaling.
	CPUThreshold string `json:"cpuThreshold,omitempty" yaml:"cpuThreshold,omitempty"`
	// MemoryThreshold is the threshold to monitor to manage scaling.
	MemoryThreshold string `json:"memoryThreshold,omitempty" yaml:"memoryThreshold,omitempty"`
	// AutoscaleMetric defines external metrics for the autoscaler
	AutoscaleMetrics []AutoscaleMetric `json:"autoscaleMetrics,omitempty" yaml:"autoscaleMetrics,omitempty"`
	// DeploymentStrategy is the strategy to use when updating a deployment, defaults to 'RollingUpdate'
	DeploymentStrategy string `json:"deploymentStrategy,omitEmpty" yaml:"deploymentStrategy,omitempty"`

	// Pod level options
	// ActiveDeadlineSeconds is the grace period allowed for service start.
	ActiveDeadlineSeconds string `json:"activeDeadlineSeconds,omitempty" yaml:"activeDeadlineSeconds,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	// RestartPolicy determines if the app should be restarted in the event of failure.
	RestartPolicy string `json:"restartPolicy,omitempty" yaml:"restartPolicy,omitempty"`
	// Volumes represent pod wide volumes attachments.
	Volumes []Volume `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	// HostAliases represents additional entries to the pod's hosts file.
	HostAliases []HostAlias `json:"hostAliases,omitempty" yaml:"hostAliases,omitempty"`

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

	// DatadogHost is the datadog agent host to use for notifications.
	DatadogHost string `json:"datadogHost,omitempty" yaml:"datadogHost,omitempty"`

	// Route53TTL is the amount of time in seconds that DNS records for this service should be cached
	Route53TTL *int64 `json:"route53TTL,omitempty" yaml:"route53TTL,omitempty"`

	// Storage configures a persistent volume for the service
	Storage []StorageConfig `json:"storage,omitempty" yaml:"storage,omitempty"`

	// ServiceLBConnectionIdleTimeout is the time for the load balancer to close connection when no data is sent over the connection, default is 60 seconds
	ServiceLBConnectionIdleTimeout string `json:"serviceLBConnectionIdleTimeout,omitempty" yaml:"serviceLBConnectionIdleTimeout,omitempty"`

	// Protocol is the protocol of the server
	Protocol Protocol `json:"protocol,omitempty" yaml:"protocol,omitempty"`

	// ProxyProtocol enables or disables the proxy protocol on the load balancer
	ProxyProtocol Flag `json:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty"`
	// HTTPRedirect tells whether to include the http redirect server
	HTTPRedirect Flag `json:"httpRedirect,omitempty" yaml:"httpRedirect"`

	// BrowserAuthentication tells whether or not the service should authenticate visitors as Blend employees
	BrowserAuthentication Flag `json:"browserAuthentication,omitempty" yaml:"browserAuthentication,omitEmpty"`

}

// Config returns the request config as a full config
func (r RequestConfig) Config() *Config {
	return &Config{
		SourcePath:               r.SourcePath,
		BuildMode:                r.BuildMode,
		ServiceConfigPath:        r.ServiceConfigPath,
		GitRef:                   r.GitRef,
		GitRemote:                r.GitRemote,
		GitBase:                  r.GitBase,
		Dockerfile:               r.Dockerfile,
		FQDN:                     r.FQDN,
		SANs:                     r.SANs,
		FileMountPath:            r.FileMountPath,
		Accessibility:            r.Accessibility,
		LoadBalancerSourceRanges: r.LoadBalancerSourceRanges,
		ContainerImage:           r.ContainerImage,
		Labels:                   r.Labels,
		ContainerPort:            r.ContainerPort,
		ContainerProto:           r.ContainerProto,

		AWSRegion:               r.AWSRegion,
		AWSPolicyARN:            r.AWSPolicyARN,
		AWSCredentialsMountPath: r.AWSCredentialsMountPath,

		Autoscale:          r.Autoscale,
		Replicas:           r.Replicas,
		MinReplicas:        r.MinReplicas,
		MaxReplicas:        r.MaxReplicas,
		CPUThreshold:       r.CPUThreshold,
		MemoryThreshold:    r.MemoryThreshold,
		AutoscaleMetrics:   r.AutoscaleMetrics,
		DeploymentStrategy: r.DeploymentStrategy,

		ActiveDeadlineSeconds: r.ActiveDeadlineSeconds,
		NodeSelector:          r.NodeSelector,
		RestartPolicy:         r.RestartPolicy,
		Volumes:               r.Volumes,
		HostAliases:           r.HostAliases,

		JobActiveDeadlineSeconds: r.JobActiveDeadlineSeconds,

		ProgressDeadlineSeconds: r.ProgressDeadlineSeconds,

		Args:                           r.Args,
		Command:                        r.Command,
		Env:                            r.Env,
		LivenessProbe:                  r.LivenessProbe,
		ReadinessProbe:                 r.ReadinessProbe,
		Ports:                          r.Ports,
		Resources:                      r.Resources,
		VolumeMounts:                   r.VolumeMounts,
		WorkingDir:                     r.WorkingDir,
		DatadogHost:                    r.DatadogHost,
		Route53TTL:                     r.Route53TTL,
		Storage:                        r.Storage,
		ServiceLBConnectionIdleTimeout: r.ServiceLBConnectionIdleTimeout,
		Protocol:                       r.Protocol,

		ProxyProtocol: r.ProxyProtocol,
		HTTPRedirect:  r.HTTPRedirect,

		BrowserAuthentication: r.BrowserAuthentication,
	}
}
