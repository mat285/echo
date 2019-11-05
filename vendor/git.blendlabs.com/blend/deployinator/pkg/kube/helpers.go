package kube

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.blendlabs.com/blend/deployinator/pkg/core"
	"github.com/blend/go-sdk/env"
	exception "github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/ref"
	jsonpatch "github.com/evanphx/json-patch"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	v1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	"k8s.io/kubernetes/pkg/kubelet/apis"
)

// GetNamespace returns the default kubernetes namespace.
func GetNamespace(namespace ...string) string {
	if env.Env().Has(EnvVarNamespace) {
		return env.Env().String(EnvVarNamespace)
	}
	if len(namespace) > 0 {
		return namespace[0]
	}
	return DefaultNamespace
}

// CreateConfigMapDataPatch creates a json patchset for a given configmap's data.
func CreateConfigMapDataPatch(data map[string]*string) []byte {
	dataJSON, _ := json.Marshal(data)
	return []byte(fmt.Sprintf(`{ "data" : %s }`, string(dataJSON)))
}

// CreateDeploymentMetadataPatch creates a json patchset for a given deployment's labels.
func CreateDeploymentMetadataPatch(dep *v1beta1.Deployment) []byte {
	labelsData, _ := json.Marshal(dep.Labels)
	annotationsData, _ := json.Marshal(dep.Annotations)
	podSpecLabelsData, _ := json.Marshal(dep.Spec.Template.Labels)
	podSpecAnnotationsData, _ := json.Marshal(dep.Spec.Template.Annotations)
	return []byte(
		fmt.Sprintf(`{ "metadata" : { "labels" : %s, "annotations": %s }, "spec" : { "template" : { "metadata" : { "labels" : %s, "annotations" : %s } } } }`,
			string(labelsData),
			string(annotationsData),
			string(podSpecLabelsData),
			string(podSpecAnnotationsData),
		),
	)
}

// CreateDeploymentLabelsPatch creates a json patchset for a given deployment's labels.
func CreateDeploymentLabelsPatch(dep *v1beta1.Deployment) []byte {
	labelsData, _ := json.Marshal(dep.Labels)
	return []byte(fmt.Sprintf(`{ "metadata" : { "labels" : %s } }`, string(labelsData)))
}

// CreateDeploymentAnnotationsPatch creates a json patchset for a given deployment's annotations.
func CreateDeploymentAnnotationsPatch(dep *v1beta1.Deployment) []byte {
	annotationsData, _ := json.Marshal(dep.Annotations)
	return []byte(fmt.Sprintf(`{ "metadata" : { "annotations" : %s } }`, string(annotationsData)))
}

// CreateDeploymentPodSpecMetadataPatch creates a deployment pod spec annotations patch.
func CreateDeploymentPodSpecMetadataPatch(dep *v1beta1.Deployment) []byte {
	labelsData, _ := json.Marshal(dep.Spec.Template.Labels)
	annotationsData, _ := json.Marshal(dep.Spec.Template.Annotations)
	return []byte(fmt.Sprintf(`{ "spec" : { "template" : { "metadata" : { "labels" : %s, "annotations" : %s } } } }`, string(labelsData), string(annotationsData)))
}

// CreateDeploymentPodSpecLabelsPatch creates a deployment pod spec annotations patch.
func CreateDeploymentPodSpecLabelsPatch(dep *v1beta1.Deployment) []byte {
	labelsData, _ := json.Marshal(dep.Spec.Template.Labels)
	return []byte(fmt.Sprintf(`{ "spec" : { "template" : { "metadata" : { "labels" : %s } } } }`, string(labelsData)))
}

// CreateDeploymentPodSpecAnnotationsPatch creates a deployment pod spec annotations patch.
func CreateDeploymentPodSpecAnnotationsPatch(dep *v1beta1.Deployment) []byte {
	annotationsData, _ := json.Marshal(dep.Spec.Template.Annotations)
	return []byte(fmt.Sprintf(`{ "spec" : { "template" : { "metadata" : { "annotations" : %s } } } }`, string(annotationsData)))
}

// CreateDeploymentReplicasPatch creates a deployment replicas patch.
func CreateDeploymentReplicasPatch(replicas int) []byte {
	return []byte(fmt.Sprintf(`{ "spec" : { "replicas" : %d } }`, replicas))
}

// CreateJobLabelsPatch creates a json patchset for a given job's labels.
func CreateJobLabelsPatch(job *batchv1.Job) []byte {
	labelsData, _ := json.Marshal(job.Labels)
	return []byte(fmt.Sprintf(`{ "metadata" : { "labels" : %s } }`, string(labelsData)))
}

// CreateServiceSelectorPatch creates a service selector patch.
func CreateServiceSelectorPatch(labels map[string]string) []byte {
	selectorData, _ := json.Marshal(labels)
	return []byte(fmt.Sprintf(`{ "spec" : { "selector" : %s } }`, string(selectorData)))
}

// CreateServiceVersionSelectorPatch creates a service selector patch.
func CreateServiceVersionSelectorPatch(version string) []byte {
	return []byte(fmt.Sprintf(`{ "spec" : { "selector" : { "blend-version" : "%s" } } }`, version))
}

// CredentialsDeploymentName returns the credentials deployment name for a given scheduled task.
func CredentialsDeploymentName(serviceName string) string {
	return fmt.Sprintf("%s-credentials", serviceName)
}

// ServiceDeployLabel returns the label for the previously deployed node of a service
func ServiceDeployLabel(serviceName string) string {
	return fmt.Sprintf("deploy-%s", serviceName)
}

// ServiceCredentialsSecretName returns the credentials (e.g. vault token) secrets bucket name for a given service.
func ServiceCredentialsSecretName(serviceName string) string {
	return fmt.Sprintf("%s-credentials", serviceName)
}

// ServiceEnvVarsSecretName returns the env variable secrets bucket name for a given service.
func ServiceEnvVarsSecretName(serviceName string) string {
	return fmt.Sprintf("%s-env-vars", serviceName)
}

// ServiceFilesSecretName returns the files secrets bucket name for a given service.
func ServiceFilesSecretName(serviceName string) string {
	return fmt.Sprintf("%s-files", serviceName)
}

// ServiceCertsSecretName returns the certs secrets bucket name for a given service.
func ServiceCertsSecretName(serviceName string) string {
	return fmt.Sprintf("%s-certs", serviceName)
}

// ServiceVaultPolicyName returns the vault policy name for the service
func ServiceVaultPolicyName(serviceName string) string {
	return fmt.Sprintf("service-%s", serviceName)
}

// ServiceVaultAdminPolicyName returns the vault policy name for the service admin
func ServiceVaultAdminPolicyName(serviceName string) string {
	return fmt.Sprintf("serviceadmin-%s", serviceName) // no hyphen on purpose to avoid name collisions
}

// ProjectEnvVarsSecretName returns the env variable secrets bucket name for a given project.
func ProjectEnvVarsSecretName(projectName string) string {
	return fmt.Sprintf("project-%s-env-vars", projectName)
}

// ProjectFilesSecretName returns the files secrets bucket name for a given project.
func ProjectFilesSecretName(projectName string) string {
	return fmt.Sprintf("project-%s-files", projectName)
}

// ProjectVaultAdminPolicyName returns the vault policy name for the project admin
func ProjectVaultAdminPolicyName(projectName string) string {
	return fmt.Sprintf("projectadmin-%s", projectName)
}

// DatabaseBackupRestorePolicyName returns the vault policy name for the database backups and restores
func DatabaseBackupRestorePolicyName(serviceName string) string {
	return fmt.Sprintf("databasebackuprestore-%s", serviceName) // no hyphen on purpose to avoid name collisions
}

// UserVaultPolicyName returns the vault policy name for the user
func UserVaultPolicyName(username string) string {
	return fmt.Sprintf("user-%s", username)
}

// ServiceSecretNames returns the names of the secrets for the service
func ServiceSecretNames(serviceName string) []string {
	return append(
		ServiceManagedSecretsNames(serviceName),
		SecretEtcdClientTLS,
		SecretExternalWildcard,
		SecretEventCollectorConfig,
		WardenCertsSecretName(serviceName),
		SecretWardenProxyTLS,
	)
}

// ServiceManagedSecretsNames returns the names of the secrets that are managed on a per service basis
func ServiceManagedSecretsNames(serviceName string) []string {
	return []string{
		ServiceEnvVarsSecretName(serviceName),
		ServiceFilesSecretName(serviceName),
		ServiceCertsSecretName(serviceName),
	}
}

// ProjectManagedSecretsNames returns the names of the secrets that are managed on a per project basis
func ProjectManagedSecretsNames(projectName string) []string {
	return []string{
		ProjectEnvVarsSecretName(projectName),
		ProjectFilesSecretName(projectName),
	}
}

// WardenCertsSecretName returns the name of the secret containing a service's Warden secrets
func WardenCertsSecretName(serviceName string) string {
	return fmt.Sprintf("%s-warden-tls", serviceName)
}

// ServiceStatusConfigMapName returns the name of the configmap which we write service status to
func ServiceStatusConfigMapName(serviceName string) string {
	return fmt.Sprintf("%s-status", serviceName)
}

func envVarFromSecret(name, secretName, secretKey string, optional bool) v1.EnvVar {
	return v1.EnvVar{
		Name: name,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{Name: secretName},
				Key:                  secretKey,
				Optional:             &optional,
			},
		},
	}
}

// EnvVarFromSecretOptional returns the optional env var referenced from secret.
func EnvVarFromSecretOptional(name, secretName, secretKey string) v1.EnvVar {
	return envVarFromSecret(name, secretName, secretKey, true)
}

// EnvVarFromSecret returns the env var referenced from secret.
func EnvVarFromSecret(name, secretName, secretKey string) v1.EnvVar {
	return envVarFromSecret(name, secretName, secretKey, false)
}

// EnvVarFromPod returns the env var referenced from one of the pod fields.
func EnvVarFromPod(name, fieldPath string) v1.EnvVar {
	return v1.EnvVar{
		Name: name,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: fieldPath,
			},
		},
	}
}

// VolumeProjectFilesName returns the name for the project files volume
func VolumeProjectFilesName(project string) string {
	return fmt.Sprintf("%s-%s", VolumePlaintextSecretFiles, project)
}

// IsNotFoundError returns if an error is a kubernetes not found error.
func IsNotFoundError(err error) bool {
	return kubeErrors.IsNotFound(core.ExceptionUnwrap(err))
}

// IgnoreNotFound turns not found errors into nil
func IgnoreNotFound(err error) error {
	if err != nil && !IsNotFoundError(err) {
		return err
	}
	return nil
}

// IgnoreConflict ignores the conflict error
func IgnoreConflict(err error) error {
	if err != nil && !kubeErrors.IsConflict(err) {
		return err
	}
	return nil
}

// ClusterEnvProvider is a type that provides a service environment
type ClusterEnvProvider interface {
	IsMinikubeEnv() bool
}

// GetMasterAffinity gets the affinity for running on master nodes some more nodejs level bs
func GetMasterAffinity(config ClusterEnvProvider) *v1.Affinity {
	if config.IsMinikubeEnv() {
		return nil
	}
	return &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					v1.NodeSelectorTerm{
						MatchExpressions: []v1.NodeSelectorRequirement{
							v1.NodeSelectorRequirement{
								Key:      LabelNodeLabelRole,
								Operator: v1.NodeSelectorOpIn,
								Values: []string{
									LabelNodeLabelRoleMaster,
								},
							},
						},
					},
				},
			},
		},
	}
}

// ServiceConfig is a type that describes a service
type ServiceConfig interface {
	DoesDockerBuild() bool
	GetServiceName(...string) string
}

// GetDeployAffinity gets the affinity for running on the previously deployed node and builder nodes
func GetDeployAffinity(config ServiceConfig) *v1.Affinity {
	if !config.DoesDockerBuild() {
		return nil
	}
	return &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
				v1.PreferredSchedulingTerm{
					Weight: int32(100),
					Preference: v1.NodeSelectorTerm{
						MatchExpressions: []v1.NodeSelectorRequirement{
							v1.NodeSelectorRequirement{
								Key:      ServiceDeployLabel(config.GetServiceName()),
								Operator: v1.NodeSelectorOpExists,
							},
						},
					},
				},
			},
		},
	}
}

// GetBuilderNodeSelector gets the selector necessary for running on a builder node
func GetBuilderNodeSelector(config ClusterEnvProvider) map[string]string {
	if config.IsMinikubeEnv() {
		return nil
	}
	return map[string]string{
		LabelNodeAutoscalingRole: LabelNodeAutoscalingRoleBuilderNode,
	}
}

// GetBuilderNodeTolerations gets the tolerations necessary for running on a builder node
func GetBuilderNodeTolerations() []v1.Toleration {
	return []v1.Toleration{
		v1.Toleration{
			Key:      LabelBlendNodeRoleBuilder,
			Operator: v1.TolerationOpExists,
		},
	}
}

// GetBuilderNodeTaint gets the taint for a builder node
func GetBuilderNodeTaint() v1.Taint {
	return v1.Taint{
		Key:    LabelBlendNodeRoleBuilder,
		Effect: v1.TaintEffectNoSchedule,
	}
}

// GetCoreNodeSelector gets the selector necessary for running on a builder node
func GetCoreNodeSelector(config ClusterEnvProvider) map[string]string {
	if config.IsMinikubeEnv() {
		return nil
	}
	return map[string]string{
		LabelNodeAutoscalingRole: LabelNodeAutoscalingRoleCoreNode,
	}
}

// GetCoreNodeTolerations gets the tolerations necessary for running on a builder node
func GetCoreNodeTolerations() []v1.Toleration {
	return []v1.Toleration{
		v1.Toleration{
			Key:      LabelBlendNodeRoleCore,
			Operator: v1.TolerationOpExists,
		},
	}
}

// GetCoreNodeTaint gets the taint for a builder node
func GetCoreNodeTaint() v1.Taint {
	return v1.Taint{
		Key:    LabelBlendNodeRoleCore,
		Effect: v1.TaintEffectNoSchedule,
	}
}

// GetMasterSelector gets the selector necessary for running on a master
func GetMasterSelector(config ClusterEnvProvider) map[string]string {
	if config.IsMinikubeEnv() {
		return nil
	}
	return map[string]string{
		LabelNodeAutoscalingRole: LabelNodeAutoscalingRoleMaster,
	}
}

// GetMasterTolerations gets the tolerations necessary for running on a master
func GetMasterTolerations(config ClusterEnvProvider) []v1.Toleration {
	if config.IsMinikubeEnv() {
		return nil
	}
	return []v1.Toleration{
		v1.Toleration{
			Key:    LabelNodeRoleMaster,
			Effect: v1.TaintEffectNoSchedule,
		},
	}
}

// GetAllTolerations gets the tolerations for running anywhere
func GetAllTolerations(config ClusterEnvProvider) []v1.Toleration {
	if config.IsMinikubeEnv() {
		return nil
	}
	return append(GetMasterTolerations(config), append(GetBuilderNodeTolerations(), GetCoreNodeTolerations()...)...)
}

// GetColocationAntiAffinity gets the anti-affinity for spreading the pods across different nodes and availability zones
func GetColocationAntiAffinity(labelSelector *metav1.LabelSelector) *v1.PodAntiAffinity {
	return &v1.PodAntiAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
			{
				Weight: int32(20), // attempt to spread to different azs
				PodAffinityTerm: v1.PodAffinityTerm{
					LabelSelector: labelSelector,
					TopologyKey:   apis.LabelZoneFailureDomain,
				},
			},
			{
				Weight: int32(80), // definitely not in the same node
				PodAffinityTerm: v1.PodAffinityTerm{
					LabelSelector: labelSelector,
					TopologyKey:   apis.LabelHostname,
				},
			},
		},
	}
}

// GetPodDNSConfig gets the default pod dns config (/etc/resolv.conf).
// Currently, we set `ndots: 2` as most of our in-cluster queries are `<name>.<namespace>`
func GetPodDNSConfig() *v1.PodDNSConfig {
	return &v1.PodDNSConfig{
		Options: []v1.PodDNSConfigOption{
			{
				Name:  "ndots",
				Value: ref.String("2"),
			},
		},
	}
}

// DockerImage returns the full reference for remote image
func DockerImage(host, image, tag string) string {
	return fmt.Sprintf("%s/%s:%s", host, image, tag)
}

// GetHostEtcdTLSConfig returns pertinent information for etcd tls config
func GetHostEtcdTLSConfig() *EtcdTLSConfig {
	return &EtcdTLSConfig{
		CA:   PathKubeCACert,
		Cert: fmt.Sprintf("%s/%s", PathHostTLSBasePath, FileEtcdClientCert),
		Key:  fmt.Sprintf("%s/%s", PathHostTLSBasePath, FileEtcdClientKey),
	}
}

// EtcdTLSVolume is the volume for etcd tls secrets
func EtcdTLSVolume() v1.Volume {
	return v1.Volume{
		Name: VolumeEtcdTLS,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: SecretEtcdClientTLS,
			},
		},
	}
}

// EtcdTLSVolumeMount is a volume mount for etcd tls secrets
func EtcdTLSVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      VolumeEtcdTLS,
		MountPath: PathHostTLSBasePath,
		ReadOnly:  true,
	}
}

// EtcdTLSSecrets returns the secret for etcd client tls
func EtcdTLSSecrets(namespace string) v1.ObjectReference {
	return v1.ObjectReference{
		Kind:      KindSecret,
		Namespace: namespace,
		Name:      SecretEtcdClientTLS,
	}
}

// ImagePullSecrets gets the image pull secrets for the cluster registry
func ImagePullSecrets() []v1.LocalObjectReference {
	return []v1.LocalObjectReference{{Name: RegistryAuthSecretName}}
}

// ParseResourceList parses the map into a resource list for kube
func ParseResourceList(m map[v1.ResourceName]string) (v1.ResourceList, error) {
	r := v1.ResourceList{}
	var err error
	for name, strValue := range m {
		r[name], err = resource.ParseQuantity(strValue)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

// MustParseResourceList parses the resource list, if it fails then panic
func MustParseResourceList(m map[v1.ResourceName]string) v1.ResourceList {
	l, err := ParseResourceList(m)
	if err != nil {
		panic(err)
	}
	return l
}

// SecretFromLiterals creates a secret from the literals
func SecretFromLiterals(name string, literals map[string]string) *v1.Secret {
	secret := &v1.Secret{
		Type:       v1.SecretTypeOpaque,
		StringData: literals,
	}
	secret.Name = name
	return secret
}

// SecretFromBytes creates a secret from map[string][]byte
func SecretFromBytes(name string, data map[string][]byte) *v1.Secret {
	secret := &v1.Secret{
		Type: v1.SecretTypeOpaque,
		Data: data,
	}
	secret.Name = name
	return secret
}

// MinikubeStorageClass returns a storage class for minikube use (host path storage)
func MinikubeStorageClass(objectMeta metav1.ObjectMeta) *storagev1.StorageClass {
	objectMeta.Annotations = map[string]string{
		"storageclass.beta.kubernetes.io/is-default-class": "true",
	}
	objectMeta.Labels = map[string]string{
		"addonmanager.kubernetes.io/mode": "Reconcile",
	}
	return &storagev1.StorageClass{
		ObjectMeta:  objectMeta,
		Provisioner: "k8s.io/minikube-hostpath",
	}
}

// ContainerStarted returns true if container has started
func ContainerStarted(status v1.ContainerStatus) bool {
	return status.State.Running != nil || status.State.Terminated != nil
}

// SortSecretsByCreationTime sorts the kube secrets by creation time
func SortSecretsByCreationTime(slice []v1.Secret, reverse ...bool) {
	descend := false
	if len(reverse) > 0 {
		descend = reverse[0]
	}
	sort.Slice(slice, func(i, j int) bool {
		si := slice[i]
		sj := slice[j]

		iTime := si.GetCreationTimestamp().Time
		less := iTime.Before(sj.GetCreationTimestamp().Time)
		if descend {
			return !less
		}
		return less
	})
}

func mergedSecretData(secret *v1.Secret) map[string][]byte {
	if secret == nil {
		return nil
	}
	ret := map[string][]byte{}
	// copy Data
	for k, v := range secret.Data {
		ret[k] = v
	}
	// overwrite with StringData
	for k, v := range secret.StringData {
		ret[k] = []byte(v)
	}
	return ret
}

// NodeModifyFunc is a function type that modifies a node
type NodeModifyFunc func(node *v1.Node)

// PatchNode patches a node using a modify function
func PatchNode(client *Client, node *v1.Node, modify NodeModifyFunc) error {
	oldData, err := json.Marshal(node)
	if err != nil {
		return exception.New(err)
	}
	modify(node)
	newData, err := json.Marshal(node)
	if err != nil {
		return exception.New(err)
	}
	patchBytes, err := jsonpatch.CreateMergePatch(oldData, newData)
	if err != nil {
		return exception.New(err)
	}
	return client.PatchNode(node.Name, patchBytes)
}

// RetryOnConflict is a wrapper around `retry.RetryOnConflict` that unwraps exception.Ex error type from `fn` for conflict checking
func RetryOnConflict(fn func() error) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return core.ExceptionUnwrap(fn())
	})
	return exception.New(err)
}

// PreStopDelayHook is a prestop hook that sleep for the specified `delay` to delay SIGTERM from reaching the container
func PreStopDelayHook(delay time.Duration) *v1.Lifecycle {
	return &v1.Lifecycle{
		PreStop: &v1.Handler{
			Exec: &v1.ExecAction{
				Command: []string{
					"sleep",
					strconv.Itoa(int(delay.Seconds())),
				},
			},
		},
	}
}

// DatadogADAnnotationKey returns the annotation key based on the container name and config key for datadog autodiscovery
func DatadogADAnnotationKey(containerName string, key DatadogADConfigKey) string {
	return fmt.Sprintf("ad.datadoghq.com/%s.%s", containerName, key)
}

// MustAnnotateDatadogAD annotates an ObjectMeta with datadog autodiscovery config, panic on error
func MustAnnotateDatadogAD(objectMeta *metav1.ObjectMeta, containerName string, checkNames []string, instances []map[string]interface{}) {
	if len(checkNames) != len(instances) {
		panic(exception.New("Different numbers of check names and instances"))
	}
	if objectMeta.Annotations == nil {
		objectMeta.Annotations = make(map[string]string)
	}
	checkNamesJSON, err := json.Marshal(checkNames)
	if err != nil {
		panic(exception.New(err))
	}
	instancesJSON, err := json.Marshal(instances)
	if err != nil {
		panic(exception.New(err))
	}
	initConfigsJSON := fmt.Sprintf("[%s]", strings.Trim(strings.Repeat("{},", len(checkNames)), ","))
	objectMeta.Annotations[DatadogADAnnotationKey(containerName, DatadogADCheckNames)] = string(checkNamesJSON)
	objectMeta.Annotations[DatadogADAnnotationKey(containerName, DatadogADInstances)] = string(instancesJSON)
	objectMeta.Annotations[DatadogADAnnotationKey(containerName, DatadogADInitConfigs)] = initConfigsJSON
}

// AnnotateScalingDownDeployment annotates scaling-down deployment with the current number of replicas
func AnnotateScalingDownDeployment(d *appsv1beta1.Deployment) {
	if d.Annotations == nil {
		d.Annotations = make(map[string]string)
	}
	var replicas int32
	if d.Spec.Replicas != nil {
		replicas = *d.Spec.Replicas
	} else {
		replicas = d.Status.Replicas
	}
	d.Annotations[AnnotationScaleDownFrom] = strconv.FormatInt(int64(replicas), 10)
}

// CalculateMaxUnavailable returns a MaxUnavailable value for use in a pod disruption budget spec
func CalculateMaxUnavailable(replicas int) intstr.IntOrString {
	if replicas < 2 {
		return intstr.FromInt(replicas)
	}
	return intstr.FromString("50%") // allow half to be out of service
}

// SanitizeObjectMetaForCreate removes extra fields filled in by the api to allow an object to be created
func SanitizeObjectMetaForCreate(meta metav1.ObjectMeta) metav1.ObjectMeta {
	meta.ResourceVersion = ""
	meta.UID = ""
	meta.SelfLink = ""
	return meta
}

// CreateProjectSecrets creates the project secrets if they don't exist
func CreateProjectSecrets(project, namespace string, client *Client) error {
	return createBlendServiceSecrets(client, namespace, ProjectManagedSecretsNames(project)...)
}

// CreateServiceSecrets creates the service secrets if they don't exist
func CreateServiceSecrets(service, namespace string, client *Client) error {
	return createBlendServiceSecrets(client, namespace, ServiceManagedSecretsNames(service)...)
}

func createBlendServiceSecrets(client *Client, namespace string, names ...string) error {
	for _, name := range names {
		err := client.CreateSecretIfNotExists(&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels: map[string]string{
					LabelServiceSecret: name,
				},
			},
			Type: v1.SecretTypeOpaque,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// PodIsEvicted returns true if the reported pod status is due to an eviction.
// source from https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/eviction/helpers.go#L1016
func PodIsEvicted(podStatus v1.PodStatus) bool {
	return podStatus.Phase == v1.PodFailed && podStatus.Reason == PodStatusReasonEvicted
}

// GenerateObjectFromTemplate returns a runtime object parsed from a template
func GenerateObjectFromTemplate(path string, vars map[string]interface{}) (runtime.Object, *schema.GroupVersionKind, error) {
	processedTemplate, err := core.Template(path, vars)
	if err != nil {
		return nil, nil, exception.New(err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, groupVersionKind, err := decode(processedTemplate.Bytes(), nil, nil)
	if err != nil {
		return nil, nil, exception.New(err).WithMessagef("Error while parsing k8s yaml spec (%s)", path)
	}

	return obj, groupVersionKind, nil
}
