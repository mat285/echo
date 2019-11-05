package kube

import (
	"time"
)

const (
	// EnvVarKubectlConfig is the env var we use to set the kubectl config.
	EnvVarKubectlConfig = "KUBECTL_CONFIG"
	// EnvVarNamespace is an environment variable that overrides the default namespace.
	EnvVarNamespace = "KUBE_NAMESPACE"
	// EnvVarClientQPS is an env var for setting the queries per second of the kube client
	EnvVarClientQPS = "KUBE_CLIENT_QPS"
	// EnvVarClientBurst is an env var for setting the burst (throttling) of the kube client
	EnvVarClientBurst = "KUBE_CLIENT_BURST"
	// EnvVarAPIServiceHost is an env var containing an address to the kube api server
	EnvVarAPIServiceHost = "KUBERNETES_SERVICE_HOST"
)

const (
	// DefaultNamespace is the default kube namespace.
	DefaultNamespace = "blend"
)

// Labels
const (
	// ReservedLabelPrefix is a prefix which is not legal in user-supplied label names
	ReservedLabelPrefix = "blend-"
	// LabelRole is a label.
	LabelRole = "blend-role"
	// LabelDatabase is the database name label
	LabelDatabase = "blend-database"
	// LabelDatabaseBackup is the database backup label
	LabelDatabaseBackup = "blend-database-backup"
	// LabelProject is the project name label
	LabelProject = "blend-project"
	// LabelService is the service name label
	LabelService = "blend-service"
	// LabelTeam is the team name label
	LabelTeam = "blend-team"
	// LabelServiceEnv is the service environment of a service.
	LabelServiceEnv = "blend-service-env"
	// LabelEnv is the environment of a service.
	LabelEnv = "blend-env"
	// LabelVersion is the version label for pods, is typically `git_${CURRENT_REF}`.
	LabelVersion = "blend-version"
	// LabelDeploy is a deploy id for individual deploys.
	LabelDeploy = "blend-deploy"
	// LabelCreatedBy is the label that shows who created the resource
	LabelCreatedBy = "blend-created-by"
	// LabelCreatedAt is the label that shoes when the resource was created
	LabelCreatedAt = "blend-created-at"
	// LabelDeployedBy is the label that shows who deployed the resource
	LabelDeployedBy = "blend-deployed-by"
	// LabelDeployedAt is the label that shoes when the resource was deployed
	LabelDeployedAt = "blend-deployed-at"
	// LabelAutoDeploy is a constant.
	LabelAutoDeploy = "blend-auto-deploy"
	// LabelBuildMode is a constant.
	LabelBuildMode = "blend-build-mode"
	// LabelAccessibility is the accessibility of the service
	LabelAccessibility = "blend-accessibility"
	// LabelApp is a constant.
	LabelApp = "k8s-app"
	// LabelAddon is a constant.
	LabelAddon = "k8s-addon"
	// LabelPlainApp is just app
	LabelPlainApp = "app"
	// LabelPlainRole is just role
	LabelPlainRole = "role"
	// LabelGitRef is a label that denotes what ref was built for a deploy.
	LabelGitRef = "blend-git-ref"
	// LabelGitBase is the label that denotes what git base was used for a deploy
	LabelGitBase = "blend-git-base"
	// LabelJobName is a descriptor for job pods, generally the value will be "deployinator-builder".
	LabelJobName = "blend-job-name"
	// LabelPinPod specifies that the deploy pod should be pinned and not deleted
	LabelPinPod = "blend-pin-pod"

	// LabelUpdatedAt is the label denoting when a resource has been updated
	LabelUpdatedAt = "blend-updated-at"

	// LabelServiceSecret labels secrets for blend services
	LabelServiceSecret = "blend-service-secret"

	// LabelNodeRoleMater is a the master node-role label for tolerations/taints
	LabelNodeRoleMaster = "node-role.kubernetes.io/master"
	// LabelBlendNodeRoleBuilder is the builder node-role label for tolerations/taints
	LabelBlendNodeRoleBuilder = "blend-builder-node"
	// LabelBlendNodeRoleBuilder is the core node-role label for tolerations/taints
	LabelBlendNodeRoleCore = "blend-core-node"

	// LabelServiceAutoScalingRole is a label used to demarcate autoscalers
	// for services, otherwise referred to in the Kubernetes literature as
	// a horizontal pod autoscaler (HPA)
	LabelServiceAutoScalingRole = "service-auto-scaler"

	// From the defunct well_known_labels.go
	LabelNodeLabelRole            = "kubernetes.io/role"
	LabelNodeLabelRoleMaster      = "master"
	LabelNodeLabelRoleNode        = "node"
	LabelNodeLabelRoleBuilderNode = "builder-node"
	LabelNodeLabelRoleCoreNode    = "core-node"

	LabelNodeAutoscalingRole            = "role"
	LabelNodeAutoscalingRoleBuilderNode = LabelNodeLabelRoleBuilderNode
	LabelNodeAutoscalingRoleCoreNode    = LabelNodeLabelRoleCoreNode
	LabelNodeAutoscalingRoleNode        = LabelNodeLabelRoleNode
	LabelNodeAutoscalingRoleMaster      = LabelNodeLabelRoleMaster
)

// Annotations
const (
	// AnnotationDeployedBy is a constant.
	AnnotationDeployedBy = "blend.com/build-deployed-by"
	// AnnotationsDeploy is an annotation that gives a unique id to each resource created by a deploy.
	AnnotationsDeploy = "blend.com/build-deploy"
	// AnnotationGitRefHash is a constant.
	AnnotationGitRefHash = "blend.com/build-git-ref-hash"
	// AnnotationScaleDownFrom is an annotation that indicates the number of replicas before the deployment is scaled down by the scaledown job
	AnnotationScaleDownFrom = "blend.com/scale-down-from"
	// AnnotationRevision is an annotation that indicates the revision number of a replica set
	AnnotationRevision = "deployment.kubernetes.io/revision"
	// AnnotationClusterAutoscalerSafeToEvict is an annotation that prevents cluster autoscaler from evicting the pod if set to false
	AnnotationClusterAutoscalerSafeToEvict = "cluster-autoscaler.kubernetes.io/safe-to-evict"

	// IAM Role stuff

	// AnnotationNamespacePermittedIAMRoles are the permitted roles for namespaces
	AnnotationNamespacePermittedIAMRoles = "iam.amazonaws.com/permitted"
	// AnnotationIAMPodRole is the role the pod can assume
	AnnotationIAMPodRole = "iam.amazonaws.com/role"
)

const (
	// RoleService is a role representing a Kubernetes service, which is an
	// abstraction which allows Kubernetes to expose a group of pods as a
	// network service.
	RoleService = "service"
	// RoleServiceInstance is a role representing a particular instance of
	// a Kubernetes service.
	RoleServiceInstance = "service-instance"
	// RoleServiceMapping is a role.
	RoleServiceMapping = "service-mapping"
	// RoleTask is for task template tasks.
	RoleTask = "task"
	// RoleJob is a role
	RoleJob = "job"
	// RoleDeployment is a role.
	RoleDeployment = "deployment"
	// RoleIngress is a role.
	RoleIngress = "ingress"
	// RoleDaemonSet is a role.
	RoleDaemonSet = "daemon-set"
	// RoleCronJob is a role
	RoleCronJob = "cron-job"
	// RoleDatabase is a role
	RoleDatabase = "database"
)

// CertFile is a special name enum for cert files.
type CertFile string

const (
	// CertFileCert is the filename for the tls cert.
	CertFileCert CertFile = "tls.crt"
	// CertFileKey is the filename for the tls key.
	CertFileKey CertFile = "tls.key"
)

const (
	// VolumePlaintextSecretFiles is a volume name.
	VolumePlaintextSecretFiles = "plaintext-files"
	// VolumeSecretsAgent is a volume name.
	VolumeSecretsAgent = "secrets-agent"
	// VolumeAWSCredentials is a volume name.
	VolumeAWSCredentials = "aws-credentials"
	// VolumeDockerSock is a volume name.
	VolumeDockerSock = "docker-sock"
	// VolumeSSHKeys is a volume name.
	VolumeSSHKeys = "ssh-keys"
	// VolumeTLSCerts is a volume name.
	VolumeTLSCerts = "tls-certs"
	// VolumeEtcdTLS is a volume name.
	VolumeEtcdTLS = "etcd-tls-volume"
	// VolumeKubeCACert is a voluime name.
	VolumeKubeCACert = "kube-ca-cert"

	// PathAWSCredentials is a mount path.
	PathAWSCredentials = "/var/aws-credentials"
	// PathSecretsAgent is a mount path.
	PathSecretsAgent = "/var/secrets-agent"
	// PathTLSCerts is a mount path
	PathTLSCerts = "/var/tls-certs"
	// PathKubeCACert is the path to the kube ca for the pod's service account
	PathKubeCACert = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

	// PathHostTLSBasePath is where tls secrets are mounted on hosts
	PathHostTLSBasePath = "/srv/kubernetes"

	// SecretExternalWildcard is the external wildcard secret.
	SecretExternalWildcard = "external-wildcard"
	// SecretInternalWildcard is an internal secret used by the proxy.
	SecretInternalWildcard = "internal-wildcard"
	// SecretEtcdClientTLS is the etcd client tls secret
	SecretEtcdClientTLS = "etcd-client-tls"

	// FileEtcdClientCert is the etcd client cert file
	FileEtcdClientCert = "etcd-client.crt"
	// FileEtcdClientKey is the etcd client key file
	FileEtcdClientKey = "etcd-client.key"
	// FileEtcdClientCACert is the etcd client CA cert file
	FileEtcdClientCACert = "etcd-client-ca.crt"
)

const (
	// PathWardenCerts is where Warden TLS certs are mounted
	PathWardenCerts = "/var/warden-certs"
	// VolumeWardenCerts is the name of the Warden TLS certs volume
	VolumeWardenCerts = "warden-certs"
	// WardenCACert is the file name of the Warden CA cert
	WardenCACert = "warden.ca.crt"

	// SecretWardenDeployinatorTLS is the secret for deployinator's Warden certs
	SecretWardenDeployinatorTLS = "deployinator-warden-tls"

	// SecretWardenProxyTLS is the secret for warden-proxy certs
	SecretWardenProxyTLS = "warden-proxy-tls"
)

const (
	// RegistryHtpaswdSecretName is the secret that holds docker registry authentication
	RegistryHtpaswdSecretName = "registry-htpasswd"
	// RegistryAuthSecretName is the secret where the auth plaintext lives
	RegistryAuthSecretName = "registry-auth-secret"
)

const (
	// KindSecret is a secret shhh don't tell anyone
	KindSecret = "Secret"
	// KindClusterRole is a cluster role
	KindClusterRole = "ClusterRole"
	// KindRole is a role
	KindRole = "Role"
)

const (
	// APIGroupRBAC is a kube api group
	APIGroupRBAC = "rbac.authorization.k8s.io"
)

var (
	// DefaultPollInterval is the default polling interval
	DefaultPollInterval = 1 * time.Second
	// DefaultPollTimeout is the default polling timeout
	DefaultPollTimeout = 5 * time.Minute

	testPollInterval   = 10 * time.Millisecond
	testPollTimeout    = 100 * time.Second
	deletePollInterval = defaultDeletePollInterval
	deletePollTimeout  = defaultDeletePollTimeout
)

const (
	defaultDeletePollInterval = 5 * time.Second
	defaultDeletePollTimeout  = 10 * time.Minute
)

const (
	// VolumeEventCollector is the volume name of the event collector
	VolumeEventCollector = "collector-socket-path"
	// MountPathEventCollector is the default mount path for the event collector.
	// It also happens to be the host path as well.
	MountPathEventCollector = "/var/run/event-collector"
)

const (
	// SecretEventCollectorConfig is the name of the config secret for the event collector
	SecretEventCollectorConfig = "event-collector-config"
	// FileEventCollectorUnixConfig is the file we store the unix config in
	FileEventCollectorUnixConfig = "config_unix.yml"
	// FileEventCollectorTCPConfig is the file we store the TCP config in
	FileEventCollectorTCPConfig = "config_tcp.yml"
	// VolumeEventCollectorConfig is the volume name of the event collector
	VolumeEventCollectorConfig = "collector-config"
	// MountPathEventCollectorConfig is the default mount path for the event collector.
	MountPathEventCollectorConfig = "/var/run/secrets/event-collector-config"
)

const (
	// PercentageBuilderNodes is the desired percentage of nodes that are builder nodes
	PercentageBuilderNodes = 40
	// MinNumNonBuilderNodes is the minimum number of non builder nodes
	MinNumNonBuilderNodes = 1
)

const (
	maxSecretsPerService = 5
)

// Predefined kube priority classes
const (
	// These requires the pod to be in kube-system
	PriotiyClassNameSystemNodeCritical    string = "system-node-critical"
	PriotiyClassNameSystemClusterCritical string = "system-cluster-critical"
	// These are just priority classes with very high priority values
	PriotiyClassNameBlendSystemNodeCritical    string = "blend-system-node-critical"
	PriotiyClassNameBlendSystemClusterCritical string = "blend-system-cluster-critical"
)

const (
	// PodStatusReasonEvicted is the status reason for evicted pods
	PodStatusReasonEvicted string = "Evicted"
)
