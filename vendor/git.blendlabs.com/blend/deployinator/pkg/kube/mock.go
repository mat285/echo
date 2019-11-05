package kube

import (
	"fmt"
	"reflect"

	exception "github.com/blend/go-sdk/exception"
	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	certmanagerclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	icertmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1alpha1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	v1beta1 "k8s.io/api/apps/v1beta1"
	autov2beta1 "k8s.io/api/autoscaling/v2beta1"
	apisbatchv1 "k8s.io/api/batch/v1"
	apisbatchv1beta1 "k8s.io/api/batch/v1beta1"
	certs "k8s.io/api/certificates/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1beta1 "k8s.io/api/extensions/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	iapiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	iadmissionregistrationv1beta1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
	appsv1beta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	autoscalingv2beta1 "k8s.io/client-go/kubernetes/typed/autoscaling/v2beta1"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	batchv1beta1 "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	certsv1beta1 "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	eventsv1beta1 "k8s.io/client-go/kubernetes/typed/events/v1beta1"
	extensionsv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	ipolicyv1beta1 "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
	irbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	istoragev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	"k8s.io/client-go/rest"
	apiregistrationv1beta1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1beta1"
	kubeAggregator "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	iapiregistrationv1beta1 "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/typed/apiregistration/v1beta1"
)

// MockKubeInterface is a mock
type MockKubeInterface struct {
	Responses chan interface{}
	Errors    chan error
}

// MockAggregatorClientSet is a mocked aggregator clientset
type MockAggregatorClientSet struct {
	kubeAggregator.Interface
	*MockAPIRegistrationV1Beta1
}

// MockAPIRegistrationV1Beta1 is a mock
type MockAPIRegistrationV1Beta1 struct {
	iapiregistrationv1beta1.ApiregistrationV1beta1Interface
	*MockAPIService
}

// MockAPIService mocks
type MockAPIService struct {
	iapiregistrationv1beta1.APIServiceInterface
	MockKubeInterface
}

// MockAPIExtensionsClientSet is a mocked aggregator clientset
type MockAPIExtensionsClientSet struct {
	apiextensionsclient.Interface
	*MockAPIExtensionsV1Beta1
}

// MockAPIExtensionsV1Beta1 is a mock
type MockAPIExtensionsV1Beta1 struct {
	iapiextensionsv1beta1.ApiextensionsV1beta1Interface
	*MockCustomResourceDefinitions
}

// MockCustomResourceDefinitions mocks
type MockCustomResourceDefinitions struct {
	iapiextensionsv1beta1.CustomResourceDefinitionInterface
	MockKubeInterface
}

// MockCertManagerClientSet is a mocked aggregator clientset
type MockCertManagerClientSet struct {
	certmanagerclient.Interface
	*MockCertManagerV1Alpha1
}

// MockCertManagerV1Alpha1 is a mock
type MockCertManagerV1Alpha1 struct {
	icertmanagerv1alpha1.CertmanagerV1alpha1Interface
	*MockClusterIssuers
	*MockIssuers
	*MockCertManagerCertificates // can't use MockCertificates because of naming conflict :(
}

// MockClusterIssuers mocks
type MockClusterIssuers struct {
	icertmanagerv1alpha1.ClusterIssuerInterface
	MockKubeInterface
}

// MockIssuers mocks
type MockIssuers struct {
	icertmanagerv1alpha1.IssuerInterface
	MockKubeInterface
}

// MockCertManagerCertificates mocks
type MockCertManagerCertificates struct {
	icertmanagerv1alpha1.CertificateInterface
	MockKubeInterface
}

// MockClientSet is a mocked clientset
type MockClientSet struct {
	kubernetes.Interface
	*MockAppsV1beta1
	*MockCoreV1
	*MockAutoscalingV2beta1
	*MockExtensionsV1beta1
	*MockRbac
	*MockCertificates
	*MockStorage
	*MockBatchV1beta1
	*MockBatchV1
	*MockPolicyV1beta1
	*MockAdmissionregistrationV1beta1
}

// MockAppsV1beta1 is a mock
type MockAppsV1beta1 struct {
	appsv1beta1.AppsV1beta1Interface
	*MockDeployments
	*MockStatefulSets
}

// MockCoreV1 is a mock
type MockCoreV1 struct {
	corev1.CoreV1Interface
	*MockServices
	*MockServiceAccounts
	*MockPods
	*MockSecrets
	*MockPersistentVolumes
	*MockPersistentVolumeClaims
	*MockConfigMaps
	*MockNamespaces
	*MockEvents
	*MockNodes
}

// MockAutoscalingV2beta1 is a mock
type MockAutoscalingV2beta1 struct {
	autoscalingv2beta1.AutoscalingV2beta1Interface
	*MockHorizontalPodAutoscalers
}

// MockExtensionsV1beta1 ia a mock
type MockExtensionsV1beta1 struct {
	extensionsv1beta1.ExtensionsV1beta1Interface
	*MockIngresses
	*MockReplicaSets
	*MockDaemonSets
	*MockDeploymentsExtension
}

// MockRbac is a mock
type MockRbac struct {
	irbacv1.RbacV1Interface
	*MockRoleBindings
	*MockRoles
	*MockClusterRoleBindings
	*MockClusterRoles
}

// MockStorage is a mock
type MockStorage struct {
	istoragev1.StorageV1Interface
	*MockStorageClasses
}

// MockCertificates is a mock
type MockCertificates struct {
	certsv1beta1.CertificatesV1beta1Interface
	*MockCertificateSigningRequests
}

// MockBatchV1beta1 is a mock
type MockBatchV1beta1 struct {
	batchv1beta1.BatchV1beta1Interface
	*MockCronJobs
}

// MockBatchV1 is a mock
type MockBatchV1 struct {
	batchv1.BatchV1Interface
	*MockJobs
}

// MockPolicyV1beta1 is a mock
type MockPolicyV1beta1 struct {
	ipolicyv1beta1.PolicyV1beta1Interface
	*MockPodDisruptionBudgets
}

type MockAdmissionregistrationV1beta1 struct {
	iadmissionregistrationv1beta1.AdmissionregistrationV1beta1Interface
	*MockValidatingWebhookConfigurations
}

// MockDeployments is a mock
type MockDeployments struct {
	appsv1beta1.DeploymentInterface
	MockKubeInterface
}

// MockDeploymentsExtension is a mock
type MockDeploymentsExtension struct {
	extensionsv1beta1.DeploymentInterface
	MockKubeInterface
}

// MockStatefulSets is a mock
type MockStatefulSets struct {
	appsv1beta1.StatefulSetInterface
	MockKubeInterface
}

// MockNamespaces is a mock
type MockNamespaces struct {
	corev1.NamespaceInterface
	MockKubeInterface
}

// MockEvents is a mock
type MockEvents struct {
	corev1.EventInterface
	MockKubeInterface
}

// MockServices is a mock
type MockServices struct {
	corev1.ServiceInterface
	MockKubeInterface
}

// MockServiceAccounts is a mock
type MockServiceAccounts struct {
	corev1.ServiceAccountInterface
	MockKubeInterface
}

// MockPods is a mock
type MockPods struct {
	corev1.PodInterface
	MockKubeInterface
}

// MockNodes is a mock
type MockNodes struct {
	corev1.NodeInterface
	MockKubeInterface
}

// MockSecrets is a mock
type MockSecrets struct {
	corev1.SecretInterface
	MockKubeInterface
}

// MockPersistentVolumes is a mock
type MockPersistentVolumes struct {
	corev1.PersistentVolumeInterface
	MockKubeInterface
}

// MockPersistentVolumeClaims is a mock
type MockPersistentVolumeClaims struct {
	corev1.PersistentVolumeClaimInterface
	MockKubeInterface
}

// MockConfigMaps is a mock
type MockConfigMaps struct {
	corev1.ConfigMapInterface
	MockKubeInterface
}

// MockHorizontalPodAutoscalers is a mock
type MockHorizontalPodAutoscalers struct {
	autoscalingv2beta1.HorizontalPodAutoscalerInterface
	MockKubeInterface
}

// MockIngresses is a mock
type MockIngresses struct {
	extensionsv1beta1.IngressInterface
	MockKubeInterface
}

// MockReplicaSets is a mock
type MockReplicaSets struct {
	extensionsv1beta1.ReplicaSetInterface
	MockKubeInterface
}

// MockDaemonSets is a mock
type MockDaemonSets struct {
	extensionsv1beta1.DaemonSetInterface
	MockKubeInterface
}

// MockRoles is a mock
type MockRoles struct {
	irbacv1.RoleInterface
	MockKubeInterface
}

// MockClusterRoles is a mock
type MockClusterRoles struct {
	irbacv1.ClusterRoleInterface
	MockKubeInterface
}

// MockClusterRoleBindings is a mock
type MockClusterRoleBindings struct {
	irbacv1.ClusterRoleBindingInterface
	MockKubeInterface
}

// MockRoleBindings is a mock
type MockRoleBindings struct {
	irbacv1.RoleBindingInterface
	MockKubeInterface
}

// MockCertificateSigningRequests is a mock
type MockCertificateSigningRequests struct {
	certsv1beta1.CertificateSigningRequestInterface
	MockKubeInterface
}

// MockStorageClasses is a mock
type MockStorageClasses struct {
	istoragev1.StorageClassInterface
	MockKubeInterface
}

// MockCronJobs is a mock
type MockCronJobs struct {
	batchv1beta1.CronJobInterface
	MockKubeInterface
}

// MockJobs is a mock
type MockJobs struct {
	batchv1.JobInterface
	MockKubeInterface
}

// MockPodDisruptionBudgets is a mock
type MockPodDisruptionBudgets struct {
	ipolicyv1beta1.PodDisruptionBudgetInterface
	MockKubeInterface
}

// MockValidatingWebhookConfigurations is a mock
type MockValidatingWebhookConfigurations struct {
	iadmissionregistrationv1beta1.ValidatingWebhookConfigurationInterface
	MockKubeInterface
}

func typeError(v interface{}) error {
	name := ""
	if v != nil {
		name = reflect.TypeOf(v).Name()
	}
	return exception.New(fmt.Sprintf("Invalid type for response `%s` `%v`", name, v))
}

// NewMockClient mocks a new kube client
func NewMockClient() (*Client, chan interface{}, chan error) {
	resps := make(chan interface{}, 30)
	errs := make(chan error, 30)
	return &Client{
		clientset:              NewMockClientSet(resps, errs),
		aggregatorClientset:    NewMockAggregatorClientSet(resps, errs),
		apiextensionsClientset: NewMockAPIExtensionsClientSet(resps, errs),
		certmanagerClientset:   NewMockCertManagerClientSet(resps, errs),
		config:                 &rest.Config{},
	}, resps, errs
}

// NewMockAggregatorClientSet returns a new mock aggregator clientset
func NewMockAggregatorClientSet(resps chan interface{}, errs chan error) kubeAggregator.Interface {
	return &MockAggregatorClientSet{
		MockAPIRegistrationV1Beta1: NewMockAPIRegistrationV1Beta1(resps, errs),
	}
}

// ApiregistrationV1beta1 returns an interface
func (m *MockAggregatorClientSet) ApiregistrationV1beta1() iapiregistrationv1beta1.ApiregistrationV1beta1Interface {
	return m.MockAPIRegistrationV1Beta1
}

// APIServices returns an interface
func (m *MockAPIRegistrationV1Beta1) APIServices() iapiregistrationv1beta1.APIServiceInterface {
	return m.MockAPIService
}

// NewMockAPIExtensionsClientSet returns a new mock aggregator clientset
func NewMockAPIExtensionsClientSet(resps chan interface{}, errs chan error) apiextensionsclient.Interface {
	return &MockAPIExtensionsClientSet{
		MockAPIExtensionsV1Beta1: NewMockAPIExtensionsV1Beta1(resps, errs),
	}
}

// ApiextensionsV1beta1 returns an interface
func (m *MockAPIExtensionsClientSet) ApiextensionsV1beta1() iapiextensionsv1beta1.ApiextensionsV1beta1Interface {
	return m.MockAPIExtensionsV1Beta1
}

// CustomResourceDefinitions returns an interface
func (m *MockAPIExtensionsV1Beta1) CustomResourceDefinitions() iapiextensionsv1beta1.CustomResourceDefinitionInterface {
	return m.MockCustomResourceDefinitions
}

// NewMockCertManagerClientSet returns a new mock aggregator clientset
func NewMockCertManagerClientSet(resps chan interface{}, errs chan error) certmanagerclient.Interface {
	return &MockCertManagerClientSet{
		MockCertManagerV1Alpha1: NewMockCertManagerV1Alpha1(resps, errs),
	}
}

// CertmanagerV1alpha1 returns an interface
func (m *MockCertManagerClientSet) CertmanagerV1alpha1() icertmanagerv1alpha1.CertmanagerV1alpha1Interface {
	return m.MockCertManagerV1Alpha1
}

// ClusterIssuers returns an interface
func (m *MockCertManagerV1Alpha1) ClusterIssuers() icertmanagerv1alpha1.ClusterIssuerInterface {
	return m.MockClusterIssuers
}

// Issuers returns an interface
func (m *MockCertManagerV1Alpha1) Issuers(namespace string) icertmanagerv1alpha1.IssuerInterface {
	return m.MockIssuers
}

// Certificates returns an interface
func (m *MockCertManagerV1Alpha1) Certificates(namespace string) icertmanagerv1alpha1.CertificateInterface {
	return m.MockCertManagerCertificates
}

// NewMockClientSet returns a new mock clientset
func NewMockClientSet(resps chan interface{}, errs chan error) kubernetes.Interface {
	return &MockClientSet{
		MockAppsV1beta1:                  NewMockAppsV1beta1(resps, errs),
		MockCoreV1:                       NewMockCoreV1(resps, errs),
		MockAutoscalingV2beta1:           NewMockAutoscalingV2beta1(resps, errs),
		MockExtensionsV1beta1:            NewMockExtensionsV1beta1(resps, errs),
		MockRbac:                         NewMockRbac(resps, errs),
		MockCertificates:                 NewMockCertificates(resps, errs),
		MockStorage:                      NewMockStorage(resps, errs),
		MockBatchV1beta1:                 NewMockBatchV1beta1(resps, errs),
		MockBatchV1:                      NewMockBatchV1(resps, errs),
		MockPolicyV1beta1:                NewMockPolicyV1beta1(resps, errs),
		MockAdmissionregistrationV1beta1: NewMockAdmissionregistrationV1beta1(resps, errs),
	}
}

// NewMockKubeInterface returns
func NewMockKubeInterface(resps chan interface{}, errs chan error) MockKubeInterface {
	return MockKubeInterface{
		Responses: resps,
		Errors:    errs,
	}
}

// AppsV1beta1 returns an interface
func (m *MockClientSet) AppsV1beta1() appsv1beta1.AppsV1beta1Interface {
	return m.MockAppsV1beta1
}

// CoreV1 returns
func (m *MockClientSet) CoreV1() corev1.CoreV1Interface {
	return m.MockCoreV1
}

// AutoscalingV2beta1 returns
func (m *MockClientSet) AutoscalingV2beta1() autoscalingv2beta1.AutoscalingV2beta1Interface {
	return m.MockAutoscalingV2beta1
}

// ExtensionsV1beta1 returns
func (m *MockClientSet) ExtensionsV1beta1() extensionsv1beta1.ExtensionsV1beta1Interface {
	return m.MockExtensionsV1beta1
}

// RbacV1 returns
func (m *MockClientSet) RbacV1() irbacv1.RbacV1Interface {
	return m.MockRbac
}

// Certificates returns
func (m *MockClientSet) Certificates() certsv1beta1.CertificatesV1beta1Interface {
	return m.MockCertificates
}

// StorageV1 returns
func (m *MockClientSet) StorageV1() istoragev1.StorageV1Interface {
	return m.MockStorage
}

// BatchV1beta1 returns
func (m *MockClientSet) BatchV1beta1() batchv1beta1.BatchV1beta1Interface {
	return m.MockBatchV1beta1
}

// BatchV1 returns
func (m *MockClientSet) BatchV1() batchv1.BatchV1Interface {
	return m.MockBatchV1
}

// PolicyV1beta1 returns
func (m *MockClientSet) PolicyV1beta1() ipolicyv1beta1.PolicyV1beta1Interface {
	return m.MockPolicyV1beta1
}

// AdmissionregistrationV1beta1 returns
func (m *MockClientSet) AdmissionregistrationV1beta1() iadmissionregistrationv1beta1.AdmissionregistrationV1beta1Interface {
	return m.MockAdmissionregistrationV1beta1
}

// Events returns
// This function is deprecated but it is ambiguous with CoreV1().Events(namespace) and go complains if we don't implement it
func (m *MockClientSet) Events() eventsv1beta1.EventsV1beta1Interface {
	return nil
}

//*************************** APIRegistrationV1Beta1 *******************************

// NewMockAPIRegistrationV1Beta1 returns
func NewMockAPIRegistrationV1Beta1(resps chan interface{}, errs chan error) *MockAPIRegistrationV1Beta1 {
	return &MockAPIRegistrationV1Beta1{
		MockAPIService: NewMockAPIService(resps, errs),
	}
}

// NewMockAPIService returns
func NewMockAPIService(resps chan interface{}, errs chan error) *MockAPIService {
	return &MockAPIService{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// Get gets
func (m *MockAPIService) Get(name string, opts metav1.GetOptions) (*apiregistrationv1beta1.APIService, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*apiregistrationv1beta1.APIService); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//*****MockDeployments*****************************************************************************
//***********************************AppsV1Beta1************************************
//**********************************************************************************

// NewMockAppsV1beta1 returns
func NewMockAppsV1beta1(resps chan interface{}, errs chan error) *MockAppsV1beta1 {
	return &MockAppsV1beta1{
		MockDeployments:  NewMockDeployments(resps, errs),
		MockStatefulSets: NewMockStatefulSets(resps, errs),
	}
}

// NewMockDeployments returns
func NewMockDeployments(resps chan interface{}, errs chan error) *MockDeployments {
	return &MockDeployments{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// Deployments returns a mock
func (m *MockAppsV1beta1) Deployments(namespace string) appsv1beta1.DeploymentInterface {
	return m.MockDeployments
}

// NewMockStatefulSets returns
func NewMockStatefulSets(resps chan interface{}, errs chan error) *MockStatefulSets {
	return &MockStatefulSets{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// StatefulSets returns a mock
func (m *MockAppsV1beta1) StatefulSets(namespace string) appsv1beta1.StatefulSetInterface {
	return m.MockStatefulSets
}

//************Deployments**************

// Get gets
func (m *MockDeployments) Get(name string, options metav1.GetOptions) (*v1beta1.Deployment, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1beta1.Deployment); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockDeployments) List(options metav1.ListOptions) (*v1beta1.DeploymentList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1beta1.DeploymentList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockDeployments) Create(dep *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockDeployments) Update(dep *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockDeployments) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// DeleteCollections batch deletes
func (m *MockDeployments) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return <-m.Errors
}

// Patch patches
func (m *MockDeployments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v1beta1.Deployment, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1beta1.Deployment); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Watch watches
func (m *MockDeployments) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(watch.Interface); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Rollback rolls back
func (m *MockDeploymentsExtension) Rollback(deploymentRollback *metav1beta1.DeploymentRollback) error {
	return <-m.Errors
}

//************StatefulSets**************

// Get gets
func (m *MockStatefulSets) Get(name string, options metav1.GetOptions) (*v1beta1.StatefulSet, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1beta1.StatefulSet); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockStatefulSets) List(options metav1.ListOptions) (*v1beta1.StatefulSetList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1beta1.StatefulSetList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockStatefulSets) Create(dep *v1beta1.StatefulSet) (*v1beta1.StatefulSet, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockStatefulSets) Update(dep *v1beta1.StatefulSet) (*v1beta1.StatefulSet, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockStatefulSets) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Patch patches
// TODO: uncomment when needed
// func (m *MockStatefulSets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v1beta1.StatefulSet, error) {
// 	response := <-m.Responses
// 	err := <-m.Errors
// 	if out, ok := response.(*v1beta1.StatefulSet); ok || response == nil {
// 		return out, err
// 	}
// 	panic(typeError(response))
// }

//*****************************************

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//**************************************CoreV1**************************************
//**********************************************************************************

// NewMockCoreV1 returns
func NewMockCoreV1(resps chan interface{}, errs chan error) *MockCoreV1 {
	return &MockCoreV1{
		MockServices:               NewMockServices(resps, errs),
		MockServiceAccounts:        NewMockServiceAccounts(resps, errs),
		MockPods:                   NewMockPods(resps, errs),
		MockSecrets:                NewMockSecrets(resps, errs),
		MockPersistentVolumes:      NewMockPersistentVolumes(resps, errs),
		MockPersistentVolumeClaims: NewMockPersistentVolumeClaims(resps, errs),
		MockConfigMaps:             NewMockConfigMaps(resps, errs),
		MockNamespaces:             NewMockNamespaces(resps, errs),
		MockEvents:                 NewMockEvents(resps, errs),
		MockNodes:                  NewMockNodes(resps, errs),
	}
}

// NewMockServices returns
func NewMockServices(resps chan interface{}, errs chan error) *MockServices {
	return &MockServices{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockServiceAccounts returns
func NewMockServiceAccounts(resps chan interface{}, errs chan error) *MockServiceAccounts {
	return &MockServiceAccounts{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockPods returns
func NewMockPods(resps chan interface{}, errs chan error) *MockPods {
	return &MockPods{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockNodes returns
func NewMockNodes(resps chan interface{}, errs chan error) *MockNodes {
	return &MockNodes{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockSecrets returns
func NewMockSecrets(resps chan interface{}, errs chan error) *MockSecrets {
	return &MockSecrets{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockPersistentVolumes returns
func NewMockPersistentVolumes(resps chan interface{}, errs chan error) *MockPersistentVolumes {
	return &MockPersistentVolumes{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockPersistentVolumeClaims returns
func NewMockPersistentVolumeClaims(resps chan interface{}, errs chan error) *MockPersistentVolumeClaims {
	return &MockPersistentVolumeClaims{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockConfigMaps returns
func NewMockConfigMaps(resps chan interface{}, errs chan error) *MockConfigMaps {
	return &MockConfigMaps{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockNamespaces returns
func NewMockNamespaces(resps chan interface{}, errs chan error) *MockNamespaces {
	return &MockNamespaces{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockEvents returns
func NewMockEvents(resps chan interface{}, errs chan error) *MockEvents {
	return &MockEvents{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// Services returns a mock
func (m *MockCoreV1) Services(namespace string) corev1.ServiceInterface {
	return m.MockServices
}

// ServiceAccounts returns a mock
func (m *MockCoreV1) ServiceAccounts(namespace string) corev1.ServiceAccountInterface {
	return m.MockServiceAccounts
}

// Pods returns a mock
func (m *MockCoreV1) Pods(namespace string) corev1.PodInterface {
	return m.MockPods
}

// Nodes returns a mock
func (m *MockCoreV1) Nodes() corev1.NodeInterface {
	return m.MockNodes
}

// Secrets returns a mock
func (m *MockCoreV1) Secrets(namespace string) corev1.SecretInterface {
	return m.MockSecrets
}

// PersistentVolumes returns a mock
func (m *MockCoreV1) PersistentVolumes() corev1.PersistentVolumeInterface {
	return m.MockPersistentVolumes
}

// PersistentVolumeClaims returns a mock
func (m *MockCoreV1) PersistentVolumeClaims(namespace string) corev1.PersistentVolumeClaimInterface {
	return m.MockPersistentVolumeClaims
}

// ConfigMaps returns a mock
func (m *MockCoreV1) ConfigMaps(namespace string) corev1.ConfigMapInterface {
	return m.MockConfigMaps
}

// Namespaces returns a mock
func (m *MockCoreV1) Namespaces() corev1.NamespaceInterface {
	return m.MockNamespaces
}

// Events returns a mock
func (m *MockCoreV1) Events(namespace string) corev1.EventInterface {
	return m.MockEvents
}

//****************Services*****************

// Get gets
func (m *MockServices) Get(name string, options metav1.GetOptions) (*v1.Service, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.Service); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockServices) List(options metav1.ListOptions) (*v1.ServiceList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.ServiceList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockServices) Create(dep *v1.Service) (*v1.Service, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockServices) Update(dep *v1.Service) (*v1.Service, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockServices) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Patch patches
func (m *MockServices) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v1.Service, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.Service); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//*****************************************

//*************ServiceAccounts*************

// Get gets
func (m *MockServiceAccounts) Get(name string, options metav1.GetOptions) (*v1.ServiceAccount, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.ServiceAccount); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockServiceAccounts) Create(dep *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockServiceAccounts) Update(dep *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockServiceAccounts) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// List lists
func (m *MockServiceAccounts) List(options metav1.ListOptions) (*v1.ServiceAccountList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.ServiceAccountList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//*****************************************

//******************Pods*******************

// Get gets
func (m *MockPods) Get(name string, options metav1.GetOptions) (*v1.Pod, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.Pod); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockPods) List(options metav1.ListOptions) (*v1.PodList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.PodList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockPods) Create(dep *v1.Pod) (*v1.Pod, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockPods) Update(dep *v1.Pod) (*v1.Pod, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockPods) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Watch watches
func (m *MockPods) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(watch.Interface); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//*****************************************

//******************Nodes*******************

// Get gets
func (m *MockNodes) Get(name string, options metav1.GetOptions) (*v1.Node, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.Node); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockNodes) List(options metav1.ListOptions) (*v1.NodeList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.NodeList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update updates
func (m *MockNodes) Update(node *v1.Node) (*v1.Node, error) {
	return node, <-m.Errors
}

// Patch patches
func (m *MockNodes) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v1.Node, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.Node); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete updates
func (m *MockNodes) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

//*****************************************

//*****************Secrets*****************

// Get gets
func (m *MockSecrets) Get(name string, options metav1.GetOptions) (*v1.Secret, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.Secret); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockSecrets) List(options metav1.ListOptions) (*v1.SecretList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.SecretList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockSecrets) Create(dep *v1.Secret) (*v1.Secret, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockSecrets) Update(dep *v1.Secret) (*v1.Secret, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockSecrets) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

//*****************************************

//*******************PVs*******************

// Get gets
func (m *MockPersistentVolumes) Get(name string, options metav1.GetOptions) (*v1.PersistentVolume, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.PersistentVolume); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockPersistentVolumes) Create(dep *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	return dep, <-m.Errors
}

//*****************************************

//*******************PVCs******************

// Get gets
func (m *MockPersistentVolumeClaims) Get(name string, options metav1.GetOptions) (*v1.PersistentVolumeClaim, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.PersistentVolumeClaim); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockPersistentVolumeClaims) List(options metav1.ListOptions) (*v1.PersistentVolumeClaimList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.PersistentVolumeClaimList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update updates
func (m *MockPersistentVolumeClaims) Update(dep *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.PersistentVolumeClaim); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete deletes
func (m *MockPersistentVolumeClaims) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// DeleteCollection batch deletes
func (m *MockPersistentVolumeClaims) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return <-m.Errors
}

// Create creates
func (m *MockPersistentVolumeClaims) Create(dep *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	return dep, <-m.Errors
}

//*****************************************

//****************ConfigMaps***************

// Create creates
func (m *MockConfigMaps) Create(input *v1.ConfigMap) (*v1.ConfigMap, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.ConfigMap); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Get gets
func (m *MockConfigMaps) Get(name string, options metav1.GetOptions) (*v1.ConfigMap, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.ConfigMap); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete deletes
func (m *MockConfigMaps) Delete(input string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Update updates
func (m *MockConfigMaps) Update(input *v1.ConfigMap) (*v1.ConfigMap, error) {
	return input, <-m.Errors
}

// Patch patches
func (m *MockConfigMaps) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v1.ConfigMap, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.ConfigMap); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteCollection batch deletes
func (m *MockConfigMaps) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return <-m.Errors
}

//*****************************************

//***************Namespaces*****************

// Create creates
func (m *MockNamespaces) Create(input *v1.Namespace) (*v1.Namespace, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.Namespace); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Get gets
func (m *MockNamespaces) Get(name string, options metav1.GetOptions) (*v1.Namespace, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.Namespace); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockNamespaces) List(options metav1.ListOptions) (*v1.NamespaceList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.NamespaceList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete deletes
func (m *MockNamespaces) Delete(name string, options *metav1.DeleteOptions) error {
	err := <-m.Errors
	return err
}

// Update updates
func (m *MockNamespaces) Update(input *v1.Namespace) (*v1.Namespace, error) {
	return input, <-m.Errors
}

// Finalize finalizes
func (m *MockNamespaces) Finalize(namespace *v1.Namespace) (*v1.Namespace, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.Namespace); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//*****************************************

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//************************************Autoscaling***********************************
//**********************************************************************************

// NewMockAutoscalingV2beta1 returns
func NewMockAutoscalingV2beta1(resps chan interface{}, errs chan error) *MockAutoscalingV2beta1 {
	return &MockAutoscalingV2beta1{
		MockHorizontalPodAutoscalers: NewMockHorizontalPodAutoscalers(resps, errs),
	}
}

// NewMockHorizontalPodAutoscalers returns
func NewMockHorizontalPodAutoscalers(resps chan interface{}, errs chan error) *MockHorizontalPodAutoscalers {
	return &MockHorizontalPodAutoscalers{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// HorizontalPodAutoscalers returns
func (m *MockAutoscalingV2beta1) HorizontalPodAutoscalers(namespace string) autoscalingv2beta1.HorizontalPodAutoscalerInterface {
	return m.MockHorizontalPodAutoscalers
}

//*******************HPAS******************

// Get gets
func (m *MockHorizontalPodAutoscalers) Get(name string, options metav1.GetOptions) (*autov2beta1.HorizontalPodAutoscaler, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autov2beta1.HorizontalPodAutoscaler); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockHorizontalPodAutoscalers) List(options metav1.ListOptions) (*autov2beta1.HorizontalPodAutoscalerList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autov2beta1.HorizontalPodAutoscalerList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockHorizontalPodAutoscalers) Create(dep *autov2beta1.HorizontalPodAutoscaler) (*autov2beta1.HorizontalPodAutoscaler, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockHorizontalPodAutoscalers) Update(dep *autov2beta1.HorizontalPodAutoscaler) (*autov2beta1.HorizontalPodAutoscaler, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockHorizontalPodAutoscalers) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

//*****************************************

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//********************************Extensionsv1beta1*********************************
//**********************************************************************************

// NewMockExtensionsV1beta1 returns
func NewMockExtensionsV1beta1(resps chan interface{}, errs chan error) *MockExtensionsV1beta1 {
	return &MockExtensionsV1beta1{
		MockIngresses:            NewMockIngresses(resps, errs),
		MockReplicaSets:          NewMockReplicaSets(resps, errs),
		MockDaemonSets:           NewMockDaemonSets(resps, errs),
		MockDeploymentsExtension: NewMockDeploymentsExtension(resps, errs),
	}
}

// NewMockIngresses returns
func NewMockIngresses(resps chan interface{}, errs chan error) *MockIngresses {
	return &MockIngresses{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockReplicaSets returns
func NewMockReplicaSets(resps chan interface{}, errs chan error) *MockReplicaSets {
	return &MockReplicaSets{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockDaemonSets returns
func NewMockDaemonSets(resps chan interface{}, errs chan error) *MockDaemonSets {
	return &MockDaemonSets{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockDeploymentsExtension returns
func NewMockDeploymentsExtension(resps chan interface{}, errs chan error) *MockDeploymentsExtension {
	return &MockDeploymentsExtension{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// Ingresses returns
func (m *MockExtensionsV1beta1) Ingresses(namespace string) extensionsv1beta1.IngressInterface {
	return m.MockIngresses
}

// ReplicaSets returns
func (m *MockExtensionsV1beta1) ReplicaSets(namespace string) extensionsv1beta1.ReplicaSetInterface {
	return m.MockReplicaSets
}

// DaemonSets returns
func (m *MockExtensionsV1beta1) DaemonSets(namespace string) extensionsv1beta1.DaemonSetInterface {
	return m.MockDaemonSets
}

// Deployments returns
func (m *MockExtensionsV1beta1) Deployments(namespace string) extensionsv1beta1.DeploymentInterface {
	return m.MockDeploymentsExtension
}

//*****************Ingresses***************

// Get gets
func (m *MockIngresses) Get(name string, options metav1.GetOptions) (*metav1beta1.Ingress, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*metav1beta1.Ingress); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockIngresses) List(options metav1.ListOptions) (*metav1beta1.IngressList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*metav1beta1.IngressList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockIngresses) Create(dep *metav1beta1.Ingress) (*metav1beta1.Ingress, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockIngresses) Update(dep *metav1beta1.Ingress) (*metav1beta1.Ingress, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockIngresses) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

//*****************************************

//****************ReplicaSets**************

// Get gets
func (m *MockReplicaSets) Get(name string, options metav1.GetOptions) (*metav1beta1.ReplicaSet, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*metav1beta1.ReplicaSet); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockReplicaSets) List(options metav1.ListOptions) (*metav1beta1.ReplicaSetList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*metav1beta1.ReplicaSetList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockReplicaSets) Create(dep *metav1beta1.ReplicaSet) (*metav1beta1.ReplicaSet, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockReplicaSets) Update(dep *metav1beta1.ReplicaSet) (*metav1beta1.ReplicaSet, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockReplicaSets) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

//*****************************************

//****************DaemonSets***************

// Get gets
func (m *MockDaemonSets) Get(name string, options metav1.GetOptions) (*metav1beta1.DaemonSet, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*metav1beta1.DaemonSet); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockDaemonSets) List(options metav1.ListOptions) (*metav1beta1.DaemonSetList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*metav1beta1.DaemonSetList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockDaemonSets) Create(dep *metav1beta1.DaemonSet) (*metav1beta1.DaemonSet, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockDaemonSets) Update(dep *metav1beta1.DaemonSet) (*metav1beta1.DaemonSet, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockDaemonSets) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// DeleteCollection batch deletes
func (m *MockDaemonSets) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return <-m.Errors
}

//*****************************************

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//**************************************Rbac****************************************
//**********************************************************************************

// NewMockRbac returns
func NewMockRbac(resps chan interface{}, errs chan error) *MockRbac {
	return &MockRbac{
		MockRoles:               NewMockRoles(resps, errs),
		MockRoleBindings:        NewMockRoleBindings(resps, errs),
		MockClusterRoles:        NewMockClusterRoles(resps, errs),
		MockClusterRoleBindings: NewMockClusterRoleBindings(resps, errs),
	}
}

// NewMockRoles returns
func NewMockRoles(resps chan interface{}, errs chan error) *MockRoles {
	return &MockRoles{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockRoleBindings returns
func NewMockRoleBindings(resps chan interface{}, errs chan error) *MockRoleBindings {
	return &MockRoleBindings{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockClusterRoles returns
func NewMockClusterRoles(resps chan interface{}, errs chan error) *MockClusterRoles {
	return &MockClusterRoles{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockClusterRoleBindings returns
func NewMockClusterRoleBindings(resps chan interface{}, errs chan error) *MockClusterRoleBindings {
	return &MockClusterRoleBindings{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// Roles returns
func (m *MockRbac) Roles(namespace string) irbacv1.RoleInterface {
	return m.MockRoles
}

// ClusterRoles returns
func (m *MockRbac) ClusterRoles() irbacv1.ClusterRoleInterface {
	return m.MockClusterRoles
}

// RoleBindings returns
func (m *MockRbac) RoleBindings(namespace string) irbacv1.RoleBindingInterface {
	return m.MockRoleBindings
}

// ClusterRoleBindings returns
func (m *MockRbac) ClusterRoleBindings() irbacv1.ClusterRoleBindingInterface {
	return m.MockClusterRoleBindings
}

//*******************Roles*****************

// Get gets
func (m *MockRoles) Get(name string, options metav1.GetOptions) (*rbacv1.Role, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.Role); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockRoles) Create(role *rbacv1.Role) (*rbacv1.Role, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.Role); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update creates
func (m *MockRoles) Update(role *rbacv1.Role) (*rbacv1.Role, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.Role); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete deletes
func (m *MockRoles) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// List lists
func (m *MockRoles) List(options metav1.ListOptions) (*rbacv1.RoleList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.RoleList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteCollection batch deletes
func (m *MockRoles) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return <-m.Errors
}

//***************ClusterRoles**************

// Get gets
func (m *MockClusterRoles) Get(name string, options metav1.GetOptions) (*rbacv1.ClusterRole, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.ClusterRole); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockClusterRoles) Create(role *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.ClusterRole); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update creates
func (m *MockClusterRoles) Update(role *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.ClusterRole); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete deletes
func (m *MockClusterRoles) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

//*****************************************

//*******************CRBs******************

// Get gets
func (m *MockClusterRoleBindings) Get(name string, options metav1.GetOptions) (*rbacv1.ClusterRoleBinding, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.ClusterRoleBinding); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockClusterRoleBindings) Create(crb *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.ClusterRoleBinding); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete deletes
func (m *MockClusterRoleBindings) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Update updates
func (m *MockClusterRoleBindings) Update(rb *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.ClusterRoleBinding); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//*******************RBs******************

// Get gets
func (m *MockRoleBindings) Get(name string, options metav1.GetOptions) (*rbacv1.RoleBinding, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.RoleBinding); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockRoleBindings) Create(crb *rbacv1.RoleBinding) (*rbacv1.RoleBinding, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.RoleBinding); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete deletes
func (m *MockRoleBindings) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Update updates
func (m *MockRoleBindings) Update(rb *rbacv1.RoleBinding) (*rbacv1.RoleBinding, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*rbacv1.RoleBinding); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteCollection batch deletes
func (m *MockRoleBindings) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return <-m.Errors
}

//**********************************************************************************
//**********************************Certificates************************************
//**********************************************************************************

// NewMockCertificates returns certificates mock
func NewMockCertificates(resps chan interface{}, errs chan error) *MockCertificates {
	return &MockCertificates{
		MockCertificateSigningRequests: NewMockCertificateSigningRequests(resps, errs),
	}
}

// NewMockCertificateSigningRequests returns csr mock
func NewMockCertificateSigningRequests(resps chan interface{}, errs chan error) *MockCertificateSigningRequests {
	return &MockCertificateSigningRequests{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// CertificateSigningRequests returns csr interface mock
func (m *MockCertificates) CertificateSigningRequests() certsv1beta1.CertificateSigningRequestInterface {
	return m.MockCertificateSigningRequests
}

// Create creates
func (m *MockCertificateSigningRequests) Create(csr *certs.CertificateSigningRequest) (*certs.CertificateSigningRequest, error) {
	return csr, <-m.Errors
}

// Get gets
func (m *MockCertificateSigningRequests) Get(name string, options metav1.GetOptions) (*certs.CertificateSigningRequest, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*certs.CertificateSigningRequest); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete deletes
func (m *MockCertificateSigningRequests) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// UpdateApproval updates approval
func (m *MockCertificateSigningRequests) UpdateApproval(csr *certs.CertificateSigningRequest) (*certs.CertificateSigningRequest, error) {
	return csr, <-m.Errors
}

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//*************************************Storage**************************************
//**********************************************************************************

// NewMockStorage returns
func NewMockStorage(resps chan interface{}, errs chan error) *MockStorage {
	return &MockStorage{
		MockStorageClasses: NewMockStorageClasses(resps, errs),
	}
}

// NewMockStorageClasses returns
func NewMockStorageClasses(resps chan interface{}, errs chan error) *MockStorageClasses {
	return &MockStorageClasses{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// StorageClasses returns
func (m *MockStorage) StorageClasses() istoragev1.StorageClassInterface {
	return m.MockStorageClasses
}

//*******************CRBs******************

// Get gets
func (m *MockStorageClasses) Get(name string, options metav1.GetOptions) (*storagev1.StorageClass, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*storagev1.StorageClass); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update updates
func (m *MockStorageClasses) Update(dep *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*storagev1.StorageClass); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete deletes
func (m *MockStorageClasses) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Create creates
func (m *MockStorageClasses) Create(dep *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	return dep, <-m.Errors
}

//*****************************************

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//**************************************Batch***************************************
//**********************************************************************************

// NewMockBatchV1beta1 returns
func NewMockBatchV1beta1(resps chan interface{}, errs chan error) *MockBatchV1beta1 {
	return &MockBatchV1beta1{
		MockCronJobs: NewMockCronJobs(resps, errs),
	}
}

// NewMockBatchV1 returns
func NewMockBatchV1(resps chan interface{}, errs chan error) *MockBatchV1 {
	return &MockBatchV1{
		MockJobs: NewMockJobs(resps, errs),
	}
}

// NewMockCronJobs returns
func NewMockCronJobs(resps chan interface{}, errs chan error) *MockCronJobs {
	return &MockCronJobs{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// NewMockJobs returns
func NewMockJobs(resps chan interface{}, errs chan error) *MockJobs {
	return &MockJobs{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// CronJobs returns
func (m *MockBatchV1beta1) CronJobs(namespace string) batchv1beta1.CronJobInterface {
	return m.MockCronJobs
}

// Jobs returns
func (m *MockBatchV1) Jobs(namespace string) batchv1.JobInterface {
	return m.MockJobs
}

//*****************CronJobs****************

// Get gets
func (m *MockCronJobs) Get(name string, options metav1.GetOptions) (*apisbatchv1beta1.CronJob, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*apisbatchv1beta1.CronJob); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockCronJobs) List(options metav1.ListOptions) (*apisbatchv1beta1.CronJobList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*apisbatchv1beta1.CronJobList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update updates
func (m *MockCronJobs) Update(dep *apisbatchv1beta1.CronJob) (*apisbatchv1beta1.CronJob, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*apisbatchv1beta1.CronJob); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Delete deletes
func (m *MockCronJobs) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// DeleteCollection batch deletes
func (m *MockCronJobs) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return <-m.Errors
}

// Create creates
func (m *MockCronJobs) Create(dep *apisbatchv1beta1.CronJob) (*apisbatchv1beta1.CronJob, error) {
	return dep, <-m.Errors
}

//*****************Jobs****************

// Get gets
func (m *MockJobs) Get(name string, options metav1.GetOptions) (*apisbatchv1.Job, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*apisbatchv1.Job); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockJobs) List(options metav1.ListOptions) (*apisbatchv1.JobList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*apisbatchv1.JobList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update updates
func (m *MockJobs) Update(dep *apisbatchv1.Job) (*apisbatchv1.Job, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockJobs) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Create creates
func (m *MockJobs) Create(dep *apisbatchv1.Job) (*apisbatchv1.Job, error) {
	return dep, <-m.Errors
}

// Watch watches
func (m *MockJobs) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(watch.Interface); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Patch patches
func (m *MockJobs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*apisbatchv1.Job, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*apisbatchv1.Job); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//*****************************************

//*****************Events******************

// Search searches
func (m *MockEvents) Search(scheme *runtime.Scheme, objOrRef runtime.Object) (*v1.EventList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*v1.EventList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//*****************************************

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//*************************************Policy**************************************
//**********************************************************************************

// NewMockPolicyV1beta1 returns
func NewMockPolicyV1beta1(resps chan interface{}, errs chan error) *MockPolicyV1beta1 {
	return &MockPolicyV1beta1{
		MockPodDisruptionBudgets: NewMockPodDisruptionBudgets(resps, errs),
	}
}

// NewMockPodDisruptionBudgets returns
func NewMockPodDisruptionBudgets(resps chan interface{}, errs chan error) *MockPodDisruptionBudgets {
	return &MockPodDisruptionBudgets{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// PodDisruptionBudgets returns
func (m *MockPolicyV1beta1) PodDisruptionBudgets(namespace string) ipolicyv1beta1.PodDisruptionBudgetInterface {
	return m.MockPodDisruptionBudgets
}

//******************PDBs*******************

// Get gets
func (m *MockPodDisruptionBudgets) Get(name string, options metav1.GetOptions) (*policyv1beta1.PodDisruptionBudget, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*policyv1beta1.PodDisruptionBudget); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// List lists
func (m *MockPodDisruptionBudgets) List(options metav1.ListOptions) (*policyv1beta1.PodDisruptionBudgetList, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*policyv1beta1.PodDisruptionBudgetList); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Create creates
func (m *MockPodDisruptionBudgets) Create(dep *policyv1beta1.PodDisruptionBudget) (*policyv1beta1.PodDisruptionBudget, error) {
	return dep, <-m.Errors
}

// Update updates
func (m *MockPodDisruptionBudgets) Update(dep *policyv1beta1.PodDisruptionBudget) (*policyv1beta1.PodDisruptionBudget, error) {
	return dep, <-m.Errors
}

// Delete deletes
func (m *MockPodDisruptionBudgets) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

//*****************************************

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//***************************************CRD****************************************
//**********************************************************************************

// NewMockAPIExtensionsV1Beta1 returns
func NewMockAPIExtensionsV1Beta1(resps chan interface{}, errs chan error) *MockAPIExtensionsV1Beta1 {
	return &MockAPIExtensionsV1Beta1{
		MockCustomResourceDefinitions: NewMockCustomResourceDefinitions(resps, errs),
	}
}

// NewMockCustomResourceDefinitions returns
func NewMockCustomResourceDefinitions(resps chan interface{}, errs chan error) *MockCustomResourceDefinitions {
	return &MockCustomResourceDefinitions{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// Get gets
func (m *MockCustomResourceDefinitions) Get(name string, opts metav1.GetOptions) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*apiextensionsv1beta1.CustomResourceDefinition); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update updates
func (m *MockCustomResourceDefinitions) Update(crd *apiextensionsv1beta1.CustomResourceDefinition) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	return crd, <-m.Errors
}

// Delete deletes
func (m *MockCustomResourceDefinitions) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Create creates
func (m *MockCustomResourceDefinitions) Create(dep *apiextensionsv1beta1.CustomResourceDefinition) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	return dep, <-m.Errors
}

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//**************************Validating Webhook Configuration***********************
//**********************************************************************************

// NewMockAdmissionregistrationV1beta1 returns
func NewMockAdmissionregistrationV1beta1(resps chan interface{}, errs chan error) *MockAdmissionregistrationV1beta1 {
	return &MockAdmissionregistrationV1beta1{
		MockValidatingWebhookConfigurations: NewMockValidatingWebhookConfigurations(resps, errs),
	}
}

// ValidatingWebhookConfigurations returns
func (m *MockAdmissionregistrationV1beta1) ValidatingWebhookConfigurations() iadmissionregistrationv1beta1.ValidatingWebhookConfigurationInterface {
	return m.MockValidatingWebhookConfigurations
}

// NewMockValidatingWebhookConfigurations returns
func NewMockValidatingWebhookConfigurations(resps chan interface{}, errs chan error) *MockValidatingWebhookConfigurations {
	return &MockValidatingWebhookConfigurations{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// Get gets
func (m *MockValidatingWebhookConfigurations) Get(name string, opts metav1.GetOptions) (*admissionregistrationv1beta1.ValidatingWebhookConfiguration, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*admissionregistrationv1beta1.ValidatingWebhookConfiguration); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update updates
func (m *MockValidatingWebhookConfigurations) Update(v *admissionregistrationv1beta1.ValidatingWebhookConfiguration) (*admissionregistrationv1beta1.ValidatingWebhookConfiguration, error) {
	return v, <-m.Errors
}

// Delete deletes
func (m *MockValidatingWebhookConfigurations) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Create creates
func (m *MockValidatingWebhookConfigurations) Create(v *admissionregistrationv1beta1.ValidatingWebhookConfiguration) (*admissionregistrationv1beta1.ValidatingWebhookConfiguration, error) {
	return v, <-m.Errors
}

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//**********************************Cert Manager************************************
//**********************************************************************************

// NewMockCertManagerV1Alpha1 returns
func NewMockCertManagerV1Alpha1(resps chan interface{}, errs chan error) *MockCertManagerV1Alpha1 {
	return &MockCertManagerV1Alpha1{
		MockClusterIssuers:          NewMockClusterIssuers(resps, errs),
		MockIssuers:                 NewMockIssuers(resps, errs),
		MockCertManagerCertificates: NewMockCertManagerCertificates(resps, errs),
	}
}

// NewMockClusterIssuers returns
func NewMockClusterIssuers(resps chan interface{}, errs chan error) *MockClusterIssuers {
	return &MockClusterIssuers{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// Get gets
func (m *MockClusterIssuers) Get(name string, opts metav1.GetOptions) (*certmanagerv1alpha1.ClusterIssuer, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*certmanagerv1alpha1.ClusterIssuer); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update updates
func (m *MockClusterIssuers) Update(crd *certmanagerv1alpha1.ClusterIssuer) (*certmanagerv1alpha1.ClusterIssuer, error) {
	return crd, <-m.Errors
}

// Delete deletes
func (m *MockClusterIssuers) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Create creates
func (m *MockClusterIssuers) Create(dep *certmanagerv1alpha1.ClusterIssuer) (*certmanagerv1alpha1.ClusterIssuer, error) {
	return dep, <-m.Errors
}

// NewMockIssuers returns
func NewMockIssuers(resps chan interface{}, errs chan error) *MockIssuers {
	return &MockIssuers{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// Get gets
func (m *MockIssuers) Get(name string, opts metav1.GetOptions) (*certmanagerv1alpha1.Issuer, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*certmanagerv1alpha1.Issuer); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update updates
func (m *MockIssuers) Update(crd *certmanagerv1alpha1.Issuer) (*certmanagerv1alpha1.Issuer, error) {
	return crd, <-m.Errors
}

// Delete deletes
func (m *MockIssuers) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Create creates
func (m *MockIssuers) Create(dep *certmanagerv1alpha1.Issuer) (*certmanagerv1alpha1.Issuer, error) {
	return dep, <-m.Errors
}

// NewMockCertManagerCertificates returns
func NewMockCertManagerCertificates(resps chan interface{}, errs chan error) *MockCertManagerCertificates {
	return &MockCertManagerCertificates{
		MockKubeInterface: NewMockKubeInterface(resps, errs),
	}
}

// Get gets
func (m *MockCertManagerCertificates) Get(name string, opts metav1.GetOptions) (*certmanagerv1alpha1.Certificate, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*certmanagerv1alpha1.Certificate); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// Update updates
func (m *MockCertManagerCertificates) Update(crd *certmanagerv1alpha1.Certificate) (*certmanagerv1alpha1.Certificate, error) {
	return crd, <-m.Errors
}

// Delete deletes
func (m *MockCertManagerCertificates) Delete(name string, options *metav1.DeleteOptions) error {
	return <-m.Errors
}

// Create creates
func (m *MockCertManagerCertificates) Create(dep *certmanagerv1alpha1.Certificate) (*certmanagerv1alpha1.Certificate, error) {
	return dep, <-m.Errors
}

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************

//**********************************************************************************
//*************************************Helpers**************************************
//**********************************************************************************

// SetTestPollTimeouts sets the test polling timeouts
func SetTestPollTimeouts() {
	deletePollInterval = testPollInterval
	deletePollTimeout = testPollTimeout
}

// UnsetTestPollTimeouts unsets the test polling timeouts
func UnsetTestPollTimeouts() {
	deletePollInterval = defaultDeletePollInterval
	deletePollTimeout = defaultDeletePollTimeout
}

//**********************************************************************************
//**********************************************************************************
//**********************************************************************************
