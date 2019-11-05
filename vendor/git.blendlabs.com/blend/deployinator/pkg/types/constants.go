package types

import (
	"fmt"
	"time"

	coreapi "k8s.io/kubernetes/pkg/apis/core"
)

const (
	// CurrentSchemaVersion is the current schema version.
	CurrentSchemaVersion = "1.0.0"

	// BlendSystemNamespace is the namespace builder tasks and deployinator itself runs in.
	BlendSystemNamespace = "blend-system"
	//BlendNamespace is the namespace for all blend applications
	BlendNamespace = "blend"
	//EventCheckpointerNamespace is the namespace for the event-checkpointer addon
	EventCheckpointerNamespace = "event-consumer"

	// SentryNamespace is the namespace for sentry addon
	SentryNamespace = "sentry"

	// NameDeployinator is the deployinator name
	NameDeployinator = "deployinator"
	// NameDeployinatorWorker is the deployinator worker name
	NameDeployinatorWorker = "deployinator-worker"
	// NameProxy is the proxy name
	NameProxy = "proxy"
	// NameCredentials is the proxy name
	NameCredentials = "credentials"
	// NameClusterUpdater is the name of the updater
	NameClusterUpdater = "cluster-updater"

	// NameSecrulesConfigMap is the name of the config map in the
	// blend-system namespace that corresponds to the secrules job state
	NameSecrulesConfigMap = "secrules"

	// CredentialsSuffix is the suffix for all things credentials
	CredentialsSuffix = "-" + NameCredentials

	// ServiceAccountDeployinator is the deployinator service account.
	ServiceAccountDeployinator = NameDeployinator

	// SecretExternalWildcard is the secret that holds the external wildcard cert.
	SecretExternalWildcard = "external-wildcard"
	// SecretDeployinator is the deployinator secret.
	SecretDeployinator = NameDeployinator
	// SecretDeployinatorBuilder is the deployinator builder secret.
	SecretDeployinatorBuilder = "deployinator-builder"
	// SecretDeployinatorVault is the deployinator vault secret.
	SecretDeployinatorVault = "deployinator-vault"
	// SecretEnablePostgresSSL is the postgres enable ssl script secret
	SecretEnablePostgresSSL = "deployinator-enable-postgres-ssl"
	// SecretSnapshot is the snapshots for secrets
	SecretSnapshot = "snapshot"

	// AcmeDNS01ProviderName is the way for certmanager verification
	AcmeDNS01ProviderName = "aws-dns"
	// ClusterIssuerName is the cert issuer
	ClusterIssuerName = "letsencrypt"

	// ServiceEnvMinikube is an environment for minikube cluster
	ServiceEnvMinikube = "minikube"

	// DefaultBuilderWorkDir is the default working directory for builder
	DefaultBuilderWorkDir = "/app"

	// BuilderContainer is the container for builder
	BuilderContainer = "deployinator-builder"
	// DeployerContainer is the container for deployer
	DeployerContainer = "deployinator-deployer"

	// WardenProxy is the warden proxy registered with warden
	WardenProxy = "wproxy"

	// WebhookPath is the path to webhooks
	WebhookPath = "/webhook.projects"

	// ServiceNameMaxLength is the maximum length for service names
	ServiceNameMaxLength = 45 // gives us room for things like `-credentials` or `-hash`

	// VaultK8sSecretsPolicy is the policy for access k8s secrets
	VaultK8sSecretsPolicy = "k8s_secrets"
)

// Accessibility is a service accessibility.
type Accessibility string

const (
	// AccessibilityPublic is open to the world.
	AccessibilityPublic Accessibility = "public"
	// AccessibilityProtected gets an internal elb to the vpc
	AccessibilityProtected Accessibility = "protected"
	// AccessibilityPrivate is accessible on the current vpc.
	AccessibilityPrivate Accessibility = "private"
	// AccessibilityVPN is accessible on the vpn. See VPNSourceRanges
	AccessibilityVPN Accessibility = "vpn"
	// AccessibilityInternal is only accessible to other kube servies.
	AccessibilityInternal Accessibility = "internal"
	// AccessibilityNone is not accessible.
	AccessibilityNone Accessibility = "none"
)

// IsExternal returns if the accessibility is external to the cluster
func (a Accessibility) IsExternal() bool {
	return a == AccessibilityPrivate || a == AccessibilityProtected || a == AccessibilityPublic
}

// BuildMode is a meta mode for the builder.
type BuildMode string

func (bm BuildMode) String() string {
	return string(bm)
}

const (
	// BuildModeUnset indicates the build mode is unset.
	BuildModeUnset BuildMode = ""
	// BuildModeCreate creates the metadata for a service.
	BuildModeCreate BuildMode = "create"
	// BuildModeDockerPush checks out source and builds and pushes docker images, but does not configure kube resources.
	BuildModeDockerPush BuildMode = "docker-push"
	// BuildModeDeploy checks out sources, builds & pushes docker images, and provisions new resources for the app or applies a rolling update to the existing resources.
	BuildModeDeploy BuildMode = "deploy"
	// BuildModeDeploySkipBuild pulls and provisions/runs a docker image
	BuildModeDeploySkipBuild BuildMode = "deploy-skip-build"
	// BuildModeCanaryDeploy is the mode for deploying a service partially with a certain roll-out percentage
	BuildModeCanaryDeploy BuildMode = "deploy-canary"
	// BuildModeRefresh checks out sources, and applies a rolling update to the existing docker resources.
	BuildModeRefresh BuildMode = "refresh"
	// BuildModeRunTask checks out sources, builds the image, and schedules a pod to run it immediately.
	BuildModeRunTask BuildMode = "run-task"
	// BuildModeRunTaskSkipBuild checks out an image and schedules a pod to run it immediately
	BuildModeRunTaskSkipBuild BuildMode = "run-task-skip-build"
	// BuildModePrintTemplates prints the templates but does not configure or push any resources.
	BuildModePrintTemplates BuildMode = "print-templates"
	// BuildModeDown deprecates and stops kube resources.
	BuildModeDown BuildMode = "down"
	// BuildModeDelete fully deletes service.
	BuildModeDelete BuildMode = "delete"
	// BuildModeCreateSecrets creates a secrets bucket for a service.
	BuildModeCreateSecrets BuildMode = "create-secrets"
	// BuildModeDeleteSecrets creates a secrets bucket for a service.
	BuildModeDeleteSecrets BuildMode = "delete-secrets"
	// BuildModeRunProject is the mode for running a project
	BuildModeRunProject BuildMode = "run-project"
	// BuildModeLaunchDatabase is the mode for running a project
	BuildModeLaunchDatabase BuildMode = "launch-database"
	// BuildModeDeprecateDatabase is the mode for running a project
	BuildModeDeprecateDatabase BuildMode = "deprecate-database"
	// BuildModeScheduleTask is the mode for scheduling a task
	BuildModeScheduleTask BuildMode = "schedule-task"
	// BuildModeCreateBackup is the mode for creating an automatic database backup
	BuildModeCreateBackup BuildMode = "create-backup"
	// BuildModeScheduleBackup is the mode for scheduling an automatic database backup task
	BuildModeScheduleBackup BuildMode = "schedule-backup"
	// BuildModeRunBackup is the mode used to represent when a backup job is run
	BuildModeRunBackup BuildMode = "run-backup"
	// BuildModeDeprecateBackup is the mode for deprecating an automatic database backup task
	BuildModeDeprecateBackup BuildMode = "deprecate-backup"
	// BuildModeDeleteBackup is the mode for deprecating an automatic database backup task
	BuildModeDeleteBackup BuildMode = "delete-backup"
	// BuildModeCreateRestore is the mode for creating an automatic database backup
	BuildModeCreateRestore BuildMode = "create-restore"
	// BuildModeRunRestore is the mode for running an automatic database restore task
	BuildModeRunRestore BuildMode = "run-restore"
	// BuildModeDeleteRestore is the mode for running an automatic database restore task
	BuildModeDeleteRestore BuildMode = "delete-restore"
)

const (
	// ContainerPortNone denotes a container shouldn't mount an external port.
	ContainerPortNone = "none"
)

// ServiceType determines how we provision the service.
type ServiceType string

const (
	// ServiceTypeUnset is the default service type.
	ServiceTypeUnset ServiceType = ""
	// ServiceTypeDeployment creates a new deployment + service + ingress on provisioning.
	ServiceTypeDeployment ServiceType = "deployment"
	// ServiceTypeTaskTemplate  creates a new task pod for each "deploy".
	ServiceTypeTaskTemplate ServiceType = "task-template"
	// ServiceTypeDefault is the default provision mode.
	ServiceTypeDefault ServiceType = ServiceTypeDeployment
)

// Flag is a state indicator of enabled or disabled.
type Flag string

const (
	// FlagUnset is the zero state.
	FlagUnset Flag = ""
	// FlagEnabled indicates something is enabled.
	FlagEnabled Flag = "enabled"
	// FlagDisabled indicates something is disabled.
	FlagDisabled Flag = "disabled"
)

const (
	// EnvVarPodName is the name of the pod
	EnvVarPodName = "POD_NAME"
	// EnvVarPodNamespace is the namespace of the pod
	EnvVarPodNamespace = "POD_NAMESPACE"
	// EnvVarNamespace is the namespace
	EnvVarNamespace = "NAMESPACE"
	// EnvVarKubernetes is the env var to inform that it is being run on k8s
	EnvVarKubernetes = "KUBERNETES"
	// EnvVarRedirectPort is the port for redirect http -> https
	EnvVarRedirectPort = "REDIRECT_PORT"
	// EnvVarCurrentRef is the current git ref
	EnvVarCurrentRef = "CURRENT_REF"
	// EnvVarProjectName is an environment variable name for the project name.
	EnvVarProjectName = "PROJECT_NAME"
	// EnvVarNodeName is the name of the node
	EnvVarNodeName = "NODE_NAME"
	// EnvVarClusterName is the cluster name
	EnvVarClusterName = "CLUSTER_NAME"
	// EnvVarCluster is the cluster name
	EnvVarCluster = "CLUSTER"
	// EnvVarClusterEnv is the env variable for cluster environment (e.g. prod, sandbox)
	EnvVarClusterEnv = "ENV"
	// EnvVarNodeEnv is the env variable for the node env (secrets agent stuff)
	EnvVarNodeEnv = "NODE_ENV"
	// EnvVarBlendLoggerJSON is the env var for the blend ts logger
	EnvVarBlendLoggerJSON = "BLEND_LOGGER_JSON"
	// EnvVarEtcdEndpoints are the etcd endpoints
	EnvVarEtcdEndpoints = "ETCD_ENDPOINTS"
	// EnvVarEtcdInitialCluster is the initial cluster
	EnvVarEtcdInitialCluster = "ETCD_INITIAL_CLUSTER"
	// EnvVarBuildMode is the build mode
	EnvVarBuildMode = "BUILD_MODE"
	// EnvVarDeployID is the deploy id
	EnvVarDeployID = "DEPLOY_ID"
	// EnvVarDeployedBy is the user it was deployed
	EnvVarDeployedBy = "DEPLOYED_BY"
	// EnvVarRegistryHost is the env variable specifying registry hostname
	EnvVarRegistryHost = "REGISTRY_HOST"
	// EnvVarDockerAuth is the env var that stores docker auth
	EnvVarDockerAuth = "DOCKER_AUTH"
	// EnvVarSnapshotBucket is the etcd backup bucket
	EnvVarSnapshotBucket = "SNAPSHOT_BUCKET"
	// EnvVarMachineRoleName is the role the machine is running under
	EnvVarMachineRoleName = "MACHINE_ROLE"
	// EnvVarAwsCredentialsMountPath is the path the aws credentials to mount
	EnvVarAwsCredentialsMountPath = "AWS_CREDENTIALS_MOUNT_PATH"
	// EnvVarInfraVaultGithubToken is the github token to login to infra vault
	EnvVarInfraVaultGithubToken = "INFRA_VAULT_GITHUB_TOKEN"
	// EnvVarBaseURL is the base deployinator url
	EnvVarBaseURL = "BASE_URL"
	// EnvVarBuildSpecJSON is the json encoded build spec
	EnvVarBuildSpecJSON = "BUILD_JSON"
	// EnvVarClusterAvailabilityZones are the availability zones of the cluster
	EnvVarClusterAvailabilityZones = "CLUSTER_AVAILABILITY_ZONES"
	// EnvVarBuilderConfigEnvVars is the set of environment variables to be defined for a particular config
	EnvVarBuilderConfigEnvVars = "CONTAINER_ENV_VARS"
	// EnvVarVPC is the vpc of the cluster
	EnvVarVPC = "VPC"
	// EnvVarSecrulesBranch specifies the branch that the secrules job will
	// pull from, from the secrules git repository.
	EnvVarSecrulesBranch = "SECRULES_BRANCH"
	// EnvVarGoogleClientID is the env variable for google oauth
	EnvVarGoogleClientID = "OAUTH_CLIENT_ID"
	// EnvVarGoogleClientSecret is the env variable for google oauth
	EnvVarGoogleClientSecret = "OAUTH_CLIENT_SECRET"
	// EnvVarGoogleSecret is the secret for oauth login
	EnvVarGoogleSecret = "OAUTH_SECRET"
	// EnvVarVaultUnsealKeys is the env variable for vault unseal keys
	EnvVarVaultUnsealKeys = "UNSEAL_KEYS"
	// EnvVarTLSCertPath is the path to the tls cert
	EnvVarTLSCertPath = "TLS_CERT_PATH"
	// EnvVarTLSKeyPath is the path to the tls key
	EnvVarTLSKeyPath = "TLS_KEY_PATH"
	// EnvVarUpstreamServer is the env var to the upstream server for proxy
	EnvVarUpstreamServer = "UPSTREAM"
	// EnvVarDisableUpgrade disables upgrading proxy connections to https
	EnvVarDisableUpgrade = "DISABLE_UPGRADE"
	// EnvVarDisableTLSProxy disables the tls proxy server in the proxy container
	EnvVarDisableTLSProxy = "DISABLE_TLS_PROXY"
	// EnvVarDatabaseName is the env var for database names
	EnvVarDatabaseName = "DATABASE_NAME"
	// EnvVarDatabaseBackupRefHash is the ref hash for database backup iamge
	EnvVarDatabaseBackupRefHash = "DATABASE_BACKUP_REF_HASH"
	// EnvVarDatabaseBackupPolicyArn is the policy arn for backing up databases
	EnvVarDatabaseBackupPolicyArn = "DATABASE_BACKUP_POLICY_ARN"
	// EnvVarDatabaseBackupBucketName is the s3 bucket to backup databases to
	EnvVarDatabaseBackupBucketName = "DATABASE_BACKUP_BUCKET"
	// EnvVarDatabaseBackupBucket is the s3 bucket to upload backups to
	EnvVarDatabaseBackupBucket = "S3_BUCKET_NAME"
	// EnvVarDatabaseBackupUser is the username for the database to backup
	EnvVarDatabaseBackupUser = "DATABASE_USER"
	// EnvVarDatabaseBackupPwd is the password for the database to backup
	EnvVarDatabaseBackupPwd = "DATABASE_PWD"
	// EnvVarDatabaseBackupDBName is the database to backup
	EnvVarDatabaseBackupDBName = "DB_SERVICE_NAME"
	// EnvVarBackupFileName is the database backup to restore from
	EnvVarBackupFileName = "BACKUP_FILE_NAME"
	// EnvVarBackupKey is the key to the database backup to restore from
	EnvVarBackupKey = "BACKUP_KEY"
	// EnvVarAWSAccountNumber is the env variable for AWS secret access key
	EnvVarAWSAccountNumber = "AWS_ACCOUNT_NUMBER"
	// EnvVarDeployNotif is the flag to notify the slack webhook when a deploy starts
	EnvVarDeployNotif = "DEPLOY_NOTIF"
	// EnvVarProjectSlackChannel is the slack channel for the project
	EnvVarProjectSlackChannel = "PROJECT_SLACK_CHANNEL"
	// EnvVarSlackChannel is the slack channel for the service
	EnvVarSlackChannel = "SLACK_CHANNEL"
	// EnvVarSlackWebhook is the slack webhook
	EnvVarSlackWebhook = "SLACK_WEBHOOK"
	// EnvVarInfosecAuditService is the review service webhook
	EnvVarInfosecAuditService = "AUDIT_SERVICE_WEBHOOK"
	// EnvVarGitRefHash is the git ref hash
	EnvVarGitRefHash = "GIT_REF_HASH"
	// EnvVarGitRef is the full git ref
	EnvVarGitRef = "GIT_REF"
	// EnvVarGitRemote is the git remote
	EnvVarGitRemote = "GIT_REMOTE"
	// EnvVarGitBase is the git base branch
	EnvVarGitBase = "GIT_BASE"
	// EnvVarServiceType is the type of the service
	EnvVarServiceType = "SERVICE_TYPE"
	// EnvVarServiceConfigPath is the service config path
	EnvVarServiceConfigPath = "SERVICE_CONFIG_PATH"
	// EnvVarSourcePath is the source path
	EnvVarSourcePath = "SOURCE_PATH"
	// EnvVarDockerRegistry is the docker registry
	EnvVarDockerRegistry = "DOCKER_REGISTRY"
	// EnvVarDockerfile is the dockerfile
	EnvVarDockerfile = "DOCKERFILE"
	// EnvVarContainerArgs are the args to the container
	EnvVarContainerArgs = "CONTAINER_ARGS"
	// EnvVarContainerWorkingDir is the working directory of the container
	EnvVarContainerWorkingDir = "CONTAINER_WORKING_DIR"
	// EnvVarContainerCommand is the container command
	EnvVarContainerCommand = "CONTAINER_COMMAND"
	// EnvVarContainerImage is the image of the container
	EnvVarContainerImage = "CONTAINER_IMAGE"
	// EnvVarDatadogNamespace is the datadog metric namespace
	EnvVarDatadogNamespace = "DATADOG_NAMESPACE"
	// EnvVarRestoredCluster is the cluster this was restored from
	EnvVarRestoredCluster = "RESTORED_CLUSTER"
	// EnvVarRestoredEnvironment is the env this was restored from
	EnvVarRestoredEnvironment = "RESTORED_ENVIRONMENT"
	// EnvVarNATIPs is the env var for nats ips
	EnvVarNATIPs = "NAT_IPS"
	// EnvVarServiceFQDN is the fqdn of the service
	EnvVarServiceFQDN = "SERVICE_FQDN"
	// EnvVarBackupNames are the backup names used for restore
	EnvVarBackupNames = "BACKUP_NAMES"
	// EnvVarPGPassword is the env var for pg password
	EnvVarPGPassword = "PGPASSWORD"
	// EnvVarPostgresPassword is the env var for postgres password
	EnvVarPostgresPassword = "POSTGRES_PASSWORD"
	// EnvVarPGUser is the env var for the pg user
	EnvVarPGUser = "PGUSER"
	// EnvVarPostgresUser is the env var for postgres user
	EnvVarPostgresUser = "POSTGRES_USER"
	// EnvVarDefaultInternalLoadBalancerSourceRanges is a long env var
	EnvVarDefaultInternalLoadBalancerSourceRanges = "DEFAULT_INTERNAL_LOAD_BALANCER_SOURCE_RANGES"
	// EnvVarRegistryStorageMaintenance is the env var for maintenance on the registry
	EnvVarRegistryStorageMaintenance = "REGISTRY_STORAGE_MAINTENANCE"
	// EnvVarDeployinatorEmail is the env var for the deployinator email address
	EnvVarDeployinatorEmail = "DEPLOYINATOR_EMAIL"
	// EnvVarProjectRun is the env var informing that the service is part of a project run
	EnvVarProjectRun = "PROJECT_RUN"

	// INFRADEV-1185 Twistlock evaluation

	// EnvVarTwistlockAddress is the env var for twistlock address
	EnvVarTwistlockAddress = "TWISTLOCK_ADDRESS"
	// EnvVarTwistlockProject is the env var for twistlock project
	EnvVarTwistlockProject = "TWISTLOCK_PROJECT"
	// EnvVarTwistlockUsername is the env var for twistlock username
	EnvVarTwistlockUsername = "TWISTLOCK_USERNAME"
	// EnvVarTwistlockPassword is the env var for twistlock password
	EnvVarTwistlockPassword = "TWISTLOCK_PASSWORD"
	// EnvVarTwistlockEnabled tells whether twistlock is enabled or not
	EnvVarTwistlockEnabled = "TWISTLOCK_ENABLED"

	// Warden

	// EnvVarWardenClientUID is the warden proxy client UID (the environment UID of the service)
	EnvVarWardenClientUID = "CLIENT_UID"
	// EnvVarWardenSSOPortalURL is the SSO login URL
	EnvVarWardenSSOPortalURL = "SSO_PORTAL_URL"
	// EnvVarWardenProxyBindAddr is the bind address for the Warden proxy
	EnvVarWardenProxyBindAddr = "PROXY_BIND_ADDR"
	// EnvVarWardenProxyUpgradeBindAddr is the bind address for the Warden proxy's HTTP upgrader
	EnvVarWardenProxyUpgradeBindAddr = "PROXY_UPGRADE_BIND_ADDR"
	// EnvVarWardenProxyUpstream is the address of the server that the proxy sits in front of
	EnvVarWardenProxyUpstream = "PROXY_UPSTREAM"
	// EnvVarWardenProxyUseProxyProtocol determines whether or not the proxy protocol headers should be set
	EnvVarWardenProxyUseProxyProtocol = "USE_PROXY_PROTOCOL"
	// EnvVarWardenRemoteAddr is the address of the Warden server
	EnvVarWardenRemoteAddr = "WARDEN_REMOTE_ADDR"
	// EnvVarWardenTLSCertPath is the path to a Warden TLS certificate
	EnvVarWardenTLSCertPath = "WARDEN_TLS_CERT_PATH"
	// EnvVarWardenTLSKeyPath is the path to the private key of a Warden TLS certificate key
	EnvVarWardenTLSKeyPath = "WARDEN_TLS_KEY_PATH"
	// EnvVarWardenCACertPath is the path to the Warden CA cert
	EnvVarWardenCACertPath = "WARDEN_CA_CERT_PATH"

	// EnvVarWardenNamespace is the namespace for warden services
	EnvVarWardenNamespace = "WARDEN_NAMESPACE"

	// EnvVarGithubEnterpriseAPIToken is the env var for ghe api token
	EnvVarGithubEnterpriseAPIToken = "GHE_API_TOKEN"

	// EnvVarSentryDSN is the DSN for the Sentry client. This determines
	// which project events to Sentry will get reported to.
	EnvVarSentryDSN = "SENTRY_DSN"
)

const (
	// BlendGithubEnterpriseHost is our blend GHE host
	BlendGithubEnterpriseHost = "git.blendlabs.com"
)

const (
	// ServiceEnvPentest is the env for pentest
	ServiceEnvPentest = "pentest"
)

const (
	// APIDefaultLogLines is the default number of lines to return from the api
	APIDefaultLogLines = 2048
	// APIUnlimitedLogLines tells the api to request all log lines from k8s
	APIUnlimitedLogLines = -1
)

const (
	// FieldPathName is the downward api field path for pod name
	FieldPathName = "metadata.name"
	// FieldPathNamespace is the downward api field path for pod namespace
	FieldPathNamespace = "metadata.namespace"
	// FieldPathNodeName is the downward api field path for node name
	FieldPathNodeName = "spec.nodeName"
	// FieldPathHostIP is the downward api field path for host ip
	FieldPathHostIP = "status.hostIP"
)

const (
	// IngressName is the name of the default ingress
	IngressName = "ingress-nginx"
	// IngressNamespace is the namespace where ingress lives
	IngressNamespace = BlendSystemNamespace

	// RegistryName is the name of the registry
	RegistryName = "kube-registry"
	// RegistryNamespace is the namespace of the registry
	RegistryNamespace = BlendSystemNamespace

	// DataDogPort is the port for datadog
	DataDogPort = 8125
	// DataDogTracePort is the port for datadog trace (APM)
	DataDogTracePort = 8126

	// VaultPort is the port for vault
	VaultPort       = 8200
	vaultPortString = "8200" // can't convert int into string in const

	// VaultClusterSize is the cluster size for vault
	VaultClusterSize = 3

	// ClusterAutoscalerName is the name of the cluster autoscaler
	ClusterAutoscalerName = "cluster-autoscaler"
	// DataDogName is the name of the dd service
	DataDogName = "dd-agent"
	// RegistryDomainName is the name of the registry used for dns
	RegistryDomainName = "registry"
	// VaultName is the name of the vault service
	VaultName = "vault"
	// VaultSecretName is the name of the vault secret for storing unseal keys
	VaultSecretName = "vault"

	// VaultAWSDebugSecretKey is the key for vault aws key debug
	VaultAWSDebugSecretKey = "secret/debug/vault/awskey"

	// VaultAuditLogPath is the mount path for vault audit log
	VaultAuditLogPath = "/var/log/vault-audit.log"

	// DataDogHost is the host of datadog on kube
	DataDogHost = DataDogName + "." + coreapi.NamespaceSystem
	// DockerBridgeHost is the default docker bridge host
	DockerBridgeHost = "172.17.0.1"
	// DefaultAwsCredentialsMountPath is the default mount path for the credentials
	DefaultAwsCredentialsMountPath = "/root/.aws"
	// MetricsServerName is the name for metric server
	MetricsServerName = "metrics-server"

	// KIAMName is the name of kiam
	KIAMName = "kiam"
	// KIAMServerName is the name of the kiam server
	KIAMServerName = KIAMName + "-server"
	// KIAMAgentName is the name of the kiam agent
	KIAMAgentName = KIAMName + "-agent"
)

const (
	// DeployinatorServerImage is the image name of the deployinator server
	DeployinatorServerImage = NameDeployinator
	// DeployinatorBuilderImage is the image name of the deployinator builder
	DeployinatorBuilderImage = "deployinator-builder"
	// DeployinatorDeployerImage is the image name of the deployinator deployer
	DeployinatorDeployerImage = "deployinator-deployer"
	// DeployinatorWorkerImage is the image name of the deployinator worker
	DeployinatorWorkerImage = NameDeployinatorWorker
	// DatabaseBackupRestoreImage is the image name of the database backup and restore task
	DatabaseBackupRestoreImage = "database-backup-restore"
	// DeployinatorProxyImage is the image name of the deployinator proxy container
	DeployinatorProxyImage = NameProxy
	// DeployinatorInitImage is the image name of the deployinator init container
	DeployinatorInitImage = "init"
	// DeployinatorCredentialsImage is the image name of the deployinator credentials container
	DeployinatorCredentialsImage = NameCredentials
	// DeployinatorCollectorImage is the image name of the deployinator credentials container
	DeployinatorCollectorImage = "event-collector"
	// ClusterUpdaterImage is the image name of the cluster updater image
	ClusterUpdaterImage = NameClusterUpdater
)

const (
	// NameSSHBox is the sshbox name
	NameSSHBox = "sshbox"
	// SSHBoxImage is the image name of the sshbox container
	SSHBoxImage = "sshbox"
)

const (
	// NameWarden is the name of Warden
	NameWarden = "warden"
	// WardenServerImage is the image name of the Warden server
	WardenServerImage = NameWarden
	// WardenProxyImage is the image name of the Warden proxy
	WardenProxyImage = "warden-proxy"
	// WardenSSOImage is the image name of the Warden SSO
	WardenSSOImage = "warden-sso"
	// WardenBackupsImage is the image name of the Warden backups job
	WardenBackupsImage = "warden-backups"
	// WardenMigrationsImage is the image name of the Warden migrations job
	WardenMigrationsImage = "warden-migrations"
)

const (
	// InfraVaultSecretPrefix is the prefix for all k8s secrets in infra vault
	InfraVaultSecretPrefix = "secret/k8s"

	// InfraVaultWardenSecretPrefix is the prefix for all warden secrets in infra vault
	InfraVaultWardenSecretPrefix = "secret/infrastructure/warden"
)

const (
	// Infosec-audit-service server URL
	InfosecAuditService = "https://infosec-audit-service.sandbox.k8s.centrio.com/api/example"
)

const (
	// EtcdDirPath is the path to the dir of etcd
	EtcdDirPath = "/var/run/etcd"
	// EtcdDataDirPath is the path to the data dir of etcd
	EtcdDataDirPath = EtcdDirPath + "/default.etcd"
	// EtcdInitialClusterToken is the initial cluster token of etcd
	EtcdInitialClusterToken = "etcd-cluster-1"

	// EtcdServerPort is the port for communication with etcd peers
	EtcdServerPort = 2380
	// EtcdClientPort is the port for communication with etcd
	EtcdClientPort = 2379
)

const (
	// BuilderDeployerVolumeName is the name of the shared volume between deployer and builder
	BuilderDeployerVolumeName = "build-config"
	// BuilderDeployerVolumeMountPath is the mount path for the shared volume between deployer and builder
	BuilderDeployerVolumeMountPath = "/build"
	// DeployConfigFile is the path of the config file for the build
	DeployConfigFile = BuilderDeployerVolumeMountPath + "/config"
	// DockerContextFile is the path of the compressed context file for docker build
	DockerContextFile = BuilderDeployerVolumeMountPath + "/ctx"
)

var (
	// VaultHost is the host of vault on kube
	VaultHost = VaultName + "." + GetBlendSystemNamespace() + ":" + vaultPortString
	// VaultAddr is the vault address
	VaultAddr = "https://" + VaultHost
)

const (
	// FluentAwsRegion specifies the kinese main region
	FluentAwsRegion = "us-east-1"
)

var (
	// FluentAwsRegionToELBAccountID maps fluent AWS region names to their corresponding Elastic Load Balancing Account ID\
	// This mapping is used during the creation of an S3 bucket policy to ensure that the region's matching Elastic Load Balancer account has permission to put access logs into the bucket.
	// More info: https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-access-logs.html#attach-bucket-policy
	// TODO: replace this mapping once deployinator provisions buckets via terraform https://www.terraform.io/docs/providers/aws/d/elb_service_account.html
	FluentAwsRegionToELBAccountID = func() map[string]string {
		return map[string]string{
			"us-east-1":      "127311923021",
			"us-east-2":      "033677994240",
			"us-west-1":      "027434742980",
			"us-west-2":      "797873946194",
			"ca-central-1":   "985666609251",
			"eu-central-1":   "054676820928",
			"eu-west-1":      "156460612806",
			"eu-west-2":      "652711504416",
			"eu-west-3":      "009996457667",
			"eu-north-1":     "897822967062",
			"ap-northeast-1": "582318560864",
			"ap-northeast-2": "600734575887",
			"ap-northeast-3": "383597477331",
			"ap-southeast-1": "114774131450",
			"ap-southeast-2": "783225319266",
			"ap-south-1":     "718504428378",
			"sa-east-1":      "507241528517",
			// "us-gov-west-1*":   "048591011584",
			// "us-gov-east-1*":   "190560391635",
			// "cn-north-1**":     "638102146993",
			// "cn-northwest-1**": "037604701340",
		}
	}
)

// GetDefaultELBAccessLogsBucketName generates a default bucket name for the elb access logs
func GetDefaultELBAccessLogsBucketName(env string) string {
	return fmt.Sprintf("blend-k8s-%s-elb-accesslogs", env)
}

var (
	// ReservedServiceNames are the names reserved for internal usage
	ReservedServiceNames = []string{
		DeployinatorServerImage,
		DeployinatorWorkerImage,
		DeployinatorBuilderImage,
		DeployinatorDeployerImage,
		DeployinatorCredentialsImage,
		DeployinatorInitImage,
		DeployinatorProxyImage,
		ClusterUpdaterImage,
		KIAMName,
		KIAMAgentName,
		KIAMServerName,
		"kiam-helper", // TODO remove this after migrating secrets ENGPROD-84
	}

	// ReservedDomainPrefixes are the prefixes reserved for our use on the cluster domain
	ReservedDomainPrefixes = []string{
		NameDeployinator,
		WebhooksPublicName,
		VaultName,
		RegistryDomainName,
	}

	// ReservedNamespaceNames are the reserved namespace names
	ReservedNamespaceNames = []string{
		GetBlendNamespace(),
		GetBlendSystemNamespace(),
		GetEtcdNamespace(),
		GetKubeSystemNamespace(),
		GetEventCheckpointerNamespace(),
		SmokeSignalNamespace,
		"instabase",
		"twistlock",
		"kube-public",
		"default",
	}
)

var (
	// VPNSourceRanges are the ranges for whitelisting vpn accessibility
	VPNSourceRanges = []string{
		"52.24.166.9/32",    // pritunl all traffic
		"50.18.206.74/32",   // open vpn dev
		"34.197.122.136/32", // Chromebooks
	}
)

const (
	// WebhooksPublicName is the name of the webhook proxy
	WebhooksPublicName = "webhooks"
)

// database related constants
const (
	// PostgresName is name used for things related to postgres db launches
	PostgresName = "postgres-db"
	// PostgresDBImage is name of Docker image with postgres version to use
	PostgresDBImage = "postgres:10"

	// DefaultDBStorage is default value for memory to give to a database
	DefaultDBStorage = 250
	// DefaultCPURequest is the default amount of cpu requested for the database in millicores
	DefaultCPURequest = "250m"
	// DefaultMemoryRequest is the default amount of memory requested for the database
	DefaultMemoryRequest = "512m"
	// DefaultCPULimit is the default max amount of cpu that can be used by the database in millicores
	DefaultCPULimit = "2"
	// DefaultMemoryLimit is the default max amount of memory that can be used by the database
	DefaultMemoryLimit = "4Gi"
	// BackupDisplayLimit is the default number of backups that will be shown on the UI
	BackupDisplayLimit = 20
	// DatabaseBackupJobActiveDeadlineSeconds is how long we give the cronjob to finish running
	DatabaseBackupJobActiveDeadlineSeconds = "86400"

	// PostgresDataVolName is the name of the volume for postgres data to be stored in
	PostgresDataVolName = "postgres-storage"
	// PostgresDataVolPath is the path to the postgres data volume
	PostgresDataVolPath = PostgresMountPath + "/data"
	// PostgresMountPath is the path to the custom postgres mount point
	PostgresMountPath = "/postgresql"

	// PostgresPwdLen is length of password to create
	PostgresPwdLen = 16
	// PostgresSecret is the secret for postgres database creds
	PostgresSecret = "postgres-secrets"
	// PostgresUsernameEnvVar is the username environment variable
	PostgresUsernameEnvVar = "POSTGRES_USER"
	// PostgresPasswordEnvVar is the password environment variable
	PostgresPasswordEnvVar = "POSTGRES_PASSWORD"
	// PostgresUser is the username of the default postgres admin user
	PostgresUser = "postgres_admin"

	// PostgresEnableSSLVolName is the name of the volume for the enable ssl script to be stored in
	PostgresEnableSSLVolName = "postgres-enable-ssl"
	// PostgresServerTLSSecret is the name of the secret containing the server key and cert
	PostgresServerTLSSecret = "tls-secret"
	// PostgresServerCert is the name of the volume containing the server key and cert
	PostgresServerKeyCert = "postgres-server-key-cert"
	// PostgresSSLPath is the path to the default SSL folder
	PostgresSSLPath = "/ssl_files"
	// PostgresConfigSecret is the name of the secret containing the postgres configs
	PostgresConfigSecret = "config"
	// PostgresConfigFile is the file name of the postgres config file
	PostgresConfigFile = "postgresql.conf"
	// PostgresConfigFileMountPath is the path to the config file
	PostgresConfigFileMountPath = "/var/lib/postgres"

	// SecretDatabaseBackup is the secret holding aws resources for database backup tasks
	SecretDatabaseBackup = "database-backup-secrets"
	// DatabaseBackupName is the name for the database backup addon
	DatabaseBackupName = "database-backup-restore"
	// MetricsServerTLSSecret is the secret holding the metrics server key and cert
	MetricsServerTLSSecret = "metrics-server-tls"
)

const (
	// RandomStringLength is the length of a random string appended to names
	RandomStringLength = 8

	// ELBSecurityGroupPrefix is the prefix for the default elb security group
	ELBSecurityGroupPrefix = "k8s-elb"
)

// Smoke Signal related constants
const (
	// SmokeSignalTokenName is an env variable that contains a smoke signal auth token
	SmokeSignalTokenName = "SMOKE_SIGNAL_TOKEN"
	// SmokeSignalEndpointName is an env variable that contains the smoke signal endpoint
	SmokeSignalEndpointName = "SMOKE_SIGNAL_ENDPOINT"
	// SmokeSignalCheckAPIPrefix is the prefix to the check API
	SmokeSignalCheckAPIPrefix = "/api/check"
	// SmokeSignalAuthHeaderKey is the key for the auth header needed for protected operations
	SmokeSignalAuthHeaderKey = "smoke-signal-auth"
	// DeployCheckFile is the path of the config file for the build
	DeployCheckFilePath = BuilderDeployerVolumeMountPath + "/smoke-signal-check"
	// SmokeSignalNamespace is the namespace for Smoke Signal
	SmokeSignalNamespace = "smoke-signal"
)

const (
	// EndpointRemovalGracePeriod is the (overestimated) grace period for the pod to be removed from the endpoint
	EndpointRemovalGracePeriod = 15 * time.Second
	// ProxyShutdownGracePeriod is the grace period for the proxy server to shutdown
	ProxyShutdownGracePeriod = 15 * time.Second
)

const (
	// PagerDutyEmail is the email to send to for pagerduty
	PagerDutyEmail = "k8s-email@blend.pagerduty.com"
)

const (
	// DeployNotificationsComplete represents only notifying on complete, which is the default
	DeployNotificationsComplete = "complete"
	// DeployNotificationsAll represents notifications on start and success/failure
	DeployNotificationsAll = "all"
	// DeployNotificationsSuccess represents notifications only on success
	DeployNotificationsSuccess = "success"
	// DeployNotificationsFailure represents notifications only on failure
	DeployNotificationsFailure = "failure"
)

const (
	// LabelTeam is the label name for a team
	LabelTeam = "blend-team"
	// TeamNameUnassigned is the default team name
	TeamNameUnassigned = "unassigned"
)
