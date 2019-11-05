package types

import (
	"fmt"
	"strings"
	"time"

	"git.blendlabs.com/blend/deployinator/pkg/kube"
	"github.com/blend/go-sdk/collections"
	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/stringutil"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/validation"
)

// ComposeConfigs composes a given list of configs
func ComposeConfigs(configs ...*Config) *Config {
	final := &Config{}
	for _, config := range configs {
		if config != nil {
			final = final.InheritFrom(config.Config())
		}
	}
	return final
}

// MergeConfigs is an alias to compose configs.
func MergeConfigs(configs ...*Config) *Config {
	return ComposeConfigs(configs...)
}

// Config represents the full configuration variable set.
type Config struct {
	// SchemaVersion represents the current schema version.
	SchemaVersion string `json:"schemaVersion,omitempty" yaml:"schemaVersion,omitempty"`
	// KubectlConfig is the path to the kubectl config for the builder.
	KubectlConfig string ` json:"kubectlConfig,omitempty" yaml:"kubectlConfig,omitempty"`
	// ServiceType determines how we should provision the service, defaults to `deployment`.
	ServiceType ServiceType `json:"serviceType,omitempty" yaml:"serviceType,omitempty"`
	// SourcePath determines where in the builder filesystem we should check out source code.
	SourcePath string `json:"sourcePath,omitempty" yaml:"sourcePath,omitempty"`
	// BuildMode is the builder build mode.
	BuildMode BuildMode `json:"buildMode,omitempty" yaml:"buildMode,omitempty"`
	// DeployID is a unique identifier for a given deploy.
	DeployID string `json:"deployID,omitempty" yaml:"deployID,omitempty"`
	// CreatedBy is the target (optional) that created the resource.
	CreatedBy string `json:"createdBy,omitempty" yaml:"createdBy,omitempty"`
	// DeployedBy is the target (optional) that started the build.
	DeployedBy string `json:"deployedBy,omitempty" yaml:"deployedBy,omitempty"`
	// DeployedAt is the time the deploy occurred.
	DeployedAt time.Time `json:"deployedAt,omitempty" yaml:"deployedAt,omitempty"`
	// ClusterName is the current cluster name.
	ClusterName string `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`
	// Cluster is the current cluster short name.
	Cluster string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	// ServiceConfigPath is the file name to read the service configuration variables out of.
	ServiceConfigPath string `json:"serviceConfigPath,omitempty" yaml:"serviceConfigPath,omitempty"`
	// ServiceEnv is the current service environment.
	ServiceEnv string `json:"serviceEnv,omitempty" yaml:"serviceEnv,omitempty"`
	// ServiceName is the name of the service.
	ServiceName string `json:"serviceName,omitempty" yaml:"serviceName,omitempty"`
	//Labels are the the user custom labels
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// GitRef is the branch / tag / sha1 hash to clone.
	GitRef string `json:"gitRef,omitempty" yaml:"gitRef,omitempty"`
	// GitRemote is the git repo to pull code from.
	GitRemote string `json:"gitRemote,omitempty" yaml:"gitRemote,omitempty"`
	// GitBase is the base branch
	GitBase string `json:"gitBase,omitempty" yaml:"gitBase,omitempty"`
	// GitRefHash is a typically a runtime value that is the current
	// short sha1 hash of specified `git-ref`.
	GitRefHash string `json:"gitRefHash,omitempty" yaml:"gitRefHash,omitempty"`
	// ECRAuthentication indicates that ecr is needed for docker build
	ECRAuthentication Flag `json:"ecrAuthentication,omitempty" yaml:"ecrAuthentication,omitempty"`
	// Dockerfile is the docker image file to build
	Dockerfile string `json:"dockerfile,omitempty" yaml:"dockerfile,omitempty"`
	// DockerRegistry is the registry we should push docker files to.
	// It is used in the form ${DOCKER_REGISTRY}/${SERVICE_NAME}:${DOCKER_TAG}
	DockerRegistry string `json:"dockerRegistry,omitempty" yaml:"dockerRegistry,omitempty"`
	// DockerTag is the tag used in the container image defaults.
	DockerTag string `json:"dockerTag,omitempty" yaml:"dockerTag,omitempty"`
	// Namespace is the current namespace, generally this resolves to
	// the default value of "blend" but you can set it for yourself if you want.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	// FQDN is the fully qualified domain name we should map to the app.
	// If unset, the default is ${SERVICE_NAME}.${CLUSTER_NAME}.
	FQDN string `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`
	// SANs are the subject alternative names for the service. If FQDN is unset, these are ignored
	SANs []string `json:"sans,omitempty" yaml:"sans,omitempty"`
	// FileMountPath determines the mount path for files from secrets.
	FileMountPath string `json:"fileMountPath,omitempty" yaml:"fileMountPath,omitempty"`
	// Accessibility is the accessibility of the app, either public, private, internal or none.
	Accessibility Accessibility `json:"accessibility,omitempty" yaml:"accessibility,omitempty"`
	// LoadBalancerSourceRanges will limit the ips of our accessibility:public service
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty" yaml:"loadBalancerSourceRanges,omitempty"`
	// ContainerImage is the image we're going to associate with the kube deployment.
	ContainerImage string `json:"containerImage,omitempty" yaml:"containerImage,omitempty"`
	// DisallowDeploy indicates whether a service can only be deployed from a project
	DisallowDeploy bool `json:"disallowDeploy" yaml:"disallowDeploy"`

	// ContainerPort is the port to expose on the kube deployment, it can be either an integer port value or the string "none".
	// It is a simplified configuration for the `Port` field on the container spec. If you need more options, use `Ports`.
	ContainerPort string `json:"containerPort,omitempty" yaml:"containerPort,omitempty"`
	// ContainerProto determines what proto to use for the port, the default is "TCP".
	ContainerProto string `json:"containerProto,omitempty" yaml:"containerProto,omitempty"`

	// AWSRegion is the current aws region.
	AWSRegion string `json:"awsRegion,omitempty" yaml:"awsRegion,omitempty"`
	// AWSPolicyARN sets an aws policy arn to be provisioned for the running service.
	// If it is set, a sidecar container will be provisioned to mount credentials for the given policy arn.
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
	// DeploymentStrategy is the strategy to use when updating a deployment. This can be 'RollingUpdate' or 'Recreate'
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
	// TerminationGracePeriodSeconds is the time between when a pod is sent SIGTERM and SIGKILL
	TerminationGracePeriodSeconds string `json:"terminationGracePeriodSeconds,omitempty" yaml:"terminationGracePeriodSeconds,omitempty"`
	// HostAliases represents additional entries to the pod's hosts file.
	HostAliases []HostAlias `json:"hostAliases,omitempty" yaml:"hostAliases,omitempty"`

	// Job level options
	// JobActiveDeadlineSeconds is the amount of time a job may be active before being terminated.
	JobActiveDeadlineSeconds string `json:"jobActiveDeadlineSeconds,omitempty" yaml:"jobActiveDeadlineSeconds,omitempty"`

	// ProgressDeadlineSeconds is the amount of time a deployment has to make progress before being counted as a failure on update
	ProgressDeadlineSeconds string `json:"progressDeadlineSeconds,omitempty" yaml:"progressDeadlineSeconds,omitempty"`

	// ELBAccessLogsBucket is the name of the AWS S3 bucket used for elastic load balancing access logs
	ELBAccessLogsBucket string `json:"elbAccessLogsBucket,omitempty" yaml:"elbAccessLogsBucket,omitempty"`

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

	// SidecarResources are the resource requirements for the sidecar containers
	SidecarResources map[string]*ResourceRequirements `json:"sidecarResources,omitempty" yaml:"sidecarResources,omitempty"`

	// SlackChannel is a slack channel to use for notifications.
	SlackChannel string `json:"slackChannel,omitempty" yaml:"slackChannel,omitempty"`
	// Emails is a list of email addresses use for notifications.
	Emails []string `json:"emails,omitempty" yaml:"emails,omitempty"`
	// DatadogHost is the datadog agent host to use for notifications.
	DatadogHost string `json:"datadogHost,omitempty" yaml:"datadogHost,omitempty"`

	// Route53TTL is the amount of time in seconds that DNS records for this service should be cached
	Route53TTL *int64 `json:"route53TTL,omitempty" yaml:"route53TTL,omitempty"`

	// ProjectName is the name of the project being built
	ProjectName string `json:"projectName,omitempty" yaml:"projectName,omitempty"`
	// ProjectSlackChannel is the channel for a project
	ProjectSlackChannel string `json:"projectSlackChannel,omitempty" yaml:"projectSlackChannel,omitempty"`
	// ProjectRun indicates this is part of a project
	ProjectRun bool `json:"projectRun,omitempty" yaml:"projectRun,omitempty"`

	// Schedule indicates in Cron format how often a task should be run
	Schedule string `json:"schedule,omitempty" yaml:"schedule,omitempty"`

	// DatabaseName is the name of the database
	DatabaseName string `json:"databaseName,omitempty" yaml:"databaseName,omitempty"`
	// DBStorage is the storage capacity of the database
	DBStorage int `json:"dbStorage,omitempty,string" yaml:"dbStorage,omitempty"`
	// DBRelaunch is the flag to indicate that a database is being restored
	DBRelaunch bool `json:"dbRelaunch,omitempty,string" yaml:"dbRelaunch,omitempty"`
	// DBRestore is the flag to indicate that a database is being restored
	DBRestore bool `json:"dbRestore,omitempty,string" yaml:"dbRestore,omitempty"`
	// DBBackupFileName is the flag to indicate that a database is being restored
	DBBackupFileName string `json:"dbBackupFileName,omitempty" yaml:"dbBackupFileName,omitempty"`

	// DeployNotifications determines whether to send Slack notifications for starting deploys
	DeployNotifications string `json:"deployNotifications,omitempty" yaml:"deployNotifications,omitempty"`

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
	// Options from environment variables

	// TODO: remove aws credentials from config?
	// AWSAccessKey is the AWS access key. Environment variable: AWS_ACCESS_KEY
	AWSAccessKey string `json:"-" yaml:"-"`
	// AWSSecretKey is the AWS secret key. Environment variable: AWS_SECRET_KEY
	AWSSecretKey string `json:"-" yaml:"-"`
	// RegistryHost is the registry host. Environment variable name: REGISTRY_HOST
	RegistryHost string `json:"registryHost,omitempty" yaml:"registryHost,omitempty"`

	// SmokeSignalCheckPath is the relative path to the check yaml
	SmokeSignalCheckPath string `json:"smokeSignalCheckPath,omitempty" yaml:"smokeSignalCheckPath,omitempty"`

	// Projects are the projects that the service is part of
	Projects []string `json:"-" yaml:"-"`

	// BrowserAuthentication tells whether or not the service should authenticate visitors as Blend employees
	BrowserAuthentication Flag `json:"browserAuthentication,omitempty" yaml:"browserAuthentication,omitempty"`
}

// Config returns the config and implements config provider.
func (c Config) Config() *Config {
	return &c
}

// ResourceLocator returns a name/namespace pair for use with the kube client.
func (c Config) ResourceLocator() (name, namespace string) {
	name = c.GetServiceName()
	namespace = c.GetNamespace()
	if c.GetBuildMode() == BuildModeLaunchDatabase {
		name = c.GetDatabaseName()
		namespace = DefaultNamespace
	}
	return
}

// VaultContext returns the vault context for the config.
func (c Config) VaultContext() map[string]interface{} {
	return map[string]interface{}{
		"serviceName": c.ServiceName,
		"serviceEnv":  c.GetServiceEnv(),
		"namespace":   c.GetNamespace(),
		"clusterName": c.GetClusterName(),
	}
}

// ShouldProvisionKubeDeployment returns if the service should be provisioned with a deployment.
func (c Config) ShouldProvisionKubeDeployment() bool {
	return c.GetServiceType() == ServiceTypeDeployment
}

// ShouldProvisionKubeTask returns if the service should be provisioned with a task pod.
func (c Config) ShouldProvisionKubeTask() bool {
	return c.GetServiceType() == ServiceTypeTaskTemplate
}

// ShouldProvisionKubeService returns if the service should be provisioned.
func (c Config) ShouldProvisionKubeService() bool {
	return c.GetAccessibility() != AccessibilityNone
}

// ShouldProvisionKubeAutoscaler returns if the horizontal pod autoscaler should be provisioned.
func (c Config) ShouldProvisionKubeAutoscaler() bool {
	return stringutil.EqualsCaseless(string(c.GetAutoscale()), string(FlagEnabled))
}

// ShouldProvisionKubeIngress returns if the ingress should be provisioned.
func (c Config) ShouldProvisionKubeIngress() bool {
	return c.GetAccessibility() == AccessibilityPrivate
}

// ShouldProvisionServiceLoadBalancer returns if the ingress should be provisioned.
func (c Config) ShouldProvisionServiceLoadBalancer() bool {
	return c.GetAccessibility() == AccessibilityPublic
}

// HasContainerPort returns if the config specifies a valid container port.
func (c Config) HasContainerPort() bool {
	containerPort := c.GetContainerPort()
	return len(containerPort) > 0 && containerPort != ContainerPortNone
}

// ShouldAuthECR returns if the docker build requires ecr access
func (c Config) ShouldAuthECR() bool {
	return c.GetECRAuthentication() == FlagEnabled
}

// Validate validates the config.
func (c Config) Validate() error {
	if len(c.ServiceName) == 0 && len(c.ProjectName) == 0 && len(c.DatabaseName) == 0 {
		return exception.New("`SERVICE_NAME`, PROJECT_NAME`, or `DATABASE_NAME` must be set, cannot continue")
	}
	for _, m := range c.AutoscaleMetrics {
		if !m.IsValid() {
			return exception.New("`autoscaleMetric.name` and `autoscaleMetric.threshold` must both be set")
		}
	}

	if ok, errs := c.HasValidLabels(); !ok {
		return exception.New(errs)
	}
	return nil
}

// IsProjectRun returns whether this is a project run
func (c Config) IsProjectRun() bool {
	return c.GetProjectRun()
}

// GetSchemaVersion returns the current schema version or a default.
func (c Config) GetSchemaVersion(defaults ...string) string {
	return thisOrThatOrDefault(c.SchemaVersion, "", defaults...)
}

// GetKubectlConfig returns the service config filename or a default.
func (c Config) GetKubectlConfig(defaults ...string) string {
	return thisOrThatOrDefault(c.KubectlConfig, "", defaults...)
}

// GetServiceType returns a value or a default.
func (c Config) GetServiceType(defaults ...ServiceType) ServiceType {
	if len(c.ServiceType) > 0 {
		return c.ServiceType
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return ServiceTypeUnset
}

// GetSourcePath returns a value or a default.
func (c Config) GetSourcePath(defaults ...string) string {
	return thisOrThatOrDefault(c.SourcePath, "", defaults...)
}

// GetBuildMode returns the build mode or a default.
func (c Config) GetBuildMode(defaults ...BuildMode) BuildMode {
	if len(c.BuildMode) > 0 {
		return c.BuildMode
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return BuildModeUnset
}

// DoesDockerBuild returns if the build mode does docker build
func (c Config) DoesDockerBuild() bool {
	switch c.BuildMode {
	case BuildModeDeploy,
		BuildModeCanaryDeploy,
		BuildModeRunTask,
		BuildModeScheduleTask,
		BuildModeDockerPush:
		return true
	default:
		return false
	}
}

// GetDeployID gets a deploy identifier or a default.
func (c Config) GetDeployID(defaults ...string) string {
	return thisOrThatOrDefault(c.DeployID, "", defaults...)
}

// GetAndSetNonEmptyDeployID returns the deploy id if there is one, or creates a new one, sets it on the config and returns it
func (c *Config) GetAndSetNonEmptyDeployID(defaults ...string) string {
	id := c.GetDeployID(defaults...)
	if len(id) == 0 {
		c.DeployID = rand.String(RandomStringLength)
		id = c.DeployID
	}
	return id
}

// GetCreatedBy returns the deployed by target or a default.
func (c Config) GetCreatedBy(defaults ...string) string {
	return thisOrThatOrDefault(c.CreatedBy, "", defaults...)
}

// GetDeployedBy returns the deployed by target or a default.
func (c Config) GetDeployedBy(defaults ...string) string {
	return thisOrThatOrDefault(c.DeployedBy, "", defaults...)
}

// GetDeployedAt returns the deployed at timestamp or a default.
func (c Config) GetDeployedAt(defaults ...time.Time) time.Time {
	if !c.DeployedAt.IsZero() {
		return c.DeployedAt
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return time.Time{}
}

// GetClusterName returns a value or a default if unset.
func (c Config) GetClusterName(defaults ...string) string {
	return thisOrThatOrDefault(c.ClusterName, "", defaults...)
}

// GetCluster returns a value or a default if unset.
func (c Config) GetCluster(defaults ...string) string {
	return thisOrThatOrDefault(c.Cluster, "", defaults...)
}

// GetServiceConfigPath returns the service config path or a default.
func (c Config) GetServiceConfigPath(defaults ...string) string {
	return thisOrThatOrDefault(c.ServiceConfigPath, "", defaults...)
}

// GetServiceEnv returns a value or a default if unset.
func (c Config) GetServiceEnv(defaults ...string) string {
	return thisOrThatOrDefault(c.ServiceEnv, "", defaults...)
}

// GetServiceName returns a value or a default if unset.
func (c Config) GetServiceName(defaults ...string) string {
	return thisOrThatOrDefault(c.ServiceName, "", defaults...)
}

// GetLabels returns the label map or a default map if unset
func (c Config) GetLabels(defaults ...map[string]string) map[string]string {
	if c.Labels != nil && len(c.Labels) > 0 {
		return c.Labels
	}
	if len(defaults) > 0 && defaults[0] != nil {
		return defaults[0]
	}
	return map[string]string{}
}

// GetProjectName returns a value of a default if unset
func (c Config) GetProjectName(defaults ...string) string {
	return thisOrThatOrDefault(c.ProjectName, "", defaults...)
}

// GetDatabaseName returns a value of a default if unset
func (c Config) GetDatabaseName(defaults ...string) string {
	return thisOrThatOrDefault(c.DatabaseName, "", defaults...)
}

// GetDBStorage returns a value of a default if unset
func (c Config) GetDBStorage(defaults ...int) int {
	return thisOrThatOrDefaultNonzeroInt(c.DBStorage, 0, defaults...)
}

// GetDBRelaunch returns a value of a default if unset
func (c Config) GetDBRelaunch(defaults ...bool) bool {
	return thisOrThatOrDefaultBool(c.DBRelaunch, false, defaults...)
}

// GetDBRestore returns a value of a default if unset
func (c Config) GetDBRestore(defaults ...bool) bool {
	return thisOrThatOrDefaultBool(c.DBRestore, false, defaults...)
}

// GetDBBackupFileName returns a value of a default if unset
func (c Config) GetDBBackupFileName(defaults ...string) string {
	return thisOrThatOrDefault(c.DBBackupFileName, "", defaults...)
}

// GetProjectSlackChannel returns a value of a default if unset
func (c Config) GetProjectSlackChannel(defaults ...string) string {
	return thisOrThatOrDefault(c.ProjectSlackChannel, "", defaults...)
}

// GetProjectRun returns a value of a default if unset
func (c Config) GetProjectRun(defaults ...bool) bool {
	return thisOrThatOrDefaultBool(c.ProjectRun, false, defaults...)
}

// GetGitRemote gets the git remote url or the default, `git.blendlabs.com/blend/${SERVICE_NAME}`.
func (c Config) GetGitRemote(defaults ...string) string {
	return thisOrThatOrDefault(c.GitRemote, "", defaults...)
}

// GetGitRef gets the git ref or the default, `master`.
func (c Config) GetGitRef(defaults ...string) string {
	return thisOrThatOrDefault(c.GitRef, "", defaults...)
}

// GetGitBase gets the git base or the default, ``.
func (c Config) GetGitBase(defaults ...string) string {
	return thisOrThatOrDefault(c.GitBase, "", defaults...)
}

// GetGitRefHash returns the current short sha1 hash for the given git ref.
func (c Config) GetGitRefHash(defaults ...string) string {
	return thisOrThatOrDefault(c.GitRefHash, "", defaults...)
}

// GetECRAuthentication returns whether ecr authentication is enabled or disabled
func (c Config) GetECRAuthentication(defaults ...Flag) Flag {
	return thisOrThatOrDefaultFlag(c.ECRAuthentication, "", defaults...)
}

// GetDockerfile returns a value or a default if unset.
func (c Config) GetDockerfile(defaults ...string) string {
	return thisOrThatOrDefault(c.Dockerfile, "", defaults...)
}

// GetDockerRegistry returns a value or a default if unset.
func (c Config) GetDockerRegistry(defaults ...string) string {
	return thisOrThatOrDefault(c.DockerRegistry, "", defaults...)
}

// GetDockerTag returns the current short sha1 hash for the given git ref, or "latest" if unset.
func (c Config) GetDockerTag(defaults ...string) string {
	return thisOrThatOrDefault(c.DockerTag, "", defaults...)
}

// GetNamespace returns a value or a default if unset.
func (c Config) GetNamespace(defaults ...string) string {
	return thisOrThatOrDefault(c.Namespace, "", defaults...)
}

// GetReplicas returns a value or a default if unset.
func (c Config) GetReplicas(defaults ...string) string {
	return thisOrThatOrDefault(c.Replicas, "", defaults...)
}

// GetFQDN returns a value or a default if unset.
func (c Config) GetFQDN(defaults ...string) string {
	return thisOrThatOrDefault(c.FQDN, "", defaults...)
}

// GetSANs returns the subject alternative names
func (c Config) GetSANs(defaults ...[]string) []string {
	return nonEmptySliceOrNil(thisOrThatOrDefaultSlice(c.SANs, nil, defaults...))
}

// GetDomainNames returns the domain names for the service. FQDN + SANs
func (c Config) GetDomainNames() []string {
	return append([]string{c.GetFQDN()}, c.GetSANs()...)
}

// GetFileMountPath returns the file mount path or a default.
func (c Config) GetFileMountPath(defaults ...string) string {
	return thisOrThatOrDefault(c.FileMountPath, "", defaults...)
}

// GetAccessibility returns a value or a default if unset.
func (c Config) GetAccessibility(defaults ...Accessibility) Accessibility {
	if len(c.Accessibility) > 0 {
		return c.Accessibility
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return ""
}

// GetLoadBalancerSourceRanges gets the whitelisted list of CIDRs for a load balancer
func (c Config) GetLoadBalancerSourceRanges(defaults ...[]string) []string {
	return nonEmptySliceOrNil(thisOrThatOrDefaultSlice(c.LoadBalancerSourceRanges, nil, defaults...))
}

// GetContainerImage returns a value or a default if unset.
func (c Config) GetContainerImage(defaults ...string) string {
	return thisOrThatOrDefault(c.ContainerImage, "", defaults...)
}

// GetContainerPort returns a value or a default if unset.
func (c Config) GetContainerPort(defaults ...string) string {
	return thisOrThatOrDefault(c.ContainerPort, "", defaults...)
}

// GetContainerProto returns a value or a default if unset.
func (c Config) GetContainerProto(defaults ...string) string {
	return thisOrThatOrDefault(c.ContainerProto, "", defaults...)
}

// GetAWSRegion returns a value or a default if unset.
func (c Config) GetAWSRegion(defaults ...string) string {
	return thisOrThatOrDefault(c.AWSRegion, "", defaults...)
}

// GetAWSPolicyARN returns a value or a default if unset.
func (c Config) GetAWSPolicyARN(defaults ...string) string {
	return thisOrThatOrDefault(c.AWSPolicyARN, "", defaults...)
}

// GetAWSCredentialsMountPath returns the path where we mount aws credentials
func (c Config) GetAWSCredentialsMountPath(defaults ...string) string {
	return thisOrThatOrDefault(c.AWSCredentialsMountPath, DefaultAwsCredentialsMountPath, defaults...)
}

// GetAutoscale returns a value or a default if unset.
func (c Config) GetAutoscale(defaults ...Flag) Flag {
	return thisOrThatOrDefaultFlag(c.Autoscale, "", defaults...)
}

// GetMinReplicas returns a value or a default if unset.
func (c Config) GetMinReplicas(defaults ...string) string {
	return thisOrThatOrDefault(c.MinReplicas, "", defaults...)
}

// GetMaxReplicas returns a value or a default if unset.
func (c Config) GetMaxReplicas(defaults ...string) string {
	return thisOrThatOrDefault(c.MaxReplicas, "", defaults...)
}

// GetCPUThreshold returns a value or a default if unset.
func (c Config) GetCPUThreshold(defaults ...string) string {
	return thisOrThatOrDefault(c.CPUThreshold, "", defaults...)
}

// GetMemoryThreshold returns a value or a default if unset.
func (c Config) GetMemoryThreshold(defaults ...string) string {
	return thisOrThatOrDefault(c.MemoryThreshold, "", defaults...)
}

// GetAutoscaleMetrics returns a value or a default if unset.
func (c Config) GetAutoscaleMetrics(defaults ...[]AutoscaleMetric) []AutoscaleMetric {
	if len(c.AutoscaleMetrics) > 0 {
		return c.AutoscaleMetrics
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
}

// GetDeploymentStrategy returns a value or a default if unset.
func (c Config) GetDeploymentStrategy(defaults ...string) string {
	return thisOrThatOrDefault(c.DeploymentStrategy, string(appsv1beta1.RollingUpdateDeploymentStrategyType), defaults...)
}

// Pod level options

// GetActiveDeadlineSeconds returns a value or a default if unset.
func (c Config) GetActiveDeadlineSeconds(defaults ...string) string {
	return thisOrThatOrDefault(c.ActiveDeadlineSeconds, "", defaults...)
}

// GetNodeSelector returns a value or default if unset.
func (c Config) GetNodeSelector(defaults ...map[string]string) map[string]string {
	final := map[string]string{}
	for key, value := range c.NodeSelector {
		final[key] = value
	}

	if len(defaults) > 0 {
		for key, value := range defaults[0] {
			if _, hasKey := final[key]; !hasKey {
				final[key] = value
			}
		}
	}

	return final
}

// GetRestartPolicy returns a value or a default if unset.
func (c Config) GetRestartPolicy(defaults ...string) string {
	return thisOrThatOrDefault(c.RestartPolicy, "", defaults...)
}

// GetVolumes gets (optional) volumes to attach to the pod.
func (c Config) GetVolumes(defaults ...Volume) []Volume {
	if len(c.Volumes) > 0 {
		return c.Volumes
	}
	if len(defaults) > 0 {
		return defaults
	}
	return nil
}

// GetTerminationGracePeriodSeconds returns the termination grace period
func (c Config) GetTerminationGracePeriodSeconds(defaults ...string) string {
	return thisOrThatOrDefault(c.TerminationGracePeriodSeconds, "", defaults...)
}

// Job level options

// GetJobActiveDeadlineSeconds returns a value or a default if unset.
func (c Config) GetJobActiveDeadlineSeconds(defaults ...string) string {
	return thisOrThatOrDefault(c.JobActiveDeadlineSeconds, "", defaults...)
}

// GetHostAliases returns a value or a default if unset.
func (c Config) GetHostAliases(defaults ...HostAlias) []HostAlias {
	if len(c.HostAliases) > 0 {
		return c.HostAliases
	}
	if len(defaults) > 0 {
		return defaults
	}
	return nil
}

// Deployment level options

// GetProgressDeadlineSeconds returns a value or a default if unset
func (c Config) GetProgressDeadlineSeconds(defaults ...string) string {
	return thisOrThatOrDefault(c.ProgressDeadlineSeconds, "", defaults...)
}

// Container level options

// GetArgs returns a value or default if unset.
func (c Config) GetArgs(defaults ...[]string) []string {
	return nonEmptySliceOrNil(thisOrThatOrDefaultSlice(c.Args, nil, defaults...))
}

// GetCommand returns a value or default if unset.
func (c Config) GetCommand(defaults ...[]string) []string {
	return nonEmptySliceOrNil(thisOrThatOrDefaultSlice(c.Command, nil, defaults...))
}

// GetELBAccessLogsBucketName returns the configured bucket name in AWS S3 used for elastic load balancer access logs
func (c Config) GetELBAccessLogsBucketName(defaults ...string) string {
	return thisOrThatOrDefault(c.ELBAccessLogsBucket, "", defaults...)
}

// GetEnv returns the env vars coalesced with the defaults.
func (c Config) GetEnv(defaults ...EnvVar) []EnvVar {
	ourVars := map[string]bool{}
	var coalesced []EnvVar
	for _, envVar := range c.Env {
		ourVars[envVar.Name] = true
		coalesced = append(coalesced, envVar)
	}

	for _, envVar := range defaults {
		if _, hasVar := ourVars[envVar.Name]; !hasVar {
			coalesced = append(coalesced, envVar)
		}
	}

	return coalesced
}

// GetLivenessProbe returns a value or a default if unset.
func (c Config) GetLivenessProbe(defaults ...*Probe) *Probe {
	if c.LivenessProbe != nil {
		return c.LivenessProbe
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
}

// GetPorts returns a value or a default if unset.
func (c Config) GetPorts(defaults ...ContainerPort) []ContainerPort {
	if len(c.Ports) > 0 {
		return c.Ports
	}
	if len(defaults) > 0 {
		return defaults
	}
	return nil
}

// GetReadinessProbe returns a value or a default if unset.
func (c Config) GetReadinessProbe(defaults ...*Probe) *Probe {
	if c.ReadinessProbe != nil {
		return c.ReadinessProbe
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
}

// GetResources returns a value or a default if unset.
func (c Config) GetResources(defaults ...*ResourceRequirements) *ResourceRequirements {
	if c.Resources != nil {
		return c.Resources
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
}

// GetSidecarResources returns a value or a default if unset.
func (c Config) GetSidecarResources(defaults ...map[string]*ResourceRequirements) map[string]*ResourceRequirements {
	if c.SidecarResources != nil && len(c.SidecarResources) > 0 {
		return c.SidecarResources
	}
	if len(defaults) > 0 && defaults[0] != nil {
		return defaults[0]
	}
	return map[string]*ResourceRequirements{}
}

// GetVolumeMounts gets (optional) volume mounts to attach to the pod.
func (c Config) GetVolumeMounts(defaults ...VolumeMount) []VolumeMount {
	if len(c.VolumeMounts) > 0 {
		return c.VolumeMounts
	}
	if len(defaults) > 0 {
		return defaults
	}
	return nil
}

// GetWorkingDir gets a value or a default if unset.
func (c Config) GetWorkingDir(defaults ...string) string {
	return thisOrThatOrDefault(c.WorkingDir, "", defaults...)
}

// GetEmails gets the notification email recipients.
func (c Config) GetEmails(defaults ...string) []string {
	return nonEmptySliceOrNil(thisOrThatOrDefaultSlice(c.Emails, nil, defaults))
}

// GetSlackChannel gets the notification slack channel.
func (c Config) GetSlackChannel(defaults ...string) string {
	return thisOrThatOrDefault(c.SlackChannel, "", defaults...)
}

// GetDeployNotifications gets the string that determines whether or not to notify when deployed.
func (c Config) GetDeployNotifications(defaults ...string) string {
	return thisOrThatOrDefault(c.DeployNotifications, DeployNotificationsComplete, defaults...)
}

// GetDatadogHost gets the datadog host to use for stats etc.
func (c Config) GetDatadogHost(defaults ...string) string {
	return thisOrThatOrDefault(c.DatadogHost, "", defaults...)
}

// GetRoute53TTL gets the DNS time to live
func (c Config) GetRoute53TTL(defaults int64) int64 {
	if c.Route53TTL == nil {
		return defaults
	}
	return *c.Route53TTL
}

// GetAWSAccessKey gets the aws access key.
func (c Config) GetAWSAccessKey(defaults ...string) string {
	return thisOrThatOrDefault(c.AWSAccessKey, "", defaults...)
}

// GetAWSSecretKey gets the aws secret key.
func (c Config) GetAWSSecretKey(defaults ...string) string {
	return thisOrThatOrDefault(c.AWSSecretKey, "", defaults...)
}

// GetStorage returns the persistent storage requested for the service
func (c Config) GetStorage(defaults ...[]StorageConfig) []StorageConfig {
	if len(c.Storage) > 0 {
		return c.Storage
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
}

// GetServiceLBConnectionIdleTimeout gets the service load balancer idle timeout period
func (c Config) GetServiceLBConnectionIdleTimeout(defaults ...string) string {
	return thisOrThatOrDefault(c.ServiceLBConnectionIdleTimeout, "", defaults...)
}

// ShouldProvisionStorage returns if the service needs storage
func (c Config) ShouldProvisionStorage() bool {
	return len(c.GetStorage()) > 0
}

// GetProtocol returns the service protocol
func (c Config) GetProtocol(defaults ...string) Protocol {
	return Protocol(thisOrThatOrDefault(string(c.Protocol), "", defaults...))
}

// GetProxyProtocol returns whether proxy protocol is enabled
func (c Config) GetProxyProtocol(defaults ...Flag) Flag {
	return thisOrThatOrDefaultFlag(c.ProxyProtocol, "", defaults...)
}

// IsHTTPService returns if the service is listening on http
func (c Config) IsHTTPService() bool {
	return c.GetProtocol(string(ProtocolHTTP)) == ProtocolHTTP
}

// ShouldEnableProxyProtocol returns whether proxy protocol should be enabled
func (c Config) ShouldEnableProxyProtocol() bool {
	return c.IsHTTPService() || c.GetProxyProtocol() == FlagEnabled
}

// GetHTTPRedirect returns whether http redirect is enabled
func (c Config) GetHTTPRedirect(defaults ...Flag) Flag {
	return thisOrThatOrDefaultFlag(c.HTTPRedirect, "", defaults...)
}

// ShouldEnableHTTPRedirect returns whether to enable the http redirect
func (c Config) ShouldEnableHTTPRedirect() bool {
	return c.IsHTTPService() || c.GetHTTPRedirect() == FlagEnabled
}

// ShouldEnableTLSProxy returns whether the tls proxy server should be enabled
func (c Config) ShouldEnableTLSProxy() bool {
	return c.IsHTTPService()
}

// ShouldIncludeProxySidecar returns whether the proxy sidecar should be included
func (c Config) ShouldIncludeProxySidecar() bool {
	return c.IsServiceTypeLoadBalancer() && (c.ShouldEnableHTTPRedirect() || c.ShouldEnableTLSProxy())
}

// GetWardenProxy returns whether the warden proxy is enabled
func (c Config) GetWardenProxy(defaults ...Flag) Flag {
	return thisOrThatOrDefaultFlag(c.BrowserAuthentication, "", defaults...)
}

// ShouldIncludeWardenProxySidecar returns whether the warden proxy sidecar should be included
func (c Config) ShouldIncludeWardenProxySidecar() bool {
	return c.GetWardenProxy() == FlagEnabled
}

// GetRegistryHost returns the registry host.
func (c Config) GetRegistryHost(defaults ...string) string {
	if len(c.RegistryHost) > 0 {
		return c.RegistryHost
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	if c.ServiceEnv == ServiceEnvMinikube {
		return env.Env().String(EnvVarRegistryHost)
	}
	return fmt.Sprintf("registry.%s", c.GetClusterName())
}

// GetSmokeSignalCheckPath returns the smoke signal check path
func (c Config) GetSmokeSignalCheckPath(defaults ...string) string {
	return thisOrThatOrDefault(c.SmokeSignalCheckPath, "", defaults...)
}

// HasUniqueEnvVars returns whether a given config has unique environment
// variables. Note that this checks for uniqueness by the name of each
// environment variable
func (c Config) HasUniqueEnvVars() bool {
	varSet := collections.SetOfString{}

	for _, envVar := range c.Env {
		if varSet.Contains(envVar.Name) {
			return false
		}
		varSet.Add(envVar.Name)
	}
	return true
}

// GetVersion returns the version for the config.
func (c Config) GetVersion() string {
	if len(c.GitRefHash) > 0 {
		return fmt.Sprintf("git_%s", c.GitRefHash)
	}
	return ""
}

// String returns a string representation of the var set.
func (c Config) String() string {
	contents, _ := kube.YAMLEncode(c)
	return contents
}

// GetDisallowDeploy returns whether a service can only be deployed
// through a project deployment
func (c Config) GetDisallowDeploy() bool {
	return c.DisallowDeploy
}

// GetIdentifier returns service name for services and scheduled tasks
// and service name + deploy id for tasks
func (c *Config) GetIdentifier() string {
	if c.GetServiceType() == ServiceTypeTaskTemplate {
		switch c.GetBuildMode() {
		case BuildModeScheduleTask,
			BuildModeScheduleBackup,
			BuildModeCreateBackup,
			BuildModeDeprecateBackup:
			return c.GetServiceName()
		default:
			return fmt.Sprintf("%s-%s", c.GetServiceName(), c.GetDeployID())
		}
	}

	return c.GetServiceName()
}

// IsServiceTypeLoadBalancer returns if this service has a load balancer
func (c Config) IsServiceTypeLoadBalancer() bool {
	return c.ServiceType == ServiceTypeDeployment && (c.Accessibility == AccessibilityPublic || c.Accessibility == AccessibilityProtected || c.Accessibility == AccessibilityVPN)
}

// GetSchedule returns the schedule in Cron format.
func (c Config) GetSchedule(defaults ...string) string {
	return thisOrThatOrDefault(c.Schedule, "", defaults...)
}

// IsScheduledTask returns if this task runs on a schedule
func (c Config) IsScheduledTask() bool {
	return len(c.Schedule) > 0
}

// IsTask returns if the service is a task
func (c Config) IsTask() bool {
	return c.GetServiceType() == ServiceTypeTaskTemplate
}

// IsExternallyAccessible returns if the service accessibility is VPN, private, protected, or public
func (c Config) IsExternallyAccessible() bool {
	switch c.GetAccessibility() {
	case AccessibilityVPN,
		AccessibilityPrivate,
		AccessibilityProtected,
		AccessibilityPublic:
		return c.ServiceType == ServiceTypeDeployment
	default:
		return false
	}
}

// IsDeploymentStrategyRecreate returns if the deployment strategy type is Recreate
func (c Config) IsDeploymentStrategyRecreate() bool {
	return stringutil.EqualsCaseless(c.GetDeploymentStrategy(), string(appsv1beta1.RecreateDeploymentStrategyType))
}

// IAMRoleName returns the name of the iam role
func (c Config) IAMRoleName() string {
	return IAMRoleName(c.GetServiceName(), c.GetCluster())
}

// InheritFrom from returns a coalesced service config from another config.
func (c Config) InheritFrom(other *Config) *Config {
	return &Config{
		SchemaVersion:            c.GetSchemaVersion(other.SchemaVersion),
		KubectlConfig:            c.GetKubectlConfig(other.KubectlConfig),
		BuildMode:                c.GetBuildMode(other.BuildMode),
		SourcePath:               c.GetSourcePath(other.SourcePath),
		ServiceType:              c.GetServiceType(other.ServiceType),
		Labels:                   c.GetLabels(other.Labels),
		DeployID:                 c.GetDeployID(other.DeployID),
		CreatedBy:                c.GetCreatedBy(other.CreatedBy),
		DeployedBy:               c.GetDeployedBy(other.DeployedBy),
		DeployedAt:               c.GetDeployedAt(other.DeployedAt),
		DeployNotifications:      c.GetDeployNotifications(other.DeployNotifications),
		ClusterName:              c.GetClusterName(other.ClusterName),
		Cluster:                  c.GetCluster(other.Cluster),
		ServiceConfigPath:        c.GetServiceConfigPath(other.ServiceConfigPath),
		ServiceEnv:               c.GetServiceEnv(other.ServiceEnv),
		ServiceName:              c.GetServiceName(other.ServiceName),
		GitRemote:                c.GetGitRemote(other.GitRemote),
		GitRef:                   c.GetGitRef(other.GitRef),
		GitBase:                  c.GetGitBase(other.GitBase),
		GitRefHash:               c.GetGitRefHash(other.GitRefHash),
		ECRAuthentication:        c.GetECRAuthentication(other.ECRAuthentication),
		Dockerfile:               c.GetDockerfile(other.Dockerfile),
		DockerRegistry:           c.GetDockerRegistry(other.DockerRegistry),
		DockerTag:                c.GetDockerTag(other.DockerTag),
		Namespace:                c.GetNamespace(other.Namespace),
		FQDN:                     c.GetFQDN(other.FQDN),
		SANs:                     c.GetSANs(other.SANs),
		FileMountPath:            c.GetFileMountPath(other.FileMountPath),
		Accessibility:            c.GetAccessibility(other.Accessibility),
		LoadBalancerSourceRanges: c.GetLoadBalancerSourceRanges(other.LoadBalancerSourceRanges),
		ContainerImage:           c.GetContainerImage(other.ContainerImage),

		ProjectName:         c.GetProjectName(other.ProjectName),
		ProjectSlackChannel: c.GetProjectSlackChannel(other.ProjectSlackChannel),
		ProjectRun:          c.GetProjectRun(other.ProjectRun),

		Schedule: c.GetSchedule(other.Schedule),

		DatabaseName:     c.GetDatabaseName(other.DatabaseName),
		DBStorage:        c.GetDBStorage(other.DBStorage),
		DBRelaunch:       c.GetDBRelaunch(other.DBRelaunch),
		DBRestore:        c.GetDBRestore(other.DBRestore),
		DBBackupFileName: c.GetDBBackupFileName(other.DBBackupFileName),

		ContainerPort:  c.GetContainerPort(other.ContainerPort),
		ContainerProto: c.GetContainerProto(other.ContainerProto),

		AWSRegion:               c.GetAWSRegion(other.AWSRegion),
		AWSPolicyARN:            c.GetAWSPolicyARN(other.AWSPolicyARN),
		AWSCredentialsMountPath: c.GetAWSCredentialsMountPath(other.AWSCredentialsMountPath),

		Autoscale:          c.GetAutoscale(other.Autoscale),
		Replicas:           c.GetReplicas(other.Replicas),
		MinReplicas:        c.GetMinReplicas(other.MinReplicas),
		MaxReplicas:        c.GetMaxReplicas(other.MaxReplicas),
		CPUThreshold:       c.GetCPUThreshold(other.CPUThreshold),
		MemoryThreshold:    c.GetMemoryThreshold(other.MemoryThreshold),
		AutoscaleMetrics:   c.GetAutoscaleMetrics(other.AutoscaleMetrics),
		DeploymentStrategy: c.GetDeploymentStrategy(other.DeploymentStrategy),

		ActiveDeadlineSeconds: c.GetActiveDeadlineSeconds(other.ActiveDeadlineSeconds),
		NodeSelector:          c.GetNodeSelector(other.NodeSelector),
		RestartPolicy:         c.GetRestartPolicy(other.RestartPolicy),
		Volumes:               c.GetVolumes(other.Volumes...),
		TerminationGracePeriodSeconds: c.GetTerminationGracePeriodSeconds(other.TerminationGracePeriodSeconds),
		HostAliases:                   c.GetHostAliases(other.HostAliases...),

		JobActiveDeadlineSeconds: c.GetJobActiveDeadlineSeconds(other.JobActiveDeadlineSeconds),

		ProgressDeadlineSeconds: c.GetProgressDeadlineSeconds(other.ProgressDeadlineSeconds),

		Args:             c.GetArgs(other.Args),
		Command:          c.GetCommand(other.Command),
		Env:              c.GetEnv(other.Env...),
		LivenessProbe:    c.GetLivenessProbe(other.LivenessProbe),
		Ports:            c.GetPorts(other.Ports...),
		ReadinessProbe:   c.GetReadinessProbe(other.ReadinessProbe),
		Resources:        c.GetResources(other.Resources),
		SidecarResources: c.GetSidecarResources(other.SidecarResources),
		VolumeMounts:     c.GetVolumeMounts(other.VolumeMounts...),
		WorkingDir:       c.GetWorkingDir(other.WorkingDir),

		Emails:       c.GetEmails(other.Emails...),
		SlackChannel: c.GetSlackChannel(other.SlackChannel),
		DatadogHost:  c.GetDatadogHost(other.DatadogHost),

		AWSAccessKey: c.GetAWSAccessKey(other.AWSAccessKey),
		AWSSecretKey: c.GetAWSSecretKey(other.AWSSecretKey),
		RegistryHost: c.GetRegistryHost(other.RegistryHost),

		Storage:       c.GetStorage(other.Storage),
		Protocol:      c.GetProtocol(string(other.Protocol)),
		ProxyProtocol: c.GetProxyProtocol(other.ProxyProtocol),
		HTTPRedirect:  c.GetHTTPRedirect(other.HTTPRedirect),

		SmokeSignalCheckPath:           c.GetSmokeSignalCheckPath(other.SmokeSignalCheckPath),
		ELBAccessLogsBucket:            c.GetELBAccessLogsBucketName(other.ELBAccessLogsBucket),
		ServiceLBConnectionIdleTimeout: c.GetServiceLBConnectionIdleTimeout(other.ServiceLBConnectionIdleTimeout),

		BrowserAuthentication: c.GetWardenProxy(other.BrowserAuthentication),
	}
}

// IsMinikubeEnv returns if the service env is minikube
func (c Config) IsMinikubeEnv() bool {
	return c.GetServiceEnv() == ServiceEnvMinikube
}

// HasValidLabels checks if user defined labels are valid
func (c Config) HasValidLabels() (ok bool, errs error) {
	for name, value := range c.GetLabels() {
		if IsReservedLabel(name) {
			return false, exception.New(
				fmt.Sprintf("Label %s is not valid, name must not have prefix `blend-` or clash with other generated label names", name))
		}
		if errs := validation.IsQualifiedName(name); len(errs) > 0 {
			return false, exception.New(errs)
		}
		if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
			return false, exception.New(errs)
		}
	}
	return true, nil
}

func thisOrThatOrDefaultBool(this, that bool, defaults ...bool) bool {
	if this != false {
		return this
	}
	if len(defaults) > 0 && defaults[0] != false {
		return defaults[0]
	}
	return that
}

func thisOrThatOrDefaultNonzeroInt(this, that int, defaults ...int) int {
	if this != 0 {
		return this
	}
	if len(defaults) > 0 && defaults[0] != 0 {
		return defaults[0]
	}
	return that
}

func thisOrThatOrDefault(this, that string, defaults ...string) string {
	if len(this) > 0 {
		return this
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return that
}

func thisOrThatOrDefaultFlag(this, that Flag, defaults ...Flag) Flag {
	if len(this) > 0 {
		return this
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return that
}

func thisOrThatOrDefaultSlice(this, that []string, defaults ...[]string) []string {
	if len(this) > 0 {
		return this
	}
	if len(defaults) > 0 && len(defaults[0]) > 0 {
		return defaults[0]
	}
	return that
}

func nonEmptySliceOrNil(slice []string) []string {
	if len(strings.Join(slice, "")) == 0 { // make sure the slice does not contain only empty strings
		return nil
	}
	return slice
}
