package types

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"git.blendlabs.com/blend/deployinator/pkg/core"
	"git.blendlabs.com/blend/deployinator/pkg/kube"
	"github.com/blend/go-sdk/env"
	exception "github.com/blend/go-sdk/exception"
	"k8s.io/kubernetes/pkg/apis/apps"
	coreapi "k8s.io/kubernetes/pkg/apis/core"
)

var (
	// DefaultBlendNamespaces are the namespaces in a blend cluster
	DefaultBlendNamespaces = []string{GetBlendSystemNamespace(), GetBlendNamespace(), GetEventCheckpointerNamespace()}
)

// GetBlendSystemNamespace returns the blend system namespace
func GetBlendSystemNamespace() string {
	return BlendSystemNamespace
}

// GetBlendNamespace returns the blend system namespace
func GetBlendNamespace() string {
	return BlendNamespace
}

// GetKubeSystemNamespace returns the kube system namespace
func GetKubeSystemNamespace() string {
	return coreapi.NamespaceSystem
}

// GetEtcdNamespace returns the etcd namespace
func GetEtcdNamespace() string {
	return GetKubeSystemNamespace()
}

// GetEventCheckpointerNamespace returns the namespace for the event-checkpointer addon
func GetEventCheckpointerNamespace() string {
	return EventCheckpointerNamespace
}

// GetDeployinatorAWSRoleName is the role name for deployinator
func GetDeployinatorAWSRoleName(clusterName string) string {
	return fmt.Sprintf("deployinator.%s", clusterName)
}

// MastersRoleName is the role name of the masters
func MastersRoleName(clusterName string) string {
	return fmt.Sprintf("masters.%s", clusterName)
}

// NodesRoleName is the role name of the masters
func NodesRoleName(clusterName string) string {
	return fmt.Sprintf("nodes.%s", clusterName)
}

// BuildersRoleName is the role name of the masters
func BuildersRoleName(clusterName string) string {
	return fmt.Sprintf("builder-nodes.%s", clusterName)
}

// CoreNodesRoleName is the role name of the masters
func CoreNodesRoleName(clusterName string) string {
	return fmt.Sprintf("core-nodes.%s", clusterName)
}

// KIAMServerRoleName is the name of the role for kiam server
func KIAMServerRoleName(clusterName string) string {
	return fmt.Sprintf("%s.%s", KIAMServerName, clusterName)
}

// BastionsSecGroupName is the security group name of the bastion
func BastionsSecGroupName(clusterName string) string {
	return fmt.Sprintf("bastions.%s", clusterName)
}

// BuilderImageFromEnv returns the builder image from the registry and ref env vars
func BuilderImageFromEnv() string {
	return kube.DockerImage(env.Env().String(EnvVarRegistryHost), DeployinatorBuilderImage, env.Env().String(EnvVarCurrentRef))
}

// DeployerImageFromEnv returns the deployer image from the registry and ref env vars
func DeployerImageFromEnv() string {
	return kube.DockerImage(env.Env().String(EnvVarRegistryHost), DeployinatorDeployerImage, env.Env().String(EnvVarCurrentRef))
}

// SSHBoxImageFromEnv returns the SSHBox image from the registry and ref env vars
func SSHBoxImageFromEnv() string {
	return kube.DockerImage(env.Env().String(EnvVarRegistryHost), SSHBoxImage, env.Env().String(EnvVarCurrentRef))
}

// WebhookURL returns the url of the webhook hooks for github
func WebhookURL(deployinatorURL string) string {
	return fmt.Sprintf("%s%s", deployinatorURL, WebhookPath)
}

// PublicWebhookDomainName is the domain name for public webhooks
func PublicWebhookDomainName(clusterName string) string {
	return fmt.Sprintf("%s.%s", WebhooksPublicName, clusterName)
}

// PublicWebhookURL is the url for github to send webhooks to
func PublicWebhookURL(clusterName, project string) string {
	return fmt.Sprintf("https://%s/%s", PublicWebhookDomainName(clusterName), project)
}

// DeployinatorDomainName is the domain name for Deployinator
func DeployinatorDomainName(clusterName string) string {
	return fmt.Sprintf("%s.%s", NameDeployinator, clusterName)
}

// DeployinatorURL returns the base deployinator url
func DeployinatorURL(clusterName string) string {
	return fmt.Sprintf("https://deployinator.%s", clusterName)
}

// DatabaseBackupPolicyName returns the name of the aws policy for the database backup
func DatabaseBackupPolicyName(clusterName string) string {
	return fmt.Sprintf("%s-%s", DatabaseBackupName, clusterName)
}

// DatabaseConfigSecret returns the name of the config secret for the database
func DatabaseConfigSecret(databaseName string) string {
	return fmt.Sprintf("%s-%s", databaseName, PostgresConfigSecret)
}

// IAMRoleName returns the name of the IAMRole
func IAMRoleName(serviceName, clusterShortName string) string {
	return fmt.Sprintf("k8s-%s-%s", clusterShortName, serviceName)
}

// UserSystemTarget is the target for the system user
func UserSystemTarget() string {
	return UserTarget(UserSystem)
}

// UserAutodeployTarget is the target for the auto deploy user
func UserAutodeployTarget() string {
	return UserTarget(UserAutodeploy)
}

// RandomAZ returns a random availability zone within the cluster, or an error if there are none
func RandomAZ() (string, error) {
	err := env.Env().Require(EnvVarClusterAvailabilityZones)
	if err != nil {
		return "", exception.New(err)
	}
	zones := strings.Split(env.Env().String(EnvVarClusterAvailabilityZones), ",")
	validated := []string{}
	for _, zone := range zones {
		if len(zone) > 0 {
			validated = append(validated, zone)
		}
	}
	if len(validated) < 1 {
		return "", exception.New(fmt.Sprintf("No available AZs"))
	}
	n := rand.Intn(len(validated))
	return validated[n], nil
}

// KubernetesELBSecurityGroup is the security group for all k8s elbs in the cluster
func KubernetesELBSecurityGroup(clusterName string) string {
	return fmt.Sprintf("%s-%s", ELBSecurityGroupPrefix, clusterName)
}

// VPNAccessibilitySourceRanges returns the source ranges for the vpn accessibility level
func VPNAccessibilitySourceRanges() []string {
	ipList := strings.TrimSpace(env.Env().String(EnvVarNATIPs))
	nats := []string{}
	if len(ipList) > 0 {
		nats = strings.Split(ipList, ",")
	}
	ret := []string{}
	for _, nat := range nats {
		nat := strings.TrimSpace(nat)
		if len(nat) == 0 {
			continue
		}
		if !strings.Contains(nat, "/") {
			nat = nat + "/32" // nat cidr from ip
		}
		ret = append(ret, nat)
	}
	return append(ret, VPNSourceRanges...)
}

/* Vault-related functions */

// InfraVaultSecretPath returns the root path for secrets for a cluster, no trailing slash
func InfraVaultSecretPath(clusterName string) string {
	return fmt.Sprintf("%s/%s", InfraVaultSecretPrefix, clusterName)
}

// InfraVaultWardenSecretPath returns the root path for secrets for a cluster, no trailing slash
func InfraVaultWardenSecretPath() string {
	return fmt.Sprintf("%s", InfraVaultWardenSecretPrefix)
}

func timedStringPair(base string) (string, string) {
	return base, fmt.Sprintf("%s%s", base, core.TimeString())
}

// VaultRootTokenSecretKey is the secret key in infravault that contains the root token. The first returned string should be read from
// both should be written to
func VaultRootTokenSecretKey(clusterName string) (string, string) {
	return timedStringPair(fmt.Sprintf("%s/vault", InfraVaultSecretPath(clusterName)))
}

// EncryptionConfigSecretKey is the secret key in infravault that contains the kube encryption-at-rest config. The first returned string should be read from
// both should be written to
func EncryptionConfigSecretKey(clusterName string) (string, string) {
	return timedStringPair(fmt.Sprintf("%s/encryptionconfig", InfraVaultSecretPath(clusterName)))
}

// VaultPeerName returns the name of a vault instance
func VaultPeerName(id int) string {
	return fmt.Sprintf("%s-%d", VaultName, id)
}

// VaultPeerHost returns the host of a vault instance
func VaultPeerHost(id int) string {
	return fmt.Sprintf("%s.%s.%s:%d", VaultPeerName(id), VaultName, GetBlendSystemNamespace(), VaultPort)
}

// ForEachVaultPeer runs the function `fn` for each vault instance in the cluster
func ForEachVaultPeer(fn func(id int) error) error {
	for i := 0; i < VaultClusterSize; i++ {
		if err := fn(i); err != nil {
			return exception.New(err)
		}
	}
	return nil
}

// BuildSpecFromEnv creates a build spec from the environment
func BuildSpecFromEnv() (*BuildSpec, error) {
	if err := env.Env().Require(EnvVarBuildSpecJSON); err != nil {
		return nil, err
	}
	buildSpec := &BuildSpec{}
	err := json.Unmarshal(env.Env().Bytes(EnvVarBuildSpecJSON), buildSpec)
	if err != nil {
		return nil, exception.New(err)
	}
	return buildSpec, nil
}

// ValidateNonReservedDomain returns an error if the domain is not allowed
func ValidateNonReservedDomain(fqdn, clusterName string) error {
	fqdn = strings.ToLower(fqdn)
	clusterName = strings.ToLower(clusterName)
	for _, p := range ReservedDomainPrefixes {
		if fqdn == fmt.Sprintf("%s.%s", p, clusterName) {
			return exception.New(fmt.Errorf("Domain %s is reserved", fqdn))
		}
	}
	return nil
}

// DefaultInternalLoadBalancerSourceRanges returns the default source ranges
func DefaultInternalLoadBalancerSourceRanges() []string {
	e := env.Env().String(EnvVarDefaultInternalLoadBalancerSourceRanges)
	if len(e) > 0 {
		return strings.Split(e, ",")
	}
	return nil
}

// RegistryMaintenanceReadonlyValue returns the value for the env var to set registry as read only or not
func RegistryMaintenanceReadonlyValue(readonly bool) string {
	// gotta do this because docker is broken on individual vars
	// https://github.com/docker/distribution/issues/1736
	return fmt.Sprintf("readonly:\n  enabled: %s", strconv.FormatBool(readonly))
}

// DeployinatorEmail uses the environment to get the email for deployinator to use
func DeployinatorEmail() string {
	return env.Env().String(EnvVarDeployinatorEmail)
}

// IsReservedLabel checks if a label conficts with any of the generated label names
func IsReservedLabel(label string) bool {
	reservedLabels := map[string]bool{
		kube.LabelApp:                        true,
		kube.LabelAddon:                      true,
		kube.LabelNodeRoleMaster:             true,
		kube.LabelNodeLabelRole:              true,
		kube.LabelNodeLabelRoleMaster:        true,
		kube.LabelNodeLabelRoleBuilderNode:   true,
		kube.LabelNodeLabelRoleCoreNode:      true,
		apps.DefaultDeploymentUniqueLabelKey: true,
	}
	allowedLabels := map[string]bool{
		// LabelTeam is added during the team labelling migration and should be allowed
		kube.LabelTeam: true,
	}
	if _, ok := reservedLabels[label]; ok || strings.HasPrefix(label, kube.ReservedLabelPrefix) {
		allowed, ok := allowedLabels[label]
		return !ok || !allowed
	}
	return false
}
