package kube

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"
	"time"

	"git.blendlabs.com/blend/deployinator/pkg/core"
	"git.blendlabs.com/blend/deployinator/pkg/kube/selector"
	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/ref"
	"github.com/blend/go-sdk/uuid"
	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	certmanagerclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	autov2beta1 "k8s.io/api/autoscaling/v2beta1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	certsv1beta1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubetypes "k8s.io/apimachinery/pkg/types"
	kuberand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	apiregistrationv1beta1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1beta1"
	kubeAggregator "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	"k8s.io/kubernetes/pkg/serviceaccount"
)

var (
	defaultLock sync.Mutex
	defaultKube *Client
)

// Default returns the default kube client.
func Default() *Client {
	return defaultKube
}

// InitDefault init's the default kube client.
func InitDefault() error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	SetDefault(client)
	return nil
}

// SetDefault sets the default kube client
func SetDefault(client *Client) {
	defaultLock.Lock()
	defer defaultLock.Unlock()
	defaultKube = client
}

// UseContext sets current context of the kubernetes api config at `configPath` to `contextName`.
// This is based on https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/config/use_context.go#L64
func UseContext(configPath, contextName string) error {
	pathOptions := &clientcmd.ClientConfigLoadingRules{ExplicitPath: configPath}
	config, err := pathOptions.GetStartingConfig()
	if err != nil {
		return exception.New(err)
	}
	config.CurrentContext = contextName
	return exception.New(clientcmd.ModifyConfig(pathOptions, *config, true))
}

// getConfig returns a config
func getConfig() (*rest.Config, error) {
	if !env.Env().Has(EnvVarKubectlConfig) {
		return nil, exception.New(fmt.Sprintf("%s not set", EnvVarKubectlConfig))
	}
	configPath := env.Env().String(EnvVarKubectlConfig)
	core.NewLoggerFromEnvOrAll().Debugf("kubectl using config location: %s", configPath)
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	return config, exception.New(err)
}

// NewClient creates a new client.
func NewClient() (*Client, error) {
	var config *rest.Config
	var err error
	if env.Env().Has(EnvVarKubectlConfig) {
		config, err = getConfig()
		if err != nil {
			return nil, exception.New(err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, exception.New(err)
		}
	}

	// tune the config
	config.QPS = float32(env.Env().MustFloat64(EnvVarClientQPS, 50))
	config.Burst = env.Env().MustInt(EnvVarClientBurst, 100)

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, exception.New(err)
	}

	// aggregator clientset
	aggregatorClientset, err := kubeAggregator.NewForConfig(config)
	if err != nil {
		return nil, exception.New(err)
	}

	// apiextensions clientset
	apiextensionsClientset, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, exception.New(err)
	}

	// cert-manager clientset
	certmanagerClientset, err := certmanagerclient.NewForConfig(config)
	if err != nil {
		return nil, exception.New(err)
	}

	return &Client{
		clientset:              clientset,
		aggregatorClientset:    aggregatorClientset,
		apiextensionsClientset: apiextensionsClientset,
		certmanagerClientset:   certmanagerClientset,
		config:                 config,
	}, nil
}

// Client is a wrapper for the kube clientset.
type Client struct {
	clientset              kubernetes.Interface
	aggregatorClientset    kubeAggregator.Interface
	apiextensionsClientset apiextensionsclient.Interface
	certmanagerClientset   certmanagerclient.Interface
	config                 *rest.Config
}

// Clientset returns the underlying clientset.
func (c *Client) Clientset() kubernetes.Interface {
	return c.clientset
}

// Config returns the config
func (c *Client) Config() *rest.Config {
	return c.config
}

// --------------------------------------------------------------------------------
// Get Functions
// --------------------------------------------------------------------------------

// GetDeployment returns a kube deployment.
func (c *Client) GetDeployment(name string, namespace ...string) (*appsv1beta1.Deployment, error) {
	return c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetStatefulSet returns a kube StatefulSet.
func (c *Client) GetStatefulSet(name string, namespace ...string) (*appsv1beta1.StatefulSet, error) {
	return c.clientset.AppsV1beta1().StatefulSets(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetPodDisruptionBudget returns a kube PodDisruptionBudget.
func (c *Client) GetPodDisruptionBudget(name string, namespace ...string) (*policyv1beta1.PodDisruptionBudget, error) {
	return c.clientset.PolicyV1beta1().PodDisruptionBudgets(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetPodDisruptionBudgetsWithSelector returns kube PodDisruptionBudgets with the selector.
func (c *Client) GetPodDisruptionBudgetsWithSelector(selector string, namespace ...string) (*policyv1beta1.PodDisruptionBudgetList, error) {
	return c.clientset.PolicyV1beta1().PodDisruptionBudgets(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetDeploymentsWithSelector gets deployments with a given label selector.
func (c *Client) GetDeploymentsWithSelector(selector string, namespace ...string) (*appsv1beta1.DeploymentList, error) {
	return c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// WatchDeploymentsWithSelector watch pods with a given selector.
func (c *Client) WatchDeploymentsWithSelector(selector string, namespace ...string) (watch.Interface, error) {
	return c.WatchDeploymentsWithSelectorResourceVersion(selector, "", namespace...)
}

// WatchDeploymentsWithSelectorResourceVersion watch pods with a given selector.
func (c *Client) WatchDeploymentsWithSelectorResourceVersion(selector string, resourceVersion string, namespace ...string) (watch.Interface, error) {
	return c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Watch(apismetav1.ListOptions{
		LabelSelector:        selector,
		IncludeUninitialized: true,
		Watch:                true,
		ResourceVersion:      resourceVersion,
	})
}

// GetStatefulSetsWithSelector gets deployments with a given label selector.
func (c *Client) GetStatefulSetsWithSelector(selector string, namespace ...string) (*appsv1beta1.StatefulSetList, error) {
	return c.clientset.AppsV1beta1().StatefulSets(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetAutoscaler returns an auto scaler.
func (c *Client) GetAutoscaler(name string, namespace ...string) (*autov2beta1.HorizontalPodAutoscaler, error) {
	return c.clientset.AutoscalingV2beta1().HorizontalPodAutoscalers(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetAutoscalerWithSelector returns an auto scaler with a given label selector.
func (c *Client) GetAutoscalerWithSelector(selector string, namespace ...string) (*autov2beta1.HorizontalPodAutoscalerList, error) {
	return c.clientset.AutoscalingV2beta1().HorizontalPodAutoscalers(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetService returns a kube service.
func (c *Client) GetService(name string, namespace ...string) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetAPIService returns a kube APIService.
func (c *Client) GetAPIService(name string) (*apiregistrationv1beta1.APIService, error) {
	return c.aggregatorClientset.ApiregistrationV1beta1().APIServices().Get(name, apismetav1.GetOptions{})
}

// GetValidatingWebhookConfiguration returns a kube ValidatingWebhookConfiguration.
func (c *Client) GetValidatingWebhookConfiguration(name string) (*admissionregistrationv1beta1.ValidatingWebhookConfiguration, error) {
	return c.clientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(name, apismetav1.GetOptions{})
}

// GetNamespace returns a kube Namespace.
func (c *Client) GetNamespace(name string) (*corev1.Namespace, error) {
	return c.clientset.CoreV1().Namespaces().Get(name, apismetav1.GetOptions{})
}

// GetNamespacesWithSelector gets the namespaces with selector
func (c *Client) GetNamespacesWithSelector(selector string) (*corev1.NamespaceList, error) {
	return c.clientset.CoreV1().Namespaces().List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetServiceAccount returns a kube service account.
func (c *Client) GetServiceAccount(name string, namespace ...string) (*corev1.ServiceAccount, error) {
	return c.clientset.CoreV1().ServiceAccounts(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetServiceAccountsWithSelector gets services with a given selector.
func (c *Client) GetServiceAccountsWithSelector(selector string, namespace ...string) (*corev1.ServiceAccountList, error) {
	return c.clientset.CoreV1().ServiceAccounts(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetServicesWithSelector gets services with a given selector.
func (c *Client) GetServicesWithSelector(selector string, namespace ...string) (*corev1.ServiceList, error) {
	return c.clientset.CoreV1().Services(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetIngress returns a kube ingress.
func (c *Client) GetIngress(name string, namespace ...string) (*extv1beta1.Ingress, error) {
	return c.clientset.ExtensionsV1beta1().Ingresses(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetIngressesWithSelector gets ingresses with a given selector.
func (c *Client) GetIngressesWithSelector(selector string, namespace ...string) (*extv1beta1.IngressList, error) {
	return c.clientset.ExtensionsV1beta1().Ingresses(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetPod gets a pod with a given name.
func (c *Client) GetPod(name string, namespace ...string) (*corev1.Pod, error) {
	return c.clientset.CoreV1().Pods(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetPodsWithSelector gets pods with a given selector.
func (c *Client) GetPodsWithSelector(selector string, namespace ...string) (*corev1.PodList, error) {
	return c.clientset.CoreV1().Pods(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// WatchPodsWithSelector watch pods with a given selector.
func (c *Client) WatchPodsWithSelector(selector string, namespace ...string) (watch.Interface, error) {
	return c.WatchPodsWithSelectorResourceVersion(selector, "", namespace...)
}

// WatchPodsWithSelectorResourceVersion watch pods with a given selector.
func (c *Client) WatchPodsWithSelectorResourceVersion(selector string, resourceVersion string, namespace ...string) (watch.Interface, error) {
	return c.clientset.CoreV1().Pods(GetNamespace(namespace...)).Watch(apismetav1.ListOptions{
		LabelSelector:        selector,
		IncludeUninitialized: true,
		Watch:                true,
		ResourceVersion:      resourceVersion,
	})
}

// GetReplicaSet gets a replica set with a given name.
func (c *Client) GetReplicaSet(name string, namespace ...string) (*extv1beta1.ReplicaSet, error) {
	return c.clientset.ExtensionsV1beta1().ReplicaSets(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetReplicaSetsWithSelector returns the replica sets that match a selector.
func (c *Client) GetReplicaSetsWithSelector(selector string, namespace ...string) (*extv1beta1.ReplicaSetList, error) {
	return c.clientset.ExtensionsV1beta1().ReplicaSets(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetSecret gets a secret.
func (c *Client) GetSecret(name string, namespace ...string) (*corev1.Secret, error) {
	secret, err := c.clientset.CoreV1().Secrets(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
	return secret, exception.New(err)
}

// GetSecretsWithSelector gets the secrets with the given selector
func (c *Client) GetSecretsWithSelector(selector string, namespace ...string) (*corev1.SecretList, error) {
	return c.clientset.CoreV1().Secrets(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetRole gets a role with a given name.
func (c *Client) GetRole(name string, namespace ...string) (*rbacv1.Role, error) {
	return c.clientset.RbacV1().Roles(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetRoleBinding gets a RoleBinding with a given name.
func (c *Client) GetRoleBinding(name string, namespace ...string) (*rbacv1.RoleBinding, error) {
	return c.clientset.RbacV1().RoleBindings(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetClusterRole gets the cluster role
func (c *Client) GetClusterRole(name string) (*rbacv1.ClusterRole, error) {
	return c.clientset.RbacV1().ClusterRoles().Get(name, apismetav1.GetOptions{})
}

// GetClusterRoleBinding gets a RoleBinding with a given name.
func (c *Client) GetClusterRoleBinding(name string) (*rbacv1.ClusterRoleBinding, error) {
	return c.clientset.RbacV1().ClusterRoleBindings().Get(name, apismetav1.GetOptions{})
}

// GetRolesWithSelector gets a role with a given selector.
func (c *Client) GetRolesWithSelector(selector string, namespace ...string) (*rbacv1.RoleList, error) {
	return c.clientset.RbacV1().Roles(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetLatestPodLogs gets the latest logs from the pod, since the last call to this function
func (c *Client) GetLatestPodLogs(pod, container, namespace string, since *time.Time) (io.ReadCloser, error) {
	var sinceTime *apismetav1.Time
	if since != nil {
		sinceTime = &apismetav1.Time{Time: *since}
	}
	req := c.Clientset().CoreV1().Pods(GetNamespace(namespace)).GetLogs(pod, &corev1.PodLogOptions{
		Container: container,
		Follow:    true,
		SinceTime: sinceTime,
	})
	return req.Stream()
}

// GetDaemonSet gets a daemonset with a given name.
func (c *Client) GetDaemonSet(name string, namespace ...string) (*extv1beta1.DaemonSet, error) {
	return c.clientset.ExtensionsV1beta1().DaemonSets(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetDaemonSetsWithSelector gets daemonsets with selector.
func (c *Client) GetDaemonSetsWithSelector(selector string, namespace ...string) (*extv1beta1.DaemonSetList, error) {
	return c.clientset.ExtensionsV1beta1().DaemonSets(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetPersistentVolume returns the persistent volume with the given name
func (c *Client) GetPersistentVolume(name string) (*corev1.PersistentVolume, error) {
	return c.clientset.CoreV1().PersistentVolumes().Get(name, apismetav1.GetOptions{})
}

// GetPersistentVolumeClaim gets the volume claim with the name
func (c *Client) GetPersistentVolumeClaim(name string, namespace ...string) (*corev1.PersistentVolumeClaim, error) {
	return c.clientset.CoreV1().PersistentVolumeClaims(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetPersistentVolumeClaimsWithSelector gets volume claims by label selector
func (c *Client) GetPersistentVolumeClaimsWithSelector(selector string, namespace ...string) (*corev1.PersistentVolumeClaimList, error) {
	return c.clientset.CoreV1().PersistentVolumeClaims(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetStorageClass gets the storage class by name
func (c *Client) GetStorageClass(name string) (*storagev1.StorageClass, error) {
	return c.clientset.StorageV1().StorageClasses().Get(name, apismetav1.GetOptions{})
}

// GetCertificateSigningRequest gets a csr
func (c *Client) GetCertificateSigningRequest(name string) (*certsv1beta1.CertificateSigningRequest, error) {
	return c.clientset.Certificates().CertificateSigningRequests().Get(name, apismetav1.GetOptions{})
}

// GetConfigMap gets the config map
func (c *Client) GetConfigMap(name string, namespace ...string) (*corev1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetCronJob gets the cron job
func (c *Client) GetCronJob(name string, namespace ...string) (*batchv1beta1.CronJob, error) {
	return c.clientset.BatchV1beta1().CronJobs(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetJob gets the job
func (c *Client) GetJob(name string, namespace ...string) (*batchv1.Job, error) {
	return c.clientset.BatchV1().Jobs(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetJobsWithSelector gets the jobs
func (c *Client) GetJobsWithSelector(selector string, namespace ...string) (*batchv1.JobList, error) {
	return c.clientset.BatchV1().Jobs(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// WatchJobsWithSelector watch pods with a given selector.
func (c *Client) WatchJobsWithSelector(selector string, namespace ...string) (watch.Interface, error) {
	return c.WatchJobsWithSelectorResourceVersion(selector, "", namespace...)
}

// WatchJobsWithSelectorResourceVersion watch pods with a given selector.
func (c *Client) WatchJobsWithSelectorResourceVersion(selector string, resourceVersion string, namespace ...string) (watch.Interface, error) {
	return c.clientset.BatchV1().Jobs(GetNamespace(namespace...)).Watch(apismetav1.ListOptions{
		LabelSelector:        selector,
		IncludeUninitialized: true,
		Watch:                true,
		ResourceVersion:      resourceVersion,
	})
}

// GetCronJobsWithSelector gets the cron jobs
func (c *Client) GetCronJobsWithSelector(selector string, namespace ...string) (*batchv1beta1.CronJobList, error) {
	return c.clientset.BatchV1beta1().CronJobs(GetNamespace(namespace...)).List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetNode gets the node
func (c *Client) GetNode(name string) (*corev1.Node, error) {
	return c.clientset.CoreV1().Nodes().Get(name, apismetav1.GetOptions{})
}

// GetNodesWithSelector gets the nodes
func (c *Client) GetNodesWithSelector(selector string) (*corev1.NodeList, error) {
	return c.clientset.CoreV1().Nodes().List(apismetav1.ListOptions{
		LabelSelector: selector,
	})
}

// GetCustomResourceDefinition gets the crd
func (c *Client) GetCustomResourceDefinition(name string) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	return c.apiextensionsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, apismetav1.GetOptions{})
}

// GetIssuer gets the cluster issuer
func (c *Client) GetIssuer(name string, namespace ...string) (*certmanagerv1alpha1.Issuer, error) {
	return c.certmanagerClientset.CertmanagerV1alpha1().Issuers(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// GetClusterIssuer gets the cluster issuer
func (c *Client) GetClusterIssuer(name string) (*certmanagerv1alpha1.ClusterIssuer, error) {
	return c.certmanagerClientset.CertmanagerV1alpha1().ClusterIssuers().Get(name, apismetav1.GetOptions{})
}

// GetCertificate gets the certificate
func (c *Client) GetCertificate(name string, namespace ...string) (*certmanagerv1alpha1.Certificate, error) {
	return c.certmanagerClientset.CertmanagerV1alpha1().Certificates(GetNamespace(namespace...)).Get(name, apismetav1.GetOptions{})
}

// --------------------------------------------------------------------------------
// Exists Functions
// --------------------------------------------------------------------------------

// DeploymentExists returns if a resource exists.
func (c *Client) DeploymentExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetDeployment(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// DeploymentWithSelectorExists returns if a resource exists.
func (c *Client) DeploymentWithSelectorExists(selector string, namespace ...string) (bool, error) {
	deploymentList, err := c.GetDeploymentsWithSelector(selector, namespace...)
	if err != nil {
		return false, exception.New(err)
	}
	return len(deploymentList.Items) > 0, nil
}

// StatefulSetExists returns if a resource exists.
func (c *Client) StatefulSetExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetStatefulSet(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// PodDisruptionBudgetExists returns if a resource exists.
func (c *Client) PodDisruptionBudgetExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetPodDisruptionBudget(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// AutoscalerExists returns if a resource exists.
func (c *Client) AutoscalerExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetAutoscaler(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// ServiceExists returns if a resource exists.
func (c *Client) ServiceExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetService(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// APIServiceExists returns if an apiservice exists
func (c *Client) APIServiceExists(name string) (bool, error) {
	_, err := c.GetAPIService(name)
	if err != nil && !IsNotFoundError(err) {
		return false, exception.New(err)
	}
	return err == nil, nil
}

// ValidatingWebhookConfigurationExists returns if an apiservice exists
func (c *Client) ValidatingWebhookConfigurationExists(name string) (bool, error) {
	_, err := c.GetValidatingWebhookConfiguration(name)
	if err != nil && !IsNotFoundError(err) {
		return false, exception.New(err)
	}
	return err == nil, nil
}

// NamespaceExists returns if a resource exists.
func (c *Client) NamespaceExists(name string) (bool, error) {
	_, err := c.GetNamespace(name)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// ServiceAccountExists returns if a resource exists.
func (c *Client) ServiceAccountExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetServiceAccount(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// IngressExists returns if a resource exists.
func (c *Client) IngressExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetIngress(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// ReplicaSetExists returns if a resource exists.
func (c *Client) ReplicaSetExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetReplicaSet(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// PodExists returns if a resource exists.
func (c *Client) PodExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetPod(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// SecretExists returns if a resource exists.
func (c *Client) SecretExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetSecret(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// DaemonSetExists returns if a resource exists.
func (c *Client) DaemonSetExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetDaemonSet(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// DaemonSetWithSelectorExists returns if a resource exists.
func (c *Client) DaemonSetWithSelectorExists(selector string, namespace ...string) (bool, error) {
	dsList, err := c.GetDaemonSetsWithSelector(selector, namespace...)
	if err != nil {
		return false, exception.New(err)
	}
	return len(dsList.Items) > 0, nil
}

// RoleBindingExists returns if a resource exists.
func (c *Client) RoleBindingExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetRoleBinding(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// ClusterRoleBindingExists returns if a resource exists.
func (c *Client) ClusterRoleBindingExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetClusterRoleBinding(name)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// RoleExists returns if a resource exists.
func (c *Client) RoleExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetRole(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// ClusterRoleExists returns if a resource exists.
func (c *Client) ClusterRoleExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetClusterRole(name)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// StorageClassExists returns if a StorageClassExists
func (c *Client) StorageClassExists(name string) (bool, error) {
	_, err := c.GetStorageClass(name)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// PersistentVolumeClaimExists returns if the pvc exists
func (c *Client) PersistentVolumeClaimExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetPersistentVolumeClaim(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// CronJobExists returns if the cron job exists
func (c *Client) CronJobExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetCronJob(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, exception.New(err)
	}
	return err == nil, nil
}

// JobExists returns if the job exists
func (c *Client) JobExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetJob(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, exception.New(err)
	}
	return err == nil, nil
}

// CustomResourceDefinitionExists returns if a crd exists.
func (c *Client) CustomResourceDefinitionExists(name string) (bool, error) {
	_, err := c.GetCustomResourceDefinition(name)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// IssuerExists returns if a cluster issuer exists.
func (c *Client) IssuerExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetIssuer(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// ClusterIssuerExists returns if a cluster issuer exists.
func (c *Client) ClusterIssuerExists(name string) (bool, error) {
	_, err := c.GetClusterIssuer(name)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// CertificateExists returns if a certificate exists.
func (c *Client) CertificateExists(name string, namespace ...string) (bool, error) {
	_, err := c.GetCertificate(name, namespace...)
	if err != nil && !IsNotFoundError(err) {
		return false, err
	}
	return err == nil, nil
}

// --------------------------------------------------------------------------------
// Create Functions
// --------------------------------------------------------------------------------

// CreateDeployment creates a kube deployment.
func (c *Client) CreateDeployment(dep *appsv1beta1.Deployment, namespace ...string) error {
	_, err := c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Create(dep)
	return exception.New(err)
}

// CreateStatefulSet creates a kube StatefulSet.
func (c *Client) CreateStatefulSet(ss *appsv1beta1.StatefulSet) error {
	_, err := c.clientset.AppsV1beta1().StatefulSets(ss.Namespace).Create(ss)
	return exception.New(err)
}

// CreatePodDisruptionBudget create a pdb
func (c *Client) CreatePodDisruptionBudget(pdb *policyv1beta1.PodDisruptionBudget) error {
	_, err := c.clientset.PolicyV1beta1().PodDisruptionBudgets(pdb.Namespace).Create(pdb)
	return exception.New(err)
}

// CreateAutoscaler creates a kube horizontal pod autoscaler.
func (c *Client) CreateAutoscaler(autoscaler *autov2beta1.HorizontalPodAutoscaler, namespace ...string) error {
	_, err := c.clientset.AutoscalingV2beta1().HorizontalPodAutoscalers(GetNamespace(namespace...)).Create(autoscaler)
	return exception.New(err)
}

// CreateService creates a kube service.
func (c *Client) CreateService(svc *corev1.Service, namespace ...string) error {
	_, err := c.clientset.CoreV1().Services(GetNamespace(namespace...)).Create(svc)
	return exception.New(err)
}

// CreateAPIService creates a kube APIService.
func (c *Client) CreateAPIService(svc *apiregistrationv1beta1.APIService) error {
	_, err := c.aggregatorClientset.ApiregistrationV1beta1().APIServices().Create(svc)
	return exception.New(err)
}

// CreateValidatingWebhookConfiguration creates a validating webhook configuration
func (c *Client) CreateValidatingWebhookConfiguration(v *admissionregistrationv1beta1.ValidatingWebhookConfiguration) error {
	_, err := c.clientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Create(v)
	return exception.New(err)
}

// CreateNamespace creates a kube Namespace.
func (c *Client) CreateNamespace(ns *corev1.Namespace) error {
	_, err := c.clientset.CoreV1().Namespaces().Create(ns)
	return exception.New(err)
}

// CreateServiceAccount creates a kube service account.
func (c *Client) CreateServiceAccount(sa *corev1.ServiceAccount) error {
	_, err := c.clientset.CoreV1().ServiceAccounts(sa.Namespace).Create(sa)
	return exception.New(err)
}

// CreateIngress creates a kube ingress.
func (c *Client) CreateIngress(ing *extv1beta1.Ingress, namespace ...string) error {
	_, err := c.clientset.ExtensionsV1beta1().Ingresses(GetNamespace(namespace...)).Create(ing)
	return exception.New(err)
}

// CreateReplicaSet creates a kube replica set.
func (c *Client) CreateReplicaSet(rs *extv1beta1.ReplicaSet, namespace ...string) error {
	_, err := c.clientset.ExtensionsV1beta1().ReplicaSets(GetNamespace(namespace...)).Create(rs)
	return exception.New(err)
}

// CreatePod creates a kube pod.
func (c *Client) CreatePod(pod *corev1.Pod, namespace ...string) error {
	_, err := c.clientset.CoreV1().Pods(GetNamespace(namespace...)).Create(pod)
	return exception.New(err)
}

// CreateSecret creates a given secret.
func (c *Client) CreateSecret(secret *corev1.Secret, namespace ...string) error {
	_, err := c.clientset.CoreV1().Secrets(GetNamespace(namespace...)).Create(secret)
	return exception.New(err)
}

// CreateSecretIfNotExists creates the secret if it doesn't already exists
func (c *Client) CreateSecretIfNotExists(secret *corev1.Secret, namespace ...string) error {
	exists, err := c.SecretExists(secret.Name, namespace...)
	if err != nil {
		return exception.New(err)
	}
	if !exists {
		return c.CreateSecret(secret, namespace...)
	}
	return nil
}

// CreateClusterRole creates the cluster role
func (c *Client) CreateClusterRole(clusterRole *rbacv1.ClusterRole) error {
	_, err := c.clientset.RbacV1().ClusterRoles().Create(clusterRole)
	return exception.New(err)
}

// CreateRole creates the role
func (c *Client) CreateRole(role *rbacv1.Role, namespace ...string) error {
	_, err := c.clientset.RbacV1().Roles(GetNamespace(namespace...)).Create(role)
	return exception.New(err)
}

// CreateClusterRoleBinding creates a cluster role binding.
func (c *Client) CreateClusterRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding) error {
	_, err := c.clientset.RbacV1().ClusterRoleBindings().Create(clusterRoleBinding)
	return exception.New(err)
}

// CreateRoleBinding creates a role binding.
func (c *Client) CreateRoleBinding(roleBinding *rbacv1.RoleBinding, namespace ...string) error {
	_, err := c.clientset.RbacV1().RoleBindings(GetNamespace(namespace...)).Create(roleBinding)
	return exception.New(err)
}

// CreateDaemonSet creates a kube daemonset.
func (c *Client) CreateDaemonSet(daemonSet *extv1beta1.DaemonSet, namespace ...string) error {
	_, err := c.clientset.ExtensionsV1beta1().DaemonSets(GetNamespace(namespace...)).Create(daemonSet)
	return exception.New(err)
}

// CreateConfigMap creates a kube configmap.
func (c *Client) CreateConfigMap(configMap *corev1.ConfigMap, namespace ...string) error {
	_, err := c.clientset.CoreV1().ConfigMaps(GetNamespace(namespace...)).Create(configMap)
	return exception.New(err)
}

// CreatePersistentVolume creates a persistent volume for storage
func (c *Client) CreatePersistentVolume(vol *corev1.PersistentVolume) error {
	_, err := c.clientset.CoreV1().PersistentVolumes().Create(vol)
	return exception.New(err)
}

// CreatePersistentVolumeClaim creates a persistent volume claim
func (c *Client) CreatePersistentVolumeClaim(claim *corev1.PersistentVolumeClaim, namespace ...string) error {
	_, err := c.clientset.CoreV1().PersistentVolumeClaims(GetNamespace(namespace...)).Create(claim)
	return exception.New(err)
}

// CreateStorageClass creates a storage class
func (c *Client) CreateStorageClass(class *storagev1.StorageClass) error {
	_, err := c.clientset.StorageV1().StorageClasses().Create(class)
	return exception.New(err)
}

// CreateCertificateSigningRequest creates a csr
func (c *Client) CreateCertificateSigningRequest(csr *certsv1beta1.CertificateSigningRequest) error {
	_, err := c.clientset.Certificates().CertificateSigningRequests().Create(csr)
	return exception.New(err)
}

// CreateCronJob creates the cron job
func (c *Client) CreateCronJob(job *batchv1beta1.CronJob, namespace ...string) error {
	_, err := c.clientset.BatchV1beta1().CronJobs(GetNamespace(namespace...)).Create(job)
	return exception.New(err)
}

// CreateJob creates a kube job.
func (c *Client) CreateJob(job *batchv1.Job, namespace ...string) (string, error) {
	kubeJob, err := c.clientset.BatchV1().Jobs(GetNamespace(namespace...)).Create(job)
	if kubeJob == nil {
		return "", exception.New(err)
	}
	return kubeJob.ResourceVersion, exception.New(err)
}

// CreateCustomResourceDefinition creates a crd
func (c *Client) CreateCustomResourceDefinition(crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	_, err := c.apiextensionsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	return exception.New(err)
}

// CreateIssuer creates an issuer
func (c *Client) CreateIssuer(issuer *certmanagerv1alpha1.Issuer, namespace ...string) error {
	_, err := c.certmanagerClientset.CertmanagerV1alpha1().Issuers(GetNamespace(namespace...)).Create(issuer)
	return exception.New(err)
}

// CreateClusterIssuer creates a cluster issuer
func (c *Client) CreateClusterIssuer(ci *certmanagerv1alpha1.ClusterIssuer) error {
	_, err := c.certmanagerClientset.CertmanagerV1alpha1().ClusterIssuers().Create(ci)
	return exception.New(err)
}

// CreateCertificate creates a certificate
func (c *Client) CreateCertificate(cert *certmanagerv1alpha1.Certificate, namespace ...string) error {
	_, err := c.certmanagerClientset.CertmanagerV1alpha1().Certificates(GetNamespace(namespace...)).Create(cert)
	return exception.New(err)
}

// --------------------------------------------------------------------------------
// Update Functions
// --------------------------------------------------------------------------------

// UpdateDeployment updates a kube deployment.
func (c *Client) UpdateDeployment(dep *appsv1beta1.Deployment, namespace ...string) error {
	_, err := c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Update(dep)
	return exception.New(err)
}

// UpdateStatefulSet updates a kube StatefulSet.
func (c *Client) UpdateStatefulSet(ss *appsv1beta1.StatefulSet) error {
	if ss.Spec.Template.Annotations == nil {
		ss.Spec.Template.Annotations = map[string]string{}
	}
	now, err := time.Now().MarshalText()
	if err != nil {
		return exception.New(err)
	}
	ss.Spec.Template.Annotations["CreatedAt"] = string(now)
	_, err = c.clientset.AppsV1beta1().StatefulSets(ss.Namespace).Update(ss)
	return err
}

// UpdatePodDisruptionBudget updates a pdb.
func (c *Client) UpdatePodDisruptionBudget(pdb *policyv1beta1.PodDisruptionBudget) error {
	theirs, err := c.GetPodDisruptionBudget(pdb.Name, pdb.Namespace)
	if err != nil {
		return exception.New(err)
	}
	// Update stuff as needed
	theirs.Spec = pdb.Spec
	theirs.Annotations = pdb.Annotations
	_, err = c.clientset.PolicyV1beta1().PodDisruptionBudgets(pdb.Namespace).Update(theirs)
	return err
}

// UpdateDaemonSet updates a kube daemonset.
func (c *Client) UpdateDaemonSet(daemonSet *extv1beta1.DaemonSet, namespace ...string) error {
	_, err := c.clientset.ExtensionsV1beta1().DaemonSets(GetNamespace(namespace...)).Update(daemonSet)
	return err
}

// UpdateAutoscaler updates a kube horizontal autoscaler.
func (c *Client) UpdateAutoscaler(autoscaler *autov2beta1.HorizontalPodAutoscaler, namespace ...string) error {
	_, err := c.clientset.AutoscalingV2beta1().HorizontalPodAutoscalers(GetNamespace(namespace...)).Update(autoscaler)
	return err
}

// UpdateService updates a kube service.
func (c *Client) UpdateService(svc *corev1.Service, namespace ...string) error {
	_, err := c.clientset.CoreV1().Services(GetNamespace(namespace...)).Update(svc)
	return exception.New(err)
}

// mergeServicePorts merges existing and new service ports while trying to keep assigned node ports.
// this is to avoid any changes to the elb for the service, which could cause a brief period of
// elb to backend connection errors.
func (c *Client) mergeServicePorts(theirs, ours []corev1.ServicePort) []corev1.ServicePort {
	var merged []corev1.ServicePort

	// merge ports, trying to keep NodePort intact
	nodePorts := map[int32]int32{}
	for _, port := range theirs {
		nodePorts[port.Port] = port.NodePort
	}
	for _, port := range ours {
		theirNodePort, ok := nodePorts[port.Port]
		if port.NodePort == 0 && ok {
			port.NodePort = theirNodePort
		}
		merged = append(merged, port)
	}
	return merged
}

// UpdateServiceWithRetry gets the service and then selectively updates its fields.
func (c *Client) UpdateServiceWithRetry(ours *corev1.Service) error {
	return RetryOnConflict(func() error {
		theirs, err := c.GetService(ours.Name, ours.Namespace)
		if err != nil {
			return exception.New(err)
		}
		// TODO Keep adding in here fields you would like to change
		// We need to update theirs -- contraint on service updating
		theirs.Spec.Selector = ours.Spec.Selector
		theirs.Spec.Ports = c.mergeServicePorts(theirs.Spec.Ports, ours.Spec.Ports)
		theirs.Labels = ours.Labels
		theirs.Annotations = ours.Annotations
		theirs.Spec.SessionAffinity = ours.Spec.SessionAffinity
		theirs.Spec.Type = ours.Spec.Type
		theirs.Spec.LoadBalancerSourceRanges = ours.Spec.LoadBalancerSourceRanges
		return c.UpdateService(theirs, theirs.Namespace)
	})
}

// UpdateAPIService Updates a kube APIService.
func (c *Client) UpdateAPIService(svc *apiregistrationv1beta1.APIService) error {
	theirs, err := c.GetAPIService(svc.Name)
	if err != nil {
		return exception.New(err)
	}
	theirs.Spec = svc.Spec
	_, err = c.aggregatorClientset.ApiregistrationV1beta1().APIServices().Update(theirs)
	return exception.New(err)
}

// UpdateValidatingWebhookConfiguration updates a kube service.
func (c *Client) UpdateValidatingWebhookConfiguration(v *admissionregistrationv1beta1.ValidatingWebhookConfiguration) error {
	_, err := c.clientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(v)
	return exception.New(err)
}

// UpdateNamespace updates a kube Namespace.
func (c *Client) UpdateNamespace(ns *corev1.Namespace) error {
	_, err := c.clientset.CoreV1().Namespaces().Update(ns)
	return err
}

// UpdateServiceAccount updates a kube service account.
func (c *Client) UpdateServiceAccount(sa *corev1.ServiceAccount) error {
	_, err := c.clientset.CoreV1().ServiceAccounts(sa.Namespace).Update(sa)
	return err
}

// UpdateIngress updates a kube ingress.
func (c *Client) UpdateIngress(ing *extv1beta1.Ingress, namespace ...string) error {
	_, err := c.clientset.ExtensionsV1beta1().Ingresses(GetNamespace(namespace...)).Update(ing)
	return err
}

// UpdateReplicaSet updates a kube replica set.
func (c *Client) UpdateReplicaSet(rs *extv1beta1.ReplicaSet, namespace ...string) error {
	_, err := c.clientset.ExtensionsV1beta1().ReplicaSets(GetNamespace(namespace...)).Update(rs)
	return err
}

// UpdatePod updates a kube pod.
func (c *Client) UpdatePod(pod *corev1.Pod, namespace ...string) error {
	_, err := c.clientset.CoreV1().Pods(GetNamespace(namespace...)).Update(pod)
	return err
}

// UpdateSecret updates a given secret.
func (c *Client) UpdateSecret(secret *corev1.Secret, namespace ...string) error {
	_, err := c.clientset.CoreV1().Secrets(GetNamespace(namespace...)).Update(secret)
	return err
}

// UpdateConfigMap updates a given secret.
func (c *Client) UpdateConfigMap(secret *corev1.ConfigMap, namespace ...string) error {
	_, err := c.clientset.CoreV1().ConfigMaps(GetNamespace(namespace...)).Update(secret)
	return err
}

// UpdateRoleBinding updates a given role binding.
func (c *Client) UpdateRoleBinding(rb *rbacv1.RoleBinding, namespace ...string) error {
	_, err := c.clientset.RbacV1().RoleBindings(GetNamespace(namespace...)).Update(rb)
	return err
}

// UpdateClusterRoleBinding updates a given role binding.
func (c *Client) UpdateClusterRoleBinding(rb *rbacv1.ClusterRoleBinding) error {
	_, err := c.clientset.RbacV1().ClusterRoleBindings().Update(rb)
	return err
}

// UpdateRole updates a given role.
func (c *Client) UpdateRole(rb *rbacv1.Role, namespace ...string) error {
	_, err := c.clientset.RbacV1().Roles(GetNamespace(namespace...)).Update(rb)
	return err
}

// UpdateClusterRole updates a given role.
func (c *Client) UpdateClusterRole(rb *rbacv1.ClusterRole) error {
	_, err := c.clientset.RbacV1().ClusterRoles().Update(rb)
	return err
}

// UpdateStorageClass creates a storage class
func (c *Client) UpdateStorageClass(class *storagev1.StorageClass) error {
	_, err := c.clientset.StorageV1().StorageClasses().Update(class)
	return err
}

// UpdatePersistentVolumeClaim updates the pvc
func (c *Client) UpdatePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) error {
	_, err := c.clientset.CoreV1().PersistentVolumeClaims(GetNamespace(pvc.Namespace)).Update(pvc)
	return exception.New(err)
}

// UpdateCronJob updates the cron job
func (c *Client) UpdateCronJob(job *batchv1beta1.CronJob, namespace ...string) error {
	_, err := c.clientset.BatchV1beta1().CronJobs(GetNamespace(namespace...)).Update(job)
	return exception.New(err)
}

// UpdateJob updates the job
func (c *Client) UpdateJob(job *batchv1.Job, namespace ...string) error {
	_, err := c.clientset.BatchV1().Jobs(GetNamespace(namespace...)).Update(job)
	return exception.New(err)
}

// UpdateNode updates the node
func (c *Client) UpdateNode(node *corev1.Node) error {
	_, err := c.clientset.CoreV1().Nodes().Update(node)
	return exception.New(err)
}

// UpdateCustomResourceDefinition updates a crd
func (c *Client) UpdateCustomResourceDefinition(crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	_, err := c.apiextensionsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Update(crd)
	return exception.New(err)
}

// UpdateIssuer updates an issuer
func (c *Client) UpdateIssuer(issuer *certmanagerv1alpha1.Issuer, namespace ...string) error {
	_, err := c.certmanagerClientset.CertmanagerV1alpha1().Issuers(GetNamespace(namespace...)).Update(issuer)
	return exception.New(err)
}

// UpdateClusterIssuer updates a cluster issuer
func (c *Client) UpdateClusterIssuer(ci *certmanagerv1alpha1.ClusterIssuer) error {
	_, err := c.certmanagerClientset.CertmanagerV1alpha1().ClusterIssuers().Update(ci)
	return exception.New(err)
}

// UpdateCertificate updates a certificate
func (c *Client) UpdateCertificate(cert *certmanagerv1alpha1.Certificate, namespace ...string) error {
	_, err := c.certmanagerClientset.CertmanagerV1alpha1().Certificates(GetNamespace(namespace...)).Update(cert)
	return exception.New(err)
}

// --------------------------------------------------------------------------------
// Delete Functions
// --------------------------------------------------------------------------------

// DeleteDeployment deletes a deployment.
func (c *Client) DeleteDeployment(name string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy,
	})
}

// DeleteDeploymentAndWait deletes a deployment.
func (c *Client) DeleteDeploymentAndWait(name string, namespace ...string) error {
	err := c.DeleteDeployment(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.DeploymentExists, name, namespace...))
	})
	return exception.New(err)
}

// DeleteDeploymentsWithSelector batch deletes deployments based on selector
func (c *Client) DeleteDeploymentsWithSelector(selector string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).DeleteCollection(&apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy},
		apismetav1.ListOptions{
			LabelSelector: selector,
		})
}

// DeleteDeploymentsWithSelectorAndWait batch deletes deployments based on selector and waits till the completion of delete
func (c *Client) DeleteDeploymentsWithSelectorAndWait(selector string, namespace ...string) error {
	err := c.DeleteDeploymentsWithSelector(selector, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.DeploymentWithSelectorExists, selector, namespace...))
	})
	return exception.New(err)
}

// DeleteStatefulSet deletes a stateful set.
func (c *Client) DeleteStatefulSet(name string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return c.clientset.AppsV1beta1().StatefulSets(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy,
	})
}

// DeleteStatefulSetAndWait deletes a StatefulSet.
func (c *Client) DeleteStatefulSetAndWait(name string, namespace ...string) error {
	err := c.DeleteStatefulSet(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.StatefulSetExists, name, namespace...))
	})
	return exception.New(err)
}

// DeletePodDisruptionBudget deletes a PodDisruptionBudget.
func (c *Client) DeletePodDisruptionBudget(name string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return c.clientset.PolicyV1beta1().PodDisruptionBudgets(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy,
	})
}

// DeletePodDisruptionBudgetAndWait deletes a PodDisruptionBudget and waits
func (c *Client) DeletePodDisruptionBudgetAndWait(name string, namespace ...string) error {
	err := c.DeletePodDisruptionBudget(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.PodDisruptionBudgetExists, name, namespace...))
	})
	return exception.New(err)
}

// DeleteAutoscaler deletes a kube horizontal autoscaler.
func (c *Client) DeleteAutoscaler(name string, namespace ...string) error {
	return c.clientset.AutoscalingV2beta1().HorizontalPodAutoscalers(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteAutoscalerAndWait deletes an autoscaler and waits.
func (c *Client) DeleteAutoscalerAndWait(name string, namespace ...string) error {
	err := c.DeleteAutoscaler(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.AutoscalerExists, name, namespace...))
	})
	return exception.New(err)
}

// DeleteService deletes a service.
func (c *Client) DeleteService(name string, namespace ...string) error {
	err := c.clientset.CoreV1().Services(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
	return exception.New(err)
}

// DeleteServiceAndWait deletes a service.
func (c *Client) DeleteServiceAndWait(name string, namespace ...string) error {
	err := c.DeleteService(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.ServiceExists, name, namespace...))
	})
	return exception.New(err)
}

// DeleteAPIService Deletes a kube APIService.
func (c *Client) DeleteAPIService(name string) error {
	err := c.aggregatorClientset.ApiregistrationV1beta1().APIServices().Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
	return exception.New(err)
}

// DeleteAPIServiceAndWait deletes a service.
func (c *Client) DeleteAPIServiceAndWait(name string) error {
	err := c.DeleteAPIService(name)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(func() (bool, error) { return c.APIServiceExists(name) })
	})
	return exception.New(err)
}

// DeleteValidatingWebhookConfiguration Deletes a kube ValidatingWebhookConfiguration.
func (c *Client) DeleteValidatingWebhookConfiguration(name string) error {
	err := c.clientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
	return exception.New(err)
}

// DeleteValidatingWebhookConfigurationAndWait deletes a service.
func (c *Client) DeleteValidatingWebhookConfigurationAndWait(name string) error {
	err := c.DeleteValidatingWebhookConfiguration(name)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(func() (bool, error) { return c.ValidatingWebhookConfigurationExists(name) })
	})
	return exception.New(err)
}

// DeleteNamespace deletes a Namespace.
// WARN: very destructive
func (c *Client) DeleteNamespace(name string) error {
	return c.clientset.CoreV1().Namespaces().Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// FinalizeNamespace removes finalizers and finalizes the namespace. Needed for some to delete
func (c *Client) FinalizeNamespace(name string) (*corev1.Namespace, error) {
	namespace, err := c.GetNamespace(name)
	if err != nil {
		return nil, err
	}
	namespace.Finalizers = []string{}
	namespace.Spec.Finalizers = []corev1.FinalizerName{}
	return c.clientset.CoreV1().Namespaces().Finalize(namespace)
}

// DeleteNamespaceAndWait deletes a Namespace.
func (c *Client) DeleteNamespaceAndWait(name string, namespace ...string) error {
	err := c.DeleteNamespace(name)
	if IgnoreConflict(IgnoreNotFound(err)) != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(func() (bool, error) { return c.NamespaceExists(name) })
	})
	return exception.New(err)
}

// DeleteServiceAccount deletes a service account.
func (c *Client) DeleteServiceAccount(name string, namespace ...string) error {
	return c.clientset.CoreV1().ServiceAccounts(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteServiceAccountAndWait deletes a deployment.
func (c *Client) DeleteServiceAccountAndWait(name string, namespace string) error {
	err := c.DeleteServiceAccount(name, namespace)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.ServiceAccountExists, name, namespace))
	})
	return exception.New(err)
}

// DeleteIngress deletes a ingress.
func (c *Client) DeleteIngress(name string, namespace ...string) error {
	return c.clientset.ExtensionsV1beta1().Ingresses(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteIngressAndWait deletes a deployment.
func (c *Client) DeleteIngressAndWait(name string, namespace string) error {
	err := c.DeleteIngress(name, namespace)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.IngressExists, name, namespace))
	})
	return exception.New(err)
}

// DeleteReplicaSet deletes a replica set.
func (c *Client) DeleteReplicaSet(name string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return c.clientset.ExtensionsV1beta1().ReplicaSets(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy,
	})
}

// DeletePod deletes a pod.
func (c *Client) DeletePod(name string, namespace ...string) error {
	return c.clientset.CoreV1().Pods(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteSecret deletes a secret.
func (c *Client) DeleteSecret(name string, namespace ...string) error {
	return c.clientset.CoreV1().Secrets(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteClusterRoleBinding deletes a cluster role binding.
func (c *Client) DeleteClusterRoleBinding(name string) error {
	return c.clientset.RbacV1().ClusterRoleBindings().Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteClusterRoleBindingAndWait deletes a deployment.
func (c *Client) DeleteClusterRoleBindingAndWait(name string) error {
	err := c.DeleteClusterRoleBinding(name)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(func() (bool, error) { return c.ClusterRoleBindingExists(name) })
	})
	return exception.New(err)
}

// DeleteRoleBinding deletes a cluster role binding.
func (c *Client) DeleteRoleBinding(name string, namespace ...string) error {
	return c.clientset.RbacV1().RoleBindings(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteRoleBindingAndWait deletes a role binding.
func (c *Client) DeleteRoleBindingAndWait(name string, namespace ...string) error {
	err := c.DeleteRoleBinding(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.RoleBindingExists, name, namespace...))
	})
	return exception.New(err)
}

// DeleteRoleBindingsWithSelector batch deletes roles based on selector
func (c *Client) DeleteRoleBindingsWithSelector(selector string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return c.clientset.RbacV1().RoleBindings(GetNamespace(namespace...)).DeleteCollection(&apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy},
		apismetav1.ListOptions{
			LabelSelector: selector,
		})
}

// DeleteClusterRole deletes a cluster role .
func (c *Client) DeleteClusterRole(name string) error {
	return c.clientset.RbacV1().ClusterRoles().Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteClusterRoleAndWait deletes a deployment.
func (c *Client) DeleteClusterRoleAndWait(name string) error {
	err := c.DeleteClusterRole(name)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(func() (bool, error) { return c.ClusterRoleExists(name) })
	})
	return exception.New(err)
}

// DeleteRole deletes a cluster role .
func (c *Client) DeleteRole(name string, namespace ...string) error {
	return c.clientset.RbacV1().Roles(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteRoleAndWait deletes a role.
func (c *Client) DeleteRoleAndWait(name string, namespace ...string) error {
	err := c.DeleteRole(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.RoleExists, name, namespace...))
	})
	return exception.New(err)
}

// DeleteRolesWithSelector batch deletes roles based on selector
func (c *Client) DeleteRolesWithSelector(selector string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return c.clientset.RbacV1().Roles(GetNamespace(namespace...)).DeleteCollection(&apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy},
		apismetav1.ListOptions{
			LabelSelector: selector,
		})
}

// DeleteDaemonSet deletes a daemonset.
func (c *Client) DeleteDaemonSet(name string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return c.clientset.ExtensionsV1beta1().DaemonSets(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy,
	})
}

// DeleteDaemonSetAndWait deletes a role binding.
func (c *Client) DeleteDaemonSetAndWait(name string, namespace ...string) error {
	err := c.DeleteDaemonSet(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.DaemonSetExists, name, namespace...))
	})
	return exception.New(err)
}

// DeleteDaemonSetsWithSelector batch deletes daemonsets based on selector
func (c *Client) DeleteDaemonSetsWithSelector(selector string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return c.clientset.ExtensionsV1beta1().DaemonSets(GetNamespace(namespace...)).DeleteCollection(&apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy},
		apismetav1.ListOptions{
			LabelSelector: selector,
		})
}

// DeleteDaemonSetsWithSelectorAndWait batch deletes daemonsets based on selector and waits till the completion of delete
func (c *Client) DeleteDaemonSetsWithSelectorAndWait(selector string, namespace ...string) error {
	err := c.DeleteDaemonSetsWithSelector(selector, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		return core.Not(core.StringPredicate(c.DaemonSetWithSelectorExists, selector, namespace...))
	})
	return exception.New(err)
}

// DeleteConfigMap deletes a configmap.
func (c *Client) DeleteConfigMap(name string, namespace ...string) error {
	return c.clientset.CoreV1().ConfigMaps(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteConfigMapsWithSelector deletes secrets with selector.
func (c *Client) DeleteConfigMapsWithSelector(selector string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return c.clientset.CoreV1().ConfigMaps(GetNamespace(namespace...)).DeleteCollection(&apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy},
		apismetav1.ListOptions{
			LabelSelector: selector,
		})
}

// DeleteCertificateSigningRequest deletes a csr.
func (c *Client) DeleteCertificateSigningRequest(name string) error {
	return c.clientset.Certificates().CertificateSigningRequests().Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteStorageClass deletes a StorageClass.
func (c *Client) DeleteStorageClass(name string) error {
	return c.clientset.StorageV1().StorageClasses().Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// DeleteStorageClassAndWait deletes a StorageClass and waits.
func (c *Client) DeleteStorageClassAndWait(name string) error {
	err := c.DeleteStorageClass(name)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		exists, err := c.StorageClassExists(name)
		return !exists, err
	})
	return exception.New(err)
}

// DeletePersistentVolumeClaim deletes the persistent volume claim
func (c *Client) DeletePersistentVolumeClaim(name string, namespace ...string) error {
	return exception.New(c.clientset.CoreV1().PersistentVolumeClaims(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	}))
}

// DeletePersistentVolumeClaimsWithSelector batch deletes pvc based on selector
func (c *Client) DeletePersistentVolumeClaimsWithSelector(selector string, namespace ...string) error {
	propagationPolicy := apismetav1.DeletePropagationForeground
	return exception.New(c.clientset.CoreV1().PersistentVolumeClaims(GetNamespace(namespace...)).DeleteCollection(&apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &propagationPolicy},
		apismetav1.ListOptions{
			LabelSelector: selector,
		}))
}

// DeletePersistentVolumeClaimAndWait deletes the pvc and waits for deletion
func (c *Client) DeletePersistentVolumeClaimAndWait(name string, namespace ...string) error {
	err := c.DeletePersistentVolumeClaim(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		exists, err := c.PersistentVolumeClaimExists(name, namespace...)
		return !exists, err
	})
	return exception.New(err)
}

// DeleteCronJob deletes the cron job
func (c *Client) DeleteCronJob(name string, namespace ...string) error {
	background := apismetav1.DeletePropagationBackground
	return exception.New(c.clientset.BatchV1beta1().CronJobs(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &background,
	}))
}

// DeleteCronJobsWithSelector batch deletes the cron jobs based on selector
func (c *Client) DeleteCronJobsWithSelector(selector string, namespace ...string) error {
	background := apismetav1.DeletePropagationBackground
	return exception.New(c.clientset.BatchV1beta1().CronJobs(GetNamespace(namespace...)).DeleteCollection(&apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &background},
		apismetav1.ListOptions{
			LabelSelector: selector,
		}))
}

// DeleteCronJobAndWait deletes the cronjob and waits
func (c *Client) DeleteCronJobAndWait(name string, namespace ...string) error {
	err := c.DeleteCronJob(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		exists, err := c.CronJobExists(name, namespace...)
		return !exists, err
	})
	return exception.New(err)
}

// DeleteJob deletes the job
func (c *Client) DeleteJob(name string, namespace ...string) error {
	background := apismetav1.DeletePropagationBackground
	return exception.New(c.clientset.BatchV1().Jobs(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &background,
	}))
}

// DeleteJobAndWait deletes the job and waits
func (c *Client) DeleteJobAndWait(name string, namespace ...string) error {
	err := c.DeleteJob(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		exists, err := c.JobExists(name, namespace...)
		return !exists, err
	})
	return exception.New(err)
}

// DeleteCustomResourceDefinition deletes the crd
func (c *Client) DeleteCustomResourceDefinition(name string) error {
	background := apismetav1.DeletePropagationBackground
	return exception.New(c.apiextensionsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &background,
	}))
}

// DeleteCustomResourceDefinitionAndWait deletes the crd and waits
func (c *Client) DeleteCustomResourceDefinitionAndWait(name string) error {
	err := c.DeleteCustomResourceDefinition(name)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		exists, err := c.CustomResourceDefinitionExists(name)
		return !exists, err
	})
	return exception.New(err)
}

// DeleteIssuer deletes the cluster issuer
func (c *Client) DeleteIssuer(name string, namespace ...string) error {
	background := apismetav1.DeletePropagationBackground
	return exception.New(c.certmanagerClientset.CertmanagerV1alpha1().Issuers(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &background,
	}))
}

// DeleteIssuerAndWait deletes the cluster issuer and waits
func (c *Client) DeleteIssuerAndWait(name string, namespace ...string) error {
	err := c.DeleteIssuer(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		exists, err := c.IssuerExists(name)
		return !exists, err
	})
	return exception.New(err)
}

// DeleteClusterIssuer deletes the cluster issuer
func (c *Client) DeleteClusterIssuer(name string) error {
	background := apismetav1.DeletePropagationBackground
	return exception.New(c.certmanagerClientset.CertmanagerV1alpha1().ClusterIssuers().Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &background,
	}))
}

// DeleteClusterIssuerAndWait deletes the cluster issuer and waits
func (c *Client) DeleteClusterIssuerAndWait(name string) error {
	err := c.DeleteClusterIssuer(name)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		exists, err := c.ClusterIssuerExists(name)
		return !exists, err
	})
	return exception.New(err)
}

// DeleteCertificate deletes the certificate
func (c *Client) DeleteCertificate(name string, namespace ...string) error {
	background := apismetav1.DeletePropagationBackground
	return exception.New(c.certmanagerClientset.CertmanagerV1alpha1().Certificates(GetNamespace(namespace...)).Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
		PropagationPolicy:  &background,
	}))
}

// DeleteCertificateAndWait deletes the certificate and waits
func (c *Client) DeleteCertificateAndWait(name string, namespace ...string) error {
	err := c.DeleteCertificate(name, namespace...)
	if err != nil {
		return exception.New(IgnoreNotFound(err))
	}
	err = wait.PollImmediate(deletePollInterval, deletePollTimeout, func() (bool, error) {
		exists, err := c.CertificateExists(name, namespace...)
		return !exists, err
	})
	return exception.New(err)
}

// DeleteNode deletes a Node.
// WARN: very destructive
func (c *Client) DeleteNode(name string) error {
	return c.clientset.CoreV1().Nodes().Delete(name, &apismetav1.DeleteOptions{
		GracePeriodSeconds: ref.Int64(0),
	})
}

// --------------------------------------------------------------------------------
// Patch Functions
// --------------------------------------------------------------------------------

// PatchDeploymentMetadata patches deployment metadata.
func (c *Client) PatchDeploymentMetadata(dep *appsv1beta1.Deployment, namespace ...string) error {
	_, err := c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Patch(dep.Name, kubetypes.MergePatchType, CreateDeploymentMetadataPatch(dep))
	return exception.New(err)
}

// PatchDeploymentAnnotations patches deployment metadata.
func (c *Client) PatchDeploymentAnnotations(dep *appsv1beta1.Deployment, namespace ...string) error {
	_, err := c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Patch(dep.Name, kubetypes.MergePatchType, CreateDeploymentAnnotationsPatch(dep))
	return exception.New(err)
}

// PatchDeploymentPodSpecMetadata patches deployment pod spec metadata.
func (c *Client) PatchDeploymentPodSpecMetadata(dep *appsv1beta1.Deployment, namespace ...string) error {
	_, err := c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Patch(dep.Name, kubetypes.MergePatchType, CreateDeploymentPodSpecMetadataPatch(dep))
	return exception.New(err)
}

// PatchDeploymentPodSpecAnnotations patches deployment pod spec metadata.
func (c *Client) PatchDeploymentPodSpecAnnotations(dep *appsv1beta1.Deployment, namespace ...string) error {
	_, err := c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Patch(dep.Name, kubetypes.MergePatchType, CreateDeploymentPodSpecAnnotationsPatch(dep))
	return exception.New(err)
}

// PatchDeploymentPodSpecLabels patches deployment pod spec metadata.
func (c *Client) PatchDeploymentPodSpecLabels(dep *appsv1beta1.Deployment, namespace ...string) error {
	_, err := c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Patch(dep.Name, kubetypes.MergePatchType, CreateDeploymentPodSpecLabelsPatch(dep))
	return exception.New(err)
}

// PatchDeploymentLabels patches deployment metadata.
func (c *Client) PatchDeploymentLabels(dep *appsv1beta1.Deployment, namespace ...string) error {
	_, err := c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Patch(dep.Name, kubetypes.MergePatchType, CreateDeploymentLabelsPatch(dep))
	return exception.New(err)
}

// PatchJobLabels patches job metadata.
func (c *Client) PatchJobLabels(job *batchv1.Job, namespace ...string) error {
	_, err := c.clientset.BatchV1().Jobs(GetNamespace(namespace...)).Patch(job.Name, kubetypes.MergePatchType, CreateJobLabelsPatch(job))
	return exception.New(err)
}

// PatchConfigMapData patches configmap data.
func (c *Client) PatchConfigMapData(name string, data map[string]*string, namespace ...string) error {
	_, err := c.clientset.CoreV1().ConfigMaps(GetNamespace(namespace...)).Patch(name, kubetypes.MergePatchType, CreateConfigMapDataPatch(data))
	return exception.New(err)
}

// PatchServiceSelector patches the service selector for a kube service.
func (c *Client) PatchServiceSelector(name string, selector map[string]string, namespace ...string) error {
	_, err := c.clientset.CoreV1().Services(GetNamespace(namespace...)).Patch(name, kubetypes.MergePatchType, CreateServiceSelectorPatch(selector))
	return exception.New(err)
}

// PatchNode patches the node
func (c *Client) PatchNode(name string, patchBytes []byte) error {
	_, err := c.clientset.CoreV1().Nodes().Patch(name, kubetypes.MergePatchType, patchBytes)
	return exception.New(err)
}

// --------------------------------------------------------------------------------
//  Upsert Functions
// --------------------------------------------------------------------------------

// mergeServiceAccountSecrets merges service account token secrets from `theirs` into `ours`
// to prevent kube token controller from creating a new token
func (c *Client) mergeServiceAccountSecrets(ours, theirs *corev1.ServiceAccount) error {
	if len(theirs.Secrets) > 0 {
		// Index our secrets in `mergedSet`
		mergedSet := sets.NewString()
		for _, secret := range ours.Secrets {
			mergedSet.Insert(secret.Name)
		}

		// Get all token secrets of `theirs`
		tokens, err := c.GetAllTokenSecretsForServiceAccount(theirs)
		if err != nil {
			return exception.New(err)
		}

		// Add their secrets that are not in `mergedSet` to ours
		for _, secret := range tokens {
			if !mergedSet.Has(secret.Name) {
				mergedSet.Insert(secret.Name)
				ours.Secrets = append(ours.Secrets, corev1.ObjectReference{
					Kind:      KindSecret,
					Name:      secret.Name,
					Namespace: secret.Namespace,
				})
			}
		}
	}
	return nil
}

// UpsertServiceAccount creates or updates a service account
func (c *Client) UpsertServiceAccount(sa *corev1.ServiceAccount, force bool) error {
	if sa == nil {
		return exception.New(fmt.Sprintf("Nil ServiceAccount"))
	}
	if force {
		err := c.DeleteServiceAccountAndWait(sa.Name, sa.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	existing, err := c.GetServiceAccount(sa.Name, sa.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	// https://github.com/kubernetes/kubernetes/blob/master/plugin/pkg/admission/serviceaccount/admission.go#L46
	if sa.Annotations == nil {
		sa.Annotations = map[string]string{}
	}
	sa.Annotations["kubernetes.io/enforce-mountable-secrets"] = "true"
	if err == nil { // sa exists
		if err := c.mergeServiceAccountSecrets(sa, existing); err != nil {
			return exception.New(err)
		}
		return c.UpdateServiceAccount(sa)
	}
	return c.CreateServiceAccount(sa)
}

// UpsertDeployment creates or updates the deployment
func (c *Client) UpsertDeployment(dep *appsv1beta1.Deployment, force bool) error {
	if dep == nil {
		return exception.New(fmt.Sprintf("Nil Deployment"))
	}
	if force {
		err := c.DeleteDeploymentAndWait(dep.Name, dep.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.DeploymentExists(dep.Name, dep.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateDeployment(dep, dep.Namespace)
	}
	return c.CreateDeployment(dep, dep.Namespace)
}

// UpsertService creates or updates the service
func (c *Client) UpsertService(srv *corev1.Service, force bool) error {
	if srv == nil {
		return exception.New(fmt.Sprintf("Nil Service"))
	}
	if force {
		err := c.DeleteService(srv.Name, srv.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.ServiceExists(srv.Name, srv.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateServiceWithRetry(srv)
	}
	return c.CreateService(srv, srv.Namespace)
}

// UpsertIngress creates or updates the ingress
func (c *Client) UpsertIngress(ing *extv1beta1.Ingress, force bool) error {
	if ing == nil {
		return exception.New(fmt.Sprintf("Nil Ingress"))
	}
	if force {
		err := c.DeleteIngressAndWait(ing.Name, ing.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.IngressExists(ing.Name, ing.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateIngress(ing, ing.Namespace)
	}
	return c.CreateIngress(ing, ing.Namespace)
}

// UpsertAutoscaler creates or updates the autoscaler
func (c *Client) UpsertAutoscaler(auto *autov2beta1.HorizontalPodAutoscaler, force bool) error {
	if auto == nil {
		return exception.New(fmt.Sprintf("Nil Autoscaler"))
	}
	if force {
		err := c.DeleteAutoscalerAndWait(auto.Name, auto.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.AutoscalerExists(auto.Name, auto.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateAutoscaler(auto, auto.Namespace)
	}
	return c.CreateAutoscaler(auto, auto.Namespace)
}

// UpsertNamespace creates or updates a namespace
func (c *Client) UpsertNamespace(ns *corev1.Namespace, force bool) error {
	if ns == nil {
		return exception.New(fmt.Sprintf("Nil Namespace"))
	}
	if force {
		err := c.DeleteNamespaceAndWait(ns.Name)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.NamespaceExists(ns.Name)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateNamespace(ns)
	}
	return c.CreateNamespace(ns)
}

// UpsertAPIService creates or updates a APIService
func (c *Client) UpsertAPIService(svc *apiregistrationv1beta1.APIService, force bool) error {
	if svc == nil {
		return exception.New(fmt.Sprintf("Nil APIService"))
	}
	if force {
		err := c.DeleteAPIServiceAndWait(svc.Name)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.APIServiceExists(svc.Name)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateAPIService(svc)
	}
	return c.CreateAPIService(svc)
}

// UpsertValidatingWebhookConfiguration creates or updates a ValidatingWebhookConfiguration
func (c *Client) UpsertValidatingWebhookConfiguration(v *admissionregistrationv1beta1.ValidatingWebhookConfiguration, force bool) error {
	if v == nil {
		return exception.New(fmt.Sprintf("Nil ValidatingWebhookConfiguration"))
	}
	if force {
		err := c.DeleteValidatingWebhookConfigurationAndWait(v.Name)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.ValidatingWebhookConfigurationExists(v.Name)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateValidatingWebhookConfiguration(v)
	}
	return c.CreateValidatingWebhookConfiguration(v)
}

// UpsertPodDisruptionBudget creates or updates a PodDisruptionBudget
func (c *Client) UpsertPodDisruptionBudget(pdb *policyv1beta1.PodDisruptionBudget, force bool) error {
	if pdb == nil {
		return exception.New(fmt.Sprintf("Nil PodDisruptionBudget"))
	}
	if force {
		err := c.DeletePodDisruptionBudgetAndWait(pdb.Name, pdb.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.PodDisruptionBudgetExists(pdb.Name, pdb.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdatePodDisruptionBudget(pdb)
	}
	return c.CreatePodDisruptionBudget(pdb)
}

// UpsertSecret creates or updates the secret
func (c *Client) UpsertSecret(srv *corev1.Secret, force bool) error {
	if srv == nil {
		return exception.New(fmt.Sprintf("Nil Secret"))
	}
	if force {
		err := c.DeleteSecret(srv.Name, srv.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.SecretExists(srv.Name, srv.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateSecret(srv, srv.Namespace)
	}
	return c.CreateSecret(srv, srv.Namespace)
}

// ReplaceStorageClass creates or updates a StorageClass
func (c *Client) ReplaceStorageClass(sc *storagev1.StorageClass) error {
	if sc == nil {
		return exception.New(fmt.Sprintf("Nil StorageClass"))
	}
	err := c.DeleteStorageClassAndWait(sc.Name)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	return c.CreateStorageClass(sc)
}

// UpsertStorageClass creates or updates a StorageClass with a force option
func (c *Client) UpsertStorageClass(sc *storagev1.StorageClass, force bool) error {
	if sc == nil {
		return exception.New(fmt.Sprintf("Nil StorageClass"))
	}
	if force {
		err := c.DeleteStorageClassAndWait(sc.Name)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.StorageClassExists(sc.Name)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateStorageClass(sc)
	}
	return c.CreateStorageClass(sc)
}

// UpsertStatefulSet creates or updates a StatefulSet
func (c *Client) UpsertStatefulSet(ss *appsv1beta1.StatefulSet, force bool) error {
	if ss == nil {
		return exception.New(fmt.Sprintf("Nil StatefulSet"))
	}
	if force {
		err := c.DeleteStatefulSetAndWait(ss.Name, ss.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.StatefulSetExists(ss.Name, ss.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateStatefulSet(ss)
	}
	return c.CreateStatefulSet(ss)
}

// UpsertPersistentVolumeClaim upserts the persistent volume claim
func (c *Client) UpsertPersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim, force bool) error {
	if pvc == nil {
		return exception.New(fmt.Sprintf("Nil Persistent Volume Claim"))
	}
	if force {
		err := c.DeletePersistentVolumeClaimAndWait(pvc.Name, pvc.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.PersistentVolumeClaimExists(pvc.Name, pvc.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdatePersistentVolumeClaim(pvc)
	}
	return c.CreatePersistentVolumeClaim(pvc)
}

// UpsertCronJob upserts the cron job
func (c *Client) UpsertCronJob(job *batchv1beta1.CronJob, force bool) error {
	if job == nil {
		return exception.New(fmt.Sprintf("Nil Cron Job"))
	}
	if force {
		err := c.DeleteCronJobAndWait(job.Name, job.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.CronJobExists(job.Name, job.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateCronJob(job, job.Namespace)
	}
	return c.CreateCronJob(job, job.Namespace)
}

// UpsertJob upserts the job
func (c *Client) UpsertJob(job *batchv1.Job, force bool) error {
	if job == nil {
		return exception.New(fmt.Sprintf("Nil Job"))
	}
	if force {
		err := c.DeleteJobAndWait(job.Name, job.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.JobExists(job.Name, job.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateJob(job, job.Namespace)
	}
	_, err = c.CreateJob(job, job.Namespace)
	return err
}

// UpsertCustomResourceDefinition upserts the crd
func (c *Client) UpsertCustomResourceDefinition(crd *apiextensionsv1beta1.CustomResourceDefinition, force bool) error {
	if crd == nil {
		return exception.New(fmt.Sprintf("Nil CustomResourceDefinition"))
	}
	if force {
		err := c.DeleteCustomResourceDefinitionAndWait(crd.Name)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	exists, err := c.CustomResourceDefinitionExists(crd.Name)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if exists {
		return c.UpdateCustomResourceDefinition(crd)
	}
	return c.CreateCustomResourceDefinition(crd)
}

// UpsertIssuer upserts the cluster issuer
func (c *Client) UpsertIssuer(issuer *certmanagerv1alpha1.Issuer, force bool) error {
	if issuer == nil {
		return exception.New(fmt.Sprintf("Nil Issuer"))
	}
	if force {
		err := c.DeleteIssuerAndWait(issuer.Name, issuer.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	old, err := c.GetIssuer(issuer.Name, issuer.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if !IsNotFoundError(err) && old != nil {
		issuer.ResourceVersion = old.ResourceVersion
		return c.UpdateIssuer(issuer, issuer.Namespace)
	}
	return c.CreateIssuer(issuer, issuer.Namespace)
}

// UpsertClusterIssuer upserts the cluster issuer
func (c *Client) UpsertClusterIssuer(ci *certmanagerv1alpha1.ClusterIssuer, force bool) error {
	if ci == nil {
		return exception.New(fmt.Sprintf("Nil ClusterIssuer"))
	}
	if force {
		err := c.DeleteClusterIssuerAndWait(ci.Name)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	old, err := c.GetClusterIssuer(ci.Name)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if !IsNotFoundError(err) && old != nil {
		ci.ResourceVersion = old.ResourceVersion
		return c.UpdateClusterIssuer(ci)
	}
	return c.CreateClusterIssuer(ci)
}

// UpsertCertificate upserts the cluster issuer
func (c *Client) UpsertCertificate(cert *certmanagerv1alpha1.Certificate, force bool) error {
	if cert == nil {
		return exception.New(fmt.Sprintf("Nil Certificate"))
	}
	if force {
		err := c.DeleteCertificateAndWait(cert.Name, cert.Namespace)
		if IgnoreNotFound(err) != nil {
			return exception.New(err)
		}
	}
	old, err := c.GetCertificate(cert.Name, cert.Namespace)
	if IgnoreNotFound(err) != nil {
		return exception.New(err)
	}
	if !IsNotFoundError(err) && old != nil {
		cert.ResourceVersion = old.ResourceVersion
		return c.UpdateCertificate(cert, cert.Namespace)
	}
	return c.CreateCertificate(cert, cert.Namespace)
}

// UpsertConfigMap upserts a config map
func (c *Client) UpsertConfigMap(configMap *v1.ConfigMap) error {
	existingConfigMap, err := c.GetConfigMap(configMap.Name, configMap.Namespace)

	// This first checks to see if a config map was found. If the config
	// map wasn't found, create one. Otherwise, check if there was an error
	// that wasn't a not found error. If those branches fall through, then
	// our only option is to update the existing config map, which must not
	// be `nil`. The last case means that there was no not found error, but
	// the config map was still `nil`, which should never happen.
	if IsNotFoundError(err) {
		// We weren't able to retrieve a valid config map so attempt to create
		// a new one
		return c.CreateConfigMap(configMap, configMap.Namespace)
	}
	if err != nil {
		// The not found error is ok, otherwise we have an actual issue
		return err
	}
	if existingConfigMap != nil {
		// Otherwise, update the config map
		return c.UpdateConfigMap(configMap, configMap.Namespace)
	}
	return exception.New("The retrieved config map was not valid")
}

// --------------------------------------------------------------------------------
// Extension Functions
// --------------------------------------------------------------------------------

// Rollback rolls back the deployment
func (c *Client) Rollback(deploymentRollback *extv1beta1.DeploymentRollback, namespace ...string) error {
	return c.clientset.ExtensionsV1beta1().Deployments(GetNamespace(namespace...)).Rollback(deploymentRollback)
}

// --------------------------------------------------------------------------------
// App Helper Functions
// --------------------------------------------------------------------------------

// DeploymentRolloutStatus returns the status of the rollout of the deployment
func (c *Client) DeploymentRolloutStatus(name string, namespace ...string) (string, bool, error) {
	deployment, err := c.GetDeployment(name, namespace...)
	if err != nil {
		return "", false, err
	}

	// source from https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/rollout_status.go
	if deployment.Generation <= deployment.Status.ObservedGeneration {
		if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
			return fmt.Sprintf("Waiting for deployment %q rollout to finish: %d out of %d new replicas have been updated...\n", deployment.Name, deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas), false, nil
		}
		if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
			return fmt.Sprintf("Waiting for deployment %q rollout to finish: %d old replicas are pending termination...\n", deployment.Name, deployment.Status.Replicas-deployment.Status.UpdatedReplicas), false, nil
		}
		if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
			return fmt.Sprintf("Waiting for deployment %q rollout to finish: %d of %d updated replicas are available...\n", deployment.Name, deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas), false, nil
		}
		return fmt.Sprintf("deployment %q successfully rolled out\n", deployment.Name), true, nil
	}
	return fmt.Sprintf("Waiting for deployment spec update to be observed...\n"), false, nil
}

// ServiceRolloutStatus returns the status of the rollout on the service level
func (c *Client) ServiceRolloutStatus(name string, replicas int, namespace ...string) (string, bool, error) {
	service, err := c.GetService(name, namespace...)
	if err != nil {
		return "", false, err
	}
	pods, err := c.GetPodsWithSelector(selector.FromLabels(service.Spec.Selector), namespace...)
	if err != nil {
		return "", false, err
	}
	count := len(pods.Items)
	// subtract evicted pods from count
	for _, pod := range pods.Items {
		if PodIsEvicted(pod.Status) {
			count--
		}
	}
	if count < replicas {
		return fmt.Sprintf("Waiting for service %q rollout to finish: %d out of %d new replicas have been updated...\n", service.Name, count, replicas), false, nil
	}
	if count > replicas {
		return fmt.Sprintf("Waiting for service %q rollout to finish: %d old replicas are pending termination...\n", service.Name, count-replicas), false, nil
	}
	return fmt.Sprintf("Service rollout complete"), true, nil
}

// PinPod adds the pin label to the pod
func (c *Client) PinPod(podName string, namespace ...string) error {
	pod, err := c.GetPod(podName, namespace...)
	if err != nil {
		return err
	}
	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}
	pod.Labels[LabelPinPod] = "true"
	if err := c.UpdatePod(pod, pod.Namespace); err != nil {
		return err
	}

	// pin parent job
	if jobName, ok := pod.Labels["job-name"]; ok {
		return c.PinJob(jobName, pod.Namespace)
	}
	return nil
}

// UnpinPod removes the pin label from the pod
func (c *Client) UnpinPod(podName string, namespace ...string) error {
	pod, err := c.GetPod(podName, namespace...)
	if err != nil {
		return err
	}
	delete(pod.Labels, LabelPinPod)
	if err := c.UpdatePod(pod, pod.Namespace); err != nil {
		return err
	}

	// unpin paraent job
	if jobName, ok := pod.Labels["job-name"]; ok {
		return c.UnpinJob(jobName, pod.Namespace)
	}
	return nil
}

// PinJob adds the pin label to the job
func (c *Client) PinJob(jobName string, namespace ...string) error {
	job, err := c.GetJob(jobName, namespace...)
	if err != nil {
		return err
	}
	if job.Labels == nil {
		job.Labels = map[string]string{}
	}
	job.Labels[LabelPinPod] = "true"
	return c.UpdateJob(job, job.Namespace)
}

// UnpinJob removes the pin label from the job
func (c *Client) UnpinJob(jobName string, namespace ...string) error {
	job, err := c.GetJob(jobName, namespace...)
	if err != nil {
		return err
	}
	delete(job.Labels, LabelPinPod)
	return c.UpdateJob(job, job.Namespace)
}

// GetServiceDeployments gets deployments that are deployinator services.
func (c *Client) GetServiceDeployments(namespace ...string) (*appsv1beta1.DeploymentList, error) {
	return c.GetDeploymentsWithSelector(selector.Equals(LabelRole, RoleService))
}

// GetPodsForDeployment gets all pods for a given deployment
func (c *Client) GetPodsForDeployment(deploymentId string, namespace ...string) (*corev1.PodList, error) {
	return c.GetPodsWithSelector(selector.Equals(LabelDeploy, deploymentId))
}

// GetPodsForService gets all pods for a given service.
func (c *Client) GetPodsForService(serviceName string, namespace ...string) (*corev1.PodList, error) {
	return c.GetPodsWithSelector(
		selector.And(
			selector.Equals(LabelService, serviceName),
			selector.Equals(LabelRole, RoleServiceInstance),
		),
	)
}

// GetPodsForTask gets all pods for a given task
func (c *Client) GetPodsForTask(serviceName string, namespace ...string) (*corev1.PodList, error) {
	return c.GetPodsWithSelector(
		selector.And(
			selector.Equals(LabelService, serviceName),
			selector.Equals(LabelRole, RoleTask),
		),
	)
}

// GetPodsForJob gets all pods for a given job
func (c *Client) GetPodsForJob(jobName string, namespace ...string) (*corev1.PodList, error) {
	return c.GetPodsWithSelector(selector.Equals("job-name", jobName), namespace...)
}

// GetJobsForTask gets all jobs for a given task
func (c *Client) GetJobsForTask(serviceName string, namespace ...string) (*batchv1.JobList, error) {
	return c.GetJobsWithSelector(selector.And(
		selector.Equals(LabelService, serviceName),
		selector.Equals(LabelRole, RoleTask),
	),
	)
}

// GetReplicaSetsForService gets the replica sets for a given service.
func (c *Client) GetReplicaSetsForService(serviceName string, namespace ...string) (*extv1beta1.ReplicaSetList, error) {
	return c.GetReplicaSetsWithSelector(selector.Equals(LabelService, serviceName), namespace...)
}

// ReplicaSetsExistForService gets the replica sets for a given service.
func (c *Client) ReplicaSetsExistForService(serviceName string, namespace ...string) bool {
	rss, err := c.GetReplicaSetsForService(serviceName, namespace...)
	return err == nil && rss != nil && len(rss.Items) > 0
}

// GetBuilds gets all job pods loaded in the system.
func (c *Client) GetBuilds(namespace ...string) (*corev1.PodList, error) {
	return c.GetPodsWithSelector(selector.Equals(LabelRole, RoleJob), namespace...)
}

// GetBuildJobs gets all deploy jobs loaded in the system.
func (c *Client) GetBuildJobs(namespace ...string) (*batchv1.JobList, error) {
	return c.GetJobsWithSelector(selector.Equals(LabelRole, RoleJob), namespace...)
}

// GetTasks gets all task jobs loaded in the system.
func (c *Client) GetTasks(namespace ...string) (*batchv1.JobList, error) {
	return c.GetJobsWithSelector(selector.Equals(LabelRole, RoleTask), namespace...)
}

// GetBuildsForService gets the job pods for a service.
func (c *Client) GetBuildsForService(serviceName string, namespace ...string) (*corev1.PodList, error) {
	return c.GetPodsWithSelector(
		selector.And(
			selector.Equals(LabelRole, RoleJob),
			selector.Equals(LabelService, serviceName),
		), namespace...)
}

// GetBuildsForProject gets the job pods for a project.
func (c *Client) GetBuildsForProject(projectName string, namespace ...string) (*corev1.PodList, error) {
	return c.GetPodsWithSelector(
		selector.And(
			selector.Equals(LabelRole, RoleJob),
			selector.Equals(LabelProject, projectName),
		), namespace...)
}

// GetScheduledDeploys get all the scheduled deploys in the namespace
func (c *Client) GetScheduledDeploys(namespace ...string) (*batchv1beta1.CronJobList, error) {
	return c.GetCronJobsWithSelector(selector.Equals(LabelRole, RoleJob), namespace...)
}

// GetScheduledDeploysForService gets all scheduled deploys for a service
func (c *Client) GetScheduledDeploysForService(serviceName string, namespace ...string) (*batchv1beta1.CronJobList, error) {
	return c.GetCronJobsWithSelector(
		selector.And(
			selector.Equals(LabelRole, RoleJob),
			selector.Equals(LabelService, serviceName),
		), namespace...)
}

func (c *Client) deleteServiceSecretKey(secret *corev1.Secret, key string) error {
	delete(secret.Data, key)
	delete(secret.StringData, key)
	return c.UpdateServiceSecret(secret)
}

func (c *Client) setServiceSecretKey(secret *corev1.Secret, key string, contents []byte) error {
	secret.Data[key] = contents
	return c.UpdateServiceSecret(secret)
}

// GetEnvVarsForService gets the secrets for a service.
func (c *Client) GetEnvVarsForService(serviceName string, namespace ...string) (*corev1.Secret, error) {
	secret, err := c.GetSecret(ServiceEnvVarsSecretName(serviceName), namespace...)
	if IsNotFoundError(err) {
		secret, err = c.createAndGetEmptyServiceSecret(ServiceEnvVarsSecretName(serviceName), namespace...)
	}
	if err != nil {
		return nil, err
	}
	secret = fillNilDataValuesForSecret(secret)
	return secret, nil
}

// EnvVarsExistForService returns if the env vars secret exists for a service.
func (c *Client) EnvVarsExistForService(serviceName string, namespace ...string) bool {
	secret, _ := c.GetEnvVarsForService(serviceName, namespace...)
	return secret != nil
}

// SetEnvVarValue sets a secret value by its key.
func (c *Client) SetEnvVarValue(serviceName, key, value string, namespace ...string) error {
	secret, err := c.GetEnvVarsForService(serviceName, namespace...)
	if err != nil {
		return err
	}
	return c.setServiceSecretKey(secret, key, []byte(value))
}

// DeleteEnvVar deletes the secret value.
func (c *Client) DeleteEnvVar(serviceName, key string, namespace ...string) error {
	secret, err := c.GetEnvVarsForService(serviceName, namespace...)
	if err != nil {
		return err
	}
	return c.deleteServiceSecretKey(secret, key)
}

// DeleteEnvVarsForService deletes the env vars secret for a service.
func (c *Client) DeleteEnvVarsForService(serviceName string, namespace ...string) error {
	return c.DeleteSecret(ServiceEnvVarsSecretName(serviceName), namespace...)
}

// GetEnvVarsForProject gets the secrets for a project.
func (c *Client) GetEnvVarsForProject(projectName string, namespace ...string) (*corev1.Secret, error) {
	secret, err := c.GetSecret(ProjectEnvVarsSecretName(projectName), namespace...)
	if IsNotFoundError(err) {
		secret, err = c.createAndGetEmptyServiceSecret(ProjectEnvVarsSecretName(projectName), namespace...)
	}
	if err != nil {
		return nil, err
	}
	secret = fillNilDataValuesForSecret(secret)
	return secret, nil
}

// EnvVarsExistForProject returns if the env vars secret exists for a project.
func (c *Client) EnvVarsExistForProject(projectName string, namespace ...string) bool {
	secret, _ := c.GetEnvVarsForProject(projectName, namespace...)
	return secret != nil
}

// SetEnvVarValueForProject sets a secret value by its key.
func (c *Client) SetEnvVarValueForProject(projectName, key, value string, namespace ...string) error {
	secret, err := c.GetEnvVarsForProject(projectName, namespace...)
	if err != nil {
		return err
	}
	return c.setServiceSecretKey(secret, key, []byte(value))
}

// DeleteEnvVarForProject deletes the secret value.
func (c *Client) DeleteEnvVarForProject(projectName, key string, namespace ...string) error {
	secret, err := c.GetEnvVarsForProject(projectName, namespace...)
	if err != nil {
		return err
	}
	return c.deleteServiceSecretKey(secret, key)
}

// DeleteEnvVarsForProject deletes the env vars secret for a project.
func (c *Client) DeleteEnvVarsForProject(projectName string, namespace ...string) error {
	return c.DeleteSecret(ProjectEnvVarsSecretName(projectName), namespace...)
}

// GetFilesForService gets the secrets for a service.
func (c *Client) GetFilesForService(serviceName string, namespace ...string) (*corev1.Secret, error) {
	secret, err := c.clientset.CoreV1().Secrets(GetNamespace(namespace...)).Get(ServiceFilesSecretName(serviceName), apismetav1.GetOptions{})
	if IsNotFoundError(err) {
		secret, err = c.createAndGetEmptyServiceSecret(ServiceFilesSecretName(serviceName), namespace...)
	}
	if err != nil {
		return nil, err
	}
	secret = fillNilDataValuesForSecret(secret)
	return secret, nil
}

// FilesExistForService returns if the files secret exists for a service.
func (c *Client) FilesExistForService(serviceName string, namespace ...string) bool {
	secret, _ := c.GetFilesForService(serviceName, namespace...)
	return secret != nil
}

// SetFileForService sets a secret value by its key.
func (c *Client) SetFileForService(serviceName, fileName string, contents []byte, namespace ...string) error {
	secret, err := c.GetFilesForService(serviceName, namespace...)
	if err != nil {
		return err
	}
	return c.setServiceSecretKey(secret, fileName, contents)
}

// DeleteFileForService deletes the secret value.
func (c *Client) DeleteFileForService(serviceName, key string, namespace ...string) error {
	secret, err := c.GetFilesForService(serviceName, namespace...)
	if err != nil {
		return err
	}
	return c.deleteServiceSecretKey(secret, key)
}

// DeleteFilesForService deletes the files secret for a service.
func (c *Client) DeleteFilesForService(serviceName string, namespace ...string) error {
	return c.DeleteSecret(ServiceFilesSecretName(serviceName), namespace...)
}

// GetFilesForProject gets the secrets for a project.
func (c *Client) GetFilesForProject(projectName string, namespace ...string) (*corev1.Secret, error) {
	secret, err := c.clientset.CoreV1().Secrets(GetNamespace(namespace...)).Get(ProjectFilesSecretName(projectName), apismetav1.GetOptions{})
	if IsNotFoundError(err) {
		secret, err = c.createAndGetEmptyServiceSecret(ProjectFilesSecretName(projectName), namespace...)
	}
	if err != nil {
		return nil, err
	}
	secret = fillNilDataValuesForSecret(secret)
	return secret, nil
}

// FilesExistForProject returns if the files secret exists for a service.
func (c *Client) FilesExistForProject(projectName string, namespace ...string) bool {
	secret, _ := c.GetFilesForProject(projectName, namespace...)
	return secret != nil
}

// SetFileForProject sets a secret value by its key.
func (c *Client) SetFileForProject(projectName, fileName string, contents []byte, namespace ...string) error {
	secret, err := c.GetFilesForProject(projectName, namespace...)
	if err != nil {
		return err
	}
	return c.setServiceSecretKey(secret, fileName, contents)
}

// DeleteFileForProject deletes the secret value.
func (c *Client) DeleteFileForProject(projectName, key string, namespace ...string) error {
	secret, err := c.GetFilesForProject(projectName, namespace...)
	if err != nil {
		return err
	}
	return c.deleteServiceSecretKey(secret, key)
}

// DeleteFilesForProject deletes the files secret for a service.
func (c *Client) DeleteFilesForProject(projectName string, namespace ...string) error {
	return c.DeleteSecret(ProjectFilesSecretName(projectName), namespace...)
}

// CertsExistForService returns if the certs secret exists for a service.
func (c *Client) CertsExistForService(serviceName string, namespace ...string) bool {
	secret, _ := c.GetCertsForService(serviceName, namespace...)
	return secret != nil
}

// GetCertsForService gets the cert secrets for a service.
func (c *Client) GetCertsForService(serviceName string, namespace ...string) (*corev1.Secret, error) {
	secret, err := c.clientset.CoreV1().Secrets(GetNamespace(namespace...)).Get(ServiceCertsSecretName(serviceName), apismetav1.GetOptions{})
	if IsNotFoundError(err) {
		secret, err = c.createAndGetEmptyServiceSecret(ServiceCertsSecretName(serviceName), namespace...)
	}
	if err != nil {
		return nil, exception.New(err)
	}
	secret = fillNilDataValuesForSecret(secret)
	return secret, nil
}

// SetCert sets a cert file value by its key.
func (c *Client) SetCert(serviceName string, fileName CertFile, contents []byte, namespace ...string) error {
	secret, err := c.GetCertsForService(serviceName, namespace...)
	if err != nil {
		return err
	}
	return c.setServiceSecretKey(secret, string(fileName), contents)
}

// DeleteCertsForService deletes the certs secret for a service.
func (c *Client) DeleteCertsForService(serviceName string, namespace ...string) error {
	return c.DeleteSecret(ServiceCertsSecretName(serviceName), namespace...)
}

// UpdateServiceSecret updates the service secret and stores previous versions
func (c *Client) UpdateServiceSecret(secret *corev1.Secret) error {
	current, err := c.GetSecret(secret.Name, secret.Namespace)
	if err != nil {
		return err
	}
	curData := mergedSecretData(current)
	data := mergedSecretData(secret)
	// compare secrets to determine if being updated
	if reflect.DeepEqual(curData, data) {
		return nil
	}

	// save current value into new secret
	id := kuberand.String(8)
	labelServiceSecret(current)
	old := &corev1.Secret{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", current.Name, id),
			Namespace: current.Namespace,
			Labels:    current.Labels,
		},
		Data:       current.Data,
		StringData: current.StringData,
		Type:       current.Type,
	}
	old = fillNilDataValuesForSecret(old)
	err = c.CreateSecret(old, old.Namespace)
	if err != nil {
		return err
	}

	// save updated values
	secret = fillNilDataValuesForSecret(secret)
	labelServiceSecret(secret)
	secret.Labels[LabelUpdatedAt] = strconv.FormatInt(time.Now().Unix(), 10) // label with update time
	err = c.UpdateSecret(secret, secret.Namespace)
	if err != nil {
		return err
	}

	// get old secrets
	secrets, err := c.GetServiceSecretVersions(secret.Name, secret.Namespace)
	if err != nil {
		return err
	}
	if len(secrets) <= maxSecretsPerService {
		return nil
	}
	SortSecretsByCreationTime(secrets)
	toDelete := secrets[:len(secrets)-maxSecretsPerService]
	for _, s := range toDelete {
		if s.Name != secret.Name {
			err = c.DeleteSecret(s.Name, s.Namespace)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RevertServiceSecret reverts the service secret to the specified version
func (c *Client) RevertServiceSecret(secretName string, namespace ...string) error {
	version, err := c.GetSecret(secretName, GetNamespace(namespace...))
	if err != nil {
		return err
	}
	secretType, ok := version.Labels[LabelServiceSecret]
	if !ok {
		return exception.New(fmt.Errorf("Could not determine secret type"))
	}
	active, err := c.GetSecret(secretType, GetNamespace(namespace...))
	if err != nil {
		return err
	}
	// remove labels so it doesn't get picked up in the previous versions during update
	delete(version.Labels, LabelServiceSecret)
	err = c.UpdateSecret(version, version.Namespace)
	if err != nil {
		return err
	}
	active.Data = version.Data
	active.StringData = version.StringData
	err = c.UpdateServiceSecret(active)
	if err != nil {
		return err
	}
	// delete the secret now since its the active version
	return c.DeleteSecret(secretName, GetNamespace(namespace...))
}

// GetServiceSecretVersions returns the previous versions of a secret
func (c *Client) GetServiceSecretVersions(secretName string, namespace ...string) ([]corev1.Secret, error) {
	list, err := c.GetSecretsWithSelector(selector.Equals(LabelServiceSecret, secretName), GetNamespace(namespace...))
	if err != nil {
		return nil, err
	}
	ret := []corev1.Secret{}
	for _, secret := range list.Items {
		if secret.Name != secretName {
			ret = append(ret, secret)
		}
	}
	return ret, nil
}

// DeleteServiceSecrets deletes all managed service secrets and versions
func (c *Client) DeleteServiceSecrets(serviceName string, namespace ...string) error {
	secrets := ServiceManagedSecretsNames(serviceName)
	return c.deleteSecretsAndVersions(secrets, namespace...)
}

// DeleteProjectSecrets deletes all managed project secrets and versions
func (c *Client) DeleteProjectSecrets(projectName string, namespace ...string) error {
	secrets := ProjectManagedSecretsNames(projectName)
	return c.deleteSecretsAndVersions(secrets, namespace...)
}

func (c *Client) deleteSecretsAndVersions(secrets []string, namespace ...string) error {
	for _, secret := range secrets {
		err := c.DeleteSecret(secret, namespace...)
		if err != nil {
			return err
		}
		versions, err := c.GetServiceSecretVersions(secret, namespace...)
		if err != nil {
			return err
		}
		for _, version := range versions {
			err = c.DeleteSecret(version.Name, version.Namespace)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) createAndGetEmptyServiceSecret(name string, namespace ...string) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      name,
			Namespace: GetNamespace(namespace...),
			Labels: map[string]string{
				LabelServiceSecret: name,
			},
		},
		Type:       corev1.SecretTypeOpaque,
		Data:       map[string][]byte{},
		StringData: map[string]string{},
	}
	return secret, c.CreateSecret(secret, GetNamespace(namespace...))
}

// ScaleDeployment sets the current required replicas count.
func (c *Client) ScaleDeployment(name string, replicas int, namespace ...string) error {
	_, err := c.clientset.AppsV1beta1().Deployments(GetNamespace(namespace...)).Patch(name, kubetypes.MergePatchType, CreateDeploymentReplicasPatch(replicas))
	return err
}

// ApproveCertificateSigningRequest approves a csr
func (c *Client) ApproveCertificateSigningRequest(csr *certsv1beta1.CertificateSigningRequest) error {
	csr.Status.Conditions = append(csr.Status.Conditions, certsv1beta1.CertificateSigningRequestCondition{
		Type:    certsv1beta1.CertificateApproved,
		Reason:  "KubeClientApprove",
		Message: "This CSR was approved by kube client.",
	})
	_, err := c.clientset.Certificates().CertificateSigningRequests().UpdateApproval(csr)
	return err
}

// GenerateTLSCertificate generates a tls server certificate signed by kube internal CA
func (c *Client) GenerateTLSCertificate(req *x509.CertificateRequest, pollInterval, pollTimeout time.Duration) ([]byte, []byte, error) {
	name := fmt.Sprintf("%s-%s", req.Subject.CommonName, uuid.V4().ToShortString())
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, req, priv)
	if err != nil {
		return nil, nil, err
	}
	block := &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr}
	kubeCSR := &certsv1beta1.CertificateSigningRequest{
		ObjectMeta: apismetav1.ObjectMeta{
			Name: name,
		},
		Spec: certsv1beta1.CertificateSigningRequestSpec{
			Request: pem.EncodeToMemory(block),
			// Key usages taken from https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/#step-2-create-a-certificate-signing-request-object-to-send-to-the-kubernetes-api
			Usages: []certsv1beta1.KeyUsage{
				certsv1beta1.UsageDigitalSignature,
				certsv1beta1.UsageKeyEncipherment,
				certsv1beta1.UsageServerAuth,
				certsv1beta1.UsageClientAuth,
			},
		},
	}
	if err := c.CreateCertificateSigningRequest(kubeCSR); err != nil {
		return nil, nil, err
	}
	defer c.DeleteCertificateSigningRequest(name)

	if err := c.ApproveCertificateSigningRequest(kubeCSR); err != nil {
		return nil, nil, err
	}
	var cert []byte
	err = wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		approved, err := c.GetCertificateSigningRequest(name)
		if err != nil {
			return false, err
		}
		if len(approved.Status.Certificate) > 0 {
			cert = approved.Status.Certificate
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}
	keyBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}
	return cert, pem.EncodeToMemory(keyBlock), nil
}

// GetAllTokenSecretsForServiceAccount returns all mountable service account tokens for a service account (normally there should only be one)
func (c *Client) GetAllTokenSecretsForServiceAccount(sa *corev1.ServiceAccount) ([]corev1.Secret, error) {
	secretList, err := c.GetSecretsWithSelector(selector.Empty(), sa.GetNamespace())
	if err != nil {
		return nil, exception.New(err)
	}
	mountables := sets.NewString()
	for _, secretRef := range sa.Secrets {
		mountables.Insert(secretRef.Name)
	}
	secrets := []corev1.Secret{}
	for _, secret := range secretList.Items {
		if serviceaccount.IsServiceAccountToken(&secret, sa) && mountables.Has(secret.Name) {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

// GetATokenSecretForServiceAccount returns any service account token secret for a service account
func (c *Client) GetATokenSecretForServiceAccount(sa *corev1.ServiceAccount) (*corev1.Secret, error) {
	secrets, err := c.GetAllTokenSecretsForServiceAccount(sa) // trade off the ability to short-circuit for reusable code
	if err != nil {
		return nil, exception.New(err)
	}
	if len(secrets) == 0 {
		return nil, exception.New(fmt.Sprintf("Service account token not found"))
	}
	return &secrets[0], nil
}

// GetEventsForPod get events for a pod
func (c *Client) GetEventsForPod(pod *corev1.Pod) (*corev1.EventList, error) {
	// https://github.com/kubernetes/dashboard/blob/38e6827ff8eda8ab5dbdda155ce632835f4f6116/src/app/backend/resource/event/common.go#L135
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(schema.GroupVersion{Version: "v1"}, &corev1.Pod{})
	return c.clientset.CoreV1().Events(GetNamespace(pod.Namespace)).Search(scheme, pod)
}

// GetAbnormalEventsForPod gets abnormal events for a given pod
func (c *Client) GetAbnormalEventsForPod(pod *corev1.Pod) ([]corev1.Event, error) {
	events, err := c.GetEventsForPod(pod)
	if err != nil {
		return nil, err
	}
	var output []corev1.Event
	for _, event := range events.Items {
		if event.Type != corev1.EventTypeNormal {
			output = append(output, event)
		}
	}
	return output, nil
}

func fillNilDataValuesForSecret(secret *corev1.Secret) *corev1.Secret {
	if secret == nil {
		secret = &corev1.Secret{}
	}
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}
	if secret.StringData == nil {
		secret.StringData = map[string]string{}
	}
	return secret
}

func labelServiceSecret(secret *corev1.Secret) {
	if secret == nil {
		return
	}
	if secret.Labels == nil {
		secret.Labels = map[string]string{}
	}
	secret.Labels[LabelServiceSecret] = secret.Name
}
